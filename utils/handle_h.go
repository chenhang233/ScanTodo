package utils

import (
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
)

const (
	LoopbackAddress = "127.0.0.1"
)

type Methods string

var MethodsMap = map[string]string{
	"GET":  "GET",
	"POST": "POST",
}

type Reader interface {
	Read(p []byte) (n int, err error)
}

func MyReadAll(r Reader) ([]byte, error) {
	b := make([]byte, 0, 512)
	for {
		if len(b) == cap(b) {
			b = append(b, 0)[:len(b)]
		}
		n, err := r.Read(b[len(b):cap(b)])
		b = b[:len(b)+n]
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			return b, err
		}
	}
}

func CheckIpv4(ip string) bool {
	reqExp := "^(1\\d{2}|2[0-5]\\d|25[0-5]|\\d\\d|[1-9])\\.(1\\d{2}|2[0-5]\\d|25[0-5]|\\d\\d|[0-9])\\.(1\\d{2}|2[0-5]\\d|25[0-5]|\\d\\d|[0-9])\\.(1\\d{2}|2[0-5]\\d|25[0-5]|\\d\\d|[0-9])$"
	cp, _ := regexp.Compile(reqExp)
	return cp.MatchString(ip)
}

func ReadIps(ips string) ([]string, error) {
	var ipsArr []string
	if !strings.Contains(ips, "-") {
		flag := CheckIpv4(ips)
		if !flag {
			return nil, fmt.Errorf("ip错误")
		}
		ipsArr = append(ipsArr, ips)
		return ipsArr, nil
	}
	sp := strings.Split(ips, "-")
	start := sp[0]
	end := sp[1]
	flag1 := CheckIpv4(start)
	flag2 := CheckIpv4(end)
	if !flag1 || !flag2 {
		return nil, errors.New("开始或结束端口错误")
	}
	starts := strings.Split(start, ".")
	ends := strings.Split(end, ".")
	var index, i int
	for {
		if starts[i] < ends[i] {
			break
		} else if starts[i] > ends[i] {
			return nil, errors.New("开始端口大于结束端口")
		} else {
			if i == 3 {
				ipsArr = append(ipsArr, start)
				return ipsArr, nil
			}
			i++
		}
	}
	fmt.Println(starts, "starts")
	fmt.Println(ends, "end")
	fmt.Println(index, "index")
	return starts, nil
}

func ReadPorts(ports string) ([]uint16, error) {
	ps := strings.Split(ports, ",")
	portList := make([]uint16, 0, 65536)
	for _, p := range ps {
		if strings.Contains(p, "-") {
			rangePort := strings.Split(p, "-")
			start, err1 := strconv.ParseUint(rangePort[0], 10, 16)
			end, err2 := strconv.ParseUint(rangePort[1], 10, 16)
			if err2 != nil || err1 != nil {
				return nil, fmt.Errorf("端口范围有问题？")
			}
			for i := start; i <= end; i++ {
				portList = append(portList, uint16(i))
			}
			break
		}
		port, err := strconv.ParseUint(p, 10, 16)
		if err != nil {
			return nil, fmt.Errorf("端口有问题？")
		}
		portList = append(portList, uint16(port))
	}
	return portList, nil
}
