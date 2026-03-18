package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type StorageEnv struct {
	ent.Schema
}

func (StorageEnv) Mixin() []ent.Mixin { return []ent.Mixin{AuditMixin{}} }

func (StorageEnv) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "s_storage_env"}}
}

func (StorageEnv) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.String("name").NotEmpty(),
		field.String("code").NotEmpty(),
		field.String("storage_type").Default("local"),
		field.Bool("is_default").Default(false),
		field.Int32("status").Default(0),
		field.JSON("config", map[string]any{}).Optional().SchemaType(map[string]string{dialect.Postgres: "jsonb"}),
		field.String("remark").Default(""),
	}
}

func (StorageEnv) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("code").Unique(),
	}
}
