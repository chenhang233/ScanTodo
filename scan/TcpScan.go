package scan

import (
	"ScanTodo/scanLog"
	"ScanTodo/utils"
	"context"
	"fmt"
	"net"
	"sync"
	"time"
)

type TcpReq struct {
	Ip      string `json:"ip"`
	Port    string `json:"port"`
	Timeout int    `json:"timeout"`
}

type TcpScan struct {
	Log      *scanLog.LogConf
	body     *TcpReq
	ipPage   int
	portPage int
}

func (t *TcpScan) Start(ctx context.Context) error {
	var err error
	body := t.body
	ips, count, err := utils.ReadIps(body.Ip)
	if err != nil {
		err = fmt.Errorf(err.Error())
		return err
	}
	ports, err := utils.ReadPorts(body.Port)
	if err != nil {
		err = fmt.Errorf("port参数错误")
	}
	t.Log.Info.Println("准备扫描的ip数量: ", count)
	t.scanIps(ips, ports)
	return nil
}

func (t *TcpScan) End(ctx context.Context) error {
	fmt.Println("请求返回之后---")
	return nil
}

func (t *TcpScan) scanIps(ip []string, ports []uint16) {
	//fmt.Sprintf("需要扫描ip总数:%v 个，总协程:%v 个，并发:%v 个，超时:%d 毫秒", count, total, pageCount, num, s.timeout)
	var ipPageGroupLen int
	var portPageGroupLen int
	utils.ComputedGroupCount(&ipPageGroupLen, len(ip), t.ipPage)
	utils.ComputedGroupCount(&portPageGroupLen, len(ports), t.portPage)

	group := sync.WaitGroup{}
	start := 0
	for i := 0; i < ipPageGroupLen; i++ {
		ips := make([]string, 0, t.ipPage)
		if i == ipPageGroupLen-1 {
			ips = append(ips, ip[start:]...)
		} else {
			ips = append(ips, ip[start:start+t.ipPage]...)
			start += t.ipPage
		}
		group.Add(1)
		go func(ips []string, ports []uint16) {
			fmt.Println(len(ips), "ips")
			for _, ip := range ips {
				t.scanIp(ip, ports, portPageGroupLen)
			}
			group.Done()
		}(ips, ports)
	}

}

func (t *TcpScan) scanIp(ip string, ports []uint16, portPageGroupLen int) {
	//group := sync.WaitGroup{}
	//for i := 0; i < portPageGroupLen; i++ {
	//	group.Add(1)
	//	go func() {
	//
	//	}()
	//}
}

func (t *TcpScan) isOpen(ip string, port uint16, timeout time.Duration) {
	dt, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", ip, port), timeout)
	if err != nil {
		fmt.Println("错误:", err)
	}
	fmt.Println(dt, "dt")
}
