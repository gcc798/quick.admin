package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type User struct {
	ent.Schema
}

func (User) Mixin() []ent.Mixin { return []ent.Mixin{AuditMixin{}} }

func (User) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "s_user"}}
}

func (User) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.String("user_name").NotEmpty(),
		field.String("nick_name").Default(""),
		field.Int32("user_type").Default(0),
		field.String("email").Default(""),
		field.String("phonenumber").Default(""),
		field.Int32("sex").Default(2),
		field.String("avatar").Default(""),
		field.String("password").Default(""),
		field.Int32("status").Default(0),
		field.Int64("sort").Default(0),
		field.String("login_ip").Default(""),
		field.Int64("login_date").Default(0),
		field.String("open_id").Default(""),
		field.String("union_id").Default(""),
		field.String("remark").Default(""),
	}
}

func (User) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_name").Unique(),
		index.Fields("email"),
		index.Fields("phonenumber"),
	}
}
