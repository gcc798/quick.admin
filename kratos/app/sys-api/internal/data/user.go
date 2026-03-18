package data

import (
	"context"

	v1 "github.com/force-c/nai-tizi/kratos/api/system/v1"
	"google.golang.org/grpc"
)

type UserRepo struct {
	conn   *grpc.ClientConn
	client v1.UserServiceClient
}

func NewUserRepo(endpoint string) (*UserRepo, error) {
	conn, err := dialRPC(endpoint)
	if err != nil {
		return nil, err
	}
	return &UserRepo{conn: conn, client: v1.NewUserServiceClient(conn)}, nil
}
func (r *UserRepo) Create(ctx context.Context, req *v1.CreateUserRequest) (*v1.MessageReply, error) {
	return r.client.CreateUser(ctx, req)
}
func (r *UserRepo) Update(ctx context.Context, req *v1.UpdateUserRequest) (*v1.MessageReply, error) {
	return r.client.UpdateUser(ctx, req)
}
func (r *UserRepo) Delete(ctx context.Context, id int64) (*v1.MessageReply, error) {
	return r.client.DeleteUser(ctx, &v1.UserIdRequest{Id: id})
}
func (r *UserRepo) BatchDelete(ctx context.Context, ids []int64) (*v1.MessageReply, error) {
	return r.client.BatchDeleteUser(ctx, &v1.BatchDeleteUsersRequest{Ids: ids})
}
func (r *UserRepo) GetByID(ctx context.Context, id int64) (*v1.UserItem, error) {
	return r.client.GetUserById(ctx, &v1.UserIdRequest{Id: id})
}
func (r *UserRepo) Import(ctx context.Context, users []*v1.CreateUserRequest) (*v1.ImportUsersReply, error) {
	return r.client.ImportUsers(ctx, &v1.BatchImportUsersRequest{Users: users})
}
func (r *UserRepo) ResetPassword(ctx context.Context, req *v1.ResetPasswordRequest) (*v1.MessageReply, error) {
	return r.client.ResetPassword(ctx, req)
}
func (r *UserRepo) Page(ctx context.Context, req *v1.PageUsersRequest) (*v1.PageUsersReply, error) {
	return r.client.PageUser(ctx, req)
}
func (r *UserRepo) ChangePassword(ctx context.Context, req *v1.ChangePasswordRequest) (*v1.MessageReply, error) {
	return r.client.ChangePassword(ctx, req)
}
func (r *UserRepo) Close() error {
	if r == nil || r.conn == nil {
		return nil
	}
	return r.conn.Close()
}
