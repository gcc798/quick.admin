package sysservicelogic

import (
	"context"

	"github.com/force-c/nai-tizi/application/sys-rpc/internal/svc"
	"github.com/force-c/nai-tizi/application/sys-rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type AttachmentDeleteLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAttachmentDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AttachmentDeleteLogic {
	return &AttachmentDeleteLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *AttachmentDeleteLogic) AttachmentDelete(in *pb.IdReq) (*pb.Ack, error) {
	if _, err := getAttachmentByID(l.ctx, l.svcCtx, in.Id); err != nil {
		return nil, err
	}
	if _, err := l.svcCtx.DB.ExecCtx(l.ctx, `update public.biz_attachment set status = 1, update_time = now(), deleted_at = now() where id = $1 and deleted_at is null`, in.Id); err != nil {
		return nil, err
	}
	return &pb.Ack{Msg: "ok"}, nil
}
