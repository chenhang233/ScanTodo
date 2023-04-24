package main

import (
	"ScanTodo/scan/ping"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"time"
)

var usage = `
Usage:

    ping [-c] [-i] [-t] host

Examples:

    # ping 命令
    ping www.baidu.com

    # ping  5次
    ping -c 5 www.baidu.com

    # ping 5次 间隔500毫秒
    ping -c 5 -i 500ms  www.baidu.com

    # ping 指定时间 10秒
    ping -t 10s www.baidu.com

    #  发送 100字节 的包
    ping -s 100 www.baidu.com
`

func main() {
	timeout := flag.Duration("t", time.Second*5, "持续时间")
	interval := flag.Duration("i", time.Second, "间隔时间")
	size := flag.Int("s", 24, "数据包内容大小")
	count := flag.Int("c", -1, "ping次数")
	ttl := flag.Int("l", 64, "包存活时间")
	flag.Usage = func() {
		print(usage)
		flag.PrintDefaults()
	}
	flag.Parse()
	if flag.NArg() == 0 {
		fmt.Println("没有目标IP !")
		flag.Usage()
		return
	}
	host := flag.Arg(0)
	pingMetadata, err := ping.NewPingMetadata(host)
	if err != nil {
		fmt.Println(err)
		return
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for i := range c {
			pingMetadata.Log.Debug.Println("接收到中断的信号:", i)
			pingMetadata.Stop()
		}
	}()
	pingMetadata.Count = *count
	pingMetadata.Size = *size
	pingMetadata.Interval = *interval
	pingMetadata.Timeout = *timeout
	pingMetadata.TTL = *ttl
	pingMetadata.OnSetup = func() {
		pingMetadata.Log.Info.Println("OnSetup callback")
	}
	pingMetadata.OnSend = func(packet *ping.Packet) {
		pingMetadata.Log.Info.Printf(fmt.Sprintf("OnSend source: %s,destination: %s,byteLen: %v,Sequence: %v,Identifier: %v,RTTS: %v",
			pingMetadata.Source, packet.IPAddr, packet.ByteLen, packet.Sequence, packet.Identifier, packet.RTT))
	}
	pingMetadata.OnReceive = func(packet *ping.Packet) {
		pingMetadata.Log.Info.Printf(fmt.Sprintf("OnReceive  source: %s, destination: %s,byteLen: %v,Sequence: %v,Identifier: %v,RTTS: %v",
			packet.IPAddr, pingMetadata.Source, packet.ByteLen, packet.Sequence, packet.Identifier, packet.RTT))
	}
	pingMetadata.OnDuplicateReceive = func(packet *ping.Packet) {
		pingMetadata.Log.Warn.Printf(fmt.Sprintf("OnDuplicateReceive  source: %s, destination: %s,byteLen: %v,Sequence: %v,Identifier: %v,RTTS: %v",
			packet.IPAddr, pingMetadata.Source, packet.ByteLen, packet.Sequence, packet.Identifier, packet.RTT))
	}
	pingMetadata.OnFinish = func(statistics *ping.Statistics) {
		pingMetadata.Log.Info.Println("OnFinish callback")

	}
	pingMetadata.Log.Info.Println(fmt.Sprintf("开始ping 地址 %s (%s): ", pingMetadata.Addr, pingMetadata.Ipaddr))
	err = pingMetadata.Run()
	if err != nil {
		pingMetadata.Log.Debug.Println("结束日志:", err)
	}
}
