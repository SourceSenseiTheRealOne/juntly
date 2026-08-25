package reference

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestSQLRepositoryReturnsCompleteLocalizedReferenceCatalog(t *testing.T) {
	database := openReferenceDatabase(t)
	repository := NewSQLRepository(database)
	ctx := context.Background()

	categories, err := repository.Categories(ctx, "pt-PT")
	if err != nil {
		t.Fatalf("categories: %v", err)
	}
	if len(categories) != 18 {
		t.Fatalf("category count = %d, want 18", len(categories))
	}
	if categories[0].Slug != "home-repairs" || categories[0].ParentID != nil || categories[0].Name != "Reparações domésticas" {
		t.Fatalf("first category = %#v", categories[0])
	}
	seen := make(map[uuid.UUID]struct{}, len(categories))
	for _, category := range categories {
		if _, exists := seen[category.ID]; exists {
			t.Fatalf("duplicate category ID %s", category.ID)
		}
		seen[category.ID] = struct{}{}
	}

	languages, err := repository.Languages(ctx, "en")
	if err != nil {
		t.Fatalf("languages: %v", err)
	}
	if len(languages) != 3 || languages[0].Code != "pt-PT" || languages[0].Name != "Portuguese" {
		t.Fatalf("languages = %#v", languages)
	}

	localities, err := repository.Localities(ctx, "es")
	if err != nil {
		t.Fatalf("localities: %v", err)
	}
	if len(localities) != 5 {
		t.Fatalf("locality count = %d, want 5", len(localities))
	}
	for _, locality := range localities {
		if locality.ID == uuid.Nil || locality.Name == "" || locality.ParishName == "" || locality.MunicipalityName == "" || locality.DistrictName == "" {
			t.Fatalf("incomplete locality = %#v", locality)
		}
	}
}

func TestSQLRepositoryRadiusOrderingAndActiveFiltering(t *testing.T) {
	database := openReferenceDatabase(t)
	ctx := context.Background()
	parentID := seededParishID(t, database, "050205")
	ids := []uuid.UUID{
		uuid.MustParse("00000000-0000-4000-8000-000000000010"),
		uuid.MustParse("00000000-0000-4000-8000-000000000011"),
		uuid.MustParse("00000000-0000-4000-8000-000000000012"),
		uuid.MustParse("00000000-0000-4000-8000-000000000013"),
		uuid.MustParse("00000000-0000-4000-8000-000000000014"),
	}
	rows := []struct {
		id        uuid.UUID
		slug      string
		latitude  float64
		longitude float64
		active    bool
	}{
		{ids[0], "test-radius-origin", 39.8, -7.0, true},
		{ids[1], "test-radius-near-a", 39.8, -7.01, true},
		{ids[2], "test-radius-near-b", 39.8, -6.99, true},
		{ids[3], "test-radius-inactive", 39.8, -7.005, false},
		{ids[4], "test-radius-out", 39.8, -7.1, true},
	}
	for _, row := range rows {
		_, err := database.ExecContext(ctx, `
			insert into public.localities
			  (id, slug, name, parent_parish_id, source, source_element_id, source_version, source_retrieved_at, latitude, longitude, active)
			values ($1, $2, $3, $4, 'synthetic-test', $5, '1', '2026-08-23T00:00:00Z', $6, $7, $8)
		`, row.id, row.slug, row.slug, parentID, "T:"+row.slug, row.latitude, row.longitude, row.active)
		if err != nil {
			t.Fatalf("insert locality %s: %v", row.slug, err)
		}
	}
	t.Cleanup(func() {
		for _, id := range ids {
			if _, err := database.ExecContext(ctx, "delete from public.localities where id = $1", id); err != nil {
				t.Errorf("cleanup locality: %v", err)
			}
		}
	})

	nearby, err := NewSQLRepository(database).NearbyLocalities(ctx, ids[0], 2, "pt-PT")
	if err != nil {
		t.Fatalf("nearby localities: %v", err)
	}
	if len(nearby) != 3 {
		t.Fatalf("nearby count = %d, want 3 (%#v)", len(nearby), nearby)
	}
	if nearby[0].ID != ids[0] || nearby[0].DistanceMeters != 0 {
		t.Fatalf("origin result = %#v", nearby[0])
	}
	if nearby[1].ID != ids[1] || nearby[2].ID != ids[2] {
		t.Fatalf("equal-distance ordering = %#v", nearby)
	}
	for _, result := range nearby {
		if result.ID == ids[3] || result.ID == ids[4] {
			t.Fatalf("inactive/out-of-range locality included: %#v", result)
		}
	}

	_, err = NewSQLRepository(database).NearbyLocalities(ctx, uuid.New(), 2, "pt-PT")
	if err != ErrNotFound {
		t.Fatalf("missing origin error = %v, want ErrNotFound", err)
	}
}

func openReferenceDatabase(t *testing.T) *sql.DB {
	t.Helper()
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is required for PostgreSQL integration tests")
	}
	database, err := sql.Open("pgx", databaseURL)
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

func seededParishID(t *testing.T, database *sql.DB, code string) uuid.UUID {
	t.Helper()
	var id uuid.UUID
	if err := database.QueryRow("select id from public.administrative_areas where source = 'caop' and external_code = $1", code).Scan(&id); err != nil {
		t.Fatalf("load seeded parish: %v", err)
	}
	return id
}
