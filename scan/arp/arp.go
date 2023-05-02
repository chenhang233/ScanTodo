package arp

import (
	"ScanTodo/scanLog"
	"bytes"
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

	var g errgroup.Group
	g.Go(func() error {
		defer m.Stop()
		return m.mainLoop(conn)
	})
	g.Go(func() error {
		defer m.Stop()
		return m.listenPacket(conn)
	})
	err := g.Wait()
	if err != nil {
		m.Log.Warn.Printf("异常结束中:", err)
	}
	return err
}

func (m *Metadata) mainLoop(conn *pcap.Handle) error {
	t1 := time.NewTicker(m.Timeout)
	t2 := time.NewTicker(m.Interval)
	defer func() {
		t1.Stop()
		t2.Stop()
	}()
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
		}
	}
}

func (m *Metadata) listenPacket(handle *pcap.Handle) error {
	ps := gopacket.NewPacketSource(handle, handle.LinkType())
	for {
		select {
		case <-m.done:
			return nil
		case p := <-ps.Packets():
			arpLayer := p.Layer(layers.LayerTypeARP)
			if arpLayer != nil {
				arp, _ := arpLayer.(*layers.ARP)
				if bytes.Equal(m.SelfDevice.Mac, arp.SourceHwAddress) {
					continue
				}
				if arp.Operation == layers.ARPReply {
					mac := net.HardwareAddr(arp.SourceHwAddress)
					m.Log.Info.Println(fmt.Sprintf("收到arp回复 (目标MAC: %v,目标IP: %v)", mac, m.TargetDevice.Ip))
					receive := m.OnReceive
					if receive != nil {
						receive(m)
					}
				}
			}
		}
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
