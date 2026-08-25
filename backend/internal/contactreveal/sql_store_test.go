package contactreveal

import (
	"context"
	"database/sql"
	"encoding/base64"
	"errors"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestSQLRevealStoreCreatesOneSameDayLeadEventAndLimitIncrement(t *testing.T) {
	database := openRevealDatabase(t)
	ctx := context.Background()
	customerID, listingID, day := seedRevealFixture(t, database)
	store := NewSQLRevealStore(database)
	first, err := store.AuthorizeAndReserve(ctx, customerID, listingID, ChannelPhone, day)
	if err != nil || len(first.Ciphertext) == 0 || len(first.Nonce) == 0 {
		t.Fatalf("first reservation = %#v, err = %v", first, err)
	}
	second, err := store.AuthorizeAndReserve(ctx, customerID, listingID, ChannelPhone, day)
	if err != nil || string(second.Ciphertext) != string(first.Ciphertext) || string(second.Nonce) != string(first.Nonce) {
		t.Fatalf("second reservation = %#v, err = %v", second, err)
	}
	var events, count int
	if err := database.QueryRowContext(ctx, `select count(*) from public.contact_reveal_events where customer_internal_user_id = $1 and listing_id = $2`, customerID, listingID).Scan(&events); err != nil {
		t.Fatalf("event count: %v", err)
	}
	if err := database.QueryRowContext(ctx, `select successful_count from public.contact_reveal_daily_limits where customer_internal_user_id = $1 and utc_day = $2`, customerID, day).Scan(&count); err != nil {
		t.Fatalf("daily count: %v", err)
	}
	if events != 1 || count != 1 {
		t.Fatalf("events/count = %d/%d", events, count)
	}
}

func TestSQLRevealStoreEnforcesDailyCapUnderConcurrency(t *testing.T) {
	database := openRevealDatabase(t)
	ctx := context.Background()
	customerID, firstListingID, day := seedRevealFixture(t, database)
	var providerID, categoryID, localityID uuid.UUID
	if err := database.QueryRowContext(ctx, `select internal_user_id, category_id, primary_locality_id from public.listings where id = $1`, firstListingID).Scan(&providerID, &categoryID, &localityID); err != nil {
		t.Fatalf("fixture listing: %v", err)
	}
	listingIDs := []uuid.UUID{firstListingID}
	for index := 1; index < 11; index++ {
		id := uuid.New()
		if _, err := database.ExecContext(ctx, `insert into public.listings (id, internal_user_id, category_id, primary_locality_id, title, description, price_type, price_minor, currency, travels_to_customer, state) values ($1, $2, $3, $4, $5, 'A synthetic active listing used only for contact reveal rate-limit tests.', 'fixed', 5000, 'EUR', true, 'active')`, id, providerID, categoryID, localityID, "Rate listing "+id.String()); err != nil {
			t.Fatalf("seed listing %d: %v", index, err)
		}
		listingIDs = append(listingIDs, id)
	}
	store := NewSQLRevealStore(database)
	start := make(chan struct{})
	errorsOut := make(chan error, len(listingIDs))
	var workers sync.WaitGroup
	for _, listingID := range listingIDs {
		listingID := listingID
		workers.Add(1)
		go func() {
			defer workers.Done()
			<-start
			_, err := store.AuthorizeAndReserve(ctx, customerID, listingID, ChannelPhone, day)
			errorsOut <- err
		}()
	}
	close(start)
	workers.Wait()
	close(errorsOut)
	successes, forbidden := 0, 0
	for err := range errorsOut {
		if err == nil {
			successes++
		} else if errors.Is(err, ErrForbidden) {
			forbidden++
		} else {
			t.Fatalf("reveal error: %v", err)
		}
	}
	if successes != 10 || forbidden != 1 {
		t.Fatalf("successes/forbidden = %d/%d", successes, forbidden)
	}
	var count, events int
	if err := database.QueryRowContext(ctx, `select successful_count from public.contact_reveal_daily_limits where customer_internal_user_id = $1 and utc_day = $2`, customerID, day).Scan(&count); err != nil {
		t.Fatalf("daily count: %v", err)
	}
	if err := database.QueryRowContext(ctx, `select count(*) from public.contact_reveal_events where customer_internal_user_id = $1 and utc_day = $2`, customerID, day).Scan(&events); err != nil {
		t.Fatalf("event count: %v", err)
	}
	if count != 10 || events != 10 {
		t.Fatalf("count/events = %d/%d", count, events)
	}
}

func openRevealDatabase(t *testing.T) *sql.DB {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL is required for PostgreSQL integration tests")
	}
	database, err := sql.Open("pgx", url)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })
	return database
}

func seedRevealFixture(t *testing.T, database *sql.DB) (uuid.UUID, uuid.UUID, time.Time) {
	t.Helper()
	ctx := context.Background()
	var categoryID, localityID uuid.UUID
	if err := database.QueryRowContext(ctx, `select id from public.service_categories where active order by id limit 1`).Scan(&categoryID); err != nil {
		t.Fatalf("category: %v", err)
	}
	if err := database.QueryRowContext(ctx, `select id from public.localities where active order by id limit 1`).Scan(&localityID); err != nil {
		t.Fatalf("locality: %v", err)
	}
	providerID, customerID, listingID := uuid.New(), uuid.New(), uuid.New()
	key := make([]byte, 32)
	for index := range key {
		key[index] = byte(index + 1)
	}
	cipher, err := NewCipher(base64.StdEncoding.EncodeToString(key))
	if err != nil {
		t.Fatalf("cipher: %v", err)
	}
	sealed, err := cipher.Encrypt([]byte("test-contact"))
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	tx, err := database.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin fixture: %v", err)
	}
	defer tx.Rollback()
	for _, user := range []struct {
		id      uuid.UUID
		subject string
	}{{providerID, "reveal_provider_" + providerID.String()}, {customerID, "reveal_customer_" + customerID.String()}} {
		if _, err := tx.ExecContext(ctx, `insert into public.internal_users (id, clerk_subject) values ($1, $2)`, user.id, user.subject); err != nil {
			t.Fatalf("user: %v", err)
		}
		if _, err := tx.ExecContext(ctx, `insert into public.user_accounts (internal_user_id, provider_enabled) values ($1, true)`, user.id); err != nil {
			t.Fatalf("account: %v", err)
		}
	}
	if _, err := tx.ExecContext(ctx, `insert into public.provider_profiles (internal_user_id, display_name, provider_type, bio, primary_locality_id, max_travel_distance_km, travels_to_customer) values ($1, 'Reveal provider', 'professional', 'Synthetic provider for reveal tests.', $2, 25, true)`, providerID, localityID); err != nil {
		t.Fatalf("profile: %v", err)
	}
	if _, err := tx.ExecContext(ctx, `insert into public.listings (id, internal_user_id, category_id, primary_locality_id, title, description, price_type, price_minor, currency, travels_to_customer, state) values ($1, $2, $3, $4, 'Reveal listing', 'A synthetic active listing used only for contact reveal persistence tests.', 'fixed', 5000, 'EUR', true, 'active')`, listingID, providerID, categoryID, localityID); err != nil {
		t.Fatalf("listing: %v", err)
	}
	if _, err := tx.ExecContext(ctx, `insert into public.provider_contact_channels (id, internal_user_id, channel, ciphertext, nonce, key_version, enabled, reveal_consent) values ($1, $2, 'phone', $3, $4, 'v1', true, true)`, uuid.New(), providerID, sealed.Ciphertext, sealed.Nonce); err != nil {
		t.Fatalf("channel: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit fixture: %v", err)
	}
	t.Cleanup(func() {
		if _, err := database.ExecContext(ctx, `delete from public.internal_users where id in ($1, $2)`, providerID, customerID); err != nil {
			t.Errorf("cleanup: %v", err)
		}
	})
	return customerID, listingID, time.Date(2026, 8, 24, 0, 0, 0, 0, time.UTC)
}
