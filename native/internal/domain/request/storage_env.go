package request

import (
	"encoding/json"

	"github.com/gcc798/quick.admin/internal/utils/pagination"
)

// CreateStorageEnvRequest 创建存储环境请求
type CreateStorageEnvRequest struct {
	EnvName     string           `json:"name" binding:"required" msg:"环境名称不能为空"`
	EnvCode     string           `json:"code" binding:"required" msg:"环境编码不能为空"`
	StorageType string           `json:"storageType" binding:"required,oneof=local minio s3 oss" msg:"存储类型必须是 local/minio/s3/oss 之一"` // 存储类型：local/minio/s3/oss
	IsDefault   bool             `json:"isDefault"`
	Status      int32            `json:"status" binding:"oneof=0 1" msg:"状态必须是0（正常）或1（停用）"` // 状态：0正常 1停用
	Config      *json.RawMessage `json:"config"`                                            // JSON 对象
	Remark      string           `json:"remark"`
}

// UpdateStorageEnvRequest 更新存储环境请求
type UpdateStorageEnvRequest struct {
	ID          int64            `json:"id"` // 环境ID（由路径参数注入）
	EnvName     string           `json:"name" binding:"required" msg:"环境名称不能为空"`
	EnvCode     string           `json:"code" binding:"required" msg:"环境编码不能为空"`
	StorageType string           `json:"storageType" binding:"required,oneof=local minio s3 oss" msg:"存储类型必须是 local/minio/s3/oss 之一"` // 存储类型：local/minio/s3/oss
	IsDefault   bool             `json:"isDefault"`
	Status      int32            `json:"status" binding:"oneof=0 1" msg:"状态必须是0（正常）或1（停用）"` // 状态：0正常 1停用
	Config      *json.RawMessage `json:"config" binding:"required" msg:"配置信息不能为空"`
	Remark      string           `json:"remark"`
}

// SetDefaultStorageEnvRequest 设置默认存储环境请求
type SetDefaultStorageEnvRequest struct {
	ID int64 `json:"id" binding:"required" msg:"环境ID不能为空" typecast:"stringInt64"`
}

// PageStorageEnvsRequest 查询存储环境列表请求（Query 参数）
type PageStorageEnvsRequest struct {
	pagination.PageQuery        // 嵌入分页参数
	Name                 string `json:"name"`
	StorageType          string `json:"storageType" binding:"omitempty,oneof=local minio s3 oss" msg:"存储类型必须是 local/minio/s3/oss 之一"` // 存储类型：local/minio/s3/oss
}

// TestStorageEnvConnectionRequest 测试存储环境连接请求
type TestStorageEnvConnectionRequest struct {
	// 可以添加额外的测试参数，目前为空
}
