package model

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ SLoginLogModel = (*customSLoginLogModel)(nil)

type (
	// SLoginLogModel is an interface to be customized, add more methods here,
	// and implement the added methods in customSLoginLogModel.
	SLoginLogModel interface {
		sLoginLogModel
		withSession(session sqlx.Session) SLoginLogModel
	}

	customSLoginLogModel struct {
		*defaultSLoginLogModel
	}
)

// NewSLoginLogModel returns a model for the database table.
func NewSLoginLogModel(conn sqlx.SqlConn) SLoginLogModel {
	return &customSLoginLogModel{
		defaultSLoginLogModel: newSLoginLogModel(conn),
	}
}

func (m *customSLoginLogModel) withSession(session sqlx.Session) SLoginLogModel {
	return NewSLoginLogModel(sqlx.NewSqlConnFromSession(session))
}
