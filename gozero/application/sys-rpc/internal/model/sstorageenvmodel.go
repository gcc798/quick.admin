package model

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ SStorageEnvModel = (*customSStorageEnvModel)(nil)

type (
	// SStorageEnvModel is an interface to be customized, add more methods here,
	// and implement the added methods in customSStorageEnvModel.
	SStorageEnvModel interface {
		sStorageEnvModel
		withSession(session sqlx.Session) SStorageEnvModel
	}

	customSStorageEnvModel struct {
		*defaultSStorageEnvModel
	}
)

// NewSStorageEnvModel returns a model for the database table.
func NewSStorageEnvModel(conn sqlx.SqlConn) SStorageEnvModel {
	return &customSStorageEnvModel{
		defaultSStorageEnvModel: newSStorageEnvModel(conn),
	}
}

func (m *customSStorageEnvModel) withSession(session sqlx.Session) SStorageEnvModel {
	return NewSStorageEnvModel(sqlx.NewSqlConnFromSession(session))
}
