package arp

import (
	"ScanTodo/scan/handle_http"
	"ScanTodo/scan/handle_http/handle_tls"
	"ScanTodo/scanLog"
	"ScanTodo/utils"
	"errors"
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"golang.org/x/sync/errgroup"
	"net"
	"strconv"
	"strings"
	"time"
)

type NetworkDevice struct {
	Name        string
	Ip          net.IP
	Description string
	Mac         net.HardwareAddr // 自定义mac地址
	MAC         net.HardwareAddr // ip对应真实mac地址(SourceDevice)
}

type IpTcpRulesStr struct {
	SourceIps        string `json:"sourceIps"`
	DestinationIps   string `json:"destinationIps"`
	SourcePorts      string `json:"sourcePorts"`
	DestinationPorts string `json:"destinationPorts"`
}

type ipTcpRules struct {
	SourceIps        [2][4]uint8
	DestinationIps   [2][4]uint8
	SourcePorts      [][2]uint16
	DestinationPorts [][2]uint16
}

// ConfigInfo 使用者配置信息
type ConfigInfo struct {
	Op              uint16        `json:"op"`
	T               time.Duration `json:"t"`
	I               time.Duration `json:"i"`
	SIP             string        `json:"sIP"`
	SMAC            string        `json:"sMAC"`
	Smac            string        `json:"Smac"`
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
	OnReceive func(*Metadata, string)
	// 配置IP/TCP监听规则
	EnableRuleIpTcp bool
	*IpTcpRulesStr
	IpTcpRules *ipTcpRules
	// 结束信号
	done chan bool
}

// packet 接收到的数据包
type packet struct {
	payloadType string
	bytes       []byte
	byteLen     int
}

const (
	Sep1 = "-"
	Sep2 = "."
	Sep3 = ","
)

var HTTPMethodsMap []string
var HTTPResponseRow string
var TLS string
var arpEnMap map[uint16]string

func New(c *ConfigInfo) (*Metadata, error) {
	var err error
	m := &Metadata{
		CurrentIndex:    -1,
		Timeout:         c.T,
		Interval:        c.I,
		Operation:       c.Op,
		SelfDevice:      &NetworkDevice{Ip: net.ParseIP(c.HostIp)},
		TargetDevice:    &NetworkDevice{Ip: net.ParseIP(c.TIP)},
		SourceDevice:    &NetworkDevice{Ip: net.ParseIP(c.SIP)},
		done:            make(chan bool, 1),
		EnableRuleIpTcp: c.EnableRuleIpTcp,
		IpTcpRulesStr:   &IpTcpRulesStr{},
		IpTcpRules:      &ipTcpRules{},
	}
	err = m.new(c)
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
	TLS = "Version: TLS"
	hw1, err := net.ParseMAC(c.SMAC)
	if err != nil {
		return errors.New("c.SMAC, " + err.Error())
	}
	m.SourceDevice.Mac = hw1
	hw12, err := net.ParseMAC(c.Smac)
	if err != nil {
		return errors.New("c.Smac, " + err.Error())
	}
	m.SourceDevice.MAC = hw12
	hw2, err := net.ParseMAC(c.TMAC)
	if err != nil {
		return errors.New("c.TMAC, " + err.Error())
	}
	m.TargetDevice.Mac = hw2
	if m.EnableRuleIpTcp {
		r := c.IpTcpRules
		m.SourceIps = r.SourceIps
		m.SourcePorts = r.SourcePorts
		m.DestinationIps = r.DestinationIps
		m.DestinationPorts = r.DestinationPorts
		err := m.ResolveIpTcpRule()
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *Metadata) ResolveIpTcpRule() error {
	var err error
	str := m.IpTcpRulesStr
	rules := m.IpTcpRules
	if str.SourceIps != "" {
		rules.SourceIps = [2][4]uint8{}
		err = m.resolveRuleIp(&rules.SourceIps, &str.SourceIps)
		if err != nil {
			return err
		}
	}
	if str.DestinationIps != "" {
		rules.DestinationIps = [2][4]uint8{}
		err = m.resolveRuleIp(&rules.DestinationIps, &str.DestinationIps)
		if err != nil {
			return err
		}
	}
	if str.SourcePorts != "" {
		rules.SourcePorts, err = m.resolveRulePorts(&str.SourcePorts)
		if err != nil {
			return err
		}
	}
	if str.DestinationPorts != "" {
		rules.DestinationPorts, err = m.resolveRulePorts(&str.DestinationPorts)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *Metadata) resolveRulePorts(portsStr *string) ([][2]uint16, error) {
	var portsArr [][2]uint16
	ports := strings.Split(*portsStr, Sep3)
	for _, port := range ports {
		p := strings.Split(port, Sep1)
		if len(p) > 2 {
			return nil, errors.New("端口范围 a-b,c-d,f")
		}
		if len(p) == 1 {
			p = append(p, p[0])
		}
		arr := [2]uint16{}
		for index, s := range p {
			cur, err := strconv.Atoi(s)
			if err != nil {
				return nil, err
			}
			arr[index] = uint16(cur)
		}
		portsArr = append(portsArr, arr)
	}
	return portsArr, nil
}

func (m *Metadata) resolveRuleIp(ipsRule *[2][4]uint8, ipsStr *string) error {
	ips := strings.Split(*ipsStr, Sep1)
	fmt.Println(ips, "ips")
	if len(ips) > 2 {
		return errors.New("ip范围 a-b")
	}
	if len(ips) == 1 {
		ips = append(ips, ips[0])
	}
	for index, ip := range ips {
		ip4Arr := strings.Split(ip, Sep2)
		if len(ip4Arr) != 4 {
			return errors.New("非ipv4协议")
		}
		arr := [4]uint8{}
		for i, v := range ip4Arr {
			a, err := strconv.Atoi(v)
			if err != nil {
				return err
			}
			arr[i] = uint8(a)
		}
		ipsRule[index] = arr
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
	if err != nil {
		return err
	}
	m.SelfDevice = dev
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
		m.Log.Warn.Printf("End of exception:", err)
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
			return err
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
				m.Log.Warn.Printf("readPacketPayload err:", err)
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
					m.Log.Error.Println("ReadHTTP error: ", err)
				}
				m.Log.Info.Println("http Row: \n", h.Row)
				m.Log.Info.Println("http Head: \n", h.Head)
				//for k, v := range h.Head {
				//	m.Log.Info.Println(fmt.Sprintf("%s:%s", k, v))
				//}
				m.Log.Info.Print(fmt.Sprintf("body:\n %s", h.Body))
			} else {
				f, tls := handle_tls.IsTlsProtocol(receive.bytes)
				if f {
					m.Log.Info.Print(fmt.Sprintf("record layer: (ContentType: %v) (Version: %v) (Length: %v)",
						tls.ContentType, tls.Version, tls.Length))
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
			// 链路层
			ethLayer := p.Layer(layers.LayerTypeEthernet)
			if ethLayer != nil {
				eth := ethLayer.(*layers.Ethernet)
				// IP层
				ipv4Layer := p.Layer(layers.LayerTypeIPv4)
				if ipv4Layer != nil {
					ipv4 := ipv4Layer.(*layers.IPv4)
					if eth.DstMAC.String() == m.SelfDevice.Mac.String() {
						if ipv4.SrcIP.Equal(m.TargetDevice.Ip) {
							data := m.PacketHandler(p)
							err := handle.WritePacketData(data)
							if err != nil {
								m.Log.Warn.Println(err)
							}
						}
					}
					m.listenIPv4(ipv4)
					// TCP 层
					tcpLayer := p.Layer(layers.LayerTypeTCP)
					if tcpLayer != nil {
						tcp := tcpLayer.(*layers.TCP)
						m.listenTCP(tcp, pks)
					}
				}
				// ARP层
				arpLayer := p.Layer(layers.LayerTypeARP)
				if arpLayer != nil {
					arp := arpLayer.(*layers.ARP)
					m.listenARP(arp)
				}
			}
		}
	}
}

func (m *Metadata) listenARP(arp *layers.ARP) {
	mac1 := net.HardwareAddr(arp.SourceHwAddress)
	mac2 := net.HardwareAddr(arp.DstHwAddress)
	reply := fmt.Sprintf("listen %s (source MAC: %v,target MAC: %v)", arpEnMap[arp.Operation], mac1, mac2)
	//m.Log.Info.Printf(reply)
	receive := m.OnReceive
	if receive != nil {
		receive(m, reply)
	}
}

func (m *Metadata) listenIPv4(ipv4 *layers.IPv4) int {
	v4Info := fmt.Sprintf("source IP: %v,target IP: %v", ipv4.SrcIP, ipv4.DstIP)
	if m.EnableRuleIpTcp {
		m.listenIPv4Filter(ipv4, &v4Info)
		return 2
	}
	m.Log.Debug.Println(v4Info)
	return 1
}

func (m *Metadata) listenTCP(tcp *layers.TCP, re chan<- *packet) {
	info := fmt.Sprintf("source port: %v,target port: %v, seq: %v,ack: %v",
		tcp.SrcPort, tcp.DstPort, tcp.Seq, tcp.Ack)
	if m.EnableRuleIpTcp {
		m.listenTCPFilter(tcp, re, &info)
		return
	}
	m.Log.Debug.Println(info)
	m.tcpPacketPayloadSendChannel(tcp, re)
}

func (m *Metadata) tcpPacketPayloadSendChannel(tcp *layers.TCP, re chan<- *packet) {
	pl := len(tcp.Payload)
	if pl > 0 {
		re <- &packet{bytes: tcp.Payload, byteLen: pl, payloadType: "TCP"}
	}
}

func (m *Metadata) listenTCPFilter(tcp *layers.TCP, re chan<- *packet, info *string) {
	sArr := m.IpTcpRules.SourcePorts
	srcP := tcp.SrcPort
	dstP := tcp.DstPort
	f1 := true
	f2 := true
	for _, arr := range sArr {
		min := arr[0]
		max := arr[1]
		if uint16(srcP) < min || uint16(srcP) > max {
			f1 = false
		}
	}
	dArr := m.IpTcpRules.DestinationPorts
	for _, arr := range dArr {
		min := arr[0]
		max := arr[1]
		if uint16(dstP) < min || uint16(dstP) > max {
			f2 = false
		}
	}
	if f1 || f2 {
		m.Log.Debug.Println(*info)
		m.tcpPacketPayloadSendChannel(tcp, re)
	}
}

func (m *Metadata) listenIPv4Filter(ipv4 *layers.IPv4, v4Info *string) {
	s1 := m.IpTcpRules.SourceIps[0]
	s2 := m.IpTcpRules.SourceIps[1]
	curS := ipv4.SrcIP.To4()
	f1 := true
	f2 := true
	for i, cur := range curS {
		if !m.CompareByte(cur, s1[i], s2[i]) {
			f1 = false
			break
		}
	}
	d1 := m.IpTcpRules.DestinationIps[0]
	d2 := m.IpTcpRules.DestinationIps[1]
	curD := ipv4.DstIP.To4()
	for i, cur := range curD {
		if !m.CompareByte(cur, d1[i], d2[i]) {
			f2 = false
			break
		}
	}
	if f1 || f2 {
		m.Log.Debug.Println(*v4Info)
	}
}

func (m *Metadata) CompareByte(by, min, max byte) bool {
	if by >= min && by <= max {
		return true
	}
	return false
}

// PacketHandler 代理包转发,修改成真正的MAC地址
func (m *Metadata) PacketHandler(packet gopacket.Packet) []byte {
	data := packet.Data()
	//layer := packet.Layer(layers.LayerTypeIPv4)
	//ipLayer := layer.(*layers.IPv4)
	//layer = packet.Layer(layers.LayerTypeEthernet)
	//ethLayer := layer.(*layers.Ethernet)
	//m.Log.Info.Println("替换之前:\n", data[:6], string(data[:6]))
	dstMac := m.SourceDevice.MAC
	for i := 0; i < len(dstMac); i++ {
		data[i] = dstMac[i]
	}
	m.Log.Info.Println(data)
	return data
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
