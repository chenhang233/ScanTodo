package arp

import (
	"ScanTodo/scanLog"
	"errors"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
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
	Operation int
	Log       *scanLog.LogConf
	// 当前发送索引
	CurrentIndex int
	// 发包间隔时间
	Interval time.Duration
	// 发包持续时间
	Timeout  time.Duration
	OnFinish func(*Metadata)
}

func New(ip string) (*Metadata, error) {
	m := &Metadata{
		SelfDevice: &NetworkDevice{Ip: net.IP(ip)},
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

}

func (m *Metadata) Run() error {
	return nil
}

/*
SendArp
Operation  ARP请求为1 ARP响应为2
*/
func (*Metadata) SendArp(handle *pcap.Handle, Operation uint16, sourceMAC, targetMAC net.HardwareAddr, sourceIp, targetIp net.IP) error {
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
		SourceProtAddress: sourceIp,
		DstHwAddress:      targetMAC,
		DstProtAddress:    targetIp,
	}
	buffer := gopacket.NewSerializeBuffer()
	var opt gopacket.SerializeOptions
	err = gopacket.SerializeLayers(buffer, opt, eth, a)
	if err != nil {
		return err
	}
	outgoingPacket := buffer.Bytes()
	err = handle.WritePacketData(outgoingPacket)
	if err != nil {
		return err
	}
	return nil
}

// GetPcapDev 获取要使用的网卡信息
func (*Metadata) getPcapDev(ip net.IP) (*NetworkDevice, error) {
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
	return nil, errors.New("网卡异常,没有匹配ip")
}
