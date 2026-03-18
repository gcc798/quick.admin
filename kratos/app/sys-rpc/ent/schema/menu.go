package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type Menu struct {
	ent.Schema
}

func (Menu) Mixin() []ent.Mixin { return []ent.Mixin{AuditMixin{}} }

func (Menu) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "s_menu"}}
}

func (Menu) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.String("menu_name").NotEmpty(),
		field.Int64("parent_id").Default(0),
		field.Int64("sort").Default(0),
		field.String("path").Default(""),
		field.String("component").Default(""),
		field.String("query").Default(""),
		field.Int32("is_frame").Default(0),
		field.Int32("is_cache").Default(0),
		field.Int32("menu_type").Default(0),
		field.Int32("visible").Default(0),
		field.Int32("status").Default(0),
		field.String("perms").Default(""),
		field.String("icon").Default(""),
		field.String("remark").Default(""),
	}
}

func (Menu) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("parent_id"),
	}
}
