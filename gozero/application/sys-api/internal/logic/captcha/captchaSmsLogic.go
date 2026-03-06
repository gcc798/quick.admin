// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package captcha

import (
	"context"

	"github.com/force-c/nai-tizi/application/sys-api/internal/svc"
	"github.com/force-c/nai-tizi/application/sys-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CaptchaSmsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCaptchaSmsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CaptchaSmsLogic {
	return &CaptchaSmsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CaptchaSmsLogic) CaptchaSms(req *types.CaptchaPhoneReq) (resp *types.CommonResp, err error) {
	return &types.CommonResp{
		Code: 500,
		Msg:  "gozero logic not implemented yet",
	}, nil
}
