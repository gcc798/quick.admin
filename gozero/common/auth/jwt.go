package auth

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v4"
)

type AccessClaims struct {
	UserID      int64    `json:"userId"`
	UserName    string   `json:"userName"`
	ClientID    string   `json:"clientId"`
	DeviceType  string   `json:"deviceType"`
	OrgID       int64    `json:"orgId"`
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions"`
	jwt.RegisteredClaims
}

func TokenFromRequest(r *http.Request, tokenHeader string) string {
	token := strings.TrimSpace(r.Header.Get(tokenHeader))
	if token == "" {
		token = strings.TrimSpace(r.URL.Query().Get(tokenHeader))
	}
	if token == "" {
		token = strings.TrimSpace(r.URL.Query().Get("token"))
	}
	return strings.TrimSpace(strings.TrimPrefix(token, "Bearer "))
}

func ParseAccessToken(tokenString, secret string) (*AccessClaims, error) {
	if strings.TrimSpace(tokenString) == "" {
		return nil, fmt.Errorf("未登录")
	}
	token, err := jwt.ParseWithClaims(tokenString, &AccessClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("不支持的签名方法")
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*AccessClaims)
	if !ok || !token.Valid {
		return nil, jwt.ErrTokenInvalidClaims
	}
	return claims, nil
}
