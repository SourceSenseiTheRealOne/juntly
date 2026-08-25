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

type ServiceCategory struct{ ent.Schema }

func (ServiceCategory) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable(),
		field.UUID("parent_id", uuid.UUID{}).Optional().Nillable(),
		field.String("slug").NotEmpty().MaxLen(80).Unique().Immutable(),
		field.Bool("active").Default(true),
		field.Int("sort_order").NonNegative(),
		field.Time("created_at").Default(utcNow).Immutable(),
		field.Time("updated_at").Default(utcNow).UpdateDefault(utcNow),
	}
}

func (ServiceCategory) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("children", ServiceCategory.Type),
		edge.From("parent", ServiceCategory.Type).Ref("children").Field("parent_id").Unique(),
		edge.To("localized_in", SupportedLocale.Type).Through("translations", ServiceCategoryTranslation.Type),
	}
}

func (ServiceCategory) Indexes() []ent.Index {
	return []ent.Index{index.Fields("parent_id", "sort_order")}
}

func (ServiceCategory) Annotations() []entschema.Annotation {
	return []entschema.Annotation{entsql.Annotation{Table: "service_categories"}}
}
