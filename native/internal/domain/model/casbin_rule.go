package model

// CasbinRule Casbin策略存储表
type CasbinRule struct {
	ID    int64  `gorm:"column:id;primaryKey" autogen:"int64" json:"id"` // 使用分布式ID
	Ptype string `gorm:"column:ptype;not null" json:"ptype"`             // 策略类型：p(策略) 或 g(角色)
	V0    string `gorm:"column:v0" json:"v0"`                            // 主体（用户或角色）
	V1    string `gorm:"column:v1" json:"v1"`                            // 域（组织ID）或角色
	V2    string `gorm:"column:v2" json:"v2"`                            // 对象（资源）或组织ID
	V3    string `gorm:"column:v3" json:"v3"`                            // 动作（操作类型）
	V4    string `gorm:"column:v4" json:"v4"`                            // 保留字段
	V5    string `gorm:"column:v5" json:"v5"`                            // 保留字段
}

func (*CasbinRule) TableName() string {
	return "casbin_rule"
}
