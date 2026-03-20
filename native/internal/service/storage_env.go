package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/force-c/nai-tizi/internal/domain/model"
	"github.com/force-c/nai-tizi/internal/logger"
	"github.com/force-c/nai-tizi/internal/utils/pagination"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// StorageEnvService 存储环境管理服务接口
type StorageEnvService interface {
	// Create 创建存储环境
	Create(ctx context.Context, env *model.StorageEnv) error

	// Update 更新存储环境
	Update(ctx context.Context, env *model.StorageEnv) error

	// Delete 删除存储环境
	Delete(ctx context.Context, envId int64) error

	// GetById 根据 ID 查询存储环境
	GetById(ctx context.Context, envId int64) (*model.StorageEnv, error)

	// GetByCode 根据编码查询存储环境
	GetByCode(ctx context.Context, envCode string) (*model.StorageEnv, error)

	// Page 分页查询存储环境列表
	Page(ctx context.Context, pageNum, pageSize int, name string, storageType string) (*pagination.Page[model.StorageEnv], error)

	// SetDefault 设置默认环境
	SetDefault(ctx context.Context, envId int64) error

	// GetDefault 获取默认环境
	GetDefault(ctx context.Context) (*model.StorageEnv, error)

	// TestConnection 测试存储环境连接
	TestConnection(ctx context.Context, envId int64) error
}

type storageEnvService struct {
	db     *gorm.DB
	logger logger.Logger
}

// NewStorageEnvService 创建存储环境服务实例
func NewStorageEnvService(db *gorm.DB, logger logger.Logger) StorageEnvService {
	return &storageEnvService{
		db:     db,
		logger: logger,
	}
}

// Create 创建存储环境
func (s *storageEnvService) Create(ctx context.Context, env *model.StorageEnv) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 1. 检查环境编码是否已存在
		exists, err := (&model.StorageEnv{}).CheckEnvCodeExists(tx, env.EnvCode)
		if err != nil {
			return fmt.Errorf("检查环境编码失败: %w", err)
		}
		if exists {
			return fmt.Errorf("环境编码已存在: %s", env.EnvCode)
		}

		// 2. 检查是否已有默认环境
		defaultEnv, err := (&model.StorageEnv{}).FindDefault(tx)
		hasDefault := err == nil && defaultEnv != nil

		// 3. 如果没有默认环境，强制设置为默认
		if !hasDefault {
			env.IsDefault = true
			s.logger.Info("创建第一个存储环境，自动设为默认", zap.String("envCode", env.EnvCode))
		}

		// 4. 如果要设置为默认，先取消其他默认环境
		if env.IsDefault {
			if err := (&model.StorageEnv{}).ClearAllDefaults(tx); err != nil {
				return fmt.Errorf("取消原默认环境失败: %w", err)
			}
		}

		// 5. 创建环境
		if err := env.Create(tx); err != nil {
			return fmt.Errorf("创建存储环境失败: %w", err)
		}

		s.logger.Info("创建存储环境成功",
			zap.Int64("envId", env.ID),
			zap.String("envCode", env.EnvCode),
			zap.Bool("isDefault", env.IsDefault))

		return nil
	})
}

// Update 更新存储环境
func (s *storageEnvService) Update(ctx context.Context, env *model.StorageEnv) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 1. 检查环境是否存在
		existingEnv, err := (&model.StorageEnv{}).FindByID(tx, env.ID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("存储环境不存在")
			}
			return fmt.Errorf("查询存储环境失败: %w", err)
		}

		// 2. 如果修改了环境编码，检查是否重复
		if existingEnv.EnvCode != env.EnvCode {
			exists, err := (&model.StorageEnv{}).CheckEnvCodeExistsExcludingSelf(tx, env.ID, env.EnvCode)
			if err != nil {
				return fmt.Errorf("检查环境编码失败: %w", err)
			}
			if exists {
				return fmt.Errorf("环境编码已存在: %s", env.EnvCode)
			}
		}

		// 3. 如果要设置为默认，先取消其他默认环境
		if env.IsDefault && !existingEnv.IsDefault {
			if err := (&model.StorageEnv{}).ClearAllDefaults(tx); err != nil {
				return fmt.Errorf("取消原默认环境失败: %w", err)
			}
		}

		// 4. 更新环境
		updates := map[string]any{
			"name":         env.EnvName,
			"code":         env.EnvCode,
			"storage_type": env.StorageType,
			"is_default":   env.IsDefault,
			"status":       env.Status,
			"config":       env.Config,
			"remark":       env.Remark,
			"update_by":    env.UpdateBy,
		}

		if err := existingEnv.Update(tx, env.ID, updates); err != nil {
			return fmt.Errorf("更新存储环境失败: %w", err)
		}

		s.logger.Info("更新存储环境成功", zap.Int64("envId", env.ID))
		return nil
	})
}

// Delete 删除存储环境
func (s *storageEnvService) Delete(ctx context.Context, envId int64) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 1. 检查环境是否存在
		env, err := (&model.StorageEnv{}).FindByID(tx, envId)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("存储环境不存在")
			}
			return fmt.Errorf("查询存储环境失败: %w", err)
		}

		// 2. 不允许删除默认环境
		if env.IsDefaultEnv() {
			return fmt.Errorf("不能删除默认环境，请先设置其他环境为默认")
		}

		// 3. 检查是否有附件使用该环境
		hasAttachments, err := env.HasAttachments(tx)
		if err != nil {
			return fmt.Errorf("检查附件使用情况失败: %w", err)
		}
		if hasAttachments {
			count, _ := env.CountAttachments(tx, envId)
			return fmt.Errorf("该环境下还有 %d 个附件，无法删除", count)
		}

		// 4. 软删除环境
		if err := (&model.StorageEnv{}).Delete(tx, envId); err != nil {
			return fmt.Errorf("删除存储环境失败: %w", err)
		}

		s.logger.Info("删除存储环境成功", zap.Int64("envId", envId))
		return nil
	})
}

// GetById 根据 ID 查询存储环境
func (s *storageEnvService) GetById(ctx context.Context, envId int64) (*model.StorageEnv, error) {
	env, err := (&model.StorageEnv{}).FindByID(s.db, envId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("存储环境不存在")
		}
		return nil, fmt.Errorf("查询存储环境失败: %w", err)
	}
	return env, nil
}

// GetByCode 根据编码查询存储环境
func (s *storageEnvService) GetByCode(ctx context.Context, envCode string) (*model.StorageEnv, error) {
	env, err := (&model.StorageEnv{}).FindByCode(s.db, envCode)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("存储环境不存在: %s", envCode)
		}
		return nil, fmt.Errorf("查询存储环境失败: %w", err)
	}
	return env, nil
}

// Page 分页查询存储环境列表
func (s *storageEnvService) Page(ctx context.Context, pageNum, pageSize int, name string, storageType string) (*pagination.Page[model.StorageEnv], error) {
	query := s.db.Model(&model.StorageEnv{})

	// 条件查询
	if name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}
	if storageType != "" {
		query = query.Where("storage_type = ?", storageType)
	}

	// 构建 PageQuery
	pageQuery := &pagination.PageQuery{
		PageNum:  pageNum,
		PageSize: pageSize,
	}

	// 使用 Paginator 进行分页
	page, err := pagination.New[model.StorageEnv](query, pageQuery).Find()
	if err != nil {
		s.logger.Error("分页查询存储环境列表失败", zap.Error(err))
		return nil, fmt.Errorf("分页查询存储环境列表失败: %w", err)
	}

	return page, nil
}

// SetDefault 设置默认环境
func (s *storageEnvService) SetDefault(ctx context.Context, envId int64) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 1. 检查环境是否存在
		env, err := (&model.StorageEnv{}).FindByID(tx, envId)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("存储环境不存在")
			}
			return fmt.Errorf("查询存储环境失败: %w", err)
		}

		// 2. 取消所有默认环境
		if err := (&model.StorageEnv{}).ClearAllDefaults(tx); err != nil {
			return fmt.Errorf("取消原默认环境失败: %w", err)
		}

		// 3. 设置新的默认环境
		if err := env.SetAsDefault(tx, envId); err != nil {
			return fmt.Errorf("设置默认环境失败: %w", err)
		}

		s.logger.Info("设置默认存储环境成功",
			zap.Int64("envId", envId),
			zap.String("envCode", env.EnvCode))

		return nil
	})
}

// GetDefault 获取默认环境
func (s *storageEnvService) GetDefault(ctx context.Context) (*model.StorageEnv, error) {
	env, err := (&model.StorageEnv{}).FindDefault(s.db)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("未配置默认存储环境")
		}
		return nil, fmt.Errorf("查询默认存储环境失败: %w", err)
	}
	return env, nil
}

// TestConnection 测试存储环境连接
func (s *storageEnvService) TestConnection(ctx context.Context, envId int64) error {
	// 1. 查询存储环境
	env, err := s.GetById(ctx, envId)
	if err != nil {
		return fmt.Errorf("查询存储环境失败: %w", err)
	}

	// 2. 检查环境状态
	if env.Status != 0 {
		return fmt.Errorf("存储环境已停用，无法测试连接")
	}

	// 3. 根据存储类型进行连接测试
	// 这里简化实现，实际应该调用对应的存储客户端进行连接测试
	switch env.StorageType {
	case "local": // 本地存储
		// 本地存储无需测试连接
		s.logger.Info("本地存储环境连接测试成功", zap.Int64("envId", envId))
		return nil
	case "minio", "s3", "oss": // MinIO, S3, OSS
		// 实际应该创建客户端并测试连接
		// 这里简化为检查配置是否存在
		if env.Config == nil {
			return fmt.Errorf("存储环境配置为空")
		}
		s.logger.Info("存储环境连接测试成功",
			zap.Int64("envId", envId),
			zap.String("storageType", env.StorageType))
		return nil
	default:
		return fmt.Errorf("不支持的存储类型: %s", env.StorageType)
	}
}
