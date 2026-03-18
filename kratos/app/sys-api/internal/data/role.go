package data

import (
	"context"

	v1 "github.com/force-c/nai-tizi/kratos/api/system/v1"
	"google.golang.org/grpc"
)

type RoleRepo struct {
	conn   *grpc.ClientConn
	client v1.RoleServiceClient
}

func NewRoleRepo(endpoint string) (*RoleRepo, error) {
	conn, err := dialRPC(endpoint)
	if err != nil {
		return nil, err
	}
	return &RoleRepo{conn: conn, client: v1.NewRoleServiceClient(conn)}, nil
}
func (r *RoleRepo) Create(ctx context.Context, req *v1.CreateRoleRequest) (*v1.RoleItem, error) {
	return r.client.CreateRole(ctx, req)
}
func (r *RoleRepo) Update(ctx context.Context, req *v1.UpdateRoleRequest) (*v1.MessageReply, error) {
	return r.client.UpdateRole(ctx, req)
}
func (r *RoleRepo) Delete(ctx context.Context, id int64) (*v1.MessageReply, error) {
	return r.client.DeleteRole(ctx, &v1.RoleIdRequest{RoleId: id})
}
func (r *RoleRepo) GetByID(ctx context.Context, id int64) (*v1.RoleItem, error) {
	return r.client.GetRole(ctx, &v1.RoleIdRequest{RoleId: id})
}
func (r *RoleRepo) Page(ctx context.Context, req *v1.PageRoleRequest) (*v1.PageRoleReply, error) {
	return r.client.PageRole(ctx, req)
}
func (r *RoleRepo) Assign(ctx context.Context, req *v1.AssignRoleToUserRequest) (*v1.MessageReply, error) {
	return r.client.AssignRoleToUser(ctx, req)
}
func (r *RoleRepo) Remove(ctx context.Context, req *v1.AssignRoleToUserRequest) (*v1.MessageReply, error) {
	return r.client.RemoveRoleFromUser(ctx, req)
}
func (r *RoleRepo) GetUserRoles(ctx context.Context, userID int64) (*v1.GetUserRolesReply, error) {
	return r.client.GetUserRoles(ctx, &v1.GetUserRolesRequest{UserId: userID})
}
func (r *RoleRepo) AddPermission(ctx context.Context, req *v1.RolePermissionRequest) (*v1.MessageReply, error) {
	return r.client.AddRolePermission(ctx, req)
}
func (r *RoleRepo) DeletePermission(ctx context.Context, req *v1.RolePermissionRequest) (*v1.MessageReply, error) {
	return r.client.DeleteRolePermission(ctx, req)
}
func (r *RoleRepo) GetPermissions(ctx context.Context, roleKey string) (*v1.GetRolePermissionsReply, error) {
	return r.client.GetRolePermissions(ctx, &v1.GetRolePermissionsRequest{RoleKey: roleKey})
}
func (r *RoleRepo) Close() error {
	if r == nil || r.conn == nil {
		return nil
	}
	return r.conn.Close()
}
