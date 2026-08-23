package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	entschema "entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

type Locality struct{ ent.Schema }

func (Locality) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable(),
		field.String("slug").NotEmpty().MaxLen(100).Unique().Immutable(),
		field.String("name").NotEmpty().MaxLen(160),
		field.UUID("parent_parish_id", uuid.UUID{}).Immutable(),
		field.String("source").NotEmpty().MaxLen(40).Immutable(),
		field.String("source_element_id").NotEmpty().MaxLen(32).Unique().Immutable(),
		field.String("source_version").NotEmpty().MaxLen(20).Immutable(),
		field.Time("source_retrieved_at").Immutable(),
		field.Float("latitude").Min(-90).Max(90).Immutable(),
		field.Float("longitude").Min(-180).Max(180).Immutable(),
		field.Bool("active").Default(true),
		field.Time("created_at").Default(utcNow).Immutable(),
		field.Time("updated_at").Default(utcNow).UpdateDefault(utcNow),
	}
}

func (Locality) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("parent_parish", AdministrativeArea.Type).Ref("localities").Field("parent_parish_id").Unique().Required().Immutable(),
		edge.From("provider_profiles", ProviderProfile.Type).Ref("service_localities"),
	}
}

func (Locality) Annotations() []entschema.Annotation {
	return []entschema.Annotation{entsql.Annotation{Table: "localities"}}
}
