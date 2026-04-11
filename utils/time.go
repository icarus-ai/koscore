package utils

import (
	"fmt"
	"time"
)

func TimeStamp() int64 { return time.Now().Unix() }

func UinTimestamp(uin uint32) string {
	now := time.Now()
	format := now.Format("0102150405")
	ms := now.Nanosecond() / 1000000
	return fmt.Sprintf("%d_%s%02d_%d", uin, format, now.Year()%100, ms)
}
