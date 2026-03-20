package model

import (
	"github.com/force-c/nai-tizi/internal/utils"
	"gorm.io/gorm"
)

// LoginLog 登录日志
type LoginLog struct {
	ID            int64           `gorm:"column:id;primaryKey" autogen:"int64" json:"id"` // 日志ID（使用分布式ID）
	UserName      string          `gorm:"column:user_name;index" json:"userName"`         // 用户名
	Ipaddr        string          `gorm:"column:ipaddr" json:"ipaddr"`                    // 登录IP
	LoginLocation string          `gorm:"column:login_location" json:"loginLocation"`     // 登录地点
	Browser       string          `gorm:"column:browser" json:"browser"`                  // 浏览器类型
	Os            string          `gorm:"column:os" json:"os"`                            // 操作系统
	Status        int32           `gorm:"column:status;default:0" json:"status"`          // 登录状态：0成功 1失败
	Msg           string          `gorm:"column:msg" json:"msg"`                          // 提示消息
	LoginTime     utils.LocalTime `gorm:"column:login_time;index" json:"loginTime"`       // 登录时间
	ClientId      string          `gorm:"column:client_id" json:"clientId"`               // 客户端ID
}

func (*LoginLog) TableName() string {
	return "s_login_log"
}

// FindByID 根据ID查询登录日志
func (*LoginLog) FindByID(db *gorm.DB, id int64) (*LoginLog, error) {
	var log LoginLog
	err := db.Where("id = ?", id).First(&log).Error
	if err != nil {
		return nil, err
	}
	return &log, nil
}

// Create 创建登录日志
func (l *LoginLog) Create(db *gorm.DB) error {
	return db.Create(l).Error
}

// Update 更新登录日志
func (l *LoginLog) Update(db *gorm.DB, id int64, updates map[string]interface{}) error {
	return db.Model(&LoginLog{}).Where("id = ?", id).Updates(updates).Error
}

// Delete 删除登录日志（物理删除）
func (*LoginLog) Delete(db *gorm.DB, id int64) error {
	return db.Unscoped().Where("id = ?", id).Delete(&LoginLog{}).Error
}

// BatchDelete 批量删除登录日志
func (*LoginLog) BatchDelete(db *gorm.DB, ids []int64) (int64, error) {
	result := db.Unscoped().Where("id IN ?", ids).Delete(&LoginLog{})
	return result.RowsAffected, result.Error
}

// IsSuccess 判断登录是否成功
func (l *LoginLog) IsSuccess() bool {
	return l.Status == 0
}

// CleanOldLogs 清理指定天数之前的日志
func (*LoginLog) CleanOldLogs(db *gorm.DB, days int) (int64, error) {
	result := db.Unscoped().Where("login_time < NOW() - INTERVAL '? days'", days).Delete(&LoginLog{})
	return result.RowsAffected, result.Error
}
