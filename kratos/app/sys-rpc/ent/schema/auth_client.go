package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type AuthClient struct {
	ent.Schema
}

func (AuthClient) Mixin() []ent.Mixin { return []ent.Mixin{AuditMixin{}} }

func (AuthClient) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "s_auth_client"}}
}

func (AuthClient) Fields() []ent.Field {
	return []ent.Field{
		field.String("client_id").NotEmpty(),
		field.String("client_key").NotEmpty(),
		field.String("client_secret").NotEmpty(),
		field.String("grant_type").Default(""),
		field.String("device_type").Default(""),
		field.Int32("status").Default(0),
		field.Int64("timeout").Default(604800),
		field.Int64("active_timeout").Default(1800),
		field.String("remark").Default(""),
	}
}

func (AuthClient) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("client_id").Unique(),
		index.Fields("client_key").Unique(),
	}
}
