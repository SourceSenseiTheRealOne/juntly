package discovery

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestSQLRepositoryOrdersRadiusThenUpdatedThenStableID(t *testing.T) {
	database := openDiscoveryDatabase(t)
	ctx := context.Background()
	categoryID, _ := seededDiscoveryReferences(t, database)
	var parishID uuid.UUID
	if err := database.QueryRow(`select parent_parish_id from public.localities where active order by id limit 1`).Scan(&parishID); err != nil {
		t.Fatalf("seeded parish: %v", err)
	}
	ownerID := uuid.New()
	originID := uuid.MustParse("00000000-0000-4000-8000-000000000101")
	nearID := uuid.MustParse("00000000-0000-4000-8000-000000000102")
	originListingID := uuid.MustParse("00000000-0000-4000-8000-000000000201")
	nearListingID := uuid.MustParse("00000000-0000-4000-8000-000000000202")
	tiedNearListingID := uuid.MustParse("00000000-0000-4000-8000-000000000203")
	fixedUpdatedAt := time.Date(2026, 8, 24, 12, 0, 0, 0, time.UTC)
	tx, err := database.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin radius seed: %v", err)
	}
	defer tx.Rollback()
	for _, locality := range []struct {
		id        uuid.UUID
		slug      string
		longitude float64
	}{
		{originID, "discovery-radius-origin", -7.0},
		{nearID, "discovery-radius-near", -7.01},
	} {
		if _, err := tx.ExecContext(ctx, `
			insert into public.localities
			  (id, slug, name, parent_parish_id, source, source_element_id, source_version, source_retrieved_at, latitude, longitude, active)
			values ($1, $2, $2, $3, 'synthetic-test', $4, '1', '2026-08-24T00:00:00Z', 39.8, $5, true)
		`, locality.id, locality.slug, parishID, "D:"+locality.slug, locality.longitude); err != nil {
			t.Fatalf("seed radius locality: %v", err)
		}
	}
	if _, err := tx.ExecContext(ctx, `insert into public.internal_users (id, clerk_subject) values ($1, $2)`, ownerID, "discovery_radius_"+ownerID.String()); err != nil {
		t.Fatalf("seed radius owner: %v", err)
	}
	if _, err := tx.ExecContext(ctx, `insert into public.user_accounts (internal_user_id, provider_enabled) values ($1, true)`, ownerID); err != nil {
		t.Fatalf("seed radius account: %v", err)
	}
	if _, err := tx.ExecContext(ctx, `
		insert into public.provider_profiles
		  (internal_user_id, display_name, provider_type, bio, primary_locality_id, max_travel_distance_km, travels_to_customer)
		values ($1, 'Radius provider', 'professional', 'Radius discovery provider profile.', $2, 25, true)
	`, ownerID, originID); err != nil {
		t.Fatalf("seed radius profile: %v", err)
	}
	for _, listing := range []struct {
		id       uuid.UUID
		locality uuid.UUID
	}{
		{originListingID, originID},
		{nearListingID, nearID},
		{tiedNearListingID, nearID},
	} {
		if _, err := tx.ExecContext(ctx, `
			insert into public.listings
			  (id, internal_user_id, category_id, primary_locality_id, title, description, price_type, price_minor, currency, travels_to_customer, state, updated_at)
			values ($1, $2, $3, $4, 'Radius plumbing', 'A public listing used to prove deterministic radius ordering.', 'fixed', 5000, 'EUR', true, 'active', $5)
		`, listing.id, ownerID, categoryID, listing.locality, fixedUpdatedAt); err != nil {
			t.Fatalf("seed radius listing: %v", err)
		}
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit radius seed: %v", err)
	}
	t.Cleanup(func() {
		if _, err := database.ExecContext(ctx, "delete from public.internal_users where id = $1", ownerID); err != nil {
			t.Errorf("cleanup radius owner: %v", err)
		}
		if _, err := database.ExecContext(ctx, "delete from public.localities where id in ($1, $2)", originID, nearID); err != nil {
			t.Errorf("cleanup radius localities: %v", err)
		}
	})

	values, err := NewSQLRepository(database).Search(ctx, Request{Locale: "pt-PT", Query: "radius", NearLocalityID: originID, RadiusKM: 2})
	if err != nil || len(values) != 3 || values[0].ID != originListingID || values[1].ID != nearListingID || values[2].ID != tiedNearListingID {
		t.Fatalf("radius ordering values/error = %#v/%v", values, err)
	}
}
