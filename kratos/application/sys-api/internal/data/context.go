package data

import (
	"context"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-kratos/kratos/v2/transport"
	khttp "github.com/go-kratos/kratos/v2/transport/http"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type authContextKey string

const (
	currentUserIDKey      authContextKey = "current-user-id"
	currentUserNameKey    authContextKey = "current-user-name"
	currentAccessTokenKey authContextKey = "current-access-token"
	currentClientIPKey    authContextKey = "current-client-ip"
	currentUserAgentKey   authContextKey = "current-user-agent"
)

func WithCurrentAuth(ctx context.Context, userID int64, userName, token string) context.Context {
	ctx = context.WithValue(ctx, currentUserIDKey, userID)
	ctx = context.WithValue(ctx, currentUserNameKey, strings.TrimSpace(userName))
	ctx = context.WithValue(ctx, currentAccessTokenKey, strings.TrimSpace(token))
	if req, ok := khttp.RequestFromServerContext(ctx); ok {
		ctx = context.WithValue(ctx, currentClientIPKey, requestClientIP(req))
		ctx = context.WithValue(ctx, currentUserAgentKey, strings.TrimSpace(req.UserAgent()))
	}
	return ctx
}

func outgoingUserInterceptor(
	ctx context.Context,
	method string,
	req any,
	reply any,
	cc *grpc.ClientConn,
	invoker grpc.UnaryInvoker,
	opts ...grpc.CallOption,
) error {
	return invoker(withOutgoingContext(ctx), method, req, reply, cc, opts...)
}

func withOutgoingContext(ctx context.Context) context.Context {
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		md = metadata.New(nil)
	} else {
		md = md.Copy()
	}
	if userID := currentUserIDFromContext(ctx); userID > 0 {
		md.Set("x-user-id", strconv.FormatInt(userID, 10))
	}
	if accessToken := currentAccessTokenFromContext(ctx); accessToken != "" {
		md.Set("x-access-token", accessToken)
	}
	if clientIP := currentClientIPFromContext(ctx); clientIP != "" {
		md.Set("x-client-ip", clientIP)
	}
	if userAgent := currentUserAgentFromContext(ctx); userAgent != "" {
		md.Set("x-user-agent", userAgent)
	}
	return metadata.NewOutgoingContext(ctx, md)
}

func currentUserIDFromContext(ctx context.Context) int64 {
	if value, ok := ctx.Value(currentUserIDKey).(int64); ok && value > 0 {
		return value
	}
	if tr, ok := transport.FromServerContext(ctx); ok {
		if userID := parseUserID(tr.RequestHeader().Get("x-user-id")); userID > 0 {
			return userID
		}
		if userID := parseAuthorizationUserID(tr.RequestHeader().Get("Authorization")); userID > 0 {
			return userID
		}
	}
	return 0
}

func CurrentUserID(ctx context.Context) int64 {
	return currentUserIDFromContext(ctx)
}

func CurrentUserName(ctx context.Context) string {
	if value, ok := ctx.Value(currentUserNameKey).(string); ok && strings.TrimSpace(value) != "" {
		return strings.TrimSpace(value)
	}
	return ""
}

func currentAccessTokenFromContext(ctx context.Context) string {
	if value, ok := ctx.Value(currentAccessTokenKey).(string); ok && strings.TrimSpace(value) != "" {
		return strings.TrimSpace(value)
	}
	if tr, ok := transport.FromServerContext(ctx); ok {
		return parseAuthorizationToken(tr.RequestHeader().Get("Authorization"))
	}
	return ""
}

func CurrentAccessToken(ctx context.Context) string {
	return currentAccessTokenFromContext(ctx)
}

func CurrentClientIP(ctx context.Context) string {
	return currentClientIPFromContext(ctx)
}

func currentClientIPFromContext(ctx context.Context) string {
	if value, ok := ctx.Value(currentClientIPKey).(string); ok && strings.TrimSpace(value) != "" {
		return strings.TrimSpace(value)
	}
	if req, ok := khttp.RequestFromServerContext(ctx); ok {
		return requestClientIP(req)
	}
	return ""
}

func currentUserAgentFromContext(ctx context.Context) string {
	if value, ok := ctx.Value(currentUserAgentKey).(string); ok && strings.TrimSpace(value) != "" {
		return strings.TrimSpace(value)
	}
	if req, ok := khttp.RequestFromServerContext(ctx); ok {
		return strings.TrimSpace(req.UserAgent())
	}
	return ""
}

func parseAuthorizationUserID(value string) int64 {
	value = parseAuthorizationToken(value)
	if value == "" {
		return 0
	}
	for _, prefix := range []string{"kratos-access-", "kratos-refresh-"} {
		if !strings.HasPrefix(value, prefix) {
			continue
		}
		rest := strings.TrimPrefix(value, prefix)
		idx := strings.Index(rest, "-")
		if idx <= 0 {
			return 0
		}
		return parseUserID(rest[:idx])
	}
	return 0
}

func parseAuthorizationToken(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	parts := strings.SplitN(value, " ", 2)
	if len(parts) == 2 {
		value = strings.TrimSpace(parts[1])
	}
	return value
}

func parseUserID(value string) int64 {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0
	}
	userID, err := strconv.ParseInt(value, 10, 64)
	if err != nil || userID <= 0 {
		return 0
	}
	return userID
}

func requestClientIP(req *http.Request) string {
	if req == nil {
		return ""
	}
	for _, key := range []string{"X-Forwarded-For", "X-Real-IP"} {
		value := strings.TrimSpace(req.Header.Get(key))
		if value == "" {
			continue
		}
		if key == "X-Forwarded-For" {
			parts := strings.Split(value, ",")
			if len(parts) > 0 {
				return strings.TrimSpace(parts[0])
			}
		}
		return value
	}
	host, _, err := net.SplitHostPort(strings.TrimSpace(req.RemoteAddr))
	if err == nil {
		return strings.TrimSpace(host)
	}
	return strings.TrimSpace(req.RemoteAddr)
}
