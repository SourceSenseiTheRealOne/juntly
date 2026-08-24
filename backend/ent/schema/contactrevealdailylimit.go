package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	entschema "entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

type ContactRevealDailyLimit struct{ ent.Schema }

func (ContactRevealDailyLimit) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable(),
		field.UUID("customer_internal_user_id", uuid.UUID{}).Immutable(),
		field.Time("utc_day").Immutable(),
		field.Int("successful_count").Min(0).Default(0),
		field.Time("created_at").Default(utcNow).Immutable(),
		field.Time("updated_at").Default(utcNow).UpdateDefault(utcNow),
	}
}

func (ContactRevealDailyLimit) Indexes() []ent.Index {
	return []ent.Index{index.Fields("customer_internal_user_id", "utc_day").Unique()}
}

func (ContactRevealDailyLimit) Annotations() []entschema.Annotation {
	return []entschema.Annotation{entsql.Annotation{Table: "contact_reveal_daily_limits"}}
}
