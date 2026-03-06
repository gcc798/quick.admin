package sysservicelogic

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/force-c/nai-tizi/application/sys-rpc/internal/svc"
	"github.com/force-c/nai-tizi/application/sys-rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type AttachmentUploadFileLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAttachmentUploadFileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AttachmentUploadFileLogic {
	return &AttachmentUploadFileLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *AttachmentUploadFileLogic) AttachmentUploadFile(in *pb.AttachmentUploadReq) (*pb.Attachment, error) {
	if in.FileName == "" || len(in.Content) == 0 {
		return nil, errors.New("文件不能为空")
	}
	var env *storageEnvRow
	var err error
	if in.EnvCode != "" {
		env, err = getStorageEnvByCode(l.ctx, l.svcCtx, in.EnvCode)
	} else {
		env, err = getDefaultStorageEnv(l.ctx, l.svcCtx)
	}
	if err != nil {
		return nil, err
	}
	if err := ensureAttachmentDir(); err != nil {
		return nil, err
	}
	var attachmentId int64
	fileExt := strings.TrimPrefix(strings.ToLower(filepath.Ext(in.FileName)), ".")
	err = l.svcCtx.DB.QueryRowCtx(l.ctx, &attachmentId, `
		insert into public.biz_attachment (env_id, file_name, file_key, file_size, file_type, file_ext, status, create_by, create_time, update_time)
		values ($1, $2, '', $3, $4, $5, 0, 0, now(), now())
		returning id
	`, env.Id, in.FileName, len(in.Content), in.ContentType, fileExt)
	if err != nil {
		return nil, err
	}
	filePath := buildAttachmentFilePath(attachmentId, in.FileName)
	if err := os.WriteFile(filePath, in.Content, 0o644); err != nil {
		return nil, err
	}
	if _, err := l.svcCtx.DB.ExecCtx(l.ctx, `update public.biz_attachment set file_key = $2 where id = $1`, attachmentId, filePath); err != nil {
		return nil, err
	}
	row, err := getAttachmentByID(l.ctx, l.svcCtx, attachmentId)
	if err != nil {
		return nil, err
	}
	return toAttachmentPB(*row), nil
}
