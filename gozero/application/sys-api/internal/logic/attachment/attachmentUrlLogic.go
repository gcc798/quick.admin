// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package attachment

import (
	"context"

	"github.com/gcc798/nai-tizi/application/sys-api/internal/svc"
	"github.com/gcc798/nai-tizi/application/sys-api/internal/types"
	"github.com/gcc798/nai-tizi/application/sys-rpc/client/sysservice"

	"github.com/zeromicro/go-zero/core/logx"
)

type AttachmentUrlLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAttachmentUrlLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AttachmentUrlLogic {
	return &AttachmentUrlLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AttachmentUrlLogic) AttachmentUrl(req *types.AttachmentUrlQueryReq) (resp *types.CommonResp, err error) {
	data, err := l.svcCtx.SysRpcClient.AttachmentUrl(l.ctx, &sysservice.AttachmentUrlQueryReq{
		AttachmentId: req.AttachmentId,
		Expires:      req.Expires,
	})
	if err != nil {
		return &types.CommonResp{Code: 500, Msg: err.Error()}, nil
	}
	return &types.CommonResp{Code: 200, Msg: "success", Data: data}, nil
}
