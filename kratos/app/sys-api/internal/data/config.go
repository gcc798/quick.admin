package data

import (
	"context"
	v1 "github.com/force-c/nai-tizi/kratos/api/system/v1"
	"google.golang.org/grpc"
)

type ConfigRepo struct {
	conn   *grpc.ClientConn
	client v1.ConfigServiceClient
}

func NewConfigRepo(endpoint string) (*ConfigRepo, error) {
	conn, err := dialRPC(endpoint)
	if err != nil {
		return nil, err
	}
	return &ConfigRepo{conn: conn, client: v1.NewConfigServiceClient(conn)}, nil
}
func (r *ConfigRepo) Create(ctx context.Context, item *v1.ConfigItem) (*v1.MessageReply, error) {
	return r.client.CreateConfig(ctx, item)
}
func (r *ConfigRepo) Page(ctx context.Context, req *v1.PageConfigRequest) (*v1.PageConfigReply, error) {
	return r.client.PageConfig(ctx, req)
}
func (r *ConfigRepo) BatchDelete(ctx context.Context, ids []int64) (*v1.MessageReply, error) {
	return r.client.BatchDeleteConfig(ctx, &v1.BatchIdsRequest{Ids: ids})
}
func (r *ConfigRepo) ByCode(ctx context.Context, code string) (*v1.ConfigListReply, error) {
	return r.client.GetConfigByCode(ctx, &v1.GetConfigByCodeRequest{Code: code})
}
func (r *ConfigRepo) Data(ctx context.Context, code string) (*v1.ConfigDataReply, error) {
	return r.client.GetConfigDataByCode(ctx, &v1.GetConfigByCodeRequest{Code: code})
}
func (r *ConfigRepo) Update(ctx context.Context, req *v1.UpdateConfigRequest) (*v1.MessageReply, error) {
	return r.client.UpdateConfig(ctx, req)
}
func (r *ConfigRepo) Get(ctx context.Context, id int64) (*v1.ConfigItem, error) {
	return r.client.GetConfigById(ctx, &v1.IdRequest{Id: id})
}
func (r *ConfigRepo) Delete(ctx context.Context, id int64) (*v1.MessageReply, error) {
	return r.client.DeleteConfig(ctx, &v1.IdRequest{Id: id})
}
func (r *ConfigRepo) Close() error {
	if r == nil || r.conn == nil {
		return nil
	}
	return r.conn.Close()
}
