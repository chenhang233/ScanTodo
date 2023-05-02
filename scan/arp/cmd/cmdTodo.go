package main

import (
	"ScanTodo/scan/arp"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"time"
)

var usage = `
Usage:

	-sIp -sMAC -tIp -tMAC -op [-t] [-i] ip

Examples:

    # 发送arp报文 
     -sIp 192.168.232.144 -sMAC 00-50-56-C0-00-08 -tIp 192.168.232.2 -tMAC 00-50-56-e3-5e-65 192.168.232.1
`

func main() {
	op := flag.Uint("op", 1, "操作 1请求2回复")
	sIp := flag.String("sIp", "", "发送方ip")
	sMAC := flag.String("sMAC", "", "发送方MAC")
	tIp := flag.String("tIp", "", "接收方ip")
	tMAC := flag.String("tMAC", "", "接收方MAC")
	timeout := flag.Duration("t", time.Hour, "最大超时时间")
	interval := flag.Duration("i", time.Minute, "间隔时间")
	flag.Usage = func() {
		print(usage)
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() == 0 {
		fmt.Println("最后一个参数本机指定网卡IP")
		flag.Usage()
		return
	}
	host := flag.Arg(0)
	m, err := arp.New(host)
	if err != nil {
		return
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for i := range c {
			m.Log.Debug.Println("接收到中断的信号:", i)
			m.Stop()
		}
	}()

	m.Operation = uint16(*op)
	m.Timeout = *timeout
	m.Interval = *interval
	m.SourceDevice.Ip = net.ParseIP(*sIp)
	hw1, err := net.ParseMAC(*sMAC)
	if err != nil {
		m.Log.Warn.Println(err)
		return
	}
	m.SourceDevice.Mac = hw1
	m.TargetDevice.Ip = net.ParseIP(*tIp)
	hw2, err := net.ParseMAC(*tMAC)
	if err != nil {
		m.Log.Warn.Println(err)
		return
	}
	m.TargetDevice.Mac = hw2
	s1 := fmt.Sprintf("\n本机信息(ip: %v,mac: %v,Description: %s)", m.SelfDevice.Ip, m.SelfDevice.Mac, m.SelfDevice.Description)
	s2 := fmt.Sprintf("\n发送信息(ip: %v,mac: %v,Description: %s)", m.SourceDevice.Ip, m.SourceDevice.Mac, m.SourceDevice.Description)
	s3 := fmt.Sprintf("\n接收信息(ip: %v,mac: %v,Description: %s)", m.TargetDevice.Ip, m.TargetDevice.Mac, m.TargetDevice.Description)
	m.Log.Info.Println(s1 + s2 + s3)
	err = m.Run()
	if err != nil {
		m.Log.Debug.Println("异常结束日志:", err)
	}
}
