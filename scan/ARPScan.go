package scan

import (
	"ScanTodo/scanLog"
	"context"
)

type ARPProxyReq struct {
	GatewayMac string `json:"gatewayMac"`
	GatewayIp  string `json:"gatewayIp"`
	SourceMac  string `json:"sourceMac"`
	SourceIp   string `json:"sourceIp"`
	TargetMac  string `json:"targetMac"`
	TargetIp   string `json:"targetIp"`
}

type ARPProxyScan struct {
	Log  *scanLog.LogConf
	body *ARPProxyReq
}

func (t *ARPProxyScan) Start(ctx context.Context) error {
	//var err error
	//body := t.body

	return nil
}

func (t *ARPProxyScan) End(ctx context.Context) error {
	return nil
}

//type ARPScan struct {
//	Log *scanLog.LogConf
//}
