package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	entschema "entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

type ServiceCategoryTranslation struct{ ent.Schema }

func (ServiceCategoryTranslation) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("category_id", uuid.UUID{}).Immutable(),
		field.String("locale").NotEmpty().MaxLen(10).Immutable(),
		field.String("name").NotEmpty().MaxLen(120),
		field.String("description").MaxLen(500).Optional().Nillable(),
	}
}

func (ServiceCategoryTranslation) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("category", ServiceCategory.Type).Field("category_id").Unique().Required().Immutable(),
		edge.To("locale_record", SupportedLocale.Type).Field("locale").Unique().Required().Immutable(),
	}
}

func (ServiceCategoryTranslation) Annotations() []entschema.Annotation {
	return []entschema.Annotation{
		field.ID("category_id", "locale"),
		entsql.Annotation{Table: "service_category_translations"},
	}
}
