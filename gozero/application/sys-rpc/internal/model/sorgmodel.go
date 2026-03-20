package model

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ SOrgModel = (*customSOrgModel)(nil)

type (
	// SOrgModel is an interface to be customized, add more methods here,
	// and implement the added methods in customSOrgModel.
	SOrgModel interface {
		sOrgModel
		withSession(session sqlx.Session) SOrgModel
	}

	customSOrgModel struct {
		*defaultSOrgModel
	}
)

// NewSOrgModel returns a model for the database table.
func NewSOrgModel(conn sqlx.SqlConn) SOrgModel {
	return &customSOrgModel{
		defaultSOrgModel: newSOrgModel(conn),
	}
}

func (m *customSOrgModel) withSession(session sqlx.Session) SOrgModel {
	return NewSOrgModel(sqlx.NewSqlConnFromSession(session))
}
