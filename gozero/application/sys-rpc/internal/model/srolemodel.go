package model

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ SRoleModel = (*customSRoleModel)(nil)

type (
	// SRoleModel is an interface to be customized, add more methods here,
	// and implement the added methods in customSRoleModel.
	SRoleModel interface {
		sRoleModel
		withSession(session sqlx.Session) SRoleModel
	}

	customSRoleModel struct {
		*defaultSRoleModel
	}
)

// NewSRoleModel returns a model for the database table.
func NewSRoleModel(conn sqlx.SqlConn) SRoleModel {
	return &customSRoleModel{
		defaultSRoleModel: newSRoleModel(conn),
	}
}

func (m *customSRoleModel) withSession(session sqlx.Session) SRoleModel {
	return NewSRoleModel(sqlx.NewSqlConnFromSession(session))
}
