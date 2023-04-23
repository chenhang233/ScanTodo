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
	timeout := flag.Duration("t", time.Second*100, "持续时间")
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
		fmt.Println("参数错误-----")
		flag.Usage()
		return
	}
	host := flag.Arg(0)
	ping, err := ping.NewPingMetadata(host)
	if err != nil {
		fmt.Println(err)
		return
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for i := range c {
			ping.Log.Debug.Println("接收到中断的信号:", i)
			ping.Stop()
		}
	}()
	ping.Count = *count
	ping.Size = *size
	ping.Interval = *interval
	ping.Timeout = *timeout
	ping.TTL = *ttl
	ping.Log.Info.Println(fmt.Sprintf("开始ping 地址 %s (%s): ", ping.Addr, ping.Ipaddr))
	err = ping.Run()
	if err != nil {
		ping.Log.Error.Println("启动错误:", err)
	}
}
