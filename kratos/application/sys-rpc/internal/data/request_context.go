package data

import (
	"context"
	"strconv"
	"strings"

	"google.golang.org/grpc/metadata"
)

func currentOperatorID(ctx context.Context) int64 {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return 0
	}
	for _, key := range []string{"x-user-id", "user-id", "userid"} {
		values := md.Get(key)
		if len(values) == 0 {
			continue
		}
		id, err := strconv.ParseInt(strings.TrimSpace(values[0]), 10, 64)
		if err == nil && id > 0 {
			return id
		}
	}
	return 0
}

func operatorID(ctx context.Context, fallback int64) int64 {
	if value := currentOperatorID(ctx); value > 0 {
		return value
	}
	if fallback > 0 {
		return fallback
	}
	return 0
}
