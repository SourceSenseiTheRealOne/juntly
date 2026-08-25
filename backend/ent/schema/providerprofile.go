package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	entschema "entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

type ProviderProfile struct{ ent.Schema }

func (ProviderProfile) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).StorageKey("internal_user_id").Immutable(),
		field.String("display_name").NotEmpty().MaxLen(100),
		field.String("provider_type").NotEmpty().MaxLen(20),
		field.String("bio").MaxLen(1000),
		field.UUID("primary_locality_id", uuid.UUID{}),
		field.Int("max_travel_distance_km").Min(0).Max(200),
		field.Bool("travels_to_customer").Default(false),
		field.Bool("receives_customer").Default(false),
		field.Bool("remote_services").Default(false),
		field.Time("created_at").Default(utcNow).Immutable(),
		field.Time("updated_at").Default(utcNow).UpdateDefault(utcNow),
	}
}

func (ProviderProfile) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("service_localities", Locality.Type).Through("service_locality_links", ProviderServiceLocality.Type),
		edge.To("spoken_languages", SpokenLanguage.Type).Through("spoken_language_links", ProviderSpokenLanguage.Type),
	}
}

func (ProviderProfile) Annotations() []entschema.Annotation {
	return []entschema.Annotation{entsql.Annotation{Table: "provider_profiles"}}
}
