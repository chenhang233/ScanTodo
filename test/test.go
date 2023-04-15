package main

import (
	"flag"
	"fmt"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"net"
	"time"
)

func timeToBytes(t time.Time) []byte {
	nsec := t.UnixNano()
	b := make([]byte, 8)
	for i := uint8(0); i < 8; i++ {
		b[i] = byte((nsec >> ((7 - i) * 8)) & 0xff)
	}
	return b
}

func main() {
	timeout := flag.Duration("t", time.Second*100, "持续时间")
	interval := flag.Duration("i", time.Second, "间隔时间")
	size := flag.Int("s", 24, "数据包大小")
	count := flag.Int("c", -1, "ping次数")
	privileged := flag.Bool("privileged", false, "")
	ttl := flag.Int("l", 64, "TTL存活时间")

	flag.Parse()

	fmt.Println(flag.NArg())
	fmt.Println(flag.Args())

	fmt.Println(*timeout, "timeout", interval, "interval", ttl, "ttl", *privileged, "privileged")
	fmt.Println(*size, "size", *count, "count")
	target, _ := net.ResolveIPAddr("ip4:icmp", "192.168.0.1")

	packet, err := icmp.ListenPacket("ip4:icmp", "192.168.0.119")
	fmt.Println(err, "err")
	t := make([]byte, 0, 5)
	t = append(t, 23, 86, 8, 227, 61, 219, 212, 24, 20, 176)
	body := &icmp.Echo{
		ID:   35006,
		Seq:  0,
		Data: t,
	}

	msg := &icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Code: 0,
		Body: body,
	}
	msgBytes, _ := msg.Marshal(nil)
	to, err := packet.WriteTo(msgBytes, target)
	fmt.Println("写", to, err)
	//bys := make([]byte, 0, 1024)
	//go func() {
	//	for {
	//		message, err := icmp.ParseMessage(1, bys)
	//		fmt.Println(string(bys), "结果", message, message, message, "-----", err)
	//	}
	//}()
}
