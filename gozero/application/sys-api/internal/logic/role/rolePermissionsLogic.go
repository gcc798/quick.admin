package role

import (
	"context"
	"strings"

	"github.com/force-c/nai-tizi/application/sys-api/internal/svc"
	"github.com/force-c/nai-tizi/application/sys-api/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type RolePermissionsLogic struct { logx.Logger; ctx context.Context; svcCtx *svc.ServiceContext }
func NewRolePermissionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RolePermissionsLogic { return &RolePermissionsLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx} }
func (l *RolePermissionsLogic) RolePermissions(req *types.RolePermissionsQueryReq) (resp *types.CommonResp, err error) {
	rows, err := l.svcCtx.Redis.SMembers(l.ctx, "casbin:role:"+req.RoleKey+":permissions").Result()
	if err == nil && len(rows) > 0 {
		data := make([][]string, 0, len(rows))
		for _, row := range rows { data = append(data, strings.Split(row, "::")) }
		return &types.CommonResp{Code: 200, Msg: "success", Data: data}, nil
	}
	return &types.CommonResp{Code: 200, Msg: "success", Data: make([][]string, 0)}, nil
}
