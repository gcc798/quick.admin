package utils

import (
	"sync"
	"time"
)

var (
	beijingTz *time.Location
	once      sync.Once
)

// GetTimeZone 获取北京时区（东八区）
func GetTimeZone() *time.Location {
	once.Do(func() {
		var err error
		beijingTz, err = time.LoadLocation("Asia/Shanghai")
		if err != nil {
			// 如果加载失败，使用 UTC+8
			beijingTz = time.FixedZone("CST", 8*3600)
		}
	})
	return beijingTz
}
