package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type UserRole struct {
	ent.Schema
}

func (UserRole) Mixin() []ent.Mixin { return []ent.Mixin{AuditMixin{}} }

func (UserRole) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "m_user_role"}}
}

func (UserRole) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.Int64("user_id"),
		field.Int64("role_id"),
	}
}

func (UserRole) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "role_id").Unique(),
	}
}
