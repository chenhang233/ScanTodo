package main

import (
	"ScanTodo/utils"
	"fmt"
)

func main() {

	//utils.HttpPost()
	dev, err := utils.GetPcapDev("192.168.0.115")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(dev)
}
