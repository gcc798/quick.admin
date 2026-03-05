package model

import (
	"encoding/json"

	"github.com/force-c/nai-tizi/internal/utils"

	"gorm.io/gorm"
)

// StorageEnv 存储环境配置
type StorageEnv struct {
	ID          int64            `gorm:"column:id;primaryKey" autogen:"int64" json:"id"`        // 使用分布式ID
	EnvName     string           `gorm:"column:name;not null" json:"name"`                      // 环境名称
	EnvCode     string           `gorm:"column:code;uniqueIndex;not null" json:"code"`          // 环境编码（唯一）
	StorageType string           `gorm:"column:storage_type" json:"storageType"`                // 存储类型：local/minio/s3/oss
	IsDefault   bool             `gorm:"column:is_default;default:false" json:"isDefault"`      // 是否默认环境
	Status      int32            `gorm:"column:status;default:0" json:"status"`                 // 状态：0正常 1停用
	Config      *json.RawMessage `gorm:"column:config;type:jsonb;not null" json:"config"`       // 存储配置（JSON格式）
	Remark      string           `gorm:"column:remark" json:"remark"`                           // 备注
	CreateBy    int64            `gorm:"column:create_by" json:"createBy"`                      // 创建人
	CreatedTime utils.LocalTime  `gorm:"column:created_time;autoCreateTime" json:"createdTime"` // 创建时间
	UpdateBy    int64            `gorm:"column:update_by" json:"updateBy"`                      // 更新人
	UpdatedTime utils.LocalTime  `gorm:"column:updated_time;autoUpdateTime" json:"updatedTime"` // 更新时间
	DeletedAt   gorm.DeletedAt   `gorm:"column:deleted_at;index" json:"-"`                      // 删除时间
}

func (*StorageEnv) TableName() string {
	return "s_storage_env"
}

// FindByID 根据ID查询存储环境
func (*StorageEnv) FindByID(db *gorm.DB, envId int64) (*StorageEnv, error) {
	var env StorageEnv
	err := db.Where("id = ?", envId).First(&env).Error
	if err != nil {
		return nil, err
	}
	return &env, nil
}

// FindByCode 根据编码查询存储环境
func (*StorageEnv) FindByCode(db *gorm.DB, envCode string) (*StorageEnv, error) {
	var env StorageEnv
	err := db.Where("code = ? AND status = 0", envCode).First(&env).Error
	if err != nil {
		return nil, err
	}
	return &env, nil
}

// FindDefault 查询默认存储环境
func (*StorageEnv) FindDefault(db *gorm.DB) (*StorageEnv, error) {
	var env StorageEnv
	err := db.Where("is_default = ? AND status = 0", true).First(&env).Error
	if err != nil {
		return nil, err
	}
	return &env, nil
}

// CheckEnvCodeExists 检查环境编码是否存在
func (*StorageEnv) CheckEnvCodeExists(db *gorm.DB, envCode string) (bool, error) {
	var count int64
	err := db.Model(&StorageEnv{}).Where("code = ?", envCode).Count(&count).Error
	return count > 0, err
}

// CheckEnvCodeExistsExcludingSelf 检查环境编码是否被其他环境占用
func (*StorageEnv) CheckEnvCodeExistsExcludingSelf(db *gorm.DB, envId int64, envCode string) (bool, error) {
	var count int64
	err := db.Model(&StorageEnv{}).Where("code = ? AND id != ?", envCode, envId).Count(&count).Error
	return count > 0, err
}

// CountAttachments 统计使用该环境的附件数量
func (*StorageEnv) CountAttachments(db *gorm.DB, envId int64) (int64, error) {
	var count int64
	err := db.Model(&Attachment{}).Where("env_id = ? AND status = 0", envId).Count(&count).Error
	return count, err
}

// ClearAllDefaults 清除所有默认标记
func (*StorageEnv) ClearAllDefaults(db *gorm.DB) error {
	return db.Model(&StorageEnv{}).Where("is_default = ?", true).Update("is_default", false).Error
}

// Create 创建存储环境
func (s *StorageEnv) Create(db *gorm.DB) error {
	return db.Create(s).Error
}

// Update 更新存储环境
func (s *StorageEnv) Update(db *gorm.DB, envId int64, updates map[string]any) error {
	return db.Model(&StorageEnv{}).Where("id = ?", envId).Updates(updates).Error
}

// SetAsDefault 设置为默认环境
func (s *StorageEnv) SetAsDefault(db *gorm.DB, envId int64) error {
	return db.Model(&StorageEnv{}).Where("id = ?", envId).Update("is_default", true).Error
}

// Delete 删除存储环境（软删除）
func (*StorageEnv) Delete(db *gorm.DB, envId int64) error {
	return db.Where("id = ?", envId).Delete(&StorageEnv{}).Error
}

// List 分页查询存储环境列表
func (*StorageEnv) List(db *gorm.DB, offset, limit int, envName string, storageType string) ([]StorageEnv, int64, error) {
	var envs []StorageEnv
	var total int64

	query := db.Model(&StorageEnv{})

	// 条件过滤
	if envName != "" {
		query = query.Where("name LIKE ?", "%"+envName+"%")
	}
	if storageType != "" {
		query = query.Where("storage_type = ?", storageType)
	}

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	if err := query.Offset(offset).Limit(limit).
		Order("is_default DESC, created_time DESC").
		Find(&envs).Error; err != nil {
		return nil, 0, err
	}

	return envs, total, nil
}

// IsActive 判断环境是否激活
func (s *StorageEnv) IsActive() bool {
	return s.Status == 0
}

// IsDefaultEnv 判断是否为默认环境
func (s *StorageEnv) IsDefaultEnv() bool {
	return s.IsDefault
}

// HasAttachments 判断是否有附件使用该环境
func (s *StorageEnv) HasAttachments(db *gorm.DB) (bool, error) {
	count, err := (&StorageEnv{}).CountAttachments(db, s.ID)
	return count > 0, err
}
