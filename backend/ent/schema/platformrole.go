package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	entschema "entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

type PlatformRole struct{ ent.Schema }

func (PlatformRole) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable(),
		field.UUID("internal_user_id", uuid.UUID{}).Immutable(),
		field.String("role").NotEmpty().MaxLen(20).Immutable(),
		field.Time("granted_at").Default(utcNow).Immutable(),
	}
}

func (PlatformRole) Indexes() []ent.Index {
	return []ent.Index{index.Fields("internal_user_id", "role").Unique()}
}

func (PlatformRole) Annotations() []entschema.Annotation {
	return []entschema.Annotation{entsql.Annotation{Table: "platform_roles"}}
}
