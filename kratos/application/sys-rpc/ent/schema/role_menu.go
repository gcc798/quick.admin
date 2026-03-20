package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type RoleMenu struct {
	ent.Schema
}

func (RoleMenu) Mixin() []ent.Mixin { return []ent.Mixin{AuditMixin{}} }

func (RoleMenu) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "m_role_menu"}}
}

func (RoleMenu) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.Int64("role_id"),
		field.Int64("menu_id"),
	}
}

func (RoleMenu) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("role_id", "menu_id").Unique(),
	}
}
