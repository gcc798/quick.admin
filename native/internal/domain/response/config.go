package response

import (
	"encoding/json"

	"github.com/force-c/nai-tizi/internal/domain/model"
	"github.com/force-c/nai-tizi/internal/utils"
)

// ConfigResponse 配置响应
type ConfigResponse struct {
	ID          int64           `json:"id"`          // 配置ID
	Name        string          `json:"name"`        // 配置名称
	Code        string          `json:"code"`        // 配置编码
	Data        json.RawMessage `json:"data"`        // 配置数据（JSON格式）
	Remark      string          `json:"remark"`      // 备注
	CreateBy    int64           `json:"createBy"`    // 创建者
	CreatedTime utils.LocalTime `json:"createdTime"` // 创建时间
	UpdateBy    int64           `json:"updateBy"`    // 更新者
	UpdatedTime utils.LocalTime `json:"updatedTime"` // 更新时间
}

// ConfigListResponse 配置列表响应
type ConfigListResponse struct {
	Total int64            `json:"total"` // 总数
	List  []ConfigResponse `json:"list"`  // 列表
}

// ConfigDataResponse 配置数据响应（仅返回data字段）
type ConfigDataResponse struct {
	Code string          `json:"code"` // 配置编码
	Data json.RawMessage `json:"data"` // 配置数据
}

// ToConfigResponse 转换为配置响应
func ToConfigResponse(config *model.Config) ConfigResponse {
	return ConfigResponse{
		ID:          config.ID,
		Name:        config.Name,
		Code:        config.Code,
		Data:        config.Data,
		Remark:      config.Remark,
		CreateBy:    config.CreateBy,
		CreatedTime: config.CreatedTime,
		UpdateBy:    config.UpdateBy,
		UpdatedTime: config.UpdatedTime,
	}
}

// ToConfigListResponse 转换为配置列表响应
func ToConfigListResponse(configs []model.Config, total int64) ConfigListResponse {
	list := make([]ConfigResponse, 0, len(configs))
	for _, config := range configs {
		list = append(list, ToConfigResponse(&config))
	}
	return ConfigListResponse{
		Total: total,
		List:  list,
	}
}
