package sysservicelogic

import (
	"context"
	"database/sql"

	"github.com/force-c/nai-tizi/application/sys-rpc/internal/svc"
	"github.com/force-c/nai-tizi/application/sys-rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type AttachmentBindLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAttachmentBindLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AttachmentBindLogic {
	return &AttachmentBindLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *AttachmentBindLogic) AttachmentBind(in *pb.AttachmentBindReq) (*pb.Ack, error) {
	if _, err := getAttachmentByID(l.ctx, l.svcCtx, in.AttachmentId); err != nil {
		return nil, err
	}
	accessURL := ""
	if in.IsPublic {
		row, err := getAttachmentByID(l.ctx, l.svcCtx, in.AttachmentId)
		if err != nil {
			return nil, err
		}
		content, contentType, err := readAttachmentContent(row)
		if err == nil {
			accessURL = buildAttachmentDataURL(content, contentType)
		}
	}
	if _, err := l.svcCtx.DB.ExecCtx(l.ctx, `
		update public.biz_attachment
		set business_type = $2, business_id = $3, business_field = $4, is_public = $5, access_url = $6, metadata = $7, expire_time = nullif($8, '')::timestamp, update_time = now()
		where id = $1 and deleted_at is null
	`, in.AttachmentId, in.BusinessType, in.BusinessId, in.BusinessField, in.IsPublic, sql.NullString{String: accessURL, Valid: accessURL != ""}, sql.NullString{String: in.MetadataJson, Valid: in.MetadataJson != ""}, in.ExpireTime); err != nil {
		return nil, err
	}
	return &pb.Ack{Msg: "ok"}, nil
}
