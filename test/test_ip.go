package main

import (
	"fmt"
	"net"
)

func main() {
	ip := net.ParseIP("192.168.0.115")
	for i, b := range ip.To4() {
		fmt.Println(i, "i", b, "b")
	}
}
