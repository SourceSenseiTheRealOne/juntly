package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	entschema "entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type SupportedLocale struct{ ent.Schema }

func (SupportedLocale) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").NotEmpty().MaxLen(10).Immutable(),
		field.Bool("active").Default(true),
		field.Int("sort_order").NonNegative(),
	}
}

func (SupportedLocale) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("translated_categories", ServiceCategory.Type).Ref("localized_in"),
		edge.From("translated_languages", SpokenLanguage.Type).Ref("localized_in"),
	}
}

func (SupportedLocale) Annotations() []entschema.Annotation {
	return []entschema.Annotation{entsql.Annotation{Table: "supported_locales"}}
}
