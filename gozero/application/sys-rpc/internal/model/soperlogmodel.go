package model

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ SOperLogModel = (*customSOperLogModel)(nil)

type (
	// SOperLogModel is an interface to be customized, add more methods here,
	// and implement the added methods in customSOperLogModel.
	SOperLogModel interface {
		sOperLogModel
		withSession(session sqlx.Session) SOperLogModel
	}

	customSOperLogModel struct {
		*defaultSOperLogModel
	}
)

// NewSOperLogModel returns a model for the database table.
func NewSOperLogModel(conn sqlx.SqlConn) SOperLogModel {
	return &customSOperLogModel{
		defaultSOperLogModel: newSOperLogModel(conn),
	}
}

func (m *customSOperLogModel) withSession(session sqlx.Session) SOperLogModel {
	return NewSOperLogModel(sqlx.NewSqlConnFromSession(session))
}
