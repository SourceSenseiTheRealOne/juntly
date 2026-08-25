package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	entschema "entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type SpokenLanguage struct{ ent.Schema }

func (SpokenLanguage) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").NotEmpty().MaxLen(10).Immutable(),
		field.Bool("active").Default(true),
		field.Int("sort_order").NonNegative(),
	}
}

func (SpokenLanguage) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("localized_in", SupportedLocale.Type).Through("translations", SpokenLanguageTranslation.Type),
		edge.From("provider_profiles", ProviderProfile.Type).Ref("spoken_languages"),
	}
}

func (SpokenLanguage) Annotations() []entschema.Annotation {
	return []entschema.Annotation{entsql.Annotation{Table: "spoken_languages"}}
}
