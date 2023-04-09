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

const (
	TcpEndStr = "tcp扫描结束"
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
	t.Log.Warn.Println("------------------启动----------------------")
	countS := fmt.Sprintf("准备扫描的ip数量: %d", count)
	t.Log.Info.Println(countS)
	utils.SendToThePrivateClientCustom(countS)
	t.scanIps(ips, ports)
	t.Log.Warn.Println("-------------------结束---------------------")
	return nil
}

func (t *TcpScan) End(ctx context.Context) error {
	fmt.Println("请求返回之后---")
	return nil
}

func (t *TcpScan) scanIps(ip []string, ports []uint16) {
	var ipPageGroupLen int
	var portPageGroupLen int
	utils.ComputedGroupCount(&ipPageGroupLen, len(ip), t.ipPage)
	utils.ComputedGroupCount(&portPageGroupLen, len(ports), t.portPage)
	str1 := fmt.Sprintf("ip go程 %d 个", ipPageGroupLen)
	str2 := fmt.Sprintf("每个ip准备访问的端口 go程 %d 个", portPageGroupLen)
	str3 := fmt.Sprintf("当前超时时间设置为 %d 毫秒", t.body.Timeout)

	t.Log.Info.Println(str1)
	t.Log.Info.Println(str2)
	t.Log.Info.Println(str3)
	utils.SendToThePrivateClientCustom(str1)
	utils.SendToThePrivateClientCustom(str2)
	utils.SendToThePrivateClientCustom(str3)

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
	t.Log.Info.Println(TcpEndStr)
	utils.SendToThePrivateClientCustom(TcpEndStr)
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
				t.isOpen("tcp", ip, port, time.Duration(t.body.Timeout))
			}
			group.Done()
		}(ip, portSlice)
	}
	group.Wait()
}

func (t *TcpScan) isOpen(network string, ip string, port uint16, timeout time.Duration) bool {
	conn, err := net.DialTimeout(network, fmt.Sprintf("%s:%d", ip, port), timeout)
	if err != nil {
		utils.SendToThePrivateClientMsgError(ip, port, "tcp", err.Error())
		return false
	}
	msg := utils.SendToThePrivateClientMsgSuccess(ip, port, "tcp")
	t.Log.Info.Println(msg)
	_ = conn.Close()
	return true
}
