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
	Log  *scanLog.LogConf
	body *TcpReq
}

func (t *TcpScan) Start(ctx context.Context) error {
	var err error
	body := t.body
	ips, err := utils.ReadIps(body.Ip)
	if err != nil {
		err = fmt.Errorf(err.Error())
		return err
	}
	ports, err := utils.ReadPorts(body.Port)
	if err != nil {
		err = fmt.Errorf("port参数错误")
	}
	fmt.Println(":ips", ips)
	fmt.Sprintf("dd %s %v", ips, ports)
	return nil
}

func (t *TcpScan) End(ctx context.Context) error {
	fmt.Println("请求返回之后---")
	return nil
}

func (t *TcpScan) scanIp(ip string, post []uint16) {
	//fmt.Sprintf("【%v】需要扫描端口总数:%v 个，总协程:%v 个，并发:%v 个，超时:%d 毫秒", ip, total, pageCount, num, s.timeout)
	group := sync.WaitGroup{}
	group.Add(1)
}

func (t *TcpScan) isOpen(ip string, port uint16, timeout time.Duration) {
	dt, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", ip, port), timeout)
	if err != nil {
		fmt.Println("错误:", err)
	}
	fmt.Println(dt, "dt")
}
