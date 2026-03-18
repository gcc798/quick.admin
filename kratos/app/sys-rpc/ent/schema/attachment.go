package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

type Attachment struct {
	ent.Schema
}

func (Attachment) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "biz_attachment"}}
}

func (Attachment) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.Int64("env_id").Default(0),
		field.String("file_name").NotEmpty(),
		field.String("file_key").Default(""),
		field.Int64("file_size").Default(0),
		field.String("file_type").Default(""),
		field.String("file_ext").Default(""),
		field.String("business_type").Default(""),
		field.String("business_id").Default(""),
		field.String("business_field").Default(""),
		field.Bool("is_public").Default(false),
		field.String("access_url").Default(""),
		field.JSON("metadata", map[string]any{}).Optional().SchemaType(map[string]string{dialect.Postgres: "jsonb"}),
		field.Int32("status").Default(0),
		field.Time("expire_time").Optional().Nillable(),
		field.Int64("create_by").Default(0),
		field.Time("create_time").Default(time.Now),
		field.Time("update_time").Default(time.Now).UpdateDefault(time.Now),
		field.Time("deleted_at").Optional().Nillable(),
	}
}
