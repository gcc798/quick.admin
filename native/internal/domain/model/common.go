package model

import "time"

// CommonFields 公共审计字段
type CommonFields struct {
	CreateBy   int64     `gorm:"column:create_by;type:bigint" json:"createBy"`                         // 创建人
	CreateTime time.Time `gorm:"column:create_time;type:timestamptz;autoCreateTime" json:"createTime"` // 创建时间
	UpdateBy   int64     `gorm:"column:update_by;type:bigint" json:"updateBy"`                         // 更新人
	UpdateTime time.Time `gorm:"column:update_time;type:timestamptz;autoUpdateTime" json:"updateTime"` // 更新时间
}
