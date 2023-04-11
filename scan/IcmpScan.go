package scan

import (
	"ScanTodo/scanLog"
	"bytes"
	"context"
	"encoding/binary"
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
	Ip      string `json:"ip"`
	Port    string `json:"port"`
	Timeout int    `json:"timeout"`
}

type IcmpScan struct {
	Log  *scanLog.LogConf
	body *IcmpReq
}

func (t *IcmpScan) Start(ctx context.Context) error {
	var (
		icmp  ICMP
		lAddr = net.IPAddr{IP: net.ParseIP("192.168.1.105")}
		rAddr = net.IPAddr{IP: net.ParseIP("192.168.1.1")}
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
	fmt.Println("字节数组", buffer.Bytes())
	if _, err := conn.Write(buffer.Bytes()); err != nil {
		fmt.Println(err.Error())
		return err
	}
	t.Log.Info.Println("send icmp packet success")
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
