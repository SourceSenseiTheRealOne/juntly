package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	entschema "entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

type UserAccount struct {
	ent.Schema
}

func (UserAccount) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).StorageKey("internal_user_id").Immutable(),
		field.Bool("provider_enabled").Default(false),
		field.Time("onboarding_completed_at").Default(utcNow).Immutable(),
		field.Time("created_at").Default(utcNow).Immutable(),
		field.Time("updated_at").Default(utcNow).UpdateDefault(utcNow),
	}
}

func (UserAccount) Annotations() []entschema.Annotation {
	return []entschema.Annotation{
		entsql.Annotation{Table: "user_accounts"},
	}
}
