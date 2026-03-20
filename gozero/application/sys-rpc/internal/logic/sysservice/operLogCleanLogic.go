package sysservicelogic

import (
	"context"

	"github.com/force-c/nai-tizi/application/sys-rpc/internal/svc"
	"github.com/force-c/nai-tizi/application/sys-rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type OperLogCleanLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewOperLogCleanLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OperLogCleanLogic {
	return &OperLogCleanLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *OperLogCleanLogic) OperLogClean(in *pb.LogCleanReq) (*pb.Ack, error) {
	days := in.Days
	if days <= 0 {
		days = 30
	}
	if _, err := l.svcCtx.DB.ExecCtx(l.ctx, `delete from public.s_oper_log where oper_time < now() - ($1 || ' day')::interval`, days); err != nil {
		return nil, err
	}
	return &pb.Ack{Msg: "ok"}, nil
}
