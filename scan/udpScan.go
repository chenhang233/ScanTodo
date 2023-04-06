package scan

import (
	"ScanTodo/scanLog"
	"context"
)

type UdpScan struct {
	Log *scanLog.LogConf
}

func (t *UdpScan) Start(ctx context.Context) error {
	return nil
}

func (t *UdpScan) End(ctx context.Context) error {
	return nil
}
