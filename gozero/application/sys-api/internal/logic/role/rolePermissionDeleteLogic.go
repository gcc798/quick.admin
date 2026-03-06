package role

import (
	"context"

	"github.com/force-c/nai-tizi/application/sys-api/internal/svc"
	"github.com/force-c/nai-tizi/application/sys-api/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type RolePermissionDeleteLogic struct { logx.Logger; ctx context.Context; svcCtx *svc.ServiceContext }
func NewRolePermissionDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RolePermissionDeleteLogic { return &RolePermissionDeleteLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx} }
func (l *RolePermissionDeleteLogic) RolePermissionDelete(req *types.RolePermissionDeleteReq) (resp *types.CommonResp, err error) {
	if err := l.svcCtx.Redis.SRem(l.ctx, "casbin:role:"+req.RoleKey+":permissions", req.Resource+"::"+req.Action).Err(); err != nil { return &types.CommonResp{Code: 500, Msg: err.Error()}, nil }
	return &types.CommonResp{Code: 200, Msg: "success", Data: "ok"}, nil
}
