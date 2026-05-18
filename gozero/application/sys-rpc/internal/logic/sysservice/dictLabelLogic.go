package sysservicelogic

import (
	"context"
	"database/sql"

	"github.com/gcc798/nai-tizi/application/sys-rpc/internal/svc"
	"github.com/gcc798/nai-tizi/application/sys-rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type DictLabelLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDictLabelLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DictLabelLogic {
	return &DictLabelLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DictLabelLogic) DictLabel(in *pb.DictLabelQueryReq) (*pb.DictLabelResp, error) {
	var label sql.NullString
	if err := l.svcCtx.DB.QueryRowCtx(l.ctx, &label, `select dict_label from public.s_dict_data where dict_type = $1 and dict_value = $2 and deleted_at is null order by sort asc, id asc limit 1`, in.DictType, in.DictValue); err != nil {
		return nil, err
	}
	return &pb.DictLabelResp{Label: label.String}, nil
}
