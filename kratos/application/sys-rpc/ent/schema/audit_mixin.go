package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
)

// AuditMixin captures the common audit fields used by most system tables.
type AuditMixin struct {
	mixin.Schema
}

func (AuditMixin) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("create_by").Default(0),
		field.Int64("update_by").Default(0),
		field.Time("created_time").Default(time.Now),
		field.Time("updated_time").Default(time.Now).UpdateDefault(time.Now),
		field.Time("deleted_at").Optional().Nillable(),
	}
}
