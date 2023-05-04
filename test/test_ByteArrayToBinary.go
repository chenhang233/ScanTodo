package main

import (
	"fmt"
	"strings"
)

func main() {
	by := []byte{71, 69, 84}
	fmt.Println(string(by))
	sb := strings.Builder{}
	fmt.Println(sb.String() == "")
	//utils.ByteArrayToBinary(bys)
}
