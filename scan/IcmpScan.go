package scan

import (
	"ScanTodo/scanLog"
	"ScanTodo/utils"
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
)

const MsgLen = 4000

type ICMP struct {
	Type        uint8
	Code        uint8
	Checksum    uint16
	Identifier  uint16
	SequenceNum uint16
	msgData     [MsgLen]uint8
}

type IcmpReq struct {
	SourceIp string `json:"sourceIp"`
	TargetIp string `json:"targetIp"`
	Timeout  int    `json:"timeout"`
}

type IcmpScan struct {
	Log  *scanLog.LogConf
	body *IcmpReq
}

func (t *IcmpScan) Start(ctx context.Context) error {
	body := t.body
	if !utils.CheckIpv4(body.SourceIp) {
		return errors.New(fmt.Sprintf("源ip: %s 格式错误", body.SourceIp))
	}
	if !utils.CheckIpv4(body.TargetIp) {
		return errors.New(fmt.Sprintf("目标ip: %s 格式错误", body.TargetIp))
	}

	var (
		icmp  ICMP
		lAddr = net.IPAddr{IP: net.ParseIP(body.SourceIp)}
		rAddr = net.IPAddr{IP: net.ParseIP(body.TargetIp)}
	)
	conn, err := net.DialIP("ip4:icmp", &lAddr, &rAddr)
	if err != nil {
		t.Log.Error.Println("打开ip4:icmp连接错误", err)
		return err
	}

	icmp.Type = 8 // 8发送 0接收
	icmp.Code = 0
	icmp.Checksum = 0
	icmp.Identifier = 0
	icmp.SequenceNum = 0
	for i := 0; i < MsgLen; i++ {
		icmp.msgData[i] = uint8(127)
	}

	var buffer bytes.Buffer

	err = binary.Write(&buffer, binary.BigEndian, icmp)
	icmp.Checksum = t.CheckSum(buffer.Bytes())
	buffer.Reset()
	err = binary.Write(&buffer, binary.BigEndian, icmp)
	if _, err := conn.Write(buffer.Bytes()); err != nil {
		fmt.Println(err.Error())
		return err
	}
	t.Log.Info.Println("send icmp packet success")
	bys := make([]byte, 0, 1024)
	read, err := conn.Read(bys)
	if err != nil {
		t.Log.Error.Println("读取错误:", err)
	}
	t.Log.Info.Println("响应", read)
	err = conn.Close()
	return nil
}

func (t *IcmpScan) End(ctx context.Context) error {
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
