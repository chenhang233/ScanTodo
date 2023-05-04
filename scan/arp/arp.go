package arp

import (
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
var arpEnMap map[uint16]string

func New(ip string) (*Metadata, error) {
	m := &Metadata{
		CurrentIndex: -1,
		Timeout:      0,
		Interval:     1,
		SelfDevice:   &NetworkDevice{Ip: net.ParseIP(ip)},
		TargetDevice: &NetworkDevice{},
		SourceDevice: &NetworkDevice{},
		done:         make(chan bool, 1),
	}
	return m, m.Resolve()
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
	arpEnMap = map[uint16]string{}
	arpEnMap[1] = "ARP请求"
	arpEnMap[2] = "ARP响应"
	HTTPMethodsMap = []string{"GET", "POST", "DELETE", "HEAD", "OPTIONS", "PUT", "TRACE"}
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
			err := m.processPacketPayload(p)
			if err != nil {
				m.Log.Warn.Printf("处理收到的包异常", err)
			}
		}
	}
}

func (m *Metadata) processPacketPayload(receive *packet) error {
	switch receive.payloadType {
	case "TCP":
		bys := receive.bytes[:4]
		if utils.Includes(HTTPMethodsMap, string(bys)) {
			_ = m.processHTTP(receive)
		}
	}
	return nil
}

func (m *Metadata) processHTTP(receive *packet) error {
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

func (m *Metadata) listenIPv4(ipv4 *layers.IPv4) {
	m.Log.Debug.Println(fmt.Sprintf("源IP: %v,目标IP: %v", ipv4.SrcIP, ipv4.DstIP))
}

func (m *Metadata) listenTCP(tcp *layers.TCP, re chan<- *packet) {
	m.Log.Debug.Println(fmt.Sprintf("源端口: %v,目标端口: %v, seq: %v,ack: %v",
		tcp.SrcPort, tcp.DstPort, tcp.Seq, tcp.Ack))
	p := tcp.Payload
	pl := len(p)
	if pl > 0 {
		re <- &packet{bytes: p, byteLen: pl, payloadType: "TCP"}
	}
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
