package sysservicelogic

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/gcc798/nai-tizi/application/sys-rpc/internal/svc"
	"github.com/gcc798/nai-tizi/application/sys-rpc/pb"
	gzsqlx "github.com/zeromicro/go-zero/core/stores/sqlx"
	"golang.org/x/crypto/bcrypt"
)

type authClientRow struct {
	ClientId      string         `db:"client_id"`
	ClientKey     string         `db:"client_key"`
	ClientSecret  string         `db:"client_secret"`
	GrantType     sql.NullString `db:"grant_type"`
	DeviceType    sql.NullString `db:"device_type"`
	Status        int64          `db:"status"`
	Timeout       int64          `db:"timeout"`
	ActiveTimeout int64          `db:"active_timeout"`
}

type userAuthRow struct {
	Id          int64          `db:"id"`
	OrgId       sql.NullInt64  `db:"org_id"`
	UserName    string         `db:"user_name"`
	NickName    sql.NullString `db:"nick_name"`
	UserType    int64          `db:"user_type"`
	Email       sql.NullString `db:"email"`
	Phonenumber sql.NullString `db:"phonenumber"`
	Avatar      sql.NullString `db:"avatar"`
	Password    sql.NullString `db:"password"`
	Status      int64          `db:"status"`
}

type menuRow struct {
	Id          int64          `db:"id"`
	MenuName    string         `db:"menu_name"`
	ParentId    int64          `db:"parent_id"`
	Sort        int64          `db:"sort"`
	Path        sql.NullString `db:"path"`
	Component   sql.NullString `db:"component"`
	Query       sql.NullString `db:"query"`
	IsFrame     int64          `db:"is_frame"`
	IsCache     int64          `db:"is_cache"`
	MenuType    int64          `db:"menu_type"`
	Visible     int64          `db:"visible"`
	Status      int64          `db:"status"`
	Perms       sql.NullString `db:"perms"`
	Icon        sql.NullString `db:"icon"`
	Remark      sql.NullString `db:"remark"`
	CreateBy    sql.NullInt64  `db:"create_by"`
	CreatedTime sql.NullTime   `db:"created_time"`
	UpdatedTime sql.NullTime   `db:"updated_time"`
}

type roleKeyRow struct {
	RoleKey sql.NullString `db:"role_key"`
}

func authenticateClient(ctx context.Context, svcCtx *svc.ServiceContext, clientKey, clientSecret, grantType string) (*authClientRow, error) {
	if clientKey == "" || clientSecret == "" {
		return nil, fmt.Errorf("clientKey和clientSecret不能为空")
	}
	if grantType == "" {
		return nil, fmt.Errorf("grantType不能为空")
	}
	var row authClientRow
	err := svcCtx.DB.QueryRowCtx(ctx, &row, `
		select client_id, client_key, client_secret, grant_type, device_type, status, timeout, active_timeout
		from public.s_auth_client
		where client_key = $1 and deleted_at is null
		limit 1
	`, clientKey)
	if err != nil {
		if errors.Is(err, gzsqlx.ErrNotFound) {
			return nil, fmt.Errorf("客户端不存在")
		}
		return nil, fmt.Errorf("查询客户端失败")
	}
	if row.ClientSecret != clientSecret {
		return nil, fmt.Errorf("客户端认证失败")
	}
	if row.Status != 0 {
		return nil, fmt.Errorf("客户端已停用")
	}
	if !grantTypeSupported(row.GrantType.String, grantType) {
		return nil, fmt.Errorf("客户端不支持该授权类型")
	}
	return &row, nil
}

func authenticatePassword(ctx context.Context, svcCtx *svc.ServiceContext, username, password string) (*userAuthRow, error) {
	if username == "" || password == "" {
		return nil, fmt.Errorf("用户名和密码不能为空")
	}
	var row userAuthRow
	err := svcCtx.DB.QueryRowCtx(ctx, &row, `
		select id, org_id, user_name, nick_name, user_type, email, phonenumber, avatar, password, status
		from public.s_user
		where user_name = $1 and deleted_at is null
		limit 1
	`, username)
	if err != nil {
		if errors.Is(err, gzsqlx.ErrNotFound) {
			return nil, fmt.Errorf("用户名或密码错误")
		}
		return nil, fmt.Errorf("登录失败")
	}
	if !row.Password.Valid || row.Password.String == "" {
		return nil, fmt.Errorf("用户名或密码错误")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(row.Password.String), []byte(password)); err != nil {
		return nil, fmt.Errorf("用户名或密码错误")
	}
	if row.Status != 0 {
		return nil, fmt.Errorf("用户已被停用")
	}
	return &row, nil
}

func grantTypeSupported(supported, actual string) bool {
	if supported == "" || supported == actual {
		return true
	}
	for _, item := range splitAndTrim(supported) {
		if item == actual {
			return true
		}
	}
	return false
}

func splitAndTrim(s string) []string {
	out := make([]string, 0)
	cur := ""
	for _, r := range s {
		if r == ',' || r == '|' || r == ' ' {
			if cur != "" {
				out = append(out, cur)
				cur = ""
			}
			continue
		}
		cur += string(r)
	}
	if cur != "" {
		out = append(out, cur)
	}
	return out
}

func getUserByID(ctx context.Context, svcCtx *svc.ServiceContext, id int64) (*pb.User, error) {
	var row struct {
		Id          int64          `db:"id"`
		UserName    string         `db:"user_name"`
		NickName    sql.NullString `db:"nick_name"`
		UserType    int64          `db:"user_type"`
		Email       sql.NullString `db:"email"`
		Phonenumber sql.NullString `db:"phonenumber"`
		Sex         int64          `db:"sex"`
		Avatar      sql.NullString `db:"avatar"`
		Status      int64          `db:"status"`
		Sort        int64          `db:"sort"`
		LoginIp     sql.NullString `db:"login_ip"`
		LoginDate   sql.NullInt64  `db:"login_date"`
		OpenId      sql.NullString `db:"open_id"`
		UnionId     sql.NullString `db:"union_id"`
		Remark      sql.NullString `db:"remark"`
		CreateBy    sql.NullInt64  `db:"create_by"`
		UpdateBy    sql.NullInt64  `db:"update_by"`
		CreatedTime sql.NullTime   `db:"created_time"`
		UpdatedTime sql.NullTime   `db:"updated_time"`
		OrgId       sql.NullInt64  `db:"org_id"`
	}
	err := svcCtx.DB.QueryRowCtx(ctx, &row, `
		select id, user_name, nick_name, user_type, email, phonenumber, sex, avatar, status, sort, login_ip, login_date, open_id, union_id, remark, create_by, update_by, created_time, updated_time, org_id
		from public.s_user where id = $1 and deleted_at is null limit 1
	`, id)
	if err != nil {
		if errors.Is(err, gzsqlx.ErrNotFound) {
			return nil, fmt.Errorf("用户不存在")
		}
		return nil, err
	}
	return &pb.User{
		UserId:      row.Id,
		UserName:    row.UserName,
		NickName:    nullString(row.NickName),
		UserType:    int32(row.UserType),
		Email:       nullString(row.Email),
		Phonenumber: nullString(row.Phonenumber),
		Sex:         int32(row.Sex),
		Avatar:      nullString(row.Avatar),
		Status:      int32(row.Status),
		Sort:        row.Sort,
		LoginIp:     nullString(row.LoginIp),
		LoginDate:   nullInt64(row.LoginDate),
		OpenId:      nullString(row.OpenId),
		UnionId:     nullString(row.UnionId),
		Remark:      nullString(row.Remark),
		CreateBy:    nullInt64(row.CreateBy),
		UpdateBy:    nullInt64(row.UpdateBy),
		CreatedAt:   nullTime(row.CreatedTime),
		UpdatedAt:   nullTime(row.UpdatedTime),
		OrgId:       nullInt64(row.OrgId),
	}, nil
}

func getAllMenus(ctx context.Context, svcCtx *svc.ServiceContext) ([]menuRow, error) {
	var rows []menuRow
	err := svcCtx.DB.QueryRowsCtx(ctx, &rows, `
		select id, menu_name, parent_id, sort, path, component, query, is_frame, is_cache, menu_type, visible, status, perms, icon, remark, create_by, created_time, updated_time
		from public.s_menu where deleted_at is null and status = 0 order by sort asc, id asc
	`)
	return rows, err
}

func getUserMenus(ctx context.Context, svcCtx *svc.ServiceContext, userId int64) ([]menuRow, error) {
	var roles []roleKeyRow
	if err := svcCtx.DB.QueryRowsCtx(ctx, &roles, `
		select r.role_key from public.s_role r
		join public.m_user_role mur on mur.role_id = r.id
		where mur.user_id = $1 and mur.deleted_at is null and r.deleted_at is null and r.status = 0
	`, userId); err != nil {
		return nil, err
	}
	for _, role := range roles {
		if role.RoleKey.Valid && role.RoleKey.String == "super_admin" {
			return getAllMenus(ctx, svcCtx)
		}
	}
	var rows []menuRow
	err := svcCtx.DB.QueryRowsCtx(ctx, &rows, `
		select distinct m.id, m.menu_name, m.parent_id, m.sort, m.path, m.component, m.query, m.is_frame, m.is_cache, m.menu_type, m.visible, m.status, m.perms, m.icon, m.remark, m.create_by, m.created_time, m.updated_time
		from public.s_menu m
		join public.m_role_menu rm on rm.menu_id = m.id and rm.deleted_at is null
		join public.m_user_role mur on mur.role_id = rm.role_id and mur.deleted_at is null
		join public.s_role r on r.id = mur.role_id
		where mur.user_id = $1 and m.deleted_at is null and r.deleted_at is null and m.status = 0 and r.status = 0
		order by m.sort asc, m.id asc
	`, userId)
	return rows, err
}

func getMenuByID(ctx context.Context, svcCtx *svc.ServiceContext, id int64) (*pb.Menu, error) {
	var row menuRow
	err := svcCtx.DB.QueryRowCtx(ctx, &row, `
		select id, menu_name, parent_id, sort, path, component, query, is_frame, is_cache, menu_type, visible, status, perms, icon, remark, create_by, created_time, updated_time
		from public.s_menu where id = $1 and deleted_at is null limit 1
	`, id)
	if err != nil {
		if errors.Is(err, gzsqlx.ErrNotFound) {
			return nil, fmt.Errorf("菜单不存在")
		}
		return nil, err
	}
	return toMenuPB(row), nil
}

func buildMenuTree(rows []menuRow, parentID int64) []*pb.Menu {
	tree := make([]*pb.Menu, 0)
	for _, row := range rows {
		if row.ParentId != parentID {
			continue
		}
		node := toMenuPB(row)
		node.Children = buildMenuTree(rows, row.Id)
		tree = append(tree, node)
	}
	return tree
}

func toMenuList(rows []menuRow) []*pb.Menu {
	list := make([]*pb.Menu, 0, len(rows))
	for _, row := range rows {
		list = append(list, toMenuPB(row))
	}
	return list
}

func toMenuPB(row menuRow) *pb.Menu {
	return &pb.Menu{
		Id:          row.Id,
		MenuName:    row.MenuName,
		ParentId:    row.ParentId,
		Sort:        row.Sort,
		Path:        nullString(row.Path),
		Component:   nullString(row.Component),
		Query:       nullString(row.Query),
		IsFrame:     row.IsFrame,
		IsCache:     row.IsCache,
		MenuType:    row.MenuType,
		Visible:     row.Visible,
		Status:      row.Status,
		Perms:       nullString(row.Perms),
		Icon:        nullString(row.Icon),
		Remark:      nullString(row.Remark),
		CreateBy:    nullInt64(row.CreateBy),
		CreatedTime: nullTime(row.CreatedTime),
		UpdatedTime: nullTime(row.UpdatedTime),
	}
}

func nullString(v sql.NullString) string {
	if v.Valid {
		return v.String
	}
	return ""
}
func nullInt64(v sql.NullInt64) int64 {
	if v.Valid {
		return v.Int64
	}
	return 0
}
func nullTime(v sql.NullTime) string {
	if v.Valid {
		return v.Time.Format("2006-01-02 15:04:05")
	}
	return ""
}
