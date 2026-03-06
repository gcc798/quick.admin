// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package captcha

import (
	"context"

	"github.com/force-c/nai-tizi/application/sys-api/internal/svc"
	"github.com/force-c/nai-tizi/application/sys-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CaptchaEnabledTypesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCaptchaEnabledTypesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CaptchaEnabledTypesLogic {
	return &CaptchaEnabledTypesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CaptchaEnabledTypesLogic) CaptchaEnabledTypes() (resp *types.CommonResp, err error) {
	typesList := make([]string, 0, 3)
	if l.svcCtx.Config.Captcha.Image.Enabled {
		typesList = append(typesList, "image")
	}
	if l.svcCtx.Config.Captcha.Sms.Enabled {
		typesList = append(typesList, "sms")
	}
	if l.svcCtx.Config.Captcha.Email.Enabled {
		typesList = append(typesList, "email")
	}

	return &types.CommonResp{
		Code: 200,
		Msg:  "success",
		Data: typesList,
	}, nil
}
