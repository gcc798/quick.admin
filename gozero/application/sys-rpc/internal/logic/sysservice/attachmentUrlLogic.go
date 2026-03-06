package sysservicelogic

import (
	"context"

	"github.com/force-c/nai-tizi/application/sys-rpc/internal/svc"
	"github.com/force-c/nai-tizi/application/sys-rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type AttachmentUrlLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAttachmentUrlLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AttachmentUrlLogic {
	return &AttachmentUrlLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *AttachmentUrlLogic) AttachmentUrl(in *pb.AttachmentUrlQueryReq) (*pb.AttachmentUrlResp, error) {
	row, err := getAttachmentByID(l.ctx, l.svcCtx, in.AttachmentId)
	if err != nil {
		return nil, err
	}
	url := ""
	if row.AccessUrl.Valid && row.AccessUrl.String != "" {
		url = row.AccessUrl.String
	} else {
		content, contentType, err := readAttachmentContent(row)
		if err == nil {
			url = buildAttachmentDataURL(content, contentType)
		} else {
			url = row.FileKey
		}
	}
	expires := in.Expires
	if expires <= 0 {
		expires = 3600
	}
	return &pb.AttachmentUrlResp{AttachmentId: in.AttachmentId, Url: url, Expires: expires}, nil
}
