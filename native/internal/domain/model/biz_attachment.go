package model

import (
	"encoding/json"

	"github.com/force-c/nai-tizi/internal/utils"
	"gorm.io/gorm"
)

// Attachment 附件
type Attachment struct {
	ID            int64            `gorm:"column:id;primaryKey" autogen:"int64" json:"id"`      // 使用分布式ID
	EnvId         int64            `gorm:"column:env_id;not null" json:"envId"`                 // 存储环境ID
	FileName      string           `gorm:"column:file_name;not null" json:"fileName"`           // 原始文件名
	FileKey       string           `gorm:"column:file_key;not null" json:"fileKey"`             // 存储路径/Key
	FileSize      int64            `gorm:"column:file_size;not null" json:"fileSize"`           // 文件大小（字节）
	FileType      string           `gorm:"column:file_type" json:"fileType"`                    // 文件类型（MIME Type）
	FileExt       string           `gorm:"column:file_ext" json:"fileExt"`                      // 文件扩展名
	BusinessType  string           `gorm:"column:business_type" json:"businessType"`            // 业务类型
	BusinessId    string           `gorm:"column:business_id" json:"businessId"`                // 业务ID
	BusinessField string           `gorm:"column:business_field" json:"businessField"`          // 业务字段
	IsPublic      bool             `gorm:"column:is_public;default:false" json:"isPublic"`      // 是否公开访问
	AccessUrl     string           `gorm:"column:access_url" json:"accessUrl"`                  // 访问URL
	Metadata      *json.RawMessage `gorm:"column:metadata;type:jsonb" json:"metadata"`          // JSON元数据
	Status        int32            `gorm:"column:status;default:0" json:"status"`               // 状态：0正常 1已删除
	ExpireTime    utils.LocalTime  `gorm:"column:expire_time" json:"expireTime"`                // 过期时间
	CreateBy      int64            `gorm:"column:create_by" json:"createBy"`                    // 创建人
	CreateTime    utils.LocalTime  `gorm:"column:create_time;autoCreateTime" json:"createTime"` // 创建时间
	UpdateTime    utils.LocalTime  `gorm:"column:update_time;autoUpdateTime" json:"updateTime"` // 更新时间
	DeletedAt     gorm.DeletedAt   `gorm:"column:deleted_at;index" json:"-"`                    // 删除时间
}

func (*Attachment) TableName() string {
	return "biz_attachment"
}
