package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	entschema "entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

type AdministrativeArea struct{ ent.Schema }

func (AdministrativeArea) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable(),
		field.String("source").NotEmpty().MaxLen(40).Immutable(),
		field.String("source_version").NotEmpty().MaxLen(20).Immutable(),
		field.String("external_code").NotEmpty().MaxLen(32).Immutable(),
		field.String("kind").NotEmpty().MaxLen(20).Immutable(),
		field.String("name").NotEmpty().MaxLen(160),
		field.UUID("parent_id", uuid.UUID{}).Optional().Nillable().Immutable(),
		field.Bool("active").Default(true),
		field.Time("created_at").Default(utcNow).Immutable(),
		field.Time("updated_at").Default(utcNow).UpdateDefault(utcNow),
	}
}

func (AdministrativeArea) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("children", AdministrativeArea.Type),
		edge.From("parent", AdministrativeArea.Type).Ref("children").Field("parent_id").Unique().Immutable(),
		edge.To("localities", Locality.Type),
	}
}

func (AdministrativeArea) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("source", "external_code").Unique(),
		index.Fields("parent_id", "kind"),
	}
}

func (AdministrativeArea) Annotations() []entschema.Annotation {
	return []entschema.Annotation{entsql.Annotation{Table: "administrative_areas"}}
}
