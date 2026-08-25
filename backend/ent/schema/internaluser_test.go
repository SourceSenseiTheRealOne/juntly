package schema_test

import (
	"os"
	"strings"
	"testing"
)

func TestInternalUserSchemaContract(t *testing.T) {
	t.Parallel()

	contents, err := os.ReadFile("internaluser.go")
	if err != nil {
		t.Fatalf("read InternalUser schema: %v", err)
	}
	schema := string(contents)

	for _, requirement := range []string{
		"type InternalUser struct",
		"field.UUID(\"id\"",
		"Default(uuid.New)",
		"field.String(\"clerk_subject\")",
		"NotEmpty()",
		"MaxLen(255)",
		"Unique()",
		"Immutable()",
		"field.Time(\"created_at\")",
		"field.Time(\"updated_at\")",
	} {
		if !strings.Contains(schema, requirement) {
			t.Errorf("schema does not contain %q", requirement)
		}
	}

	for _, prohibited := range []string{"email", "display_name", "profile"} {
		if strings.Contains(schema, prohibited) {
			t.Errorf("schema must not include %q", prohibited)
		}
	}
}
