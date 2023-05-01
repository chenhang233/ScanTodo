package scan

import (
	"ScanTodo/scanLog"
	"context"
)

type Address struct {
	GatewayMac string `json:"gatewayMac"`
	GatewayIp  string `json:"gatewayIp"`
	SourceMac  string `json:"sourceMac"`
	SourceIp   string `json:"sourceIp"`
	TargetMac  string `json:"targetMac"`
	TargetIp   string `json:"targetIp"`
}

type ARPProxyReq struct {
	Address  Address `json:"address"`
	Timeout  int64   `json:"timeout"`
	Interval int     `json:"interval"`
}

type ARPProxyScan struct {
	Log  *scanLog.LogConf
	body *ARPProxyReq
}

func (t *ARPProxyScan) Start(ctx context.Context) error {
	//body := t.body
	//a := body.Address
	//SD, err := utils.GetPcapDev(a.SourceIp)
	//if err != nil {
	//	t.Log.Error.Println(err)
	//	return err
	//}
	//
	//ghw, err := net.ParseMAC(a.GatewayMac)
	//thw, err := net.ParseMAC(a.TargetMac)
	//g4 := net.ParseIP(a.GatewayIp).To4()
	//t4 := net.ParseIP(a.TargetIp).To4()
	//if err != nil {
	//	t.Log.Error.Println(err)
	//	return err
	//}
	//
	//handle, err := pcap.OpenLive(SD.Name, 65536, true, pcap.BlockForever)
	//if err != nil {
	//	t.Log.Error.Println("监听错误: ", err)
	//	return err
	//}
	//
	//err = utils.SendArp(handle, 2, SD.Mac, thw, g4, t4)
	//if err != nil {
	//	t.Log.Error.Println("发送目标失败: ", err)
	//	return err
	//}
	//
	//err = utils.SendArp(handle, 2, SD.Mac, ghw, t4, g4)
	//if err != nil {
	//	t.Log.Error.Println("发送网关失败: ", err)
	//	return err
	//}
	//return nil
}

func (t *ARPProxyScan) End(ctx context.Context) error {
	return nil
}

//type ARPScan struct {
//	Log *scanLog.LogConf
//}
