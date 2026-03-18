package data

import (
	"context"
	"strings"

	v1 "github.com/force-c/nai-tizi/kratos/api/system/v1"
	"google.golang.org/grpc"
)

type AuthRepo struct {
	conn   *grpc.ClientConn
	client v1.AuthServiceClient
}

func NewAuthRepo(endpoint string) (*AuthRepo, error) {
	conn, err := dialRPC(endpoint)
	if err != nil {
		return nil, err
	}
	return &AuthRepo{conn: conn, client: v1.NewAuthServiceClient(conn)}, nil
}

func (r *AuthRepo) Login(ctx context.Context, req *v1.LoginRequest) (*v1.LoginReply, error) {
	return r.client.Login(ctx, req)
}

func (r *AuthRepo) Logout(ctx context.Context) (*v1.MessageReply, error) {
	return r.client.Logout(ctx, &v1.LogoutRequest{})
}

func (r *AuthRepo) RefreshToken(ctx context.Context, req *v1.RefreshTokenRequest) (*v1.RefreshTokenReply, error) {
	return r.client.RefreshToken(ctx, req)
}

func (r *AuthRepo) Me(ctx context.Context) (*v1.MeReply, error) {
	return r.client.Me(ctx, &v1.MeRequest{})
}

func (r *AuthRepo) ValidateAccessToken(ctx context.Context, token string) (*v1.ValidateAccessTokenReply, error) {
	return r.client.ValidateAccessToken(ctx, &v1.ValidateAccessTokenRequest{AccessToken: strings.TrimSpace(token)})
}

func (r *AuthRepo) CheckPermission(ctx context.Context, userID int64, resource, action string) (*v1.CheckPermissionReply, error) {
	return r.client.CheckPermission(ctx, &v1.CheckPermissionRequest{UserId: userID, Resource: resource, Action: action})
}

func (r *AuthRepo) Close() error {
	if r == nil || r.conn == nil {
		return nil
	}
	return r.conn.Close()
}
