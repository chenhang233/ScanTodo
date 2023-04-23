package utils

import (
	"time"
)

/*
TimeToBytes 按位取

	start 56 48 40 32 24 16 8 0 >> 每次取8位
*/
func TimeToBytes(t time.Time) []byte {
	nano := t.UnixNano()
	b := make([]byte, 8)
	for i := uint8(0); i < 8; i++ {
		start := (7 - i) * 8
		b[i] = byte((nano >> start) & 0xff)
	}
	BytesToTime(b)
	return b
}

func BytesToTime(b []byte) time.Time {
	var nano int64
	for i := uint8(0); i < 8; i++ {
		start := (7 - i) * 8
		nano += int64(b[i]) << start
	}
	return time.Unix(nano/1000000000, nano%1000000000)
}
