package response

import (
	"encoding/json"

	"github.com/force-c/nai-tizi/internal/utils"
)

// StorageEnvResponse 存储环境响应
type StorageEnvResponse struct {
	ID          int64            `json:"id"`
	EnvName     string           `json:"name"`
	EnvCode     string           `json:"code"`
	StorageType string           `json:"storageType"` // 存储类型：local/minio/s3/oss
	IsDefault   bool             `json:"isDefault"`
	Status      int32            `json:"status"` // 状态：0正常 1停用
	Config      *json.RawMessage `json:"config"`
	Remark      string           `json:"remark"`
	CreateBy    int64            `json:"createBy"`
	CreatedAt   utils.LocalTime  `json:"createdAt"`
	UpdateBy    int64            `json:"updateBy"`
	UpdatedAt   utils.LocalTime  `json:"updatedAt"`
}
