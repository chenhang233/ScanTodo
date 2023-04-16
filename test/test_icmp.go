package main

import (
	"fmt"
	"golang.org/x/net/icmp"
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

var Data2 = []byte("abcdefghijklmnopqrstuvwabcdefghi")

func main() {
	packet, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	fmt.Println(err, "err")
	//packet.IPv4PacketConn().SetTTL(10)
	//go read(packet)

	for i := 0; i < 5; i++ {
		write(packet)
	}
}

func write(packet *icmp.PacketConn) {
	//body := &icmp.Echo{
	//	ID:   1,
	//	Seq:  1,
	//	Data: Data2,
	//}
	//msg := &icmp.Message{
	//	Type: ipv4.ICMPTypeEcho,
	//	Code: 0,
	//	Body: body,
	//}
	//msgBytes, _ := msg.Marshal(nil)

	for {
		target, _ := net.ResolveIPAddr("ip4", "192.168.232.141")
		bys := []byte{8, 0, 209, 190, 48, 97, 0, 1, 23, 86, 87, 102, 141, 245, 90, 236, 134, 126, 78, 166, 114, 29, 77, 165, 158, 10, 214, 128, 88, 167, 60, 38}
		fmt.Println(bys, "msgBytes", target, "dst")
		err2 := packet.IPv4PacketConn().SetTTL(64)
		fmt.Println(err2, "err2")
		to, err := packet.WriteTo(bys, target)
		fmt.Println("发", err, to)
		time.Sleep(time.Second)
		break
	}

}

func read(packet *icmp.PacketConn) {
	for {
		select {
		default:
			bys := make([]byte, 1500)
			n, cm, src, err2 := packet.IPv4PacketConn().ReadFrom(bys)
			fmt.Println(n, "n", cm.String(), "cm", src.String(), "src", err2, "err2")
			fmt.Println(bys, "返回数据")
			m, _ := icmp.ParseMessage(1, bys)
			fmt.Println(m.Type, "m.Type", m.Code, "m.code")
		}
	}
}
