package sysservicelogic

import (
	"context"
	"fmt"

	"github.com/force-c/nai-tizi/application/sys-rpc/internal/svc"
	"github.com/force-c/nai-tizi/application/sys-rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type DictDetailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDictDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DictDetailLogic {
	return &DictDetailLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DictDetailLogic) DictDetail(in *pb.StringIdReq) (*pb.Dict, error) {
	id, err := parseDictID(in.Id)
	if err != nil {
		return nil, fmt.Errorf("无效的字典ID")
	}
	row, err := getDictByID(l.ctx, l.svcCtx, id)
	if err != nil {
		return nil, err
	}
	return toDictPB(*row), nil
}
