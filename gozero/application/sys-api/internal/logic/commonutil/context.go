package commonutil

import "context"

type contextKey string

const userIDContextKey contextKey = "userId"

func WithUserID(ctx context.Context, userID int64) context.Context {
	return context.WithValue(ctx, userIDContextKey, userID)
}

func UserIDFromContext(ctx context.Context) int64 {
	value := ctx.Value(userIDContextKey)
	userID, ok := value.(int64)
	if !ok {
		return 0
	}
	return userID
}
