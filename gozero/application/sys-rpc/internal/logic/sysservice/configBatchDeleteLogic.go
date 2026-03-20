package sysservicelogic

import (
	"context"
	"errors"
	"fmt"

	"github.com/force-c/nai-tizi/application/sys-rpc/internal/svc"
	"github.com/force-c/nai-tizi/application/sys-rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type ConfigBatchDeleteLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewConfigBatchDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ConfigBatchDeleteLogic {
	return &ConfigBatchDeleteLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ConfigBatchDeleteLogic) ConfigBatchDelete(in *pb.BatchIdsReq) (*pb.Ack, error) {
	if len(in.Ids) == 0 {
		return nil, errors.New("ids 不能为空")
	}
	placeholders, args := buildConfigInt64In(in.Ids, 1)
	query := fmt.Sprintf(`update public.s_config set deleted_at = now() where id in (%s) and deleted_at is null`, placeholders)
	if _, err := l.svcCtx.DB.ExecCtx(l.ctx, query, args...); err != nil {
		return nil, err
	}
	return &pb.Ack{Msg: "ok"}, nil
}
