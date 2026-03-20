package model

import (
	"crypto/md5"
	"encoding/hex"
	"strings"

	"github.com/force-c/nai-tizi/internal/utils"

	"gorm.io/gorm"
)

// AuthClient 客户端配置表
type AuthClient struct {
	ClientId      string          `gorm:"column:client_id;type:varchar(64);primaryKey;comment:客户端ID" json:"clientId"`
	ClientKey     string          `gorm:"column:client_key;type:varchar(32);uniqueIndex;not null;comment:客户端Key" json:"clientKey"`
	ClientSecret  string          `gorm:"column:client_secret;type:varchar(255);not null;comment:客户端秘钥" json:"clientSecret"`
	GrantType     string          `gorm:"column:grant_type;type:varchar(255);comment:授权类型(逗号分隔)" json:"grantType"`
	DeviceType    string          `gorm:"column:device_type;type:varchar(32);comment:设备类型" json:"deviceType"`
	Status        int             `gorm:"column:status;default:0;comment:状态(0正常 1停用)" json:"status"`
	Timeout       int64           `gorm:"column:timeout;default:604800;comment:固定超时时间(秒),默认7天" json:"timeout"`
	ActiveTimeout int64           `gorm:"column:active_timeout;default:1800;comment:活动超时时间(秒),默认30分钟" json:"activeTimeout"`
	Remark        string          `gorm:"column:remark;type:varchar(500);comment:备注" json:"remark"`
	CreateBy      int64           `gorm:"column:create_by;comment:创建者" json:"createBy"`
	CreatedTime   utils.LocalTime `gorm:"column:created_time;autoCreateTime;comment:创建时间" json:"createdTime"`
	UpdateBy      int64           `gorm:"column:update_by;comment:更新者" json:"updateBy"`
	UpdatedTime   utils.LocalTime `gorm:"column:updated_time;autoUpdateTime;comment:更新时间" json:"updatedTime"`
	DeletedAt     gorm.DeletedAt  `gorm:"column:deleted_at;index" json:"-"` // 删除时间
}

// TableName 指定表名
func (*AuthClient) TableName() string {
	return "s_auth_client"
}

// GenerateClientId 生成客户端ID (MD5(clientKey + clientSecret))
func GenerateClientId(clientKey, clientSecret string) string {
	data := clientKey + clientSecret
	hash := md5.Sum([]byte(data))
	return hex.EncodeToString(hash[:])
}

// BeforeCreate 创建前钩子 - 自动生成clientId
func (c *AuthClient) BeforeCreate(tx *gorm.DB) error {
	if c.ClientId == "" {
		c.ClientId = GenerateClientId(c.ClientKey, c.ClientSecret)
	}
	// 设置默认值
	if c.Timeout == 0 {
		c.Timeout = 604800 // 7天
	}
	if c.ActiveTimeout == 0 {
		c.ActiveTimeout = 1800 // 30分钟
	}
	return nil
}

// IsGrantTypeSupported 检查是否支持指定的授权类型
func (c *AuthClient) IsGrantTypeSupported(grantType string) bool {
	if c.GrantType == "" {
		return false
	}
	types := strings.Split(c.GrantType, ",")
	for _, t := range types {
		if strings.TrimSpace(t) == grantType {
			return true
		}
	}
	return false
}

// IsActive 检查客户端是否启用
func (c *AuthClient) IsActive() bool {
	return c.Status == 0
}

// FindByClientId 根据clientId查询
func (*AuthClient) FindByClientId(db *gorm.DB, clientId string) (*AuthClient, error) {
	var client AuthClient
	err := db.Where("client_id = ?", clientId).First(&client).Error
	if err != nil {
		return nil, err
	}
	return &client, nil
}

// FindByClientKey 根据clientKey查询
func (*AuthClient) FindByClientKey(db *gorm.DB, clientKey string) (*AuthClient, error) {
	var client AuthClient
	err := db.Where("client_key = ?", clientKey).First(&client).Error
	if err != nil {
		return nil, err
	}
	return &client, nil
}

// CheckClientKeyExists 检查clientKey是否存在
func (*AuthClient) CheckClientKeyExists(db *gorm.DB, clientKey string) (bool, error) {
	var count int64
	err := db.Model(&AuthClient{}).Where("client_key = ?", clientKey).Count(&count).Error
	return count > 0, err
}

// CheckClientKeyExistsExcludingSelf 检查clientKey是否被其他客户端占用
func (*AuthClient) CheckClientKeyExistsExcludingSelf(db *gorm.DB, clientId, clientKey string) (bool, error) {
	var count int64
	err := db.Model(&AuthClient{}).Where("client_key = ? AND client_id != ?", clientKey, clientId).Count(&count).Error
	return count > 0, err
}

// Create 创建客户端
func (c *AuthClient) Create(db *gorm.DB) error {
	return db.Create(c).Error
}

// Update 更新客户端
func (c *AuthClient) Update(db *gorm.DB) error {
	return db.Model(&AuthClient{}).Where("client_id = ?", c.ClientId).Updates(c).Error
}

// Delete 删除客户端
func (*AuthClient) Delete(db *gorm.DB, clientId string) error {
	return db.Where("client_id = ?", clientId).Delete(&AuthClient{}).Error
}

// List 分页查询客户端列表
func (*AuthClient) List(db *gorm.DB, pageNum, pageSize int, status *int, clientKey string) ([]AuthClient, int64, error) {
	var clients []AuthClient
	var total int64

	query := db.Model(&AuthClient{})

	// 条件查询
	if status != nil {
		query = query.Where("status = ?", *status)
	}
	if clientKey != "" {
		query = query.Where("client_key LIKE ?", "%"+clientKey+"%")
	}

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (pageNum - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Order("created_time DESC").Find(&clients).Error

	return clients, total, err
}

// VerifySecret 验证客户端密钥
func (c *AuthClient) VerifySecret(secret string) bool {
	return c.ClientSecret == secret
}
