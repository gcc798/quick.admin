// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package auth

import (
	"context"

	"github.com/force-c/nai-tizi/application/sys-api/internal/svc"
	"github.com/force-c/nai-tizi/application/sys-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type MeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MeLogic {
	return &MeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MeLogic) Me(token string) (resp *types.CommonResp, err error) {
	userId, err := userIDFromToken(l.svcCtx, token)
	if err != nil {
		return &types.CommonResp{
			Code: 401,
			Msg:  err.Error(),
		}, nil
	}

	return &types.CommonResp{
		Code: 200,
		Msg:  "success",
		Data: map[string]interface{}{
			"userId": userId,
		},
	}, nil
}
