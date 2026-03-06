// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user

import (
	"context"

	"github.com/force-c/nai-tizi/application/sys-api/internal/svc"
	"github.com/force-c/nai-tizi/application/sys-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UserResetPasswordLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserResetPasswordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserResetPasswordLogic {
	return &UserResetPasswordLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserResetPasswordLogic) UserResetPassword(req *types.UserPasswordPathReq) (resp *types.CommonResp, err error) {
	return &types.CommonResp{
		Code: 500,
		Msg:  "gozero logic not implemented yet",
	}, nil
}
