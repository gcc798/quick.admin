package sysservicelogic

import (
	"context"
	"fmt"
	"strings"

	"github.com/force-c/nai-tizi/application/sys-rpc/internal/svc"
	"github.com/force-c/nai-tizi/application/sys-rpc/pb"
	"github.com/zeromicro/go-zero/core/logx"
)

type UserPageLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUserPageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserPageLogic {
	return &UserPageLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}
func (l *UserPageLogic) UserPage(in *pb.UserPageReq) (*pb.UserPageResp, error) {
	pageNum, pageSize := normalizePage(in.PageNum, in.PageSize)
	where := []string{"deleted_at is null"}
	args := make([]interface{}, 0)
	if in.Username != "" {
		args = append(args, "%"+in.Username+"%")
		where = append(where, fmt.Sprintf("user_name like $%d", len(args)))
	}
	if in.Phonenumber != "" {
		args = append(args, "%"+in.Phonenumber+"%")
		where = append(where, fmt.Sprintf("phonenumber like $%d", len(args)))
	}
	if in.Status == 0 || in.Status == 1 {
		args = append(args, in.Status)
		where = append(where, fmt.Sprintf("status = $%d", len(args)))
	}
	whereSQL := strings.Join(where, " and ")
	var total int64
	if err := l.svcCtx.DB.QueryRowCtx(l.ctx, &total, "select count(1) from public.s_user where "+whereSQL, args...); err != nil {
		return nil, err
	}
	queryArgs := append(append([]interface{}{}, args...), pageSize, (pageNum-1)*pageSize)
	var rows []struct {
		Id int64 `db:"id"`
	}
	query := `select id from public.s_user where ` + whereSQL + ` order by id desc limit $` + fmt.Sprint(len(args)+1) + ` offset $` + fmt.Sprint(len(args)+2)
	if err := l.svcCtx.DB.QueryRowsCtx(l.ctx, &rows, query, queryArgs...); err != nil {
		return nil, err
	}
	records := make([]*pb.User, 0, len(rows))
	for _, row := range rows {
		user, err := getUserByID(l.ctx, l.svcCtx, row.Id)
		if err != nil {
			return nil, err
		}
		records = append(records, user)
	}
	return &pb.UserPageResp{Records: records, Page: toPageInfo(total, pageNum, pageSize)}, nil
}
