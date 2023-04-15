package main

import (
	"flag"
	"fmt"
	"time"
)

func main() {
	timeout := flag.Duration("t", time.Second*100, "持续时间")
	size := flag.Int("s", 24, "数据包大小")

	flag.Parse()

	fmt.Println(flag.NArg())
	fmt.Println(flag.Args())

	fmt.Println(*timeout, "timeout")
	fmt.Println(*size, "size")
}
