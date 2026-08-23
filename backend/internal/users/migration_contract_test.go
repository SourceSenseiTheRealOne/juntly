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

func TestCreateUserAccountsMigrationContract(t *testing.T) {
	t.Parallel()

	migrationDirectory := filepath.Join("..", "..", "..", "supabase", "migrations")
	entries, err := os.ReadDir(migrationDirectory)
	if err != nil {
		t.Fatalf("read migration directory: %v", err)
	}

	var migrationPath string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), "_create_user_accounts.sql") {
			migrationPath = filepath.Join(migrationDirectory, entry.Name())
			break
		}
	}
	if migrationPath == "" {
		t.Fatal("user-accounts migration was not found")
	}

	contents, err := os.ReadFile(migrationPath)
	if err != nil {
		t.Fatalf("read migration: %v", err)
	}
	migration := strings.ToLower(string(contents))

	for _, requirement := range []string{
		"create table public.user_accounts",
		"internal_user_id uuid primary key",
		"references public.internal_users(id) on delete cascade",
		"provider_enabled boolean not null default false",
		"onboarding_completed_at timestamptz not null",
		"created_at timestamptz not null",
		"updated_at timestamptz not null",
	} {
		if !strings.Contains(migration, requirement) {
			t.Errorf("migration does not contain %q", requirement)
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
		if strings.Contains(migration, prohibited) {
			t.Errorf("migration must not include %q", prohibited)
		}
	}
}
