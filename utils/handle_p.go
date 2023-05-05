package utils

import (
	"bytes"
	"fmt"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"io/ioutil"
	"time"
)

/*
TimeToBytes 按位取

	start 56 48 40 32 24 16 8 0 >> 每次取8位
*/
func TimeToBytes(t time.Time) []byte {
	nano := t.UnixNano()
	b := make([]byte, 8)
	for i := uint8(0); i < 8; i++ {
		start := (7 - i) * 8
		b[i] = byte((nano >> start) & 0xff)
	}
	BytesToTime(b)
	return b
}

func BytesToTime(b []byte) time.Time {
	var nano int64
	for i := uint8(0); i < 8; i++ {
		start := (7 - i) * 8
		nano += int64(b[i]) << start
	}
	return time.Unix(nano/1000000000, nano%1000000000)
}

func ComputedGroupCount(res *int, count int, pageSize int) {
	if count < pageSize {
		*res = 1
	} else {
		*res = count/pageSize + 1
	}
}

func decimalConversion(n uint8, base uint8) []uint8 {
	i := n
	var b []uint8
	for i > 0 {
		b = append(b, i%base)
		i /= base
	}
	return b
}

func Includes(arr []string, content string) bool {
	for i := range arr {
		if content == arr[i] {
			return true
		}
	}
	return false
}

func ByteArrayToBinary(bys []byte) {
	//var b []uint
	str := ""
	//for _, v := range bys {
	//fmt.Println(v, "vvv")
	//fmt.Println(decimalConversion(v, 2))
	//s := string(v)
	//str += s
	//}
	fmt.Println(str, "sss")
	fmt.Println("------------------------")
}

func GbToUtf8(s []byte) ([]byte, error) {
	reader := transform.NewReader(bytes.NewReader(s), simplifiedchinese.GBK.NewDecoder())
	d, e := ioutil.ReadAll(reader)
	if e != nil {
		return nil, e
	}
	return d, nil
}
