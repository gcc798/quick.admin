package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/force-c/nai-tizi/internal/domain/model"
	"github.com/force-c/nai-tizi/internal/infrastructure/jwt"
	logging "github.com/force-c/nai-tizi/internal/logger"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const (
	// TokenActiveKeyPrefix Token 活动状态 Redis Key 前缀（已废弃，使用 AccessToken 短期过期替代）
	TokenActiveKeyPrefix = "token:active:"

	// RefreshTokenKeyPrefix RefreshToken Redis Key 前缀
	// refresh_token:{userId}:{clientId} -> refreshToken
	RefreshTokenKeyPrefix = "refresh_token:"
)

// TokenManager Token 管理器接口
type TokenManager interface {
	// GenerateTokenPair 生成 AccessToken 和 RefreshToken
	GenerateTokenPair(ctx context.Context, user *model.User, client *model.AuthClient) (accessToken, refreshToken string, accessExpiresIn, refreshExpiresIn int64, err error)

	// ValidateAccessToken 验证 AccessToken
	ValidateAccessToken(ctx context.Context, token string) (*jwt.Claims, error)

	// RefreshAccessToken 使用 RefreshToken 刷新 AccessToken
	RefreshAccessToken(ctx context.Context, refreshToken string, client *model.AuthClient) (newAccessToken, newRefreshToken string, accessExpiresIn, refreshExpiresIn int64, err error)

	// InvalidateToken 使 Token 失效（登出时调用）
	InvalidateToken(ctx context.Context, userId int64, clientId string) error
}

type tokenManager struct {
	jwt    *jwt.Jwt
	redis  *redis.Client
	logger logging.Logger
}

// NewTokenManager 创建 TokenManager 实例
func NewTokenManager(jwtService *jwt.Jwt, redis *redis.Client, logger logging.Logger) TokenManager {
	return &tokenManager{
		jwt:    jwtService,
		redis:  redis,
		logger: logger,
	}
}

// GenerateTokenPair 生成 AccessToken 和 RefreshToken
// AccessToken: 使用 JWT，过期时间为 client.ActiveTimeout（短期）
// RefreshToken: 随机字符串，存储在 Redis，过期时间为 client.Timeout（长期）
func (m *tokenManager) GenerateTokenPair(ctx context.Context, user *model.User, client *model.AuthClient) (string, string, int64, int64, error) {
	// 1. 生成 AccessToken（JWT）
	accessToken, accessExpiresIn, err := m.jwt.GenerateToken(
		user.ID,
		user.UserName,
		client.ClientId,
		client.DeviceType,
		client.ActiveTimeout, // 使用 ActiveTimeout 作为 AccessToken 过期时间
	)
	if err != nil {
		m.logger.Error("生成 AccessToken 失败", zap.Error(err))
		return "", "", 0, 0, fmt.Errorf("生成 AccessToken 失败: %w", err)
	}

	// 2. 生成 RefreshToken（随机字符串）
	refreshToken, err := generateRandomToken(32)
	if err != nil {
		m.logger.Error("生成 RefreshToken 失败", zap.Error(err))
		return "", "", 0, 0, fmt.Errorf("生成 RefreshToken 失败: %w", err)
	}

	// 3. 将 RefreshToken 存储到 Redis
	refreshKey := m.getRefreshTokenKey(user.ID, client.ClientId)
	refreshTTL := time.Duration(client.Timeout) * time.Second

	// 存储 RefreshToken 的元数据（包含用户信息）
	refreshData := map[string]interface{}{
		"token":      refreshToken,
		"userId":     user.ID,
		"userName":   user.UserName,
		"clientId":   client.ClientId,
		"deviceType": client.DeviceType,
		"createdAt":  time.Now().Unix(),
	}

	err = m.redis.HSet(ctx, refreshKey, refreshData).Err()
	if err != nil {
		m.logger.Error("存储 RefreshToken 失败", zap.Error(err))
		return "", "", 0, 0, fmt.Errorf("存储 RefreshToken 失败: %w", err)
	}

	// 设置过期时间
	err = m.redis.Expire(ctx, refreshKey, refreshTTL).Err()
	if err != nil {
		m.logger.Warn("设置 RefreshToken 过期时间失败", zap.Error(err))
	}

	m.logger.Info("生成 Token 对成功",
		zap.Int64("userId", user.ID),
		zap.String("clientId", client.ClientId),
		zap.Int64("accessExpiresIn", accessExpiresIn),
		zap.Int64("refreshExpiresIn", client.Timeout))

	return accessToken, refreshToken, accessExpiresIn, client.Timeout, nil
}

// ValidateAccessToken 验证 AccessToken
func (m *tokenManager) ValidateAccessToken(ctx context.Context, token string) (*jwt.Claims, error) {
	claims, err := m.jwt.ValidateToken(token)
	if err != nil {
		return nil, fmt.Errorf("AccessToken 无效或已过期")
	}
	return claims, nil
}

// RefreshAccessToken 使用 RefreshToken 刷新 AccessToken
// 1. 验证 RefreshToken 是否存在且有效
// 2. 生成新的 AccessToken
// 3. 轮换 RefreshToken（生成新的，使旧的失效）
func (m *tokenManager) RefreshAccessToken(ctx context.Context, refreshToken string, client *model.AuthClient) (string, string, int64, int64, error) {
	// 1. 查找 RefreshToken（遍历所有用户的 RefreshToken）
	// 注意：这里为了性能，我们需要优化查找方式
	// 方案：使用 refreshToken 的哈希作为额外的索引
	tokenHash := generateTokenHash(refreshToken)
	indexKey := "refresh_token_index:" + tokenHash

	// 先从索引中获取用户信息
	userKey, err := m.redis.Get(ctx, indexKey).Result()
	if err != nil {
		m.logger.Warn("RefreshToken 索引不存在", zap.Error(err))
		return "", "", 0, 0, fmt.Errorf("RefreshToken 无效或已过期")
	}

	// 2. 从 Redis 获取 RefreshToken 数据
	refreshData, err := m.redis.HGetAll(ctx, userKey).Result()
	if err != nil || len(refreshData) == 0 {
		m.logger.Warn("RefreshToken 数据不存在", zap.Error(err))
		return "", "", 0, 0, fmt.Errorf("RefreshToken 无效或已过期")
	}

	// 3. 验证 RefreshToken 是否匹配
	storedToken := refreshData["token"]
	if storedToken != refreshToken {
		m.logger.Warn("RefreshToken 不匹配")
		return "", "", 0, 0, fmt.Errorf("RefreshToken 无效")
	}

	// 4. 验证 clientId 是否匹配
	if refreshData["clientId"] != client.ClientId {
		m.logger.Warn("ClientId 不匹配",
			zap.String("expected", refreshData["clientId"]),
			zap.String("actual", client.ClientId))
		return "", "", 0, 0, fmt.Errorf("客户端不匹配")
	}

	// 5. 提取用户信息
	userId := parseInt64(refreshData["userId"])
	userName := refreshData["userName"]
	deviceType := refreshData["deviceType"]

	// 6. 生成新的 AccessToken
	newAccessToken, accessExpiresIn, err := m.jwt.GenerateToken(
		userId,
		userName,
		client.ClientId,
		deviceType,
		client.ActiveTimeout,
	)
	if err != nil {
		m.logger.Error("生成新 AccessToken 失败", zap.Error(err))
		return "", "", 0, 0, fmt.Errorf("生成新 AccessToken 失败: %w", err)
	}

	// 7. 轮换 RefreshToken（生成新的）
	newRefreshToken, err := generateRandomToken(32)
	if err != nil {
		m.logger.Error("生成新 RefreshToken 失败", zap.Error(err))
		return "", "", 0, 0, fmt.Errorf("生成新 RefreshToken 失败: %w", err)
	}

	// 8. 更新 Redis 中的 RefreshToken
	refreshTTL := time.Duration(client.Timeout) * time.Second
	newRefreshData := map[string]interface{}{
		"token":      newRefreshToken,
		"userId":     userId,
		"userName":   userName,
		"clientId":   client.ClientId,
		"deviceType": deviceType,
		"createdAt":  time.Now().Unix(),
	}

	err = m.redis.HSet(ctx, userKey, newRefreshData).Err()
	if err != nil {
		m.logger.Error("更新 RefreshToken 失败", zap.Error(err))
		return "", "", 0, 0, fmt.Errorf("更新 RefreshToken 失败: %w", err)
	}

	// 重置过期时间
	err = m.redis.Expire(ctx, userKey, refreshTTL).Err()
	if err != nil {
		m.logger.Warn("设置 RefreshToken 过期时间失败", zap.Error(err))
	}

	// 9. 更新索引（删除旧的，创建新的）
	_ = m.redis.Del(ctx, indexKey).Err()
	newTokenHash := generateTokenHash(newRefreshToken)
	newIndexKey := "refresh_token_index:" + newTokenHash
	_ = m.redis.Set(ctx, newIndexKey, userKey, refreshTTL).Err()

	m.logger.Info("刷新 Token 成功",
		zap.Int64("userId", userId),
		zap.String("clientId", client.ClientId))

	return newAccessToken, newRefreshToken, accessExpiresIn, client.Timeout, nil
}

// InvalidateToken 使 Token 失效（登出时调用）
// 删除 Redis 中的 RefreshToken
func (m *tokenManager) InvalidateToken(ctx context.Context, userId int64, clientId string) error {
	refreshKey := m.getRefreshTokenKey(userId, clientId)

	// 获取 RefreshToken 以删除索引
	refreshData, err := m.redis.HGetAll(ctx, refreshKey).Result()
	if err == nil && len(refreshData) > 0 {
		token := refreshData["token"]
		if token != "" {
			tokenHash := generateTokenHash(token)
			indexKey := "refresh_token_index:" + tokenHash
			_ = m.redis.Del(ctx, indexKey).Err()
		}
	}

	// 删除 RefreshToken
	return m.redis.Del(ctx, refreshKey).Err()
}

// getRefreshTokenKey 获取 RefreshToken Redis Key
func (m *tokenManager) getRefreshTokenKey(userId int64, clientId string) string {
	return fmt.Sprintf("%s%d:%s", RefreshTokenKeyPrefix, userId, clientId)
}

// generateRandomToken 生成随机 Token
func generateRandomToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// generateTokenHash 生成 Token 的 SHA256 哈希值
func generateTokenHash(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

// parseInt64 将字符串转换为 int64
func parseInt64(s string) int64 {
	var result int64
	fmt.Sscanf(s, "%d", &result)
	return result
}
