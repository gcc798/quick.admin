package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type LoginLog struct {
	ent.Schema
}

func (LoginLog) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "s_login_log"}}
}

func (LoginLog) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.String("user_name").Default(""),
		field.String("ipaddr").Default(""),
		field.String("login_location").Default(""),
		field.String("browser").Default(""),
		field.String("os").Default(""),
		field.Int32("status").Default(0),
		field.String("msg").Default(""),
		field.Time("login_time").Default(time.Now),
		field.String("client_id").Default(""),
	}
}

func (LoginLog) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_name"),
		index.Fields("login_time"),
	}
}
