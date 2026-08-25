package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	entschema "entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

type ProviderServiceLocality struct{ ent.Schema }

func (ProviderServiceLocality) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("internal_user_id", uuid.UUID{}).Immutable(),
		field.UUID("locality_id", uuid.UUID{}).Immutable(),
	}
}

func (ProviderServiceLocality) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("profile", ProviderProfile.Type).Field("internal_user_id").Unique().Required().Immutable(),
		edge.To("locality", Locality.Type).Field("locality_id").Unique().Required().Immutable(),
	}
}

func (ProviderServiceLocality) Annotations() []entschema.Annotation {
	return []entschema.Annotation{
		field.ID("internal_user_id", "locality_id"),
		entsql.Annotation{Table: "provider_service_localities"},
	}
}
