package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	entschema "entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

type Listing struct{ ent.Schema }

func (Listing) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable(),
		field.UUID("internal_user_id", uuid.UUID{}).Immutable(),
		field.UUID("category_id", uuid.UUID{}),
		field.UUID("primary_locality_id", uuid.UUID{}),
		field.String("title").NotEmpty().MaxLen(140),
		field.String("description").NotEmpty().MaxLen(4000),
		field.Enum("price_type").Values("fixed", "hourly", "daily", "quote", "negotiable"),
		field.Int("price_minor").Optional().Nillable().Positive(),
		field.String("currency").Default("EUR").Immutable().MaxLen(3),
		field.Bool("travels_to_customer").Default(false),
		field.Bool("receives_customer").Default(false),
		field.Bool("remote_services").Default(false),
		field.Enum("state").Values("draft", "pending_review", "active", "rejected", "paused", "archived").Default("draft"),
		field.Int("revision").Default(1).Min(1),
		field.Time("created_at").Default(utcNow).Immutable(),
		field.Time("updated_at").Default(utcNow).UpdateDefault(utcNow),
	}
}

func (Listing) Annotations() []entschema.Annotation {
	return []entschema.Annotation{entsql.Annotation{Table: "listings"}}
}
