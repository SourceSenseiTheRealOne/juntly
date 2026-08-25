package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	entschema "entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type SpokenLanguageTranslation struct{ ent.Schema }

func (SpokenLanguageTranslation) Fields() []ent.Field {
	return []ent.Field{
		field.String("language_code").NotEmpty().MaxLen(10).Immutable(),
		field.String("locale").NotEmpty().MaxLen(10).Immutable(),
		field.String("name").NotEmpty().MaxLen(80),
	}
}

func (SpokenLanguageTranslation) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("language", SpokenLanguage.Type).Field("language_code").Unique().Required().Immutable(),
		edge.To("locale_record", SupportedLocale.Type).Field("locale").Unique().Required().Immutable(),
	}
}

func (SpokenLanguageTranslation) Annotations() []entschema.Annotation {
	return []entschema.Annotation{
		field.ID("language_code", "locale"),
		entsql.Annotation{Table: "spoken_language_translations"},
	}
}
