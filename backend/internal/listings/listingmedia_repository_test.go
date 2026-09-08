package listings

import (
	"context"
	"errors"
	"testing"

	entlisting "github.com/SourceSenseiTheRealOne/juntly/backend/ent/listing"
	entlistingmedia "github.com/SourceSenseiTheRealOne/juntly/backend/ent/listingmedia"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/listingmedia"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/provideraccess"
	"github.com/google/uuid"
)

func TestListingMediaEntRepositoryReservesOwnerPendingMediaWithoutPublicReference(t *testing.T) {
	client := openListingClient(t)
	ctx := context.Background()
	owner, localityIDs, categoryID := createListingProvider(t, client)
	other, _, _ := createListingProvider(t, client)
	listing, err := NewEntRepository(client).Create(ctx, owner.ID, integrationCreate(categoryID, localityIDs[0]))
	if err != nil {
		t.Fatalf("create listing: %v", err)
	}
	mediaID := uuid.New()
	request := listingmedia.UploadRequest{Ordinal: 1, ContentType: "image/webp", ByteSize: 1024, ChecksumSHA256: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}
	privateReference := "storage-internal/media/" + mediaID.String()
	repository := listingmedia.NewEntRepository(client)
	if err := repository.RequireEditable(ctx, owner.ID, listing.ID); err != nil {
		t.Fatal(err)
	}
	if err := repository.RequireEditable(ctx, other.ID, listing.ID); !errors.Is(err, provideraccess.ErrForbidden) {
		t.Fatalf("cross-owner preauthorization: %v", err)
	}
	if err := repository.ReservePending(ctx, owner.ID, listing.ID, mediaID, request, privateReference); err != nil {
		t.Fatalf("reserve: %v", err)
	}
	stored, err := client.ListingMedia.Get(ctx, mediaID)
	if err != nil || stored.ListingID != listing.ID || stored.ObjectReference != privateReference || string(stored.State) != "pending_upload" {
		t.Fatalf("stored=%#v err=%v", stored, err)
	}
	if err := repository.ReservePending(ctx, other.ID, listing.ID, uuid.New(), request, "storage-internal/media/other"); err == nil {
		t.Fatal("cross-owner reservation error=nil")
	}
	count, err := client.ListingMedia.Query().Where(entlistingmedia.ListingIDEQ(listing.ID)).Count(ctx)
	if err != nil || count != 1 {
		t.Fatalf("count=%d err=%v", count, err)
	}
	if err := client.Listing.UpdateOneID(listing.ID).SetState(entlisting.StatePendingReview).Exec(ctx); err != nil {
		t.Fatal(err)
	}
	if err := repository.RequireEditable(ctx, owner.ID, listing.ID); !errors.Is(err, provideraccess.ErrForbidden) {
		t.Fatalf("non-editable preauthorization: %v", err)
	}
}
