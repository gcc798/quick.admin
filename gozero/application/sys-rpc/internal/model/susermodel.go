package model

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ SUserModel = (*customSUserModel)(nil)

type (
	// SUserModel is an interface to be customized, add more methods here,
	// and implement the added methods in customSUserModel.
	SUserModel interface {
		sUserModel
		withSession(session sqlx.Session) SUserModel
	}

	customSUserModel struct {
		*defaultSUserModel
	}
)

// NewSUserModel returns a model for the database table.
func NewSUserModel(conn sqlx.SqlConn) SUserModel {
	return &customSUserModel{
		defaultSUserModel: newSUserModel(conn),
	}
}

func (m *customSUserModel) withSession(session sqlx.Session) SUserModel {
	return NewSUserModel(sqlx.NewSqlConnFromSession(session))
}
