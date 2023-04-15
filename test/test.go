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

type packet struct {
	bytes  []byte
	nbytes int
	ttl    int
}

func main() {
	timeout := flag.Duration("t", time.Second*100, "持续时间")
	interval := flag.Duration("i", time.Second, "间隔时间")
	size := flag.Int("s", 24, "数据包大小")
	count := flag.Int("c", -1, "ping次数")
	privileged := flag.Bool("privileged", false, "")
	ttl := flag.Int("l", 64, "TTL存活时间")
	flag.Parse()
	fmt.Println(*timeout, "timeout", interval, "interval", ttl, "ttl", *privileged, "privileged")
	fmt.Println(*size, "size", *count, "count")
	fmt.Println("----------------------")
	target, _ := net.ResolveIPAddr("ip4:icmp", "192.168.0.1")
	packet, _ := icmp.ListenPacket("ip4:icmp", "192.168.0.119")
	t := make([]byte, 0, 5)
	t = append(t, 23, 86, 8, 227, 61, 219, 212, 24,
		20, 176, 45, 23, 86, 8, 227, 61, 219, 212, 24, 20,
		176, 45, 23, 86, 8, 227, 61, 219, 212, 24, 20, 176,
		45, 176, 45, 23, 86, 8, 227, 61, 219, 212, 24, 20,
		176, 45, 23, 86, 8, 227, 61, 219, 212, 24, 20)
	body := &icmp.Echo{
		ID:   1,
		Seq:  1,
		Data: t,
	}

	msg := &icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Code: 0,
		Body: body,
	}
	msgBytes, _ := msg.Marshal(nil)
	go func() {
		to, err := packet.WriteTo(msgBytes, target)
		fmt.Println(to, "to", err, "err")
		time.Sleep(time.Second)
	}()
	for {
		select {
		default:
			bys := make([]byte, len(t)+8+ipv4.HeaderLen)
			n, cm, src, err2 := packet.IPv4PacketConn().ReadFrom(bys)
			fmt.Println(n, "n", cm.String(), "cm", src.String(), "src", err2, "err2")
			fmt.Println(bys, "返回数据")

			m, _ := icmp.ParseMessage(1, bys)
			fmt.Println(m.Type, "m.Type", m.Code, "m.code")
			switch pkt := m.Body.(type) {
			case *icmp.Echo:
				fmt.Println(pkt.ID, "pkt.ID", pkt.Seq, "pkt.Seq", pkt.Data, "返回")
			}
		}
		//time.Sleep(time.Second)
	}
}
