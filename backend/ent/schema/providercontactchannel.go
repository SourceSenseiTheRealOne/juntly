package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	entschema "entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

type ProviderContactChannel struct{ ent.Schema }

func (ProviderContactChannel) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable(),
		field.UUID("internal_user_id", uuid.UUID{}).Immutable(),
		field.Enum("channel").Values("phone", "whatsapp").Immutable(),
		field.Bytes("ciphertext").Immutable(),
		field.Bytes("nonce").Immutable(),
		field.String("key_version").NotEmpty().MaxLen(32).Immutable(),
		field.Bool("enabled").Default(false),
		field.Bool("reveal_consent").Default(false),
		field.Time("created_at").Default(utcNow).Immutable(),
		field.Time("updated_at").Default(utcNow).UpdateDefault(utcNow),
	}
}

func (ProviderContactChannel) Indexes() []ent.Index {
	return []ent.Index{index.Fields("internal_user_id", "channel").Unique()}
}

func (ProviderContactChannel) Annotations() []entschema.Annotation {
	return []entschema.Annotation{entsql.Annotation{Table: "provider_contact_channels"}}
}
