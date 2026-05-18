package sysservicelogic

import (
	"context"
	"fmt"
	"strings"

	"github.com/gcc798/nai-tizi/application/sys-rpc/internal/svc"
	"github.com/gcc798/nai-tizi/application/sys-rpc/pb"
	"github.com/zeromicro/go-zero/core/logx"
)

type UserBatchDeleteLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUserBatchDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserBatchDeleteLogic {
	return &UserBatchDeleteLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}
func (l *UserBatchDeleteLogic) UserBatchDelete(in *pb.BatchIdsReq) (*pb.Ack, error) {
	if len(in.Ids) == 0 {
		return &pb.Ack{Msg: "ok"}, nil
	}
	parts := make([]string, 0, len(in.Ids))
	args := make([]interface{}, 0, len(in.Ids))
	for i, id := range in.Ids {
		parts = append(parts, fmt.Sprintf("$%d", i+1))
		args = append(args, id)
	}
	if _, err := l.svcCtx.DB.ExecCtx(l.ctx, `update public.s_user set deleted_at = now() where id in (`+strings.Join(parts, ",")+`) and deleted_at is null`, args...); err != nil {
		return nil, err
	}
	return &pb.Ack{Msg: "ok"}, nil
}
