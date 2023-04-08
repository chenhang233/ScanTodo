package scan

import (
	"ScanTodo/scanLog"
	"ScanTodo/utils"
	"context"
	"encoding/json"
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
	t.Log.Info.Println("----------------------")
	t.Log.Info.Println("准备扫描的ip数量: ", count)
	t.scanIps(ips, ports)
	t.Log.Info.Println("----------------------")
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
			for _, ip := range ips {
				t.scanIp(ip, ports, portPageGroupLen)
			}
			group.Done()
		}(ips, ports)
	}
	group.Wait()
	fmt.Println("扫描结束")
}

func (t *TcpScan) scanIp(ip string, ports []uint16, portPageGroupLen int) {
	group := sync.WaitGroup{}
	start := 0
	for i := 0; i < portPageGroupLen; i++ {
		portSlice := make([]uint16, 0, t.portPage)
		if i == portPageGroupLen-1 {
			portSlice = append(portSlice, ports[start:]...)
		} else {
			portSlice = append(portSlice, ports[start:start+t.portPage]...)
			start += t.portPage
		}
		group.Add(1)
		go func(ip string, portSlice []uint16) {
			for _, port := range portSlice {
				t.isOpen(ip, port, time.Duration(t.body.Timeout))
			}
			group.Done()
		}(ip, portSlice)
	}
}

func (t *TcpScan) isOpen(ip string, port uint16, timeout time.Duration) bool {
	dt, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", ip, port), timeout)
	if err != nil {
		sf := fmt.Sprintf("失败: %v", err)
		js, _ := json.Marshal(sf)
		utils.HubInstance.PrivateClient.Send <- js
		return false
	}
	sf := fmt.Sprintf("成功: %v, ip: %s , 端口: %d", dt, ip, port)
	js, _ := json.Marshal(sf)
	utils.HubInstance.PrivateClient.Send <- js
	return true
}
