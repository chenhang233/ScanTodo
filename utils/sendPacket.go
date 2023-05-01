package utils

import (
	"fmt"
	"net"
)

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
