package idgen

import (
	"fmt"
	"sync"
	"time"

	"github.com/sony/sonyflake"
)

var (
	sf   *sonyflake.Sonyflake
	once sync.Once
)

// Init 初始化ID生成器
// machineID: 机器ID，用于分布式环境区分不同节点，范围 0-65535
func Init(machineID uint16) error {
	var initErr error
	once.Do(func() {
		st := sonyflake.Settings{
			StartTime: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			MachineID: func() (uint16, error) {
				return machineID, nil
			},
		}
		sf = sonyflake.NewSonyflake(st)
		if sf == nil {
			initErr = fmt.Errorf("sonyflake初始化失败")
		}
	})
	return initErr
}

// NextID 生成下一个唯一ID
func NextID() (int64, error) {
	if sf == nil {
		if err := Init(1); err != nil {
			return 0, err
		}
	}
	id, err := sf.NextID()
	if err != nil {
		return 0, fmt.Errorf("生成ID失败: %w", err)
	}
	return int64(id), nil
}

// MustNextID 生成下一个唯一ID，失败时panic
func MustNextID() int64 {
	id, err := NextID()
	if err != nil {
		panic(err)
	}
	return id
}
