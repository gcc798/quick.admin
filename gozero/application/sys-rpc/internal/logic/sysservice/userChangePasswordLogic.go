package sysservicelogic

import (
	"context"

	"github.com/force-c/nai-tizi/application/sys-rpc/internal/svc"
	"github.com/force-c/nai-tizi/application/sys-rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type UserChangePasswordLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUserChangePasswordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserChangePasswordLogic {
	return &UserChangePasswordLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UserChangePasswordLogic) UserChangePassword(in *pb.UserChangePasswordReq) (*pb.Ack, error) {
	// todo: add your logic here and delete this line

	return &pb.Ack{}, nil
}
