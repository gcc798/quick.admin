package sysservicelogic

import (
	"context"

	"github.com/force-c/nai-tizi/application/sys-rpc/internal/svc"
	"github.com/force-c/nai-tizi/application/sys-rpc/pb"
	"github.com/zeromicro/go-zero/core/logx"
)

type UserDeleteLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUserDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserDeleteLogic {
	return &UserDeleteLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}
func (l *UserDeleteLogic) UserDelete(in *pb.IdReq) (*pb.Ack, error) {
	if _, err := l.svcCtx.DB.ExecCtx(l.ctx, `update public.s_user set deleted_at = now() where id = $1 and deleted_at is null`, in.Id); err != nil {
		return nil, err
	}
	return &pb.Ack{Msg: "ok"}, nil
}
