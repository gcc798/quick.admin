package menu

import (
	"fmt"
	"strings"

	"github.com/gcc798/nai-tizi/application/sys-api/internal/svc"
	"github.com/golang-jwt/jwt/v4"
)

type accessClaims struct {
	UserId int64 `json:"userId"`
	jwt.RegisteredClaims
}

func ResolveUserID(svcCtx *svc.ServiceContext, token string) (int64, error) {
	if token == "" {
		return 0, fmt.Errorf("未登录")
	}
	token = strings.TrimSpace(strings.TrimPrefix(token, "Bearer "))
	parsed, err := jwt.ParseWithClaims(token, &accessClaims{}, func(token *jwt.Token) (interface{}, error) { return []byte(svcCtx.Config.Jwt.Secret), nil })
	if err != nil {
		return 0, fmt.Errorf("未登录")
	}
	claims, ok := parsed.Claims.(*accessClaims)
	if !ok || !parsed.Valid {
		return 0, fmt.Errorf("未登录")
	}
	return claims.UserId, nil
}
