package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type Org struct {
	ent.Schema
}

func (Org) Mixin() []ent.Mixin { return []ent.Mixin{AuditMixin{}} }

func (Org) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "s_org"}}
}

func (Org) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.Int64("parent_id").Default(0),
		field.String("ancestors").Default(""),
		field.String("org_name").NotEmpty(),
		field.String("org_code").Default(""),
		field.String("org_type").Default("company"),
		field.String("leader").Default(""),
		field.String("phone").Default(""),
		field.String("email").Default(""),
		field.Int32("status").Default(0),
		field.Int64("sort").Default(0),
		field.String("remark").Default(""),
	}
}

func (Org) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("parent_id"),
		index.Fields("org_code").Unique(),
	}
}
