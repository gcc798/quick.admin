package model

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ SMenuModel = (*customSMenuModel)(nil)

type (
	// SMenuModel is an interface to be customized, add more methods here,
	// and implement the added methods in customSMenuModel.
	SMenuModel interface {
		sMenuModel
		withSession(session sqlx.Session) SMenuModel
	}

	customSMenuModel struct {
		*defaultSMenuModel
	}
)

// NewSMenuModel returns a model for the database table.
func NewSMenuModel(conn sqlx.SqlConn) SMenuModel {
	return &customSMenuModel{
		defaultSMenuModel: newSMenuModel(conn),
	}
}

func (m *customSMenuModel) withSession(session sqlx.Session) SMenuModel {
	return NewSMenuModel(sqlx.NewSqlConnFromSession(session))
}
