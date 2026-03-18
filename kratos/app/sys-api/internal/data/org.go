package data

import (
	"context"

	v1 "github.com/force-c/nai-tizi/kratos/api/system/v1"
	"google.golang.org/grpc"
)

type OrgRepo struct {
	conn   *grpc.ClientConn
	client v1.OrgServiceClient
}

func NewOrgRepo(endpoint string) (*OrgRepo, error) {
	conn, err := dialRPC(endpoint)
	if err != nil {
		return nil, err
	}
	return &OrgRepo{conn: conn, client: v1.NewOrgServiceClient(conn)}, nil
}
func (r *OrgRepo) Create(ctx context.Context, req *v1.CreateOrgRequest) (*v1.OrgIdReply, error) {
	return r.client.CreateOrg(ctx, req)
}
func (r *OrgRepo) Update(ctx context.Context, req *v1.UpdateOrgRequest) (*v1.MessageReply, error) {
	return r.client.UpdateOrg(ctx, req)
}
func (r *OrgRepo) Delete(ctx context.Context, id int64) (*v1.MessageReply, error) {
	return r.client.DeleteOrg(ctx, &v1.OrgIdRequest{Id: id})
}
func (r *OrgRepo) BatchDelete(ctx context.Context, ids []int64) (*v1.MessageReply, error) {
	return r.client.BatchDeleteOrg(ctx, &v1.BatchDeleteOrgsRequest{Ids: ids})
}
func (r *OrgRepo) GetByID(ctx context.Context, id int64) (*v1.OrgItem, error) {
	return r.client.GetOrgById(ctx, &v1.OrgIdRequest{Id: id})
}
func (r *OrgRepo) Tree(ctx context.Context) (*v1.OrgTreeReply, error) {
	return r.client.GetOrgTree(ctx, &v1.GetOrgTreeRequest{})
}
func (r *OrgRepo) Page(ctx context.Context, req *v1.PageOrgsRequest) (*v1.PageOrgsReply, error) {
	return r.client.PageOrg(ctx, req)
}
func (r *OrgRepo) Close() error {
	if r == nil || r.conn == nil {
		return nil
	}
	return r.conn.Close()
}
