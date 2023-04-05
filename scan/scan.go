package scan

import (
	"ScanTodo/scanLog"
	"ScanTodo/utils"
	"context"
	"encoding/json"
	"fmt"
)

type ScanRepo interface {
	Start(context.Context) error
	End(context.Context) error
}

//
//type MYLog interface {
//	/*
//		LogError
//		Context 当前环境
//		msg 要打印的消息
//	*/
//	LogError(context.Context, string, error)
//}

type ScanCase struct {
	Repo ScanRepo
	Log  *scanLog.LogConf
}

func NewScanCase() (*ScanCase, error) {
	loadLog, err := scanLog.LoadLog("日志")
	if err != nil {
		panic(err)
	}
	var tc ScanRepo
	tc = &TcpScan{
		Log: loadLog,
	}
	scan := &ScanCase{Log: loadLog, Repo: tc}
	return scan, nil
}

type TcpReq struct {
	Ip      string `json:"ip"`
	Port    string `json:"port"`
	Timeout int    `json:"timeout"`
}

type TcpScan struct {
	Log *scanLog.LogConf
}

func (t *TcpScan) Start(ctx context.Context) error {
	bys := ctx.Value("body").([]byte)
	body := &TcpReq{}
	err := json.Unmarshal(bys, body)
	if err != nil {
		t.Log.Error.Println("json错误", err)
	}
	ipf := utils.CheckIpv4(body.Ip)
	if !ipf {
		err = fmt.Errorf("ip参数错误")
		return err
	}
	ports, err := utils.ReadPorts(body.Port)
	if err != nil {
		err = fmt.Errorf("port参数错误")
	}
	fmt.Println(ports, "ports")
	//if body.Ip == "" || body.Port == "" {
	//	return fmt.Errorf("参数错误")
	//}
	return nil
}

func (t *TcpScan) End(ctx context.Context) error {
	fmt.Println("请求返回之后---")
	return nil
}

type UdpScan struct {
	Log *scanLog.LogConf
}

func (t *UdpScan) Start(ctx context.Context) error {
	return nil
}

func (t *UdpScan) End(ctx context.Context) error {
	return nil
}
