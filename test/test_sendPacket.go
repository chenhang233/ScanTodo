package main

import (
	"fmt"
	"net"
)

func main() {
	hw, err2 := net.ParseMAC("00-50-56-C0-00-08")
	fmt.Println(hw, err2)
	//utils.HttpPost()
	//dev, err := utils.GetPcapDev("192.168.232.1")
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}
	//fmt.Println(dev)
}
