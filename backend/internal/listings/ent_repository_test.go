package listings

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"testing"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/SourceSenseiTheRealOne/juntly/backend/ent"
	"github.com/SourceSenseiTheRealOne/juntly/backend/ent/listing"
	"github.com/SourceSenseiTheRealOne/juntly/backend/ent/listingevent"
	"github.com/SourceSenseiTheRealOne/juntly/backend/ent/locality"
	"github.com/SourceSenseiTheRealOne/juntly/backend/ent/servicecategory"
	"github.com/SourceSenseiTheRealOne/juntly/backend/ent/spokenlanguage"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/accounts"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/providers"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/users"
	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestEntRepositoryCreatesDraftAndAtomicCreatedEvent(t *testing.T) {
	client := openListingClient(t)
	ctx := context.Background()
	owner, localityIDs, categoryID := createListingProvider(t, client)
	repository := NewEntRepository(client)
	input := integrationCreate(categoryID, localityIDs[0])

	created, err := repository.Create(ctx, owner.ID, input)
	if err != nil {
		t.Fatalf("create draft: %v", err)
	}
	if created.ID == uuid.Nil || created.State != StateDraft || created.Revision != 1 || created.CreatedAt.IsZero() || created.UpdatedAt.IsZero() {
		t.Fatalf("created listing = %#v", created)
	}
	found, err := repository.FindByOwner(ctx, owner.ID, created.ID)
	if err != nil || found == nil || found.ID != created.ID || found.State != StateDraft {
		t.Fatalf("find owner listing = %#v, err = %v", found, err)
	}
	events, err := client.ListingEvent.Query().Where(listingevent.ListingIDEQ(created.ID)).All(ctx)
	if err != nil || len(events) != 1 || string(events[0].EventType) != "created" || events[0].ToState != "draft" || events[0].Revision != 1 {
		t.Fatalf("created events = %#v, err = %v", events, err)
	}
	other, _, _ := createListingProvider(t, client)
	missing, err := repository.FindByOwner(ctx, other.ID, created.ID)
	if err != nil || missing != nil {
		t.Fatalf("cross-owner listing = %#v, err = %v", missing, err)
	}
}

func TestEntRepositoryRejectsLocalityOutsideProviderServiceAreasWithoutEvent(t *testing.T) {
	client := openListingClient(t)
	ctx := context.Background()
	owner, localityIDs, categoryID := createListingProvider(t, client)
	repository := NewEntRepository(client)
	input := integrationCreate(categoryID, localityIDs[1])
	if _, err := repository.Create(ctx, owner.ID, input); err == nil {
		t.Fatal("out-of-service-area listing error = nil")
	}
	count, err := client.Listing.Query().Where(listing.InternalUserIDEQ(owner.ID)).Count(ctx)
	if err != nil || count != 0 {
		t.Fatalf("listing count = %d, err = %v", count, err)
	}
}

func TestEntRepositoryReplacesDraftWithCASAndOneUpdatedEvent(t *testing.T) {
	client := openListingClient(t)
	ctx := context.Background()
	owner, localityIDs, categoryID := createListingProvider(t, client)
	repository := NewEntRepository(client)
	created, err := repository.Create(ctx, owner.ID, integrationCreate(categoryID, localityIDs[0]))
	if err != nil {
		t.Fatalf("create draft: %v", err)
	}
	updatedInput := integrationCreate(categoryID, localityIDs[0])
	updatedInput.Title = "Canalização local atualizada"
	updated, err := repository.ReplaceDraft(ctx, owner.ID, created.ID, created.Revision, updatedInput)
	if err != nil || updated.Revision != 2 || updated.Title != updatedInput.Title || updated.CreatedAt != created.CreatedAt {
		t.Fatalf("updated = %#v, err = %v", updated, err)
	}
	if _, err := repository.ReplaceDraft(ctx, owner.ID, created.ID, created.Revision, updatedInput); !errors.Is(err, ErrConflict) {
		t.Fatalf("stale update error = %v, want ErrConflict", err)
	}
	events, err := client.ListingEvent.Query().Where(listingevent.ListingIDEQ(created.ID)).Order(ent.Asc(listingevent.FieldRevision)).All(ctx)
	if err != nil || len(events) != 2 || string(events[1].EventType) != "updated" || events[1].Revision != 2 {
		t.Fatalf("events = %#v, err = %v", events, err)
	}
}

func openListingClient(t *testing.T) *ent.Client {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL is required for PostgreSQL integration tests")
	}
	database, err := sql.Open("pgx", url)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	client := ent.NewClient(ent.Driver(entsql.OpenDB(dialect.Postgres, database)))
	t.Cleanup(func() {
		if err := client.Close(); err != nil {
			t.Errorf("close client: %v", err)
		}
	})
	return client
}

func createListingProvider(t *testing.T, client *ent.Client) (users.InternalUser, []uuid.UUID, uuid.UUID) {
	t.Helper()
	ctx := context.Background()
	owner, _, err := users.NewService(users.NewEntRepository(client)).Reconcile(ctx, users.VerifiedIdentity{Subject: "listing_provider_" + uuid.NewString()})
	if err != nil {
		t.Fatalf("reconcile provider: %v", err)
	}
	accountsRepo := accounts.NewEntRepository(client)
	if _, err := accountsRepo.Create(ctx, owner.ID); err != nil {
		t.Fatalf("create account: %v", err)
	}
	if _, err := accountsRepo.SetProviderEnabled(ctx, owner.ID, true); err != nil {
		t.Fatalf("enable provider: %v", err)
	}
	localityIDs, err := client.Locality.Query().Where(locality.ActiveEQ(true)).Order(ent.Asc(locality.FieldID)).IDs(ctx)
	if err != nil || len(localityIDs) < 2 {
		t.Fatalf("localities = %v, err = %v", localityIDs, err)
	}
	languageIDs, err := client.SpokenLanguage.Query().Where(spokenlanguage.ActiveEQ(true)).Limit(1).IDs(ctx)
	if err != nil || len(languageIDs) != 1 {
		t.Fatalf("languages = %v, err = %v", languageIDs, err)
	}
	categoryIDs, err := client.ServiceCategory.Query().Where(servicecategory.ActiveEQ(true)).Limit(1).IDs(ctx)
	if err != nil || len(categoryIDs) != 1 {
		t.Fatalf("categories = %v, err = %v", categoryIDs, err)
	}
	_, err = providers.NewEntRepository(client).Replace(ctx, owner.ID, providers.ReplaceProfile{DisplayName: "Listing provider", ProviderType: providers.ProviderTypeProfessional, Bio: "Synthetic provider profile for listing integration.", PrimaryLocalityID: localityIDs[0], ServiceLocalityIDs: []uuid.UUID{localityIDs[0]}, MaxTravelDistanceKM: 25, TravelsToCustomer: true, LanguageCodes: languageIDs})
	if err != nil {
		t.Fatalf("create provider profile: %v", err)
	}
	t.Cleanup(func() {
		if _, err := client.Listing.Delete().Where(listing.InternalUserIDEQ(owner.ID)).Exec(ctx); err != nil {
			t.Errorf("cleanup listings: %v", err)
		}
		if err := client.InternalUser.DeleteOneID(owner.ID).Exec(ctx); err != nil {
			t.Errorf("cleanup owner: %v", err)
		}
	})
	return owner, localityIDs, categoryIDs[0]
}

func integrationCreate(categoryID, localityID uuid.UUID) CreateListing {
	price := 5000
	return CreateListing{CategoryID: categoryID, PrimaryLocalityID: localityID, Title: "Canalização local", Description: "Serviço de canalização local para pequenas reparações domésticas e manutenção.", PriceType: PriceTypeFixed, PriceMinor: &price, Currency: "EUR", TravelsToCustomer: true}
}
