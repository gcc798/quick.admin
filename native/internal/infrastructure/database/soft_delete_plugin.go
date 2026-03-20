package database

import (
	"gorm.io/gorm"
)

// SoftDeletePlugin GORM 软删除插件
// 用于全局控制是否启用软删除功能
type SoftDeletePlugin struct {
	// Enabled 为 true 时使用软删除，为 false 时使用物理删除
	Enabled bool
}

// Name 返回插件名称
func (p *SoftDeletePlugin) Name() string {
	return "SoftDeletePlugin"
}

// Initialize 初始化插件，注册 GORM 回调
func (p *SoftDeletePlugin) Initialize(db *gorm.DB) error {
	// 注册删除前的回调
	callback := db.Callback().Delete()

	if callback.Get("soft_delete:before") == nil {
		callback.Before("gorm:delete").Register("soft_delete:before", p.beforeDelete)
	}

	return nil
}

// beforeDelete 删除前的回调函数
func (p *SoftDeletePlugin) beforeDelete(db *gorm.DB) {
	// 如果禁用软删除，强制使用物理删除
	if !p.Enabled {
		// 设置 Unscoped 标志，跳过软删除
		db.Statement.Unscoped = true
	}

	// 如果已经显式调用了 Unscoped()，保持原有行为
	// GORM 会自动处理
}
