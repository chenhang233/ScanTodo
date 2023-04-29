package main

import (
	"fmt"
	"strings"
)

func main() {
	str := "80-81,90-91"
	split := strings.Split(str, ",")
	fmt.Println(split)

	for i, v := range split {
		fmt.Println(i, v)
	}
}
