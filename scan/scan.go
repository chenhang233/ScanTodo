package scan

import (
	"ScanTodo/scanLog"
	"context"
	"encoding/json"
	"os"
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

func NewScanCase(scanUseCase string, body []byte) (*ScanCase, error) {
	loadLog, err := scanLog.LoadLog(scanLog.ScansLogPath)
	if err != nil {
		panic(err)
	}
	var tc ScanRepo
	switch scanUseCase {
	case "TCP":
		req := &TcpReq{}
		err = json.Unmarshal(body, req)
		tc = &TcpScan{
			Log:      loadLog,
			body:     req,
			ipPage:   100,
			portPage: 100,
		}
	case "UDP":
		tc = &UdpScan{
			Log: loadLog,
		}
	case "ICMP":
		req := &IcmpReq{}
		err = json.Unmarshal(body, req)
		tc = &IcmpScan{
			Log:  loadLog,
			body: req,
		}
	default:
		loadLog.Error.Println("未实现 " + scanUseCase + " 这个类型")
		os.Exit(-1)
	}
	if err != nil {
		loadLog.Error.Println("NewScanCase error: ", err)
	}

	scan := &ScanCase{Log: loadLog, Repo: tc}
	return scan, nil
}
