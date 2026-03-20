package service

import (
	"context"
	"fmt"
	"time"

	"github.com/force-c/nai-tizi/internal/config"
	logging "github.com/force-c/nai-tizi/internal/logger"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const (
	// UserTokenKeyPrefix 用户 Token Redis Key 前缀（共享 Token 模式）
	// user:token:{userId}:{clientId} -> token
	UserTokenKeyPrefix = "user:token:"

	// UserTokensKeyPrefix 用户 Token 集合 Redis Key 前缀（非共享 Token 模式）
	// user:tokens:{userId}:{clientId} -> Set[token]
	UserTokensKeyPrefix = "user:tokens:"
)

// ConcurrentLoginManager 并发登录管理器接口
type ConcurrentLoginManager interface {
	// HandleConcurrentLogin 检查并处理并发登录
	// 返回：是否应该使用现有 Token，现有 Token（如果有）
	HandleConcurrentLogin(ctx context.Context, userId int64, clientId string, timeout int64) (useExisting bool, existingToken string, err error)

	// RecordLogin 记录用户登录
	RecordLogin(ctx context.Context, userId int64, clientId string, token string, timeout int64) error

	// InvalidateUserTokens 使用户的所有 Token 失效
	InvalidateUserTokens(ctx context.Context, userId int64, clientId string) error
}

type concurrentLoginManager struct {
	redis        *redis.Client
	tokenManager TokenManager
	config       *config.Config
	logger       logging.Logger
}

// NewConcurrentLoginManager 创建 ConcurrentLoginManager 实例
func NewConcurrentLoginManager(redis *redis.Client, tokenManager TokenManager, cfg *config.Config, logger logging.Logger) ConcurrentLoginManager {
	return &concurrentLoginManager{
		redis:        redis,
		tokenManager: tokenManager,
		config:       cfg,
		logger:       logger,
	}
}

// HandleConcurrentLogin 检查并处理并发登录
// 1. 如果不允许并发登录，使所有旧 Token 失效
// 2. 如果允许并发登录且 share-token=true，返回现有 Token
// 3. 如果允许并发登录且 share-token=false，允许生成新 Token
func (m *concurrentLoginManager) HandleConcurrentLogin(ctx context.Context, userId int64, clientId string, timeout int64) (bool, string, error) {
	// 不允许并发登录：使所有旧 Token 失效
	if !m.config.Auth.AllowConcurrent {
		err := m.InvalidateUserTokens(ctx, userId, clientId)
		if err != nil {
			m.logger.Warn("使旧 Token 失效失败", zap.Error(err))
		}
		return false, "", nil
	}

	// 允许并发登录且共享 Token：返回现有 Token
	if m.config.Auth.ShareToken {
		key := m.getUserTokenKey(userId, clientId)
		existingToken, err := m.redis.Get(ctx, key).Result()
		if err == nil && existingToken != "" {
			// 验证现有 AccessToken 是否仍然有效
			_, err := m.tokenManager.ValidateAccessToken(ctx, existingToken)
			if err == nil {
				// Token 有效，返回现有 Token
				return true, existingToken, nil
			}
			// Token 已失效，删除记录
			_ = m.redis.Del(ctx, key).Err()
		}
	}

	// 允许并发登录且不共享 Token：允许生成新 Token
	return false, "", nil
}

// RecordLogin 记录用户登录
// 1. 如果 share-token=true，存储单个 Token
// 2. 如果 share-token=false，将 Token 添加到集合
func (m *concurrentLoginManager) RecordLogin(ctx context.Context, userId int64, clientId string, token string, timeout int64) error {
	ttl := time.Duration(timeout) * time.Second

	if m.config.Auth.ShareToken {
		// 共享 Token 模式：存储单个 Token
		key := m.getUserTokenKey(userId, clientId)
		return m.redis.Set(ctx, key, token, ttl).Err()
	}

	// 非共享 Token 模式：将 Token 添加到集合
	key := m.getUserTokensKey(userId, clientId)
	pipe := m.redis.Pipeline()
	pipe.SAdd(ctx, key, token)
	pipe.Expire(ctx, key, ttl)
	_, err := pipe.Exec(ctx)
	return err
}

// InvalidateUserTokens 使用户的所有 Token 失效
// 1. 删除 Redis 中的用户 Token 记录
// 2. 使 RefreshToken 失效
func (m *concurrentLoginManager) InvalidateUserTokens(ctx context.Context, userId int64, clientId string) error {
	// 使 RefreshToken 失效
	_ = m.tokenManager.InvalidateToken(ctx, userId, clientId)

	if m.config.Auth.ShareToken {
		// 共享 Token 模式：删除单个 Token 记录
		key := m.getUserTokenKey(userId, clientId)
		return m.redis.Del(ctx, key).Err()
	}

	// 非共享 Token 模式：删除 Token 集合
	key := m.getUserTokensKey(userId, clientId)
	return m.redis.Del(ctx, key).Err()
}

// getUserTokenKey 获取用户 Token Redis Key（共享 Token 模式）
func (m *concurrentLoginManager) getUserTokenKey(userId int64, clientId string) string {
	return fmt.Sprintf("%s%d:%s", UserTokenKeyPrefix, userId, clientId)
}

// getUserTokensKey 获取用户 Token 集合 Redis Key（非共享 Token 模式）
func (m *concurrentLoginManager) getUserTokensKey(userId int64, clientId string) string {
	return fmt.Sprintf("%s%d:%s", UserTokensKeyPrefix, userId, clientId)
}
