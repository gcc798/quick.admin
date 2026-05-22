package model

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ SApiPermissionModel = (*customSApiPermissionModel)(nil)

type (
	// SApiPermissionModel is an interface to be customized, add more methods here,
	// and implement the added methods in customSApiPermissionModel.
	SApiPermissionModel interface {
		sApiPermissionModel
		withSession(session sqlx.Session) SApiPermissionModel
	}

	customSApiPermissionModel struct {
		*defaultSApiPermissionModel
	}
)

// NewSApiPermissionModel returns a model for the database table.
func NewSApiPermissionModel(conn sqlx.SqlConn) SApiPermissionModel {
	return &customSApiPermissionModel{
		defaultSApiPermissionModel: newSApiPermissionModel(conn),
	}
}

func (m *customSApiPermissionModel) withSession(session sqlx.Session) SApiPermissionModel {
	return NewSApiPermissionModel(sqlx.NewSqlConnFromSession(session))
}
