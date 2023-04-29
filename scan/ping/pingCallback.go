package ping

import (
	"ScanTodo/scanLog"
	"fmt"
)

type LogMeta struct {
	Log *scanLog.LogConf
}

func OnFinish(statistics *Statistics, pingMetadata *LogMeta) {
	pingMetadata.Log.Debug.Printf(fmt.Sprintf("OnFinish target: %v, PacketsSent: %v, PacketsReceive: %v, PacketLoss: %v %%,PacketsReceiveDuplicates: %v",
		statistics.Addr, statistics.PacketsSent, statistics.PacketsReceive, statistics.PacketLoss, statistics.PacketsReceiveDuplicates))
	pingMetadata.Log.Debug.Println(fmt.Sprintf("OnFinish RTTs: %v, AverageRoundTripTime: %v, MaxRoundTripTime: %v,MinRoundTripTime: %v ",
		statistics.RTTs, statistics.AverageRoundTripTime, statistics.MaxRoundTripTime, statistics.MinRoundTripTime))
}
