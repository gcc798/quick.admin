package schema

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type SystemConfig struct {
	ent.Schema
}

func (SystemConfig) Mixin() []ent.Mixin { return []ent.Mixin{AuditMixin{}} }

func (SystemConfig) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "s_config"}}
}

func (SystemConfig) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.String("name").NotEmpty(),
		field.String("code").NotEmpty(),
		field.JSON("data", json.RawMessage{}).Optional().SchemaType(map[string]string{dialect.Postgres: "jsonb"}),
		field.String("remark").Default(""),
	}
}

func (SystemConfig) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("code"),
	}
}
