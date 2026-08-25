package contactreveal

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestContactRevealMigrationContract(t *testing.T) {
	t.Parallel()
	directory := filepath.Join("..", "..", "..", "supabase", "migrations")
	entries, err := os.ReadDir(directory)
	if err != nil {
		t.Fatalf("read migrations: %v", err)
	}
	var migration string
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), "_create_contact_reveals.sql") {
			contents, readErr := os.ReadFile(filepath.Join(directory, entry.Name()))
			if readErr != nil {
				t.Fatalf("read migration: %v", readErr)
			}
			migration = strings.ToLower(string(contents))
		}
	}
	if migration == "" {
		t.Fatal("contact reveal migration not found")
	}
	for _, requirement := range []string{
		"create table public.provider_contact_channels",
		"ciphertext bytea not null",
		"nonce bytea not null",
		"key_version text not null",
		"channel in ('phone', 'whatsapp')",
		"unique (internal_user_id, channel)",
		"create table public.contact_reveal_daily_limits",
		"unique (customer_internal_user_id, utc_day)",
		"create table public.contact_reveal_events",
		"unique (customer_internal_user_id, listing_id, channel, utc_day)",
		"create index contact_reveal_events_listing_created_idx",
		"create index provider_contact_channels_owner_idx",
	} {
		if !strings.Contains(migration, requirement) {
			t.Errorf("migration missing %q", requirement)
		}
	}
	for _, prohibited := range []string{"phone_number", "whatsapp_number", "contact_value", "plaintext"} {
		if strings.Contains(migration, prohibited) {
			t.Errorf("migration must not persist %q", prohibited)
		}
	}
}
