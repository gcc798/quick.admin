package biz

import (
	"context"
	"strconv"
	"strings"

	"google.golang.org/grpc/metadata"
)

func currentUserID(ctx context.Context) int64 {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return 0
	}
	for _, key := range []string{"x-user-id", "user-id", "userid"} {
		values := md.Get(key)
		if len(values) == 0 {
			continue
		}
		id, err := strconv.ParseInt(values[0], 10, 64)
		if err == nil && id > 0 {
			return id
		}
	}
	return 0
}

func currentAccessToken(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	for _, key := range []string{"x-access-token", "authorization"} {
		values := md.Get(key)
		if len(values) == 0 {
			continue
		}
		value := strings.TrimSpace(values[0])
		if value == "" {
			continue
		}
		parts := strings.SplitN(value, " ", 2)
		if len(parts) == 2 {
			value = strings.TrimSpace(parts[1])
		}
		if value != "" {
			return value
		}
	}
	return ""
}

func currentClientIP(ctx context.Context) string {
	return firstIncomingValue(ctx, "x-client-ip", "x-forwarded-for", "x-real-ip")
}

func currentUserAgent(ctx context.Context) string {
	return firstIncomingValue(ctx, "x-user-agent", "user-agent")
}

func firstIncomingValue(ctx context.Context, keys ...string) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	for _, key := range keys {
		values := md.Get(key)
		if len(values) == 0 {
			continue
		}
		value := strings.TrimSpace(values[0])
		if value != "" {
			return value
		}
	}
	return ""
}
