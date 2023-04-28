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
     127.0.0.1

    # ping  5次
     -c 5 127.0.0.1

    # ping 5次 间隔500毫秒
     -c 5 -i 500ms  127.0.0.1

    # ping 指定持续时间 10秒
     -t 10s 127.0.0.1

    #  发送 100字节 的包
     -s 100 127.0.0.1
`

func main() {
	timeout := flag.Duration("t", time.Second*4, "持续总时间")
	interval := flag.Duration("i", time.Second, "间隔时间")
	size := flag.Int("s", 24, "数据包内容大小")
	count := flag.Int("c", 4, "ping次数")
	ttl := flag.Int("l", 64, "IPv4 包过路由器次数")
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
	fmt.Println(flag.Args())
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
		pingMetadata.Log.Info.Printf(fmt.Sprintf("OnSend source: %s,destination: %s,byteLen: %v,Sequence: %v,Identifier: %v",
			pingMetadata.Source, packet.IPAddr, packet.ByteLen, packet.Sequence, packet.Identifier))
	}
	pingMetadata.OnReceive = func(packet *ping.Packet) {
		pingMetadata.Log.Info.Printf(fmt.Sprintf("OnReceive  source: %s, destination: %s,byteLen: %v,Sequence: %v,Identifier: %v,RTT: %v",
			packet.IPAddr, pingMetadata.Source, packet.ByteLen, packet.Sequence, packet.Identifier, packet.RTT))
	}
	pingMetadata.OnDuplicateReceive = func(packet *ping.Packet) {
		pingMetadata.Log.Warn.Printf(fmt.Sprintf("OnDuplicateReceive  source: %s, destination: %s,byteLen: %v,Sequence: %v,Identifier: %v,RTTS: %v",
			packet.IPAddr, pingMetadata.Source, packet.ByteLen, packet.Sequence, packet.Identifier, packet.RTT))
	}
	pingMetadata.OnFinish = func(statistics *ping.Statistics) {
		pingMetadata.Log.Debug.Printf(fmt.Sprintf("OnFinish target: %v, PacketsSent: %v, PacketsReceive: %v, PacketLoss: %v %%,PacketsReceiveDuplicates: %v",
			statistics.Addr, statistics.PacketsSent, statistics.PacketsReceive, statistics.PacketLoss, statistics.PacketsReceiveDuplicates))
	}
	pingMetadata.Log.Info.Println(fmt.Sprintf("开始ping 地址 %s (%s): ", pingMetadata.Addr, pingMetadata.Ipaddr))
	err = pingMetadata.Run()
	if err != nil {
		pingMetadata.Log.Debug.Println("异常结束日志:", err)
	}
}
