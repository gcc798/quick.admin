package response

import "github.com/force-c/nai-tizi/internal/utils"

// AttachmentResponse 附件响应
type AttachmentResponse struct {
	AttachmentId  int64           `json:"attachmentId"`
	EnvId         int64           `json:"envId"`
	FileName      string          `json:"fileName"`
	FileKey       string          `json:"fileKey"`
	FileSize      int64           `json:"fileSize"`
	FileType      string          `json:"fileType"`
	FileExt       string          `json:"fileExt"`
	BusinessType  string          `json:"businessType"`
	BusinessId    string          `json:"businessId"`
	BusinessField string          `json:"businessField"`
	IsPublic      bool            `json:"isPublic"`
	AccessUrl     string          `json:"accessUrl"`
	Metadata      string          `json:"metadata"`
	Status        int32           `json:"status"` // 状态：0正常 1已删除
	ExpireTime    utils.LocalTime `json:"expireTime"`
	CreateBy      int64           `json:"createBy"`
	CreateTime    utils.LocalTime `json:"createTime"`
	UpdateTime    utils.LocalTime `json:"updateTime"`
}

// AttachmentURLResponse 附件 URL 响应
type AttachmentURLResponse struct {
	AttachmentId int64  `json:"attachmentId"`
	URL          string `json:"url"`
	Expires      int64  `json:"expires"` // 过期时间（秒）
}
