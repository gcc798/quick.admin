package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/force-c/nai-tizi/internal/domain/model"
	logging "github.com/force-c/nai-tizi/internal/logger"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

const (
	ClientCacheKeyPrefix = "s_auth_client:"
	ClientCacheTTL       = 30 * 24 * time.Hour
)

type ClientService interface {
	// AuthenticateClient 认证客户端（通过 clientKey + clientSecret）
	AuthenticateClient(ctx context.Context, clientKey, clientSecret, grantType string) (*model.AuthClient, error)
}

type clientService struct {
	db     *gorm.DB
	redis  *redis.Client
	logger logging.Logger
}

func NewClientService(db *gorm.DB, redis *redis.Client, logger logging.Logger) ClientService {
	return &clientService{db: db, redis: redis, logger: logger}
}

// AuthenticateClient 通过 clientKey 和 clientSecret 认证客户端
// 1. 根据 clientKey 查询客户端配置（优先从 Redis 缓存读取）
// 2. 验证 clientSecret 是否匹配
// 3. 检查客户端状态是否启用
// 4. 检查客户端是否支持请求的授权类型
func (s *clientService) AuthenticateClient(ctx context.Context, clientKey, clientSecret, grantType string) (*model.AuthClient, error) {
	if clientKey == "" || clientSecret == "" {
		return nil, fmt.Errorf("clientKey和clientSecret不能为空")
	}
	if grantType == "" {
		return nil, fmt.Errorf("grantType不能为空")
	}

	// 尝试从 Redis 缓存读取（使用 clientKey 作为缓存键）
	cacheKey := ClientCacheKeyPrefix + "key:" + clientKey
	val, err := s.redis.Get(ctx, cacheKey).Result()
	var client *model.AuthClient
	if err == nil {
		var cached model.AuthClient
		if json.Unmarshal([]byte(val), &cached) == nil {
			client = &cached
		}
	}

	// 缓存未命中，从数据库查询
	if client == nil {
		var m model.AuthClient
		c, err := m.FindByClientKey(s.db, clientKey)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, fmt.Errorf("客户端不存在")
			}
			return nil, fmt.Errorf("查询客户端失败: %w", err)
		}
		client = c
		// 写入缓存
		b, _ := json.Marshal(client)
		_ = s.redis.Set(ctx, cacheKey, string(b), ClientCacheTTL).Err()
	}

	// 验证 clientSecret
	if !client.VerifySecret(clientSecret) {
		return nil, fmt.Errorf("客户端认证失败")
	}

	// 检查客户端状态
	if !client.IsActive() {
		return nil, fmt.Errorf("客户端已停用")
	}

	// 检查是否支持该授权类型
	if !client.IsGrantTypeSupported(grantType) {
		return nil, fmt.Errorf("客户端不支持该授权类型")
	}

	return client, nil
}
