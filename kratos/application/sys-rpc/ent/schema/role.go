package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type Role struct {
	ent.Schema
}

func (Role) Mixin() []ent.Mixin { return []ent.Mixin{AuditMixin{}} }

func (Role) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "s_role"}}
}

func (Role) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.String("role_key").NotEmpty(),
		field.String("role_name").NotEmpty(),
		field.Int64("sort").Default(0),
		field.Int32("status").Default(0),
		field.Int32("data_scope").Default(1),
		field.Bool("is_system").Default(false),
		field.String("remark").Default(""),
	}
}

func (Role) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("role_key").Unique(),
	}
}
