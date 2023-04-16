package main

import (
	"flag"
	"fmt"
	"time"
)

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
}
