package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/force-c/nai-tizi/application/sys-api/internal/svc"
	"github.com/force-c/nai-tizi/application/sys-api/internal/types"
	"github.com/force-c/nai-tizi/application/sys-rpc/client/sysservice"
	"github.com/golang-jwt/jwt/v4"
	"github.com/redis/go-redis/v9"
)

const (
	refreshTokenKeyPrefix   = "refresh_token:"
	refreshTokenIndexPrefix = "refresh_token_index:"
)

type accessClaims struct {
	UserId     int64  `json:"userId"`
	UserName   string `json:"userName"`
	ClientId   string `json:"clientId"`
	DeviceType string `json:"deviceType"`
	jwt.RegisteredClaims
}

type authClient struct {
	ClientId      string
	ClientKey     string
	DeviceType    string
	Timeout       int64
	ActiveTimeout int64
}

type loginUser struct {
	Id          int64
	UserName    string
	NickName    string
	Email       string
	Phonenumber string
	Avatar      string
	UserType    int64
}

func loginWithRPC(ctx context.Context, svcCtx *svc.ServiceContext, req *types.LoginReq) (*loginUser, *authClient, error) {
	resp, err := svcCtx.SysRpcClient.AuthLogin(ctx, &sysservice.AuthLoginReq{
		ClientKey:    req.ClientKey,
		ClientSecret: req.ClientSecret,
		GrantType:    req.GrantType,
		Username:     req.Username,
		Password:     req.Password,
		Code:         req.Code,
		Phonenumber:  req.Phonenumber,
		Email:        req.Email,
		WxCode:       req.WxCode,
		Uuid:         req.Uuid,
	})
	if err != nil {
		return nil, nil, err
	}
	if resp.UserInfo == nil {
		return nil, nil, fmt.Errorf("登录失败")
	}
	return &loginUser{
			Id:          resp.UserInfo.UserId,
			UserName:    resp.UserInfo.Username,
			NickName:    resp.UserInfo.Nickname,
			Email:       resp.UserInfo.Email,
			Phonenumber: resp.UserInfo.Phonenumber,
			Avatar:      resp.UserInfo.Avatar,
			UserType:    int64(resp.UserInfo.UserType),
		}, &authClient{
			ClientId:      resp.ClientId,
			ClientKey:     resp.ClientKey,
			DeviceType:    resp.DeviceType,
			Timeout:       resp.Timeout,
			ActiveTimeout: resp.ActiveTimeout,
		}, nil
}

func buildLoginResponse(ctx context.Context, svcCtx *svc.ServiceContext, user *loginUser, client *authClient) (*types.CommonResp, error) {
	accessToken, accessExpiresIn, err := generateAccessToken(user, client, svcCtx.Config.Jwt.Secret)
	if err != nil {
		return nil, fmt.Errorf("生成Token失败")
	}
	refreshToken, err := generateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("生成Token失败")
	}
	if err := storeRefreshToken(ctx, svcCtx, user, client, refreshToken); err != nil {
		return nil, fmt.Errorf("生成Token失败")
	}
	return &types.CommonResp{Code: 200, Msg: "success", Data: map[string]interface{}{
		"accessToken":      accessToken,
		"refreshToken":     refreshToken,
		"expiresIn":        accessExpiresIn,
		"refreshExpiresIn": client.Timeout,
		"userInfo": map[string]interface{}{
			"userId":      user.Id,
			"username":    user.UserName,
			"nickname":    user.NickName,
			"phonenumber": user.Phonenumber,
			"email":       user.Email,
			"avatar":      user.Avatar,
			"userType":    user.UserType,
		},
	}}, nil
}

func refreshLoginToken(ctx context.Context, svcCtx *svc.ServiceContext, refreshToken string, clientKey, clientSecret string) (*types.CommonResp, error) {
	if refreshToken == "" {
		return nil, fmt.Errorf("RefreshToken 无效或已过期")
	}
	indexKey := refreshTokenIndexPrefix + hashToken(refreshToken)
	refreshKey, err := svcCtx.Redis.Get(ctx, indexKey).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, fmt.Errorf("RefreshToken 无效或已过期")
		}
		return nil, fmt.Errorf("RefreshToken 无效或已过期")
	}
	refreshData, err := svcCtx.Redis.HGetAll(ctx, refreshKey).Result()
	if err != nil || len(refreshData) == 0 {
		return nil, fmt.Errorf("RefreshToken 无效或已过期")
	}
	if refreshData["token"] != refreshToken {
		return nil, fmt.Errorf("RefreshToken 无效")
	}
	_ = clientSecret
	// refresh flow uses cached client metadata; clientKey must match the stored token owner
	if refreshData["clientKey"] != "" && refreshData["clientKey"] != clientKey {
		return nil, fmt.Errorf("客户端不匹配")
	}
	userId, _ := strconv.ParseInt(refreshData["userId"], 10, 64)
	userResp, err := svcCtx.SysRpcClient.UserProfile(ctx, &sysservice.IdReq{Id: userId})
	if err != nil {
		return nil, fmt.Errorf("用户不存在")
	}
	accessToken, accessExpiresIn, err := generateAccessToken(&loginUser{Id: userResp.UserId, UserName: userResp.UserName, NickName: userResp.NickName, Email: userResp.Email, Phonenumber: userResp.Phonenumber, Avatar: userResp.Avatar, UserType: int64(userResp.UserType)}, &authClient{ClientId: refreshData["clientId"], ClientKey: clientKey, DeviceType: refreshData["deviceType"], Timeout: parseInt64(refreshData["timeout"]), ActiveTimeout: parseInt64(refreshData["activeTimeout"])}, svcCtx.Config.Jwt.Secret)
	if err != nil {
		return nil, fmt.Errorf("生成新Token失败")
	}
	newRefreshToken, err := generateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("生成新Token失败")
	}
	client := &authClient{ClientId: refreshData["clientId"], ClientKey: clientKey, DeviceType: refreshData["deviceType"], Timeout: parseInt64(refreshData["timeout"]), ActiveTimeout: parseInt64(refreshData["activeTimeout"])}
	user := &loginUser{Id: userResp.UserId, UserName: userResp.UserName}
	if err := storeRefreshData(ctx, svcCtx, refreshKey, user, client, newRefreshToken); err != nil {
		return nil, fmt.Errorf("更新RefreshToken失败")
	}
	_ = svcCtx.Redis.Del(ctx, indexKey).Err()
	_ = svcCtx.Redis.Set(ctx, refreshTokenIndexPrefix+hashToken(newRefreshToken), refreshKey, time.Duration(client.Timeout)*time.Second).Err()
	return &types.CommonResp{Code: 200, Msg: "success", Data: map[string]interface{}{"accessToken": accessToken, "refreshToken": newRefreshToken, "expiresIn": accessExpiresIn, "refreshExpiresIn": client.Timeout}}, nil
}

func invalidateByToken(ctx context.Context, svcCtx *svc.ServiceContext, token string) {
	if token == "" {
		return
	}
	claims, err := parseAccessToken(token, svcCtx.Config.Jwt.Secret)
	if err != nil {
		return
	}
	refreshKey := buildRefreshKey(claims.UserId, claims.ClientId)
	refreshData, err := svcCtx.Redis.HGetAll(ctx, refreshKey).Result()
	if err == nil && len(refreshData) > 0 && refreshData["token"] != "" {
		_ = svcCtx.Redis.Del(ctx, refreshTokenIndexPrefix+hashToken(refreshData["token"])).Err()
	}
	_ = svcCtx.Redis.Del(ctx, refreshKey).Err()
}

func userIDFromToken(svcCtx *svc.ServiceContext, token string) (int64, error) {
	if token == "" {
		return 0, fmt.Errorf("未登录")
	}
	claims, err := parseAccessToken(token, svcCtx.Config.Jwt.Secret)
	if err != nil {
		return 0, fmt.Errorf("未登录")
	}
	return claims.UserId, nil
}

func generateAccessToken(user *loginUser, client *authClient, secret string) (string, int64, error) {
	expireSeconds := client.ActiveTimeout
	if expireSeconds <= 0 {
		expireSeconds = 1800
	}
	expireAt := time.Now().Add(time.Duration(expireSeconds) * time.Second)
	claims := accessClaims{UserId: user.Id, UserName: user.UserName, ClientId: client.ClientId, DeviceType: client.DeviceType, RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(expireAt), IssuedAt: jwt.NewNumericDate(time.Now()), Issuer: "NAI-TIZI-gozero"}}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", 0, err
	}
	return tokenString, expireSeconds, nil
}

func parseAccessToken(tokenString, secret string) (*accessClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &accessClaims{}, func(token *jwt.Token) (interface{}, error) { return []byte(secret), nil })
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*accessClaims)
	if !ok || !token.Valid {
		return nil, jwt.ErrTokenInvalidClaims
	}
	return claims, nil
}

func generateRefreshToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func storeRefreshToken(ctx context.Context, svcCtx *svc.ServiceContext, user *loginUser, client *authClient, refreshToken string) error {
	refreshKey := buildRefreshKey(user.Id, client.ClientId)
	if err := storeRefreshData(ctx, svcCtx, refreshKey, user, client, refreshToken); err != nil {
		return err
	}
	return svcCtx.Redis.Set(ctx, refreshTokenIndexPrefix+hashToken(refreshToken), refreshKey, time.Duration(client.Timeout)*time.Second).Err()
}

func storeRefreshData(ctx context.Context, svcCtx *svc.ServiceContext, refreshKey string, user *loginUser, client *authClient, refreshToken string) error {
	data := map[string]interface{}{"token": refreshToken, "userId": strconv.FormatInt(user.Id, 10), "userName": user.UserName, "clientId": client.ClientId, "clientKey": client.ClientKey, "deviceType": client.DeviceType, "timeout": strconv.FormatInt(client.Timeout, 10), "activeTimeout": strconv.FormatInt(client.ActiveTimeout, 10)}
	if err := svcCtx.Redis.HSet(ctx, refreshKey, data).Err(); err != nil {
		return err
	}
	return svcCtx.Redis.Expire(ctx, refreshKey, time.Duration(client.Timeout)*time.Second).Err()
}

func buildRefreshKey(userId int64, clientId string) string {
	return refreshTokenKeyPrefix + strconv.FormatInt(userId, 10) + ":" + clientId
}
func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
func parseInt64(v string) int64 { n, _ := strconv.ParseInt(v, 10, 64); return n }
