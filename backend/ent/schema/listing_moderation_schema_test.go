package schema_test

import (
	"os"
	"strings"
	"testing"
)

func TestListingModerationSchemas(t *testing.T) {
	t.Parallel()

	contracts := map[string][]string{
		"platformrole.go": {
			"type PlatformRole struct",
			"field.UUID(\"id\"",
			"Default(uuid.New)",
			"field.UUID(\"internal_user_id\"",
			"field.String(\"role\").NotEmpty().MaxLen(20)",
			"field.Time(\"granted_at\")",
			"index.Fields(\"internal_user_id\", \"role\").Unique()",
			"Table: \"platform_roles\"",
		},
		"listing.go": {
			"type Listing struct",
			"field.UUID(\"id\"",
			"Default(uuid.New)",
			"field.UUID(\"internal_user_id\"",
			"field.UUID(\"category_id\"",
			"field.UUID(\"primary_locality_id\"",
			"field.String(\"title\")",
			"field.String(\"description\")",
			"field.Enum(\"price_type\").Values(\"fixed\", \"hourly\", \"daily\", \"quote\", \"negotiable\")",
			"field.Int(\"price_minor\")",
			"Optional()",
			"Nillable()",
			"field.String(\"currency\").Default(\"EUR\")",
			"field.Bool(\"travels_to_customer\")",
			"field.Bool(\"receives_customer\")",
			"field.Bool(\"remote_services\")",
			"field.Enum(\"state\").Values(\"draft\", \"pending_review\", \"active\", \"rejected\", \"paused\", \"archived\")",
			"field.Int(\"revision\").Default(1)",
			"Table: \"listings\"",
		},
		"listingevent.go": {
			"type ListingEvent struct",
			"field.UUID(\"id\"",
			"field.UUID(\"listing_id\"",
			"field.UUID(\"actor_internal_user_id\"",
			"field.Enum(\"event_type\").Values(\"created\", \"updated\", \"submitted\", \"approved\", \"rejected\", \"paused\", \"archived\")",
			"field.String(\"from_state\")",
			"field.String(\"to_state\")",
			"field.Int(\"revision\")",
			"field.String(\"reason\")",
			"Table: \"listing_events\"",
		},
		"listingmedia.go": {
			"type ListingMedia struct",
			"field.UUID(\"id\"",
			"field.UUID(\"listing_id\"",
			"field.Int(\"ordinal\")",
			"field.String(\"content_type\")",
			"field.Int64(\"byte_size\")",
			"field.String(\"checksum_sha256\")",
			"field.String(\"object_reference\")",
			"field.Enum(\"state\").Values(\"pending_upload\", \"ready\", \"deleted\")",
			"Table: \"listing_media\"",
		},
	}

	for path, required := range contracts {
		path, required := path, required
		t.Run(path, func(t *testing.T) {
			t.Parallel()
			contents, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read schema %s: %v", path, err)
			}
			for _, requirement := range required {
				if !strings.Contains(string(contents), requirement) {
					t.Errorf("schema %s does not contain %q", path, requirement)
				}
			}
		})
	}
}

func TestListingSchemasExcludePublicContactAndStorageAuthority(t *testing.T) {
	t.Parallel()
	for _, path := range []string{"listing.go", "listingevent.go", "listingmedia.go"} {
		contents, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		value := strings.ToLower(string(contents))
		for _, prohibited := range []string{"email", "phone", "whatsapp", "address", "clerk", "token", "session", "storage_secret", "access_key", "bucket"} {
			if strings.Contains(value, prohibited) {
				t.Errorf("%s must not persist %q", path, prohibited)
			}
		}
	}
}
