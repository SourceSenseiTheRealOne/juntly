package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	entschema "entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

type InternalUser struct {
	ent.Schema
}

func (InternalUser) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable(),
		field.String("clerk_subject").NotEmpty().MaxLen(255).Unique().Immutable(),
		field.Time("created_at").Default(utcNow).Immutable(),
		field.Time("updated_at").Default(utcNow).UpdateDefault(utcNow),
	}
}

func (InternalUser) Annotations() []entschema.Annotation {
	return []entschema.Annotation{
		entsql.Annotation{Table: "internal_users"},
	}
}

func utcNow() time.Time {
	return time.Now().UTC()
}
