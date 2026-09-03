package payments

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPaymentMigrationDefinesDurableMoneyAndWebhookInvariants(t *testing.T) {
	migration := paymentMigration(t)
	for _, requirement := range []string{
		"create table public.provider_payment_accounts",
		"stripe_account_id text not null unique",
		"create table public.payment_orders",
		"gross_minor integer not null",
		"platform_fee_minor integer not null",
		"provider_net_minor integer not null",
		"unique(booking_id)",
		"stripe_checkout_session_id text unique",
		"stripe_payment_intent_id text unique",
		"stripe_invoice_id text",
		"create table public.payment_events",
		"create table public.stripe_webhook_receipts",
		"stripe_event_id text primary key",
		"create table public.payment_disputes",
	} {
		if !strings.Contains(migration, requirement) {
			t.Errorf("payment migration does not contain %q", requirement)
		}
	}
	for _, prohibited := range []string{"card_number", "bank_account", "raw_payload", "webhook_signature"} {
		if strings.Contains(migration, prohibited) {
			t.Errorf("payment migration must not persist %q", prohibited)
		}
	}
}

func paymentMigration(t *testing.T) string {
	t.Helper()
	directory := filepath.Join("..", "..", "..", "supabase", "migrations")
	entries, err := os.ReadDir(directory)
	if err != nil {
		t.Fatalf("read migrations: %v", err)
	}
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), "_create_payments.sql") {
			contents, err := os.ReadFile(filepath.Join(directory, entry.Name()))
			if err != nil {
				t.Fatalf("read payment migration: %v", err)
			}
			return strings.ToLower(string(contents))
		}
	}
	t.Fatal("payment migration not found")
	return ""
}
