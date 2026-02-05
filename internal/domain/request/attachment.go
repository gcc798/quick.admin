package request

import (
	"mime/multipart"

	"github.com/force-c/nai-tizi/internal/utils"
)

// UploadFileRequest 上传文件请求（步骤1：只上传文件）
// 参数数量：2 个（符合规范：≤ 3）
type UploadFileRequest struct {
	File    *multipart.FileHeader `form:"file" binding:"required" msg:"请选择要上传的文件"`
	EnvCode string                `form:"envCode"` // 存储环境编码（可选，不传则使用默认环境）
}

// BindAttachmentToBusinessRequest 绑定附件到业务请求（步骤2：绑定业务信息）
// 参数数量：6 个（符合规范：> 3 使用 JSON）
type BindAttachmentToBusinessRequest struct {
	BusinessType  string                 `json:"businessType" binding:"required" msg:"业务类型不能为空"` // 业务类型
	BusinessId    string                 `json:"businessId" binding:"required" msg:"业务ID不能为空"`   // 业务ID
	BusinessField string                 `json:"businessField"`                                  // 业务字段
	IsPublic      bool                   `json:"isPublic"`                                       // 是否公开
	Metadata      map[string]interface{} `json:"metadata"`                                       // 元数据（JSON对象）
	ExpireTime    *utils.LocalTime       `json:"expireTime"`                                     // 过期时间
}

// GetAttachmentURLRequest 获取附件 URL 请求（Query 参数）
type GetAttachmentURLRequest struct {
	Expires int `form:"expires" binding:"omitempty,min=0" msg:"过期时间不能为负数"` // 过期时间（秒），0 表示永久，默认 3600
}

// ListAttachmentsByBusinessRequest 根据业务查询附件列表请求（Query 参数）
type ListAttachmentsByBusinessRequest struct {
	BusinessType string `form:"businessType" binding:"required" msg:"业务类型不能为空"`
	BusinessId   string `form:"businessId" binding:"required" msg:"业务ID不能为空"`
}

// PageAttachmentsRequest 分页查询附件列表请求
type PageAttachmentsRequest struct {
	PageNum      int    `json:"pageNum" binding:"required,min=1" msg:"页码必须大于等于1"`
	PageSize     int    `json:"pageSize" binding:"required,min=1,max=100" msg:"每页数量必须是1-100之间"`
	FileName     string `json:"fileName"`
	FileType     string `json:"fileType"`
	BusinessType string `json:"businessType"`
}
