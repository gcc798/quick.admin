package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/force-c/nai-tizi/internal/constants"
	"github.com/force-c/nai-tizi/internal/domain/model"
	"github.com/force-c/nai-tizi/internal/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// storageTypeToString 将存储类型 int32 转换为字符串
func storageTypeToString(storageType int32) string {
	switch storageType {
	case constants.StorageTypeLocal:
		return "local"
	case constants.StorageTypeMinio:
		return "minio"
	case constants.StorageTypeS3:
		return "s3"
	case constants.StorageTypeOSS:
		return "oss"
	default:
		return fmt.Sprintf("unknown(%d)", storageType)
	}
}

// StorageManager 存储管理器接口
type StorageManager interface {
	// GetStorage 根据环境 ID 获取 Storage 实例
	GetStorage(envId int64) (Storage, error)

	// GetStorageByCode 根据环境编码获取 Storage 实例
	GetStorageByCode(envCode string) (Storage, error)

	// GetDefaultStorage 获取默认 Storage
	GetDefaultStorage() (Storage, error)

	// ReloadConfig 重新加载配置（从数据库）
	ReloadConfig() error

	// RegisterStorageType 注册新的存储类型
	RegisterStorageType(typeName string, factory StorageFactory) error
}

// storageManager 存储管理器实现
type storageManager struct {
	db           *gorm.DB
	logger       logger.Logger
	storages     map[int64]Storage         // envId -> Storage 实例
	envCodeMap   map[string]int64          // envCode -> envId
	defaultEnvId int64                     // 默认环境 ID
	factories    map[string]StorageFactory // storageType -> Factory
	mu           sync.RWMutex
}

// NewStorageManager 创建存储管理器
func NewStorageManager(db *gorm.DB, logger logger.Logger) StorageManager {
	return &storageManager{
		db:         db,
		logger:     logger,
		storages:   make(map[int64]Storage),
		envCodeMap: make(map[string]int64),
		factories:  make(map[string]StorageFactory),
	}
}

// RegisterStorageType 注册存储类型
func (m *storageManager) RegisterStorageType(typeName string, factory StorageFactory) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.factories[typeName]; exists {
		return fmt.Errorf("存储类型 %s 已注册", typeName)
	}

	m.factories[typeName] = factory
	m.logger.Info("注册存储类型", zap.String("type", typeName))
	return nil
}

// GetStorage 根据环境 ID 获取 Storage 实例
func (m *storageManager) GetStorage(envId int64) (Storage, error) {
	m.mu.RLock()
	storage, exists := m.storages[envId]
	m.mu.RUnlock()

	if exists {
		return storage, nil
	}

	// 从数据库加载配置并创建 Storage 实例
	return m.loadStorage(envId)
}

// GetStorageByCode 根据环境编码获取 Storage 实例
func (m *storageManager) GetStorageByCode(envCode string) (Storage, error) {
	m.mu.RLock()
	envId, exists := m.envCodeMap[envCode]
	m.mu.RUnlock()

	if exists {
		return m.GetStorage(envId)
	}

	// 从数据库查询环境 ID
	var env model.StorageEnv
	if err := m.db.Where("env_code = ? AND status = ?", envCode, 0).First(&env).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("存储环境不存在: %s", envCode)
		}
		return nil, fmt.Errorf("查询存储环境失败: %w", err)
	}

	// 缓存 envCode -> envId 映射
	m.mu.Lock()
	m.envCodeMap[envCode] = env.ID
	m.mu.Unlock()

	return m.loadStorage(env.ID)
}

// GetDefaultStorage 获取默认 Storage
func (m *storageManager) GetDefaultStorage() (Storage, error) {
	m.mu.RLock()
	defaultEnvId := m.defaultEnvId
	m.mu.RUnlock()

	if defaultEnvId > 0 {
		return m.GetStorage(defaultEnvId)
	}

	// 从数据库查询默认环境
	var env model.StorageEnv
	if err := m.db.Where("is_default = ? AND status = ?", true, 0).First(&env).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("未配置默认存储环境")
		}
		return nil, fmt.Errorf("查询默认存储环境失败: %w", err)
	}

	// 缓存默认环境 ID
	m.mu.Lock()
	m.defaultEnvId = env.ID
	m.mu.Unlock()

	return m.loadStorage(env.ID)
}

// loadStorage 从数据库加载配置并创建 Storage 实例
func (m *storageManager) loadStorage(envId int64) (Storage, error) {
	// 从数据库查询环境配置
	var env model.StorageEnv
	if err := m.db.Where("id = ? AND status = ?", envId, 0).First(&env).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("存储环境不存在: %d", envId)
		}
		return nil, fmt.Errorf("查询存储环境失败: %w", err)
	}

	// 解析配置
	var config map[string]interface{}
	if err := json.Unmarshal(*env.Config, &config); err != nil {
		return nil, fmt.Errorf("解析存储配置失败: %w", err)
	}

	// 获取工厂
	m.mu.RLock()
	factory, exists := m.factories[env.StorageType]
	m.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("不支持的存储类型: %s", env.StorageType)
	}

	// 创建 Storage 实例
	storage, err := factory.Create(config)
	if err != nil {
		return nil, fmt.Errorf("创建存储实例失败: %w", err)
	}

	// 缓存 Storage 实例
	m.mu.Lock()
	m.storages[envId] = storage
	m.envCodeMap[env.EnvCode] = envId
	if env.IsDefault {
		m.defaultEnvId = envId
	}
	m.mu.Unlock()

	m.logger.Info("加载存储环境",
		zap.Int64("envId", envId),
		zap.String("envCode", env.EnvCode),
		zap.String("storageType", env.StorageType))

	return storage, nil
}

// ReloadConfig 重新加载配置
func (m *storageManager) ReloadConfig() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 清空缓存
	m.storages = make(map[int64]Storage)
	m.envCodeMap = make(map[string]int64)
	m.defaultEnvId = 0

	m.logger.Info("重新加载存储配置")
	return nil
}
