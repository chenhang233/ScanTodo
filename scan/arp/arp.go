package arp

import (
	"ScanTodo/scan/handle_http"
	"ScanTodo/scanLog"
	"ScanTodo/utils"
	"errors"
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"golang.org/x/sync/errgroup"
	"net"
	"strings"
	"time"
)

type NetworkDevice struct {
	Name        string
	Ip          net.IP
	Description string
	Mac         net.HardwareAddr
}

type IpTcpRulesStr struct {
	SourceIps        string `json:"sourceIps"`
	DestinationIps   string `json:"destinationIps"`
	SourcePorts      string `json:"sourcePorts"`
	DestinationPorts string `json:"destinationPorts"`
}

type ipTcpRules struct {
	SourceIps        []string
	DestinationIps   []string
	SourcePorts      []string
	DestinationPorts []string
}

// ConfigInfo 使用者配置信息
type ConfigInfo struct {
	Op              uint16        `json:"op"`
	T               time.Duration `json:"t"`
	I               time.Duration `json:"i"`
	SIP             string        `json:"sIP"`
	SMAC            string        `json:"sMAC"`
	TIP             string        `json:"tIP"`
	TMAC            string        `json:"tMAC"`
	HostIp          string        `json:"hostIp"`
	EnableRuleIpTcp bool          `json:"enableRuleIpTcp"`
	IpTcpRules      IpTcpRulesStr `json:"ipTcpRules"`
}

// Metadata 数据包元数据结构体
type Metadata struct {
	// 源
	SourceDevice *NetworkDevice
	// 目标
	TargetDevice *NetworkDevice
	// 自身
	SelfDevice *NetworkDevice
	// 发请求1 回复2
	Operation uint16
	Log       *scanLog.LogConf
	// 当前发送索引
	CurrentIndex int
	// 发包间隔时间
	Interval time.Duration
	// 发包超时时间
	Timeout time.Duration
	//ARP 回调
	OnFinish  func(*Metadata)
	OnSetup   func()
	OnSend    func(*Metadata)
	OnReceive func(*Metadata)
	// 配置IP/TCP监听规则
	EnableRuleIpTcp bool
	*IpTcpRulesStr
	*ipTcpRules
	// 结束信号
	done chan bool
}

// packet 接收到的数据包
type packet struct {
	payloadType string
	bytes       []byte
	byteLen     int
}

var HTTPMethodsMap []string
var HTTPResponseRow string
var arpEnMap map[uint16]string

func New(c *ConfigInfo) (*Metadata, error) {
	m := &Metadata{
		CurrentIndex:    -1,
		Timeout:         c.T,
		Interval:        c.I,
		Operation:       c.Op,
		SelfDevice:      &NetworkDevice{Ip: net.ParseIP(c.HostIp)},
		TargetDevice:    &NetworkDevice{Ip: net.ParseIP(c.TIP)},
		SourceDevice:    &NetworkDevice{Ip: net.ParseIP(c.SIP)},
		done:            make(chan bool, 1),
		EnableRuleIpTcp: false,
		IpTcpRulesStr:   &IpTcpRulesStr{},
		ipTcpRules:      &ipTcpRules{},
	}
	err := m.new(c)
	if err != nil {
		return nil, err
	}
	return m, m.Resolve()
}

func (m *Metadata) new(c *ConfigInfo) error {
	arpEnMap = map[uint16]string{}
	arpEnMap[1] = "ARP请求"
	arpEnMap[2] = "ARP响应"
	HTTPMethodsMap = []string{"GET", "POST", "DELETE", "HEAD", "OPTIONS", "PUT", "TRACE"}
	HTTPResponseRow = "HTTP"
	hw1, err := net.ParseMAC(c.SMAC)
	if err != nil {
		return err
	}
	m.SourceDevice.Mac = hw1
	hw2, err := net.ParseMAC(c.TMAC)
	if err != nil {
		return err
	}
	m.TargetDevice.Mac = hw2
	if c.EnableRuleIpTcp {
		// 过滤规则...
	}
	return nil
}

func (m *Metadata) Resolve() error {
	loadLog, err := scanLog.LoadLog(scanLog.ARPLogPath)
	m.Log = loadLog
	if err != nil {
		return err
	}
	dev, err := m.getPcapDev(m.SelfDevice.Ip)
	m.SelfDevice = dev
	if err != nil {
		return err
	}
	return nil
}

func (m *Metadata) Stop() {
	open := true
	select {
	case open = <-m.done:
	default:
	}
	if open {
		close(m.done)
	}
}

func (m *Metadata) finish() {
	handle := m.OnFinish
	if handle != nil {
		handle(m)
	}
}

func (m *Metadata) listen() (*pcap.Handle, error) {
	conn, err := pcap.OpenLive(m.SelfDevice.Name, 65536, true, pcap.BlockForever)
	if err != nil {
		m.Log.Error.Println("网卡错误: ", err)
		return nil, err
	}
	return conn, err
}

func (m *Metadata) Run() error {
	conn, err := m.listen()
	if err != nil {
		return err
	}
	defer conn.Close()
	return m.run(conn)
}

func (m *Metadata) run(conn *pcap.Handle) error {
	defer m.finish()
	setup := m.OnSetup
	if setup != nil {
		setup()
	}
	pks := make(chan *packet, 5)
	defer close(pks)
	var g errgroup.Group
	g.Go(func() error {
		defer m.Stop()
		return m.mainLoop(conn, pks)
	})
	g.Go(func() error {
		defer m.Stop()
		return m.listenPacket(conn, pks)
	})
	err := g.Wait()
	if err != nil {
		m.Log.Warn.Printf("异常结束中:", err)
	}
	return err
}

func (m *Metadata) mainLoop(conn *pcap.Handle, pks <-chan *packet) error {
	t1 := time.NewTicker(m.Timeout)
	t2F := true
	if m.Interval < 0 {
		t2F = false
		m.Interval = time.Second
	}
	t2 := time.NewTicker(m.Interval)
	defer func() {
		t1.Stop()
		t2.Stop()
	}()
	if t2F {
		err := m.sendArp(conn, m.Operation, m.SourceDevice.Mac, m.TargetDevice.Mac, m.SourceDevice.Ip, m.TargetDevice.Ip)
		if err != nil {
			m.Log.Warn.Println("sendArp: ", err)
		}
	} else {
		t2.Stop()
	}
	for {
		select {
		case <-m.done:
			return nil
		case <-t1.C:
			return nil
		case <-t2.C:
			err := m.sendArp(conn, m.Operation, m.SourceDevice.Mac, m.TargetDevice.Mac, m.SourceDevice.Ip, m.TargetDevice.Ip)
			if err != nil {
				m.Log.Warn.Println("sendArp: ", err)
			}
			send := m.OnSend
			if send != nil {
				send(m)
			}
		case p := <-pks:
			err := m.readPacketPayload(p)
			if err != nil {
				m.Log.Warn.Printf("处理收到的包异常", err)
			}
		}
	}
}

func (m *Metadata) readPacketPayload(receive *packet) error {
	var err error
	switch receive.payloadType {
	case "TCP":
		if receive.byteLen > 10 {
			flag := string(receive.bytes[:10])
			if utils.Includes(HTTPMethodsMap, flag) || strings.Contains(flag, HTTPResponseRow) {
				h := &handle_http.Metadata{}
				err = h.ReadHTTP(receive.bytes, m.Log)
				if err != nil {
					m.Log.Error.Println("ReadHTTP: ", err)
				}
			}
		}
	}
	return nil
}

func (m *Metadata) listenPacket(handle *pcap.Handle, pks chan<- *packet) error {
	ps := gopacket.NewPacketSource(handle, handle.LinkType())
	for {
		select {
		case <-m.done:
			return nil
		case p := <-ps.Packets():
			arpLayer := p.Layer(layers.LayerTypeARP)
			if arpLayer != nil {
				arp := arpLayer.(*layers.ARP)
				m.listenARP(arp)
			}
			ipv4Layer := p.Layer(layers.LayerTypeIPv4)
			if ipv4Layer != nil {
				ipv4 := ipv4Layer.(*layers.IPv4)
				m.listenIPv4(ipv4)
				m.defaultForwardData()
			}
			tcpLayer := p.Layer(layers.LayerTypeTCP)
			if tcpLayer != nil {
				tcp := tcpLayer.(*layers.TCP)
				m.listenTCP(tcp, pks)
			}
		}
	}
}

func (m *Metadata) listenARP(arp *layers.ARP) {
	mac1 := net.HardwareAddr(arp.SourceHwAddress)
	mac2 := net.HardwareAddr(arp.DstHwAddress)
	reply := fmt.Sprintf("监听 %s (源MAC: %v,目标MAC: %v)", arpEnMap[arp.Operation], mac1, mac2)
	m.Log.Info.Printf(reply)
	receive := m.OnReceive
	if receive != nil {
		receive(m)
	}
}

func (m *Metadata) listenIPv4(ipv4 *layers.IPv4) int {
	v4Info := fmt.Sprintf("源IP: %v,目标IP: %v", ipv4.SrcIP, ipv4.DstIP)
	m.Log.Debug.Println(v4Info)
	return 1
}

func (m *Metadata) listenTCP(tcp *layers.TCP, re chan<- *packet) {
	p := tcp.Payload
	pl := len(p)
	m.Log.Debug.Println(fmt.Sprintf("源端口: %v,目标端口: %v, seq: %v,ack: %v",
		tcp.SrcPort, tcp.DstPort, tcp.Seq, tcp.Ack))
	if pl > 0 {
		re <- &packet{bytes: p, byteLen: pl, payloadType: "TCP"}
	}
}

func (m *Metadata) defaultForwardData() {

}

/*
SendArp
Operation  ARP请求为1 ARP响应为2
*/
func (*Metadata) sendArp(conn *pcap.Handle, Operation uint16, sourceMAC, targetMAC net.HardwareAddr, sourceIp, targetIp net.IP) error {
	var err error
	eth := &layers.Ethernet{
		SrcMAC:       sourceMAC,
		DstMAC:       targetMAC,
		EthernetType: layers.EthernetTypeARP,
	}
	a := &layers.ARP{
		AddrType:          layers.LinkTypeEthernet,
		Protocol:          layers.EthernetTypeIPv4,
		HwAddressSize:     uint8(6),
		ProtAddressSize:   uint8(4),
		Operation:         Operation,
		SourceHwAddress:   sourceMAC,
		SourceProtAddress: sourceIp.To4(),
		DstHwAddress:      targetMAC,
		DstProtAddress:    targetIp.To4(),
	}
	buffer := gopacket.NewSerializeBuffer()
	var opt gopacket.SerializeOptions
	err = gopacket.SerializeLayers(buffer, opt, eth, a)
	if err != nil {
		return err
	}
	outgoingPacket := buffer.Bytes()
	err = conn.WritePacketData(outgoingPacket)
	if err != nil {
		return err
	}
	return nil
}

// GetPcapDev 获取要使用的网卡信息
func (m *Metadata) getPcapDev(ip net.IP) (*NetworkDevice, error) {
	ifs, err := pcap.FindAllDevs()
	if err != nil {
		return nil, err
	}
	for _, v := range ifs {
		for _, address := range v.Addresses {
			if address.IP.Equal(ip) {
				dev := &NetworkDevice{
					Name: v.Name,
					Ip:   ip,
				}
				interfaces, err := net.Interfaces()
				if err != nil {
					return nil, err
				}
			A:
				for _, item := range interfaces {
					addrs, err := item.Addrs()
					if err != nil {
						return nil, err
					}
					for _, addr := range addrs {
						tempIp := net.ParseIP(strings.Split(addr.String(), "/")[0])
						if tempIp.Equal(ip) {
							dev.Description = item.Name
							dev.Mac = item.HardwareAddr
							break A
						}
					}
				}
				return dev, nil
			}
		}
	}
	e := "网卡异常,没有匹配ip"
	m.Log.Error.Println(e)
	return nil, errors.New(e)
}
