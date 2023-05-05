package main

import (
	"fmt"
)

func main() {
	//ip := net.ParseIP("192.168.0.115")
	//for i, b := range ip.To4() {
	//	fmt.Println(i, "i", b, "b")
	//}
	Split("a:szZ", ":")
}
func Split(line string, separator string) {
	for i, v := range line {
		fmt.Println(i, byte(v))
	}
}
