package schema_test

import (
	"os"
	"strings"
	"testing"
)

func TestContactRevealSchemasUseSeparateEncryptedVaultAndLeadState(t *testing.T) {
	t.Parallel()
	contracts := map[string][]string{
		"providercontactchannel.go": {
			"type ProviderContactChannel struct",
			"field.UUID(\"internal_user_id\"",
			"field.Enum(\"channel\").Values(\"phone\", \"whatsapp\")",
			"field.Bytes(\"ciphertext\")",
			"field.Bytes(\"nonce\")",
			"field.String(\"key_version\")",
			"field.Bool(\"enabled\")",
			"field.Bool(\"reveal_consent\")",
			"index.Fields(\"internal_user_id\", \"channel\").Unique()",
			"Table: \"provider_contact_channels\"",
		},
		"contactrevealdailylimit.go": {
			"type ContactRevealDailyLimit struct",
			"field.UUID(\"customer_internal_user_id\"",
			"field.Time(\"utc_day\")",
			"field.Int(\"successful_count\").Min(0)",
			"index.Fields(\"customer_internal_user_id\", \"utc_day\").Unique()",
			"Table: \"contact_reveal_daily_limits\"",
		},
		"contactrevealevent.go": {
			"type ContactRevealEvent struct",
			"field.UUID(\"customer_internal_user_id\"",
			"field.UUID(\"provider_internal_user_id\"",
			"field.UUID(\"listing_id\"",
			"field.Enum(\"channel\").Values(\"phone\", \"whatsapp\")",
			"field.Time(\"utc_day\")",
			"field.Time(\"revealed_at\")",
			"index.Fields(\"customer_internal_user_id\", \"listing_id\", \"channel\", \"utc_day\").Unique()",
			"Table: \"contact_reveal_events\"",
		},
	}
	for path, required := range contracts {
		contents, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		for _, value := range required {
			if !strings.Contains(string(contents), value) {
				t.Errorf("%s missing %q", path, value)
			}
		}
	}
}
