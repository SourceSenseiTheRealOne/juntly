package listingmedia

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/SourceSenseiTheRealOne/juntly/backend/ent"
	entlisting "github.com/SourceSenseiTheRealOne/juntly/backend/ent/listing"
	"github.com/SourceSenseiTheRealOne/juntly/backend/ent/listingmedia"
	"github.com/SourceSenseiTheRealOne/juntly/backend/ent/locality"
	"github.com/SourceSenseiTheRealOne/juntly/backend/ent/servicecategory"
	"github.com/SourceSenseiTheRealOne/juntly/backend/ent/spokenlanguage"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/accounts"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/listings"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/providers"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/users"
	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestEntRepositoryReservesOwnerPendingMediaWithoutPublicReference(t *testing.T) {
	client := openMediaClient(t)
	ctx := context.Background()
	owner, listingID := createMediaListing(t, client)
	repository := NewEntRepository(client)
	mediaID := uuid.New()
	request := validRequest()
	privateReference := "storage-internal/media/" + mediaID.String()

	if err := repository.ReservePending(ctx, owner.ID, listingID, mediaID, request, privateReference); err != nil {
		t.Fatalf("reserve pending media: %v", err)
	}
	stored, err := client.ListingMedia.Get(ctx, mediaID)
	if err != nil || stored.ListingID != listingID || stored.ObjectReference != privateReference || string(stored.State) != "pending_upload" {
		t.Fatalf("stored media = %#v, err = %v", stored, err)
	}

	other, _ := createMediaListing(t, client)
	if err := repository.ReservePending(ctx, other.ID, listingID, uuid.New(), request, "storage-internal/media/other"); err == nil {
		t.Fatal("cross-owner media reservation error = nil")
	}
	count, err := client.ListingMedia.Query().Where(listingmedia.ListingIDEQ(listingID)).Count(ctx)
	if err != nil || count != 1 {
		t.Fatalf("media count = %d, err = %v", count, err)
	}
}

func openMediaClient(t *testing.T) *ent.Client {
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
	t.Cleanup(func() { _ = client.Close() })
	return client
}

func createMediaListing(t *testing.T, client *ent.Client) (users.InternalUser, uuid.UUID) {
	t.Helper()
	ctx := context.Background()
	owner, _, err := users.NewService(users.NewEntRepository(client)).Reconcile(ctx, users.VerifiedIdentity{Subject: "media_provider_" + uuid.NewString()})
	if err != nil {
		t.Fatalf("reconcile owner: %v", err)
	}
	accountsRepo := accounts.NewEntRepository(client)
	if _, err := accountsRepo.Create(ctx, owner.ID); err != nil {
		t.Fatalf("create account: %v", err)
	}
	if _, err := accountsRepo.SetProviderEnabled(ctx, owner.ID, true); err != nil {
		t.Fatalf("enable provider: %v", err)
	}
	localityID, err := client.Locality.Query().Where(locality.ActiveEQ(true)).FirstID(ctx)
	if err != nil {
		t.Fatalf("locality: %v", err)
	}
	languageID, err := client.SpokenLanguage.Query().Where(spokenlanguage.ActiveEQ(true)).FirstID(ctx)
	if err != nil {
		t.Fatalf("language: %v", err)
	}
	categoryID, err := client.ServiceCategory.Query().Where(servicecategory.ActiveEQ(true)).FirstID(ctx)
	if err != nil {
		t.Fatalf("category: %v", err)
	}
	if _, err := providers.NewEntRepository(client).Replace(ctx, owner.ID, providers.ReplaceProfile{DisplayName: "Media provider", ProviderType: providers.ProviderTypeProfessional, Bio: "Synthetic provider profile for media integration.", PrimaryLocalityID: localityID, ServiceLocalityIDs: []uuid.UUID{localityID}, MaxTravelDistanceKM: 25, TravelsToCustomer: true, LanguageCodes: []string{languageID}}); err != nil {
		t.Fatalf("profile: %v", err)
	}
	price := 5000
	listing, err := listings.NewEntRepository(client).Create(ctx, owner.ID, listings.CreateListing{CategoryID: categoryID, PrimaryLocalityID: localityID, Title: "Media test listing", Description: "Synthetic listing used only to prove owner-scoped media reservations.", PriceType: listings.PriceTypeFixed, PriceMinor: &price, Currency: "EUR", TravelsToCustomer: true})
	if err != nil {
		t.Fatalf("listing: %v", err)
	}
	t.Cleanup(func() {
		_, _ = client.ListingMedia.Delete().Where(listingmedia.ListingIDEQ(listing.ID)).Exec(ctx)
		_, _ = client.Listing.Delete().Where(entlisting.IDEQ(listing.ID)).Exec(ctx)
		_ = client.InternalUser.DeleteOneID(owner.ID).Exec(ctx)
	})
	return owner, listing.ID
}
