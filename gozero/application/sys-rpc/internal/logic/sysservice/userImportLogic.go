package sysservicelogic

import (
	"context"

	"github.com/gcc798/nai-tizi/application/sys-rpc/internal/svc"
	"github.com/gcc798/nai-tizi/application/sys-rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type UserImportLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUserImportLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserImportLogic {
	return &UserImportLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UserImportLogic) UserImport(in *pb.UserImportReq) (*pb.Ack, error) {
	for _, user := range in.Users {
		if _, err := NewUserCreateLogic(l.ctx, l.svcCtx).UserCreate(user); err != nil {
			return nil, err
		}
	}
	return &pb.Ack{Msg: "ok"}, nil
}
