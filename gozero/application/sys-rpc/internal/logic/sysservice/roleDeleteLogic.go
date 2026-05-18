package sysservicelogic

import (
	"context"
	"errors"

	"github.com/gcc798/nai-tizi/application/sys-rpc/internal/svc"
	"github.com/gcc798/nai-tizi/application/sys-rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type RoleDeleteLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRoleDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RoleDeleteLogic {
	return &RoleDeleteLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RoleDeleteLogic) RoleDelete(in *pb.IdReq) (*pb.Ack, error) {
	row, err := getRoleByID(l.ctx, l.svcCtx, in.Id)
	if err != nil {
		return nil, err
	}
	if row.IsSystem {
		return nil, errors.New("系统角色不可删除")
	}
	var count int64
	if err := l.svcCtx.DB.QueryRowCtx(l.ctx, &count, `select count(1) from public.m_user_role where role_id = $1 and deleted_at is null`, in.Id); err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, errors.New("该角色已被用户使用，无法删除")
	}
	if _, err := l.svcCtx.DB.ExecCtx(l.ctx, `update public.m_role_menu set deleted_at = now() where role_id = $1 and deleted_at is null`, in.Id); err != nil {
		return nil, err
	}
	if _, err := l.svcCtx.DB.ExecCtx(l.ctx, `update public.s_role set deleted_at = now() where id = $1 and deleted_at is null`, in.Id); err != nil {
		return nil, err
	}
	return &pb.Ack{Msg: "ok"}, nil
}
