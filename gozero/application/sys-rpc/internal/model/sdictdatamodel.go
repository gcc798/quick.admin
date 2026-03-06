package model

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ SDictDataModel = (*customSDictDataModel)(nil)

type (
	// SDictDataModel is an interface to be customized, add more methods here,
	// and implement the added methods in customSDictDataModel.
	SDictDataModel interface {
		sDictDataModel
		withSession(session sqlx.Session) SDictDataModel
	}

	customSDictDataModel struct {
		*defaultSDictDataModel
	}
)

// NewSDictDataModel returns a model for the database table.
func NewSDictDataModel(conn sqlx.SqlConn) SDictDataModel {
	return &customSDictDataModel{
		defaultSDictDataModel: newSDictDataModel(conn),
	}
}

func (m *customSDictDataModel) withSession(session sqlx.Session) SDictDataModel {
	return NewSDictDataModel(sqlx.NewSqlConnFromSession(session))
}
