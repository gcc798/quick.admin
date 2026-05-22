package user

import (
	"context"

	"github.com/gcc798/nai-tizi/application/sys-api/internal/logic/commonutil"
	"github.com/gcc798/nai-tizi/application/sys-api/internal/svc"
	"github.com/gcc798/nai-tizi/application/sys-api/internal/types"
	"github.com/gcc798/nai-tizi/application/sys-rpc/client/sysservice"
	"github.com/zeromicro/go-zero/core/logx"
)

type XcxGetInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewXcxGetInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *XcxGetInfoLogic {
	return &XcxGetInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *XcxGetInfoLogic) XcxGetInfo() (resp *types.CommonResp, err error) {
	userID := commonutil.UserIDFromContext(l.ctx)
	user, err := l.svcCtx.SysRpcClient.UserDetail(l.ctx, &sysservice.IdReq{Id: userID})
	if err != nil {
		return &types.CommonResp{Code: 500, Msg: "没有权限访问用户数据"}, nil
	}
	return &types.CommonResp{Code: 200, Msg: "操作成功", Data: user}, nil
}
