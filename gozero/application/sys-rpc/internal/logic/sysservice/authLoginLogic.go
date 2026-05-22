package sysservicelogic

import (
	"context"
	"errors"

	"github.com/gcc798/nai-tizi/application/sys-rpc/internal/svc"
	"github.com/gcc798/nai-tizi/application/sys-rpc/pb"
	"github.com/zeromicro/go-zero/core/logx"
)

type AuthLoginLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAuthLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AuthLoginLogic {
	return &AuthLoginLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *AuthLoginLogic) AuthLogin(in *pb.AuthLoginReq) (*pb.AuthLoginResp, error) {
	client, err := authenticateClient(l.ctx, l.svcCtx, in.ClientKey, in.ClientSecret, in.GrantType)
	if err != nil {
		return nil, err
	}
	if in.GrantType != "password" {
		return nil, errUnsupportedGrantType()
	}
	if err := verifyImageCaptcha(l.ctx, l.svcCtx, in.Uuid, in.Code); err != nil {
		return nil, err
	}
	user, err := authenticatePassword(l.ctx, l.svcCtx, in.Username, in.Password)
	if err != nil {
		return nil, err
	}
	return &pb.AuthLoginResp{
		ClientId:      client.ClientId,
		ClientKey:     client.ClientKey,
		DeviceType:    nullString(client.DeviceType),
		Timeout:       client.Timeout,
		ActiveTimeout: client.ActiveTimeout,
		UserInfo: &pb.UserInfo{
			UserId:      user.Id,
			Username:    user.UserName,
			Nickname:    nullString(user.NickName),
			Phonenumber: nullString(user.Phonenumber),
			Email:       nullString(user.Email),
			Avatar:      nullString(user.Avatar),
			UserType:    int32(user.UserType),
			OrgId:       nullInt64(user.OrgId),
		},
	}, nil
}

func errUnsupportedGrantType() error { return errors.New("当前仅实现 password 登录") }
