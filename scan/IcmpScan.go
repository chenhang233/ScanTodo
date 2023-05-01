package scan

import (
	"ScanTodo/scan/ping"
	"ScanTodo/scanLog"
	"ScanTodo/utils"
	"context"
	"fmt"
	"sync"
	"time"
)

const MsgLen = 4000
const pageSize = 30

type ICMP struct {
	Type        uint8
	Code        uint8
	Checksum    uint16
	Identifier  uint16
	SequenceNum uint16
	msgData     [MsgLen]uint8
}

type PingMetadata struct {
	Count    int `json:"count"`
	Size     int `json:"size"`
	Interval int `json:"interval"`
	Timeout  int `json:"timeout"`
	TTL      int `json:"TTL"`
}

type IcmpReq struct {
	Ip string `json:"ip"`
}

type IcmpScan struct {
	Log  *scanLog.LogConf
	body *IcmpReq
}

func (t *IcmpScan) Start(ctx context.Context) error {
	var err error
	body := t.body
	ips, count, err := utils.ReadIps(body.Ip)
	if err != nil {
		t.Log.Error.Println(err)
		return err
	}
	countS := fmt.Sprintf("准备扫描的ip数量: %d", count)
	t.Log.Info.Println(countS)
	utils.SendToThePrivateClientCustom(countS)
	err = t.scanIps(ips)
	if err != nil {
		t.Log.Error.Println(err)
		return err
	}
	//pingMetadata, err := ping.NewPingMetadata(host)
	return nil
}

func (t *IcmpScan) End(ctx context.Context) error {
	return nil
}

func (t *IcmpScan) scanIps(ips []string) error {
	var ipPageGroupLen int
	var start int

	utils.ComputedGroupCount(&ipPageGroupLen, len(ips), pageSize)
	g := sync.WaitGroup{}

	for i := 0; i < ipPageGroupLen; i++ {
		ipSlice := make([]string, 0, pageSize)
		if i == ipPageGroupLen-1 {
			ipSlice = append(ipSlice, ips[start:]...)
		} else {
			ipSlice = append(ipSlice, ips[start:start+pageSize]...)
			start += pageSize
		}
		g.Add(1)
		go func(ipSlice []string) {
			for i := 0; i < len(ipSlice); i++ {
				ip := ipSlice[i]
				metadata, err := ping.NewPingMetadata(ip)
				if err != nil {
					t.Log.Error.Println("异常NewPingMetadata日志:", err)
				}
				metadata.Count = 4
				metadata.Size = 24
				metadata.Interval = time.Second
				metadata.Timeout = time.Second * 5
				metadata.TTL = 64
				if err != nil {
					t.Log.Error.Println("异常Resolve日志:", err)
				}
				sf := fmt.Sprintf("开始ping 地址 %s (%s): ", metadata.Addr, metadata.Ipaddr)
				t.Log.Info.Println(sf)
				utils.SendToThePrivateClientCustom(sf)
				metadata.OnFinish = func(statistics *ping.Statistics) {
					meta := &ping.LogMeta{Log: t.Log}
					ping.OnFinish(statistics, meta)
					if statistics.PacketLoss < 100 {
						utils.SendToThePrivateClientCustom(fmt.Sprintf("[成功]: 目标IP: %s, 丢包率: %v", ip, statistics.PacketLoss))
					} else {
						utils.SendToThePrivateClientCustom("[失败]: IP:" + ip)
					}
				}
				err = metadata.Run()
				if err != nil {
					t.Log.Error.Println("异常Run日志:", err)
				}
			}
			g.Done()
		}(ipSlice)
	}
	g.Wait()
	return nil
}

func (t *IcmpScan) CheckSum(data []byte) uint16 {
	var (
		sum    uint32
		length = len(data)
		index  int
	)
	for length > 1 {
		sum += uint32(data[index])<<8 + uint32(data[index+1])
		index += 2
		length -= 2
	}
	if length > 0 {
		sum += uint32(data[index])
	}
	sum += sum >> 16
	return uint16(^sum)
}
