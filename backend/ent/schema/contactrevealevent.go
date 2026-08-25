package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	entschema "entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

type ContactRevealEvent struct{ ent.Schema }

func (ContactRevealEvent) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable(),
		field.UUID("customer_internal_user_id", uuid.UUID{}).Immutable(),
		field.UUID("provider_internal_user_id", uuid.UUID{}).Immutable(),
		field.UUID("listing_id", uuid.UUID{}).Immutable(),
		field.Enum("channel").Values("phone", "whatsapp").Immutable(),
		field.Time("utc_day").Immutable(),
		field.Time("revealed_at").Default(utcNow).Immutable(),
	}
}

func (ContactRevealEvent) Indexes() []ent.Index {
	return []ent.Index{index.Fields("customer_internal_user_id", "listing_id", "channel", "utc_day").Unique()}
}

func (ContactRevealEvent) Annotations() []entschema.Annotation {
	return []entschema.Annotation{entsql.Annotation{Table: "contact_reveal_events"}}
}
