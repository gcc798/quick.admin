package auth

import "context"

type contextKey string

const (
	userIDContextKey     contextKey = "userId"
	userNameContextKey   contextKey = "userName"
	clientIDContextKey   contextKey = "clientId"
	deviceTypeContextKey contextKey = "deviceType"
	orgIDContextKey      contextKey = "orgId"
	rolesContextKey      contextKey = "roles"
	permissionsKey       contextKey = "permissions"
)

type UserContext struct {
	UserID      int64
	UserName    string
	ClientID    string
	DeviceType  string
	OrgID       int64
	Roles       []string
	Permissions []string
}

func WithUserContext(ctx context.Context, user UserContext) context.Context {
	ctx = context.WithValue(ctx, userIDContextKey, user.UserID)
	ctx = context.WithValue(ctx, userNameContextKey, user.UserName)
	ctx = context.WithValue(ctx, clientIDContextKey, user.ClientID)
	ctx = context.WithValue(ctx, deviceTypeContextKey, user.DeviceType)
	ctx = context.WithValue(ctx, orgIDContextKey, user.OrgID)
	ctx = context.WithValue(ctx, rolesContextKey, cloneStrings(user.Roles))
	ctx = context.WithValue(ctx, permissionsKey, cloneStrings(user.Permissions))
	return ctx
}

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

func UserNameFromContext(ctx context.Context) string {
	value := ctx.Value(userNameContextKey)
	userName, ok := value.(string)
	if !ok {
		return ""
	}
	return userName
}

func ClientIDFromContext(ctx context.Context) string {
	value := ctx.Value(clientIDContextKey)
	clientID, ok := value.(string)
	if !ok {
		return ""
	}
	return clientID
}

func DeviceTypeFromContext(ctx context.Context) string {
	value := ctx.Value(deviceTypeContextKey)
	deviceType, ok := value.(string)
	if !ok {
		return ""
	}
	return deviceType
}

func OrgIDFromContext(ctx context.Context) int64 {
	value := ctx.Value(orgIDContextKey)
	orgID, ok := value.(int64)
	if !ok {
		return 0
	}
	return orgID
}

func RolesFromContext(ctx context.Context) []string {
	value := ctx.Value(rolesContextKey)
	roles, ok := value.([]string)
	if !ok {
		return nil
	}
	return cloneStrings(roles)
}

func PermissionsFromContext(ctx context.Context) []string {
	value := ctx.Value(permissionsKey)
	permissions, ok := value.([]string)
	if !ok {
		return nil
	}
	return cloneStrings(permissions)
}

func cloneStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	cloned := make([]string, len(values))
	copy(cloned, values)
	return cloned
}
