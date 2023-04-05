package scan

import (
	"ScanTodo/utils"
	"context"
	"encoding/json"
	"fmt"
)

type ScanRepo interface {
	Start(context.Context) error
	End(context.Context) error
}

type MYLog interface {
	/*
		LogError
		Context 当前环境
		msg 要打印的消息
	*/
	LogError(context.Context, string, error)
}

type ScanCase struct {
	Repo ScanRepo
	Log  MYLog
}

func NewScanCase() (*ScanCase, error) {
	var lo MYLog
	lo = &utils.MyLog{}
	var tc ScanRepo
	tc = &TcpScan{
		Log: &utils.MyLog{},
	}
	scan := &ScanCase{Log: lo, Repo: tc}
	return scan, nil
}

type TcpReq struct {
	Ip      string `json:"ip"`
	Port    string `json:"port"`
	Timeout int    `json:"timeout"`
}

type TcpScan struct {
	Log MYLog
}

func (t *TcpScan) Start(ctx context.Context) error {
	bys := ctx.Value("body").([]byte)
	body := &TcpReq{}
	err := json.Unmarshal(bys, body)
	if err != nil {
		t.Log.LogError(ctx, "json 解析错误", err)
	}
	fmt.Println(body.Ip, body.Port)
	//if body.Ip == "" || body.Port == "" {
	//	return fmt.Errorf("参数错误")
	//}
	return nil
}

func (t *TcpScan) End(ctx context.Context) error {
	return nil
}
