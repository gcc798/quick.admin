package data

import (
	"context"

	v1 "github.com/force-c/nai-tizi/kratos/api/system/v1"
	"google.golang.org/grpc"
)

type HealthRepo struct {
	conn   *grpc.ClientConn
	client v1.HealthServiceClient
}

func NewHealthRepo(endpoint string) (*HealthRepo, error) {
	conn, err := dialRPC(endpoint)
	if err != nil {
		return nil, err
	}
	return &HealthRepo{
		conn:   conn,
		client: v1.NewHealthServiceClient(conn),
	}, nil
}

func (r *HealthRepo) Ping(ctx context.Context, name string) (*v1.PingReply, error) {
	return r.client.Ping(ctx, &v1.PingRequest{Name: name})
}

func (r *HealthRepo) Close() error {
	if r == nil || r.conn == nil {
		return nil
	}
	return r.conn.Close()
}
