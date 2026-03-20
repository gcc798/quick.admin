package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type DictData struct {
	ent.Schema
}

func (DictData) Mixin() []ent.Mixin { return []ent.Mixin{AuditMixin{}} }

func (DictData) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "s_dict_data"}}
}

func (DictData) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.Int64("parent_id").Default(0),
		field.Int64("sort").Default(0),
		field.String("dict_label").Default(""),
		field.String("dict_value").Default(""),
		field.String("dict_type").Default(""),
		field.Bool("is_default").Default(false),
		field.Int32("status").Default(0),
		field.String("remark").Default(""),
	}
}

func (DictData) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("parent_id"),
		index.Fields("dict_type"),
	}
}
