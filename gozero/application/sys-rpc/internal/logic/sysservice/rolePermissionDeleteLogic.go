package sysservicelogic

import (
	"context"

	"github.com/force-c/nai-tizi/application/sys-rpc/internal/svc"
	"github.com/force-c/nai-tizi/application/sys-rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type RolePermissionDeleteLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRolePermissionDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RolePermissionDeleteLogic {
	return &RolePermissionDeleteLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RolePermissionDeleteLogic) RolePermissionDelete(in *pb.RolePermissionReq) (*pb.Ack, error) {
	// todo: add your logic here and delete this line

	return &pb.Ack{}, nil
}
