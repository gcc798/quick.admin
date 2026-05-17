package model

// CasbinRule Casbin策略存储表
type CasbinRule struct {
	ID    uint   `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Ptype string `gorm:"column:ptype;type:varchar(8);not null;index" json:"ptype"`
	V0    string `gorm:"column:v0;type:varchar(128);index" json:"v0"`
	V1    string `gorm:"column:v1;type:varchar(128);index" json:"v1"`
	V2    string `gorm:"column:v2;type:varchar(255);index" json:"v2"`
	V3    string `gorm:"column:v3;type:varchar(64)" json:"v3"`
	V4    string `gorm:"column:v4;type:varchar(255)" json:"v4"`
	V5    string `gorm:"column:v5;type:varchar(255)" json:"v5"`
}

// TableName 返回数据库表名。
func (*CasbinRule) TableName() string {
	return "casbin_rule"
}
