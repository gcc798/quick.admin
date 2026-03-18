package data

import (
	"context"

	v1 "github.com/force-c/nai-tizi/kratos/api/system/v1"
	"google.golang.org/grpc"
)

type LoginLogRepo struct {
	conn   *grpc.ClientConn
	client v1.LoginLogServiceClient
}

type OperLogRepo struct {
	conn   *grpc.ClientConn
	client v1.OperLogServiceClient
}

func NewLoginLogRepo(endpoint string) (*LoginLogRepo, error) {
	conn, err := dialRPC(endpoint)
	if err != nil {
		return nil, err
	}
	return &LoginLogRepo{conn: conn, client: v1.NewLoginLogServiceClient(conn)}, nil
}

func NewOperLogRepo(endpoint string) (*OperLogRepo, error) {
	conn, err := dialRPC(endpoint)
	if err != nil {
		return nil, err
	}
	return &OperLogRepo{conn: conn, client: v1.NewOperLogServiceClient(conn)}, nil
}

func (r *LoginLogRepo) Create(ctx context.Context, item *v1.CreateLoginLogRequest) (*v1.MessageReply, error) {
	return r.client.CreateLoginLog(ctx, item)
}

func (r *LoginLogRepo) Page(ctx context.Context, req *v1.PageLoginLogRequest) (*v1.PageLogReply, error) {
	return r.client.PageLoginLog(ctx, req)
}

func (r *LoginLogRepo) BatchDelete(ctx context.Context, ids []int64) (*v1.MessageReply, error) {
	return r.client.BatchDeleteLoginLog(ctx, &v1.LogBatchIdsRequest{Ids: ids})
}

func (r *LoginLogRepo) Clean(ctx context.Context, days int32) (*v1.LogCleanReply, error) {
	return r.client.CleanLoginLog(ctx, &v1.CleanLogRequest{Days: days})
}

func (r *LoginLogRepo) Update(ctx context.Context, req *v1.UpdateLoginLogRequest) (*v1.MessageReply, error) {
	return r.client.UpdateLoginLog(ctx, req)
}

func (r *LoginLogRepo) Get(ctx context.Context, id int64) (*v1.LogItem, error) {
	return r.client.GetLoginLogById(ctx, &v1.LogIdRequest{Id: id})
}

func (r *LoginLogRepo) Delete(ctx context.Context, id int64) (*v1.MessageReply, error) {
	return r.client.DeleteLoginLog(ctx, &v1.LogIdRequest{Id: id})
}

func (r *LoginLogRepo) Close() error {
	if r == nil || r.conn == nil {
		return nil
	}
	return r.conn.Close()
}

func (r *OperLogRepo) Create(ctx context.Context, item *v1.CreateOperLogRequest) (*v1.MessageReply, error) {
	return r.client.CreateOperLog(ctx, item)
}

func (r *OperLogRepo) Page(ctx context.Context, req *v1.PageOperLogRequest) (*v1.PageLogReply, error) {
	return r.client.PageOperLog(ctx, req)
}

func (r *OperLogRepo) BatchDelete(ctx context.Context, ids []int64) (*v1.MessageReply, error) {
	return r.client.BatchDeleteOperLog(ctx, &v1.LogBatchIdsRequest{Ids: ids})
}

func (r *OperLogRepo) Clean(ctx context.Context, days int32) (*v1.LogCleanReply, error) {
	return r.client.CleanOperLog(ctx, &v1.CleanLogRequest{Days: days})
}

func (r *OperLogRepo) Update(ctx context.Context, req *v1.UpdateOperLogRequest) (*v1.MessageReply, error) {
	return r.client.UpdateOperLog(ctx, req)
}

func (r *OperLogRepo) Get(ctx context.Context, id int64) (*v1.LogItem, error) {
	return r.client.GetOperLogById(ctx, &v1.LogIdRequest{Id: id})
}

func (r *OperLogRepo) Delete(ctx context.Context, id int64) (*v1.MessageReply, error) {
	return r.client.DeleteOperLog(ctx, &v1.LogIdRequest{Id: id})
}

func (r *OperLogRepo) Close() error {
	if r == nil || r.conn == nil {
		return nil
	}
	return r.conn.Close()
}
