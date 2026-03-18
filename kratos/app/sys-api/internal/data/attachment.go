package data

import (
	"context"
	v1 "github.com/force-c/nai-tizi/kratos/api/system/v1"
	"google.golang.org/grpc"
)

type AttachmentRepo struct {
	conn   *grpc.ClientConn
	client v1.AttachmentServiceClient
}

func NewAttachmentRepo(endpoint string) (*AttachmentRepo, error) {
	conn, err := dialRPC(endpoint)
	if err != nil {
		return nil, err
	}
	return &AttachmentRepo{conn: conn, client: v1.NewAttachmentServiceClient(conn)}, nil
}
func (r *AttachmentRepo) Close() error {
	if r == nil || r.conn == nil {
		return nil
	}
	return r.conn.Close()
}
func (r *AttachmentRepo) Upload(ctx context.Context, req *v1.UploadFileRequest) (*v1.AttachmentItem, error) {
	return r.client.UploadFile(ctx, req)
}
func (r *AttachmentRepo) Bind(ctx context.Context, req *v1.BindAttachmentRequest) (*v1.MessageReply, error) {
	return r.client.BindAttachmentToBusiness(ctx, req)
}
func (r *AttachmentRepo) Get(ctx context.Context, id int64) (*v1.AttachmentItem, error) {
	return r.client.GetAttachment(ctx, &v1.AttachmentIdRequest{AttachmentId: id})
}
func (r *AttachmentRepo) List(ctx context.Context, bt, bid string) (*v1.AttachmentListReply, error) {
	return r.client.ListAttachmentsByBusiness(ctx, &v1.ListAttachmentsByBusinessRequest{BusinessType: bt, BusinessId: bid})
}
func (r *AttachmentRepo) Page(ctx context.Context, req *v1.PageAttachmentsRequest) (*v1.PageAttachmentsReply, error) {
	return r.client.PageAttachments(ctx, req)
}
func (r *AttachmentRepo) Download(ctx context.Context, id int64) (*v1.AttachmentDownloadReply, error) {
	return r.client.DownloadAttachment(ctx, &v1.AttachmentIdRequest{AttachmentId: id})
}
func (r *AttachmentRepo) URL(ctx context.Context, id int64, expires int64) (*v1.AttachmentURLReply, error) {
	return r.client.GetAttachmentURL(ctx, &v1.AttachmentURLRequest{AttachmentId: id, Expires: expires})
}
func (r *AttachmentRepo) Delete(ctx context.Context, id int64) (*v1.MessageReply, error) {
	return r.client.DeleteAttachment(ctx, &v1.AttachmentIdRequest{AttachmentId: id})
}
