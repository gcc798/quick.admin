package captcha

import (
	"context"

	"github.com/gcc798/nai-tizi/application/sys-api/internal/logic/commonutil"
	"github.com/gcc798/nai-tizi/application/sys-api/internal/svc"
	"github.com/gcc798/nai-tizi/application/sys-api/internal/types"
	"github.com/gcc798/nai-tizi/application/sys-rpc/client/sysservice"
	"github.com/zeromicro/go-zero/core/logx"
)

type ResourceSmsCodeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewResourceSmsCodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ResourceSmsCodeLogic {
	return &ResourceSmsCodeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ResourceSmsCodeLogic) ResourceSmsCode(req *types.ResourceSmsCodeReq) (resp *types.CommonResp, err error) {
	data, err := l.svcCtx.SysRpcClient.CaptchaSms(l.ctx, &sysservice.CaptchaPhoneReq{Phonenumber: req.Phone})
	if err != nil {
		return &types.CommonResp{Code: 500, Msg: err.Error()}, nil
	}
	return &types.CommonResp{Code: 200, Msg: "操作成功", Data: map[string]interface{}{
		"id":       data.Id,
		"type":     data.Type,
		"data":     commonutil.JSONStringToValue(data.DataJson),
		"expireAt": data.ExpireAt,
	}}, nil
}
