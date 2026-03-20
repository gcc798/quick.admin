package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

type OperLog struct {
	ent.Schema
}

func (OperLog) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "s_oper_log"}}
}

func (OperLog) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.String("title").Default(""),
		field.String("business_type").Default(""),
		field.String("method").Default(""),
		field.String("request_method").Default(""),
		field.String("device_type").Default(""),
		field.String("oper_name").Default(""),
		field.String("oper_url").Default(""),
		field.String("oper_ip").Default(""),
		field.String("oper_location").Default(""),
		field.String("oper_param").Default(""),
		field.String("json_result").Default(""),
		field.String("status").Default("0"),
		field.String("error_msg").Default(""),
		field.Time("oper_time").Default(time.Now),
		field.Int64("cost_time").Default(0),
		field.String("user_agent").Default(""),
	}
}
