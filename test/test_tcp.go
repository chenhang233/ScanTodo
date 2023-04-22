package main

import (
	"fmt"
	"net"
	"time"
)

func main() {
	ip, _ := net.ResolveTCPAddr("tcp", "127.0.0.1:8888")
	dialer := net.Dialer{LocalAddr: ip}
	dial, err := dialer.Dial("tcp", "192.168.232.141:8000")
	fmt.Println(err, "err", dial, "dial")
	time.Sleep(time.Second * 5)
	fmt.Println("关闭连接")
	dial.Close()
}
