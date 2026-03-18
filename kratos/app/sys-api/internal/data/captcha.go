package data

import (
	"context"
	v1 "github.com/force-c/nai-tizi/kratos/api/system/v1"
	"google.golang.org/grpc"
)

type CaptchaRepo struct {
	conn   *grpc.ClientConn
	client v1.CaptchaServiceClient
}

func NewCaptchaRepo(endpoint string) (*CaptchaRepo, error) {
	conn, err := dialRPC(endpoint)
	if err != nil {
		return nil, err
	}
	return &CaptchaRepo{conn: conn, client: v1.NewCaptchaServiceClient(conn)}, nil
}
func (r *CaptchaRepo) Image(ctx context.Context) (*v1.CaptchaReply, error) {
	return r.client.GenerateImageCaptcha(ctx, &v1.GenerateImageCaptchaRequest{})
}
func (r *CaptchaRepo) SMS(ctx context.Context, phone string) (*v1.CaptchaReply, error) {
	return r.client.SendSMSCaptcha(ctx, &v1.SendSMSCaptchaRequest{Phonenumber: phone, Phone: phone})
}
func (r *CaptchaRepo) Email(ctx context.Context, email string) (*v1.CaptchaReply, error) {
	return r.client.SendEmailCaptcha(ctx, &v1.SendEmailCaptchaRequest{Email: email})
}
func (r *CaptchaRepo) Enabled(ctx context.Context) (*v1.GetEnabledTypesReply, error) {
	return r.client.GetEnabledTypes(ctx, &v1.GetEnabledTypesRequest{})
}
func (r *CaptchaRepo) Close() error {
	if r == nil || r.conn == nil {
		return nil
	}
	return r.conn.Close()
}
