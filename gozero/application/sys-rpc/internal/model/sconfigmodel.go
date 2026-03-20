package model

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ SConfigModel = (*customSConfigModel)(nil)

type (
	// SConfigModel is an interface to be customized, add more methods here,
	// and implement the added methods in customSConfigModel.
	SConfigModel interface {
		sConfigModel
		withSession(session sqlx.Session) SConfigModel
	}

	customSConfigModel struct {
		*defaultSConfigModel
	}
)

// NewSConfigModel returns a model for the database table.
func NewSConfigModel(conn sqlx.SqlConn) SConfigModel {
	return &customSConfigModel{
		defaultSConfigModel: newSConfigModel(conn),
	}
}

func (m *customSConfigModel) withSession(session sqlx.Session) SConfigModel {
	return NewSConfigModel(sqlx.NewSqlConnFromSession(session))
}
