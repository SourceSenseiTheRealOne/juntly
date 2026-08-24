package discovery

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"testing"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestSQLRepositoryProjectsOnlyActiveListingsWithLocalizedRadiusSearch(t *testing.T) {
	database := openDiscoveryDatabase(t)
	ctx := context.Background()
	categoryID, localityID := seededDiscoveryReferences(t, database)
	ownerID, activeID, draftID := uuid.New(), uuid.New(), uuid.New()
	tx, err := database.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin discovery seed: %v", err)
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `insert into public.internal_users (id, clerk_subject) values ($1, $2)`, ownerID, "discovery_provider_"+ownerID.String()); err != nil {
		t.Fatalf("seed discovery owner: %v", err)
	}
	if _, err := tx.ExecContext(ctx, `insert into public.user_accounts (internal_user_id, provider_enabled) values ($1, true)`, ownerID); err != nil {
		t.Fatalf("seed discovery account: %v", err)
	}
	if _, err := tx.ExecContext(ctx, `
		insert into public.provider_profiles
		  (internal_user_id, display_name, provider_type, bio, primary_locality_id, max_travel_distance_km, travels_to_customer)
		values ($1, 'Public provider', 'professional', 'Discovery test provider profile.', $2, 25, true)
	`, ownerID, localityID); err != nil {
		t.Fatalf("seed discovery profile: %v", err)
	}
	if _, err := tx.ExecContext(ctx, `
		insert into public.listings
		  (id, internal_user_id, category_id, primary_locality_id, title, description, price_type, price_minor, currency, travels_to_customer, state)
		values
		  ($1, $2, $3, $4, 'Discovery active plumbing', 'A public plumbing listing for active discovery testing.', 'fixed', 5000, 'EUR', true, 'active'),
		  ($5, $2, $3, $4, 'Discovery private draft', 'A draft listing that must never appear in public discovery.', 'fixed', 5000, 'EUR', true, 'draft')
	`, activeID, ownerID, categoryID, localityID, draftID); err != nil {
		t.Fatalf("seed discovery listings: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit discovery seed: %v", err)
	}
	t.Cleanup(func() {
		if _, err := database.ExecContext(ctx, "delete from public.internal_users where id = $1", ownerID); err != nil {
			t.Errorf("cleanup discovery owner: %v", err)
		}
	})

	repository := NewSQLRepository(database)
	values, err := repository.Search(ctx, Request{Locale: "pt-PT", Query: "plumbing", NearLocalityID: localityID, RadiusKM: 1, PriceType: PriceTypeFixed, ServiceMode: ServiceModeTravelsToCustomer})
	if err != nil || len(values) != 1 || values[0].ID != activeID {
		t.Fatalf("public search values/error = %#v/%v", values, err)
	}
	value := values[0]
	if value.CategoryID != categoryID || value.PrimaryLocalityID != localityID || value.CategoryName == "" || value.LocalityName == "" || value.ProviderDisplayName != "Public provider" || value.PriceMinor == nil || *value.PriceMinor != 5000 {
		t.Fatalf("incomplete public projection = %#v", value)
	}
	found, err := repository.Get(ctx, activeID, "pt-PT")
	if err != nil || found == nil || found.ID != activeID {
		t.Fatalf("active public detail = %#v/%v", found, err)
	}
	missing, err := repository.Get(ctx, draftID, "pt-PT")
	if missing != nil || !errors.Is(err, ErrNotFound) {
		t.Fatalf("draft public detail = %#v/%v", missing, err)
	}
}

func openDiscoveryDatabase(t *testing.T) *sql.DB {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL is required for PostgreSQL integration tests")
	}
	database, err := sql.Open("pgx", url)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	t.Cleanup(func() {
		if err := database.Close(); err != nil {
			t.Errorf("close database: %v", err)
		}
	})
	return database
}

func seededDiscoveryReferences(t *testing.T, database *sql.DB) (uuid.UUID, uuid.UUID) {
	t.Helper()
	var categoryID, localityID uuid.UUID
	if err := database.QueryRow(`select id from public.service_categories where active order by id limit 1`).Scan(&categoryID); err != nil {
		t.Fatalf("seeded category: %v", err)
	}
	if err := database.QueryRow(`select id from public.localities where active order by id limit 1`).Scan(&localityID); err != nil {
		t.Fatalf("seeded locality: %v", err)
	}
	return categoryID, localityID
}
