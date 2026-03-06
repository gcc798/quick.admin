package role

import (
	"context"

	"github.com/force-c/nai-tizi/application/sys-api/internal/svc"
	"github.com/force-c/nai-tizi/application/sys-api/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type RolePermissionAddLogic struct { logx.Logger; ctx context.Context; svcCtx *svc.ServiceContext }
func NewRolePermissionAddLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RolePermissionAddLogic { return &RolePermissionAddLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx} }
func (l *RolePermissionAddLogic) RolePermissionAdd(req *types.RolePermissionReq) (resp *types.CommonResp, err error) {
	if err := l.svcCtx.Redis.SAdd(l.ctx, "casbin:role:"+req.RoleKey+":permissions", req.Resource+"::"+req.Action).Err(); err != nil { return &types.CommonResp{Code: 500, Msg: err.Error()}, nil }
	return &types.CommonResp{Code: 200, Msg: "success", Data: "ok"}, nil
}
