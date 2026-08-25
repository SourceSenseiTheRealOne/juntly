package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	entschema "entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

type ProviderSpokenLanguage struct{ ent.Schema }

func (ProviderSpokenLanguage) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("internal_user_id", uuid.UUID{}).Immutable(),
		field.String("language_code").NotEmpty().MaxLen(10).Immutable(),
	}
}

func (ProviderSpokenLanguage) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("profile", ProviderProfile.Type).Field("internal_user_id").Unique().Required().Immutable(),
		edge.To("language", SpokenLanguage.Type).Field("language_code").Unique().Required().Immutable(),
	}
}

func (ProviderSpokenLanguage) Annotations() []entschema.Annotation {
	return []entschema.Annotation{
		field.ID("internal_user_id", "language_code"),
		entsql.Annotation{Table: "provider_spoken_languages"},
	}
}
