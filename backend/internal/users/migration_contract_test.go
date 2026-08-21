package users_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCreateInternalUsersMigrationContract(t *testing.T) {
	t.Parallel()

	migrationDirectory := filepath.Join("..", "..", "..", "supabase", "migrations")
	entries, err := os.ReadDir(migrationDirectory)
	if err != nil {
		t.Fatalf("read migration directory: %v", err)
	}

	var migrationPath string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), "_create_internal_users.sql") {
			migrationPath = filepath.Join(migrationDirectory, entry.Name())
			break
		}
	}
	if migrationPath == "" {
		t.Fatal("internal-users migration was not found")
	}

	contents, err := os.ReadFile(migrationPath)
	if err != nil {
		t.Fatalf("read migration: %v", err)
	}
	migration := strings.ToLower(string(contents))

	for _, requirement := range []string{
		"create table public.internal_users",
		"id uuid primary key",
		"clerk_subject text not null unique",
		"created_at timestamptz not null",
		"updated_at timestamptz not null",
		"char_length(clerk_subject) between 1 and 255",
	} {
		if !strings.Contains(migration, requirement) {
			t.Errorf("migration does not contain %q", requirement)
		}
	}
}
