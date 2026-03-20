package commonutil

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/force-c/nai-tizi/application/sys-api/internal/svc"
	"github.com/golang-jwt/jwt/v4"
)

type accessClaims struct {
	UserId int64 `json:"userId"`
	jwt.RegisteredClaims
}

func BearerToken(r *http.Request, header string) string {
	token := r.Header.Get(header)
	return strings.TrimSpace(strings.TrimPrefix(token, "Bearer "))
}

func UserIDFromRequest(svcCtx *svc.ServiceContext, r *http.Request) (int64, error) {
	token := BearerToken(r, svcCtx.Config.Auth.TokenHeader)
	if token == "" {
		return 0, fmt.Errorf("未登录")
	}
	parsed, err := jwt.ParseWithClaims(token, &accessClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(svcCtx.Config.Jwt.Secret), nil
	})
	if err != nil {
		return 0, fmt.Errorf("未登录")
	}
	claims, ok := parsed.Claims.(*accessClaims)
	if !ok || !parsed.Valid {
		return 0, fmt.Errorf("未登录")
	}
	return claims.UserId, nil
}
