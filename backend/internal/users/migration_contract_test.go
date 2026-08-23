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

func TestCreateTaxonomyLocationsProviderProfilesMigrationContract(t *testing.T) {
	t.Parallel()

	migrationDirectory := filepath.Join("..", "..", "..", "supabase", "migrations")
	entries, err := os.ReadDir(migrationDirectory)
	if err != nil {
		t.Fatalf("read migration directory: %v", err)
	}

	var migrationPath string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), "_create_taxonomy_locations_provider_profiles.sql") {
			migrationPath = filepath.Join(migrationDirectory, entry.Name())
			break
		}
	}
	if migrationPath == "" {
		t.Fatal("taxonomy, locations, and provider-profiles migration was not found")
	}

	contents, err := os.ReadFile(migrationPath)
	if err != nil {
		t.Fatalf("read migration: %v", err)
	}
	migration := strings.ToLower(string(contents))

	for _, requirement := range []string{
		"create extension if not exists postgis",
		"create table public.supported_locales",
		"create table public.service_categories",
		"create table public.service_category_translations",
		"primary key (category_id, locale)",
		"create table public.spoken_languages",
		"create table public.spoken_language_translations",
		"primary key (language_code, locale)",
		"create table public.administrative_areas",
		"create table public.localities",
		"center geography(point, 4326) generated always as",
		"st_setsrid(st_makepoint(longitude, latitude), 4326)::geography",
		"create table public.provider_profiles",
		"internal_user_id uuid primary key",
		"references public.user_accounts(internal_user_id) on delete cascade",
		"provider_type in ('individual', 'professional', 'business')",
		"max_travel_distance_km between 0 and 200",
		"travels_to_customer or receives_customer or remote_services",
		"create table public.provider_service_localities",
		"primary key (internal_user_id, locality_id)",
		"create table public.provider_spoken_languages",
		"primary key (internal_user_id, language_code)",
		"'050205'",
		"'050510'",
		"'050518'",
		"'050520'",
		"'050521'",
		"'r5396187'",
		"'r5395738'",
		"'n371426674'",
		"'r5431477'",
		"'n440173641'",
		"'home-repairs'",
		"'computer-repair'",
	} {
		if !strings.Contains(migration, requirement) {
			t.Errorf("migration does not contain %q", requirement)
		}
	}

	for _, prohibited := range []string{
		"clerk_subject",
		"email",
		"phone",
		"whatsapp",
		"exact_address",
		"payment",
		"identity_document",
	} {
		if strings.Contains(migration, prohibited) {
			t.Errorf("migration must not include %q", prohibited)
		}
	}
}
