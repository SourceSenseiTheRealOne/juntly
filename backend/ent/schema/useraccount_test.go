package schema_test

import (
	"os"
	"strings"
	"testing"
)

func TestUserAccountSchemaContract(t *testing.T) {
	t.Parallel()

	contents, err := os.ReadFile("useraccount.go")
	if err != nil {
		t.Fatalf("read UserAccount schema: %v", err)
	}
	schema := string(contents)

	for _, requirement := range []string{
		"type UserAccount struct",
		"field.UUID(\"id\"",
		"StorageKey(\"internal_user_id\")",
		"Immutable()",
		"field.Bool(\"provider_enabled\")",
		"Default(false)",
		"field.Time(\"onboarding_completed_at\")",
		"field.Time(\"created_at\")",
		"field.Time(\"updated_at\")",
		"Table: \"user_accounts\"",
	} {
		if !strings.Contains(schema, requirement) {
			t.Errorf("schema does not contain %q", requirement)
		}
	}

	for _, prohibited := range []string{
		"clerk_subject",
		"email",
		"phone",
		"display_name",
		"profile",
		"role",
		"contact",
	} {
		if strings.Contains(schema, prohibited) {
			t.Errorf("schema must not include %q", prohibited)
		}
	}
}
