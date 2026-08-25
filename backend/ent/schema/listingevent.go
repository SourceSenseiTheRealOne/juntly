package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	entschema "entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

type ListingEvent struct{ ent.Schema }

func (ListingEvent) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable(),
		field.UUID("listing_id", uuid.UUID{}).Immutable(),
		field.UUID("actor_internal_user_id", uuid.UUID{}).Immutable(),
		field.Enum("event_type").Values("created", "updated", "submitted", "approved", "rejected", "paused", "archived").Immutable(),
		field.String("from_state").Optional().Nillable().MaxLen(32).Immutable(),
		field.String("to_state").NotEmpty().MaxLen(32).Immutable(),
		field.Int("revision").Min(1).Immutable(),
		field.String("reason").Optional().Nillable().MaxLen(500).Immutable(),
		field.Time("created_at").Default(utcNow).Immutable(),
	}
}

func (ListingEvent) Annotations() []entschema.Annotation {
	return []entschema.Annotation{entsql.Annotation{Table: "listing_events"}}
}
