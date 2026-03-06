package model

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ SAuthClientModel = (*customSAuthClientModel)(nil)

type (
	// SAuthClientModel is an interface to be customized, add more methods here,
	// and implement the added methods in customSAuthClientModel.
	SAuthClientModel interface {
		sAuthClientModel
		withSession(session sqlx.Session) SAuthClientModel
	}

	customSAuthClientModel struct {
		*defaultSAuthClientModel
	}
)

// NewSAuthClientModel returns a model for the database table.
func NewSAuthClientModel(conn sqlx.SqlConn) SAuthClientModel {
	return &customSAuthClientModel{
		defaultSAuthClientModel: newSAuthClientModel(conn),
	}
}

func (m *customSAuthClientModel) withSession(session sqlx.Session) SAuthClientModel {
	return NewSAuthClientModel(sqlx.NewSqlConnFromSession(session))
}
