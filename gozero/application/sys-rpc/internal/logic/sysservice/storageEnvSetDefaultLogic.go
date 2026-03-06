package sysservicelogic

import (
	"context"

	"github.com/force-c/nai-tizi/application/sys-rpc/internal/svc"
	"github.com/force-c/nai-tizi/application/sys-rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type StorageEnvSetDefaultLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewStorageEnvSetDefaultLogic(ctx context.Context, svcCtx *svc.ServiceContext) *StorageEnvSetDefaultLogic {
	return &StorageEnvSetDefaultLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *StorageEnvSetDefaultLogic) StorageEnvSetDefault(in *pb.StorageEnvDefaultReq) (*pb.Ack, error) {
	if _, err := l.svcCtx.DB.ExecCtx(l.ctx, `update public.s_storage_env set is_default = false where is_default = true and deleted_at is null`); err != nil {
		return nil, err
	}
	if _, err := l.svcCtx.DB.ExecCtx(l.ctx, `update public.s_storage_env set is_default = true, updated_time = now() where id = $1 and deleted_at is null`, in.Id); err != nil {
		return nil, err
	}
	return &pb.Ack{Msg: "ok"}, nil
}
