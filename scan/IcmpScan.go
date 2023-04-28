package scan

import (
	"ScanTodo/scan/ping"
	"ScanTodo/scanLog"
	"ScanTodo/utils"
	"context"
	"fmt"
	"golang.org/x/sync/errgroup"
	"time"
)

const MsgLen = 4000
const pageSize = 50

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
	var g errgroup.Group
	var start int

	utils.ComputedGroupCount(&ipPageGroupLen, len(ips), pageSize)
	for i := 0; i < ipPageGroupLen; i++ {
		ipSlice := make([]string, 0, pageSize)
		if i == ipPageGroupLen-1 {
			ipSlice = append(ips, ips[start:]...)
		} else {
			ipSlice = append(ips, ipSlice[start:start+pageSize]...)
			start += pageSize
		}
		g.Go(func() error {
			j := 0
			//l := len(ipSlice)
			metadata, err := ping.NewPingMetadata(ipSlice[j])
			if err != nil {
				t.Log.Error.Println(err)
				return err
			}
			metadata.Count = 4
			metadata.Size = 24
			metadata.Interval = time.Second
			metadata.Timeout = time.Second * 4
			metadata.TTL = 64
			metadata.Log.Info.Println(fmt.Sprintf("开始ping 地址 %s (%s): ", metadata.Addr, metadata.Ipaddr))
			err = metadata.Run()
			if err != nil {
				metadata.Log.Debug.Println("异常结束日志:", err)
			}
			return err
		})
	}

	return g.Wait()
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
