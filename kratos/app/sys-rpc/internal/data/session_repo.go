package data

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	accessTokenTTL  = 30 * time.Minute
	refreshTokenTTL = 7 * 24 * time.Hour
)

type sessionRecord struct {
	UserID    int64  `json:"userId"`
	ClientID  string `json:"clientId"`
	CreatedAt int64  `json:"createdAt"`
}

type tokenPair struct {
	AccessToken  string
	RefreshToken string
}

func (r *Resources) IssueSession(ctx context.Context, userID int64, client *AuthClientInfo) (string, string, error) {
	return r.issueSession(ctx, userID, client, true)
}

func (r *Resources) issueSession(ctx context.Context, userID int64, client *AuthClientInfo, reuseExisting bool) (string, string, error) {
	if r == nil || r.Redis == nil {
		return "", "", fmt.Errorf("redis is not configured")
	}
	clientID := ""
	if client != nil {
		clientID = strings.TrimSpace(client.ClientID)
	}
	if userID <= 0 || clientID == "" {
		return "", "", fmt.Errorf("invalid session subject")
	}
	if reuseExisting && r.Auth.ShareToken {
		if pair, err := r.currentTokenPair(ctx, userID, clientID); err == nil && pair != nil {
			return pair.AccessToken, pair.RefreshToken, nil
		}
	}
	if !r.Auth.AllowConcurrent || r.Auth.ShareToken {
		if err := r.revokeUserClientSessions(ctx, userID, clientID); err != nil {
			return "", "", err
		}
	}
	nonce, err := randomHex(16)
	if err != nil {
		return "", "", err
	}
	accessToken := fmt.Sprintf("kratos-access-%d-%s", userID, nonce)
	refreshToken := fmt.Sprintf("kratos-refresh-%d-%s", userID, nonce)
	record, err := marshalSessionRecord(sessionRecord{
		UserID:    userID,
		ClientID:  clientID,
		CreatedAt: time.Now().Unix(),
	})
	if err != nil {
		return "", "", err
	}
	accessTTL, refreshTTL := sessionTTLs(client)
	pipe := r.Redis.TxPipeline()
	pipe.Set(ctx, sessionAccessKey(accessToken), record, accessTTL)
	pipe.Set(ctx, sessionRefreshKey(refreshToken), record, refreshTTL)
	pipe.Set(ctx, sessionAccessToRefreshKey(accessToken), refreshToken, accessTTL)
	pipe.Set(ctx, sessionRefreshToAccessKey(refreshToken), accessToken, refreshTTL)
	pipe.Set(ctx, userClientAccessKey(userID, clientID), accessToken, refreshTTL)
	pipe.Set(ctx, userClientRefreshKey(userID, clientID), refreshToken, refreshTTL)
	if _, err := pipe.Exec(ctx); err != nil {
		return "", "", err
	}
	return accessToken, refreshToken, nil
}

func (r *Resources) RefreshSession(ctx context.Context, refreshToken string, client *AuthClientInfo) (int64, string, string, error) {
	if r == nil || r.Redis == nil {
		return 0, "", "", fmt.Errorf("redis is not configured")
	}
	refreshToken = strings.TrimSpace(refreshToken)
	if refreshToken == "" {
		return 0, "", "", fmt.Errorf("invalid refresh token")
	}
	recordValue, err := r.Redis.Get(ctx, sessionRefreshKey(refreshToken)).Result()
	if err != nil {
		return 0, "", "", fmt.Errorf("invalid refresh token")
	}
	record, err := unmarshalSessionRecord(recordValue)
	if err != nil || record.UserID <= 0 {
		return 0, "", "", fmt.Errorf("invalid refresh token")
	}
	clientID := ""
	if client != nil {
		clientID = strings.TrimSpace(client.ClientID)
	}
	if clientID == "" || record.ClientID != clientID {
		return 0, "", "", fmt.Errorf("客户端不匹配")
	}
	oldAccessToken, _ := r.Redis.Get(ctx, sessionRefreshToAccessKey(refreshToken)).Result()
	accessToken, newRefreshToken, err := r.issueSession(ctx, record.UserID, client, false)
	if err != nil {
		return 0, "", "", err
	}
	if err := r.RevokeSession(ctx, oldAccessToken, refreshToken); err != nil {
		return 0, "", "", err
	}
	return record.UserID, accessToken, newRefreshToken, nil
}

func (r *Resources) RevokeSession(ctx context.Context, accessToken, refreshToken string) error {
	if r == nil || r.Redis == nil {
		return fmt.Errorf("redis is not configured")
	}
	accessToken = strings.TrimSpace(accessToken)
	refreshToken = strings.TrimSpace(refreshToken)
	if accessToken != "" && refreshToken == "" {
		refreshToken, _ = r.Redis.Get(ctx, sessionAccessToRefreshKey(accessToken)).Result()
	}
	if refreshToken != "" && accessToken == "" {
		accessToken, _ = r.Redis.Get(ctx, sessionRefreshToAccessKey(refreshToken)).Result()
	}
	record := sessionRecord{}
	for _, key := range []string{sessionAccessKey(accessToken), sessionRefreshKey(refreshToken)} {
		if key == "" {
			continue
		}
		value, err := r.Redis.Get(ctx, key).Result()
		if err != nil {
			continue
		}
		parsed, err := unmarshalSessionRecord(value)
		if err == nil && parsed.UserID > 0 {
			record = parsed
			break
		}
	}
	keys := make([]string, 0, 4)
	if accessToken != "" {
		keys = append(keys, sessionAccessKey(accessToken), sessionAccessToRefreshKey(accessToken))
	}
	if refreshToken != "" {
		keys = append(keys, sessionRefreshKey(refreshToken), sessionRefreshToAccessKey(refreshToken))
	}
	if record.UserID > 0 && record.ClientID != "" {
		keys = append(keys, userClientAccessKey(record.UserID, record.ClientID), userClientRefreshKey(record.UserID, record.ClientID))
	}
	if len(keys) == 0 {
		return nil
	}
	return r.Redis.Del(ctx, keys...).Err()
}

func (r *Resources) SessionDetails(ctx context.Context, accessToken string) (*sessionRecord, error) {
	if r == nil || r.Redis == nil {
		return nil, fmt.Errorf("redis is not configured")
	}
	accessToken = strings.TrimSpace(accessToken)
	if accessToken == "" {
		return nil, fmt.Errorf("access token is required")
	}
	value, err := r.Redis.Get(ctx, sessionAccessKey(accessToken)).Result()
	if err != nil {
		return nil, fmt.Errorf("invalid access token")
	}
	record, err := unmarshalSessionRecord(value)
	if err != nil || record.UserID <= 0 {
		return nil, fmt.Errorf("invalid access token")
	}
	return &record, nil
}

func (r *Resources) SessionUserID(ctx context.Context, accessToken string) (int64, error) {
	record, err := r.SessionDetails(ctx, accessToken)
	if err != nil {
		return 0, err
	}
	return record.UserID, nil
}

func sessionAccessKey(token string) string {
	return "auth:access:" + token
}

func sessionRefreshKey(token string) string {
	return "auth:refresh:" + token
}

func sessionAccessToRefreshKey(token string) string {
	return "auth:access:refresh:" + token
}

func sessionRefreshToAccessKey(token string) string {
	return "auth:refresh:access:" + token
}

func userClientAccessKey(userID int64, clientID string) string {
	return "auth:user:access:" + strconv.FormatInt(userID, 10) + ":" + strings.TrimSpace(clientID)
}

func userClientRefreshKey(userID int64, clientID string) string {
	return "auth:user:refresh:" + strconv.FormatInt(userID, 10) + ":" + strings.TrimSpace(clientID)
}

func sessionTTLs(client *AuthClientInfo) (time.Duration, time.Duration) {
	accessTTL := accessTokenTTL
	refreshTTL := refreshTokenTTL
	if client != nil && client.ActiveTimeout > 0 {
		accessTTL = time.Duration(client.ActiveTimeout) * time.Second
	}
	if client != nil && client.Timeout > 0 {
		refreshTTL = time.Duration(client.Timeout) * time.Second
	}
	return accessTTL, refreshTTL
}

func (r *Resources) currentTokenPair(ctx context.Context, userID int64, clientID string) (*tokenPair, error) {
	accessToken, err := r.Redis.Get(ctx, userClientAccessKey(userID, clientID)).Result()
	if err != nil || strings.TrimSpace(accessToken) == "" {
		return nil, fmt.Errorf("session not found")
	}
	refreshToken, err := r.Redis.Get(ctx, userClientRefreshKey(userID, clientID)).Result()
	if err != nil || strings.TrimSpace(refreshToken) == "" {
		return nil, fmt.Errorf("session not found")
	}
	if _, err = r.SessionDetails(ctx, accessToken); err != nil {
		_ = r.RevokeSession(ctx, accessToken, refreshToken)
		return nil, err
	}
	return &tokenPair{AccessToken: accessToken, RefreshToken: refreshToken}, nil
}

func (r *Resources) revokeUserClientSessions(ctx context.Context, userID int64, clientID string) error {
	pair, err := r.currentTokenPair(ctx, userID, clientID)
	if err != nil || pair == nil {
		return nil
	}
	return r.RevokeSession(ctx, pair.AccessToken, pair.RefreshToken)
}

func marshalSessionRecord(record sessionRecord) (string, error) {
	value, err := json.Marshal(record)
	if err != nil {
		return "", err
	}
	return string(value), nil
}

func unmarshalSessionRecord(value string) (sessionRecord, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return sessionRecord{}, fmt.Errorf("empty session record")
	}
	if userID, err := strconv.ParseInt(value, 10, 64); err == nil && userID > 0 {
		return sessionRecord{UserID: userID}, nil
	}
	var record sessionRecord
	if err := json.Unmarshal([]byte(value), &record); err != nil {
		return sessionRecord{}, err
	}
	return record, nil
}

func randomHex(size int) (string, error) {
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}
