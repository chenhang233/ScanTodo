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
	var count, i int
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

	//startInts := make([]uint8, 0, 4)
	//endInts := make([]uint8, 0, 4)
	ipsArr = make([]string, 0)

	//for i, v := range starts {
	//	v2, _ := strconv.ParseUint(v, 10, 16)
	//	v3, _ := strconv.ParseUint(ends[i], 10, 16)
	//	startInts = append(startInts, uint8(v2))
	//	endInts = append(endInts, uint8(v3))
	//}
	//fmt.Println(startInts, "startInts")
	//fmt.Println(endInts, "endInts")
	fmt.Println(i, "i 需要加数的 最大位")
	endI, _ := strconv.ParseUint(ends[i], 10, 16)
	fmt.Println(endI, "endI")
	f := 3
	ipsArr = append(ipsArr, strings.Join(starts, "."))
A:
	for {
		pu, _ := strconv.ParseUint(starts[f], 10, 16)
		pu++
		for pu > 255 {
			starts[f] = "0"
			f--
			pu2, _ := strconv.ParseUint(starts[f], 10, 16)
			pu = pu2 + 1
		}
		starts[f] = strconv.FormatUint(pu, 10)
		ipsArr = append(ipsArr, strings.Join(starts, "."))
		count++
		if f == i && pu == endI {
			break A
		}
	}
	f++ // 正向走--
	for {
		pu, _ := strconv.ParseUint(starts[f], 10, 16)
		pu++

	}

	fmt.Println("ip总数 count", count)
	return ipsArr, nil
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
