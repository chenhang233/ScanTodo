package main

import (
	"ScanTodo/scan/arp"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
)

var usage = `
Usage:

	-sIp -sMAC -tIp -tMAC -op ip

Examples:

    # 发送arp报文 
     -sIp 192.168.232.144 -sMAC 00-50-56-C0-00-08 -tIp 192.168.232.2 -tMAC 00-50-56-e3-5e-65 192.168.232.1
`

func main() {
	op := flag.Int("op", 1, "操作 1请求2回复")
	sIp := flag.String("sIp", "", "发送方ip")
	sMAC := flag.String("sMAC", "", "发送方MAC")
	tIp := flag.String("tIp", "", "接收方ip")
	tMAC := flag.String("tMAC", "", "接收方MAC")
	flag.Usage = func() {
		print(usage)
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() == 0 {
		fmt.Println("没有本机IP !")
		flag.Usage()
		return
	}
	host := flag.Arg(0)
	metadata, err := arp.New(host)
	if err != nil {
		fmt.Println(err)
		return
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for i := range c {
			metadata.Log.Debug.Println("接收到中断的信号:", i)
			metadata.Stop()
		}
	}()

	metadata.Operation = *op
	metadata.SourceDevice.Ip = net.ParseIP(*sIp)
	hw1, err := net.ParseMAC(*sMAC)
	if err != nil {
		metadata.Log.Warn.Println(err)
		return
	}
	metadata.SourceDevice.Mac = hw1
	metadata.TargetDevice.Ip = net.ParseIP(*tIp)
	hw2, err := net.ParseMAC(*tMAC)
	if err != nil {
		metadata.Log.Warn.Println(err)
		return
	}
	metadata.TargetDevice.Mac = hw2

	err = metadata.Run()
	if err != nil {
		metadata.Log.Debug.Println("异常结束日志:", err)
	}
}
