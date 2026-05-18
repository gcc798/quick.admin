package sysservicelogic

import (
	"context"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gcc798/nai-tizi/application/sys-rpc/internal/svc"
	"github.com/gcc798/nai-tizi/application/sys-rpc/pb"
	gzsqlx "github.com/zeromicro/go-zero/core/stores/sqlx"
)

type attachmentRow struct {
	AttachmentId  int64          `db:"id"`
	EnvId         int64          `db:"env_id"`
	FileName      string         `db:"file_name"`
	FileKey       string         `db:"file_key"`
	FileSize      int64          `db:"file_size"`
	FileType      sql.NullString `db:"file_type"`
	FileExt       sql.NullString `db:"file_ext"`
	BusinessType  sql.NullString `db:"business_type"`
	BusinessId    sql.NullString `db:"business_id"`
	BusinessField sql.NullString `db:"business_field"`
	IsPublic      bool           `db:"is_public"`
	AccessUrl     sql.NullString `db:"access_url"`
	Metadata      sql.NullString `db:"metadata"`
	Status        int64          `db:"status"`
	ExpireTime    sql.NullTime   `db:"expire_time"`
	CreateBy      sql.NullInt64  `db:"create_by"`
	CreateTime    sql.NullTime   `db:"create_time"`
	UpdateTime    sql.NullTime   `db:"update_time"`
}

func getAttachmentByID(ctx context.Context, svcCtx *svc.ServiceContext, id int64) (*attachmentRow, error) {
	var row attachmentRow
	err := svcCtx.DB.QueryRowCtx(ctx, &row, `
		select id, env_id, file_name, file_key, file_size, file_type, file_ext, business_type, business_id, business_field, is_public, access_url, metadata, status, expire_time, create_by, create_time, update_time
		from public.biz_attachment
		where id = $1 and status = 0 and deleted_at is null
		limit 1
	`, id)
	if err != nil {
		if errors.Is(err, gzsqlx.ErrNotFound) {
			return nil, fmt.Errorf("附件不存在")
		}
		return nil, err
	}
	return &row, nil
}

func toAttachmentPB(row attachmentRow) *pb.Attachment {
	return &pb.Attachment{
		AttachmentId:  row.AttachmentId,
		EnvId:         row.EnvId,
		FileName:      row.FileName,
		FileKey:       row.FileKey,
		FileSize:      row.FileSize,
		FileType:      nullString(row.FileType),
		FileExt:       nullString(row.FileExt),
		BusinessType:  nullString(row.BusinessType),
		BusinessId:    nullString(row.BusinessId),
		BusinessField: nullString(row.BusinessField),
		IsPublic:      row.IsPublic,
		AccessUrl:     nullString(row.AccessUrl),
		MetadataJson:  nullString(row.Metadata),
		Status:        int32(row.Status),
		ExpireTime:    nullTime(row.ExpireTime),
		CreateBy:      nullInt64(row.CreateBy),
		CreateTime:    nullTime(row.CreateTime),
		UpdateTime:    nullTime(row.UpdateTime),
	}
}

func toAttachmentList(rows []attachmentRow) []*pb.Attachment {
	list := make([]*pb.Attachment, 0, len(rows))
	for _, row := range rows {
		list = append(list, toAttachmentPB(row))
	}
	return list
}

func buildAttachmentInt64In(ids []int64, start int) (string, []interface{}) {
	parts := make([]string, 0, len(ids))
	args := make([]interface{}, 0, len(ids))
	for i, id := range ids {
		parts = append(parts, fmt.Sprintf("$%d", start+i))
		args = append(args, id)
	}
	return strings.Join(parts, ", "), args
}

func attachmentRootDir() string {
	return filepath.Join("runtime", "attachments")
}

func ensureAttachmentDir() error {
	return os.MkdirAll(attachmentRootDir(), 0o755)
}

func buildAttachmentFilePath(id int64, fileName string) string {
	base := filepath.Base(fileName)
	base = strings.ReplaceAll(base, " ", "_")
	return filepath.Join(attachmentRootDir(), fmt.Sprintf("%d_%d_%s", id, time.Now().UnixNano(), base))
}

func readAttachmentContent(row *attachmentRow) ([]byte, string, error) {
	if row.FileKey == "" {
		return nil, "", fmt.Errorf("附件文件不存在")
	}
	content, err := os.ReadFile(row.FileKey)
	if err != nil {
		return nil, "", err
	}
	contentType := nullString(row.FileType)
	if contentType == "" {
		contentType = http.DetectContentType(content)
	}
	return content, contentType, nil
}

func buildAttachmentDataURL(content []byte, contentType string) string {
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	return "data:" + contentType + ";base64," + base64.StdEncoding.EncodeToString(content)
}
