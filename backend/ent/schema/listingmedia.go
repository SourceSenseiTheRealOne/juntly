package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	entschema "entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

type ListingMedia struct{ ent.Schema }

func (ListingMedia) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable(),
		field.UUID("listing_id", uuid.UUID{}).Immutable(),
		field.Int("ordinal").Min(1),
		field.String("content_type").NotEmpty().MaxLen(100),
		field.Int64("byte_size").Positive(),
		field.String("checksum_sha256").NotEmpty().MaxLen(64).Immutable(),
		field.String("object_reference").NotEmpty().MaxLen(512).Immutable(),
		field.Enum("state").Values("pending_upload", "ready", "deleted").Default("pending_upload"),
		field.Time("created_at").Default(utcNow).Immutable(),
		field.Time("updated_at").Default(utcNow).UpdateDefault(utcNow),
	}
}

func (ListingMedia) Annotations() []entschema.Annotation {
	return []entschema.Annotation{entsql.Annotation{Table: "listing_media"}}
}
