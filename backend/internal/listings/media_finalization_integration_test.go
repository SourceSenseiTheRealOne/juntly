package listings

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"image"
	"image/png"
	"strings"
	"sync"
	"testing"

	entlisting "github.com/SourceSenseiTheRealOne/juntly/backend/ent/listing"
	"github.com/SourceSenseiTheRealOne/juntly/backend/ent/listingevent"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/listingmedia"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/provideraccess"
	"github.com/google/uuid"
)

func TestListingMediaFinalizationDurableIsolationAndReplay(t *testing.T) {
	client := openListingClient(t)
	ctx := context.Background()
	owner, localities, category := createListingProvider(t, client)
	other, _, _ := createListingProvider(t, client)
	listing, err := NewEntRepository(client).Create(ctx, owner.ID, integrationCreate(category, localities[0]))
	if err != nil {
		t.Fatal(err)
	}
	repo := listingmedia.NewEntRepository(client)
	final, ok := repo.(listingmedia.FinalizationRepository)
	if !ok {
		t.Fatal("repository does not support durable finalization")
	}
	var buffer bytes.Buffer
	if err := png.Encode(&buffer, image.NewRGBA(image.Rect(0, 0, 2, 3))); err != nil {
		t.Fatal(err)
	}
	body := buffer.Bytes()
	sum := sha256.Sum256(body)
	request := listingmedia.UploadRequest{Ordinal: 1, ContentType: "image/png", ByteSize: int64(len(body)), ChecksumSHA256: hex.EncodeToString(sum[:])}
	verified, err := listingmedia.SanitizeImage(body, request)
	if err != nil {
		t.Fatal(err)
	}
	mediaID := uuid.New()
	pending := "pending/" + mediaID.String()
	reference := "verified/" + mediaID.String() + "/" + verified.ChecksumSHA256 + ".png"
	if err := repo.ReservePending(ctx, owner.ID, listing.ID, mediaID, request, pending); err != nil {
		t.Fatal(err)
	}
	if _, err := final.GetEditable(ctx, other.ID, listing.ID, mediaID); !errors.Is(err, provideraccess.ErrForbidden) {
		t.Fatal("cross-owner read accepted")
	}
	if err := final.Complete(ctx, other.ID, listing.ID, mediaID, listing.Revision, reference, verified); !errors.Is(err, provideraccess.ErrForbidden) {
		t.Fatal("cross-owner completion accepted")
	}
	media, err := final.GetEditable(ctx, owner.ID, listing.ID, mediaID)
	if err != nil || media.Ready || media.ObjectReference != pending {
		t.Fatalf("pending projection: %v", err)
	}
	if err := final.Complete(ctx, owner.ID, listing.ID, mediaID, media.Revision+1, reference, verified); err == nil {
		t.Fatal("stale revision accepted")
	}
	if err := final.Complete(ctx, owner.ID, listing.ID, mediaID, media.Revision, "verified/wrong.png", verified); err == nil {
		t.Fatal("unscoped publication reference accepted")
	}
	for range 2 {
		if err := final.Complete(ctx, owner.ID, listing.ID, mediaID, media.Revision, reference, verified); err != nil {
			t.Fatal(err)
		}
	}
	stored, err := client.ListingMedia.Get(ctx, mediaID)
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(stored)
	if err != nil || !strings.Contains(string(encoded), `"verified_object_reference":"`+reference+`"`) || stored.ObjectReference != pending || string(stored.State) != "ready" {
		t.Fatal("ready record lost verified bytes or original upload provenance")
	}
	updated, err := client.Listing.Get(ctx, listing.ID)
	if err != nil || updated.Revision != media.Revision+1 {
		t.Fatal("completion did not advance parent revision exactly once")
	}
	events, err := client.ListingEvent.Query().Where(listingevent.ListingIDEQ(listing.ID), listingevent.EventTypeEQ(listingevent.EventTypeUpdated)).Count(ctx)
	if err != nil || events != 1 {
		t.Fatal("completion audit is not idempotent")
	}
	reader, ok := repo.(listingmedia.ReadRepository)
	if !ok {
		t.Fatal("repository cannot enforce verified-media reads")
	}
	photos, err := reader.ListVerified(ctx, listingmedia.OwnerRead, owner.ID, listing.ID)
	if err != nil || len(photos) != 1 || photos[0].ObjectReference != reference {
		t.Fatal("owner cannot read verified photo")
	}
	for _, scope := range []listingmedia.ReadScope{listingmedia.PublicRead, listingmedia.OwnerRead, listingmedia.ModeratorRead} {
		if _, err := reader.ListVerified(ctx, scope, other.ID, listing.ID); !errors.Is(err, listingmedia.ErrNotFound) {
			t.Fatalf("draft exposed to unauthorized scope %s", scope)
		}
	}
	if err := client.PlatformRole.Create().SetInternalUserID(other.ID).SetRole("moderator").Exec(ctx); err != nil {
		t.Fatal(err)
	}
	if photos, err := reader.ListVerified(ctx, listingmedia.ModeratorRead, other.ID, listing.ID); err != nil || len(photos) != 1 {
		t.Fatal("durably granted moderator cannot review verified photo")
	}
	if err := client.Listing.UpdateOneID(listing.ID).SetState(entlisting.StateActive).Exec(ctx); err != nil {
		t.Fatal(err)
	}
	if photos, err := reader.ListVerified(ctx, listingmedia.PublicRead, uuid.Nil, listing.ID); err != nil || len(photos) != 1 {
		t.Fatal("active listing photo unavailable publicly")
	}
	if err := client.Listing.UpdateOneID(listing.ID).SetState(entlisting.StatePaused).Exec(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := reader.ListVerified(ctx, listingmedia.PublicRead, uuid.Nil, listing.ID); !errors.Is(err, listingmedia.ErrNotFound) {
		t.Fatal("paused listing photo remained public")
	}
	if err := client.Listing.UpdateOneID(listing.ID).SetState(entlisting.StateDraft).Exec(ctx); err != nil {
		t.Fatal(err)
	}
	verified.Bytes = append(verified.Bytes, 1)
	if err := final.Complete(ctx, owner.ID, listing.ID, mediaID, media.Revision, reference, verified); err == nil {
		t.Fatal("contradictory replay accepted")
	}
	if err := client.Listing.UpdateOneID(listing.ID).SetState(entlisting.StatePendingReview).Exec(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := final.GetEditable(ctx, owner.ID, listing.ID, mediaID); !errors.Is(err, provideraccess.ErrForbidden) {
		t.Fatal("reviewed listing remained editable")
	}
}

func TestListingMediaFinalizationConcurrentCompletionHasOneAudit(t *testing.T) {
	client := openListingClient(t)
	ctx := context.Background()
	owner, localities, category := createListingProvider(t, client)
	listing, err := NewEntRepository(client).Create(ctx, owner.ID, integrationCreate(category, localities[0]))
	if err != nil {
		t.Fatal(err)
	}
	repo := listingmedia.NewEntRepository(client)
	final, ok := repo.(listingmedia.FinalizationRepository)
	if !ok {
		t.Fatal("repository does not support durable finalization")
	}
	var buffer bytes.Buffer
	if err := png.Encode(&buffer, image.NewRGBA(image.Rect(0, 0, 2, 3))); err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(buffer.Bytes())
	request := listingmedia.UploadRequest{Ordinal: 1, ContentType: "image/png", ByteSize: int64(buffer.Len()), ChecksumSHA256: hex.EncodeToString(sum[:])}
	verified, err := listingmedia.SanitizeImage(buffer.Bytes(), request)
	if err != nil {
		t.Fatal(err)
	}
	id := uuid.New()
	if err := repo.ReservePending(ctx, owner.ID, listing.ID, id, request, "pending/"+id.String()); err != nil {
		t.Fatal(err)
	}
	ref := "verified/" + id.String() + "/" + verified.ChecksumSHA256 + ".png"
	start := make(chan struct{})
	results := make(chan error, 2)
	var wg sync.WaitGroup
	for range 2 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			results <- final.Complete(ctx, owner.ID, listing.ID, id, listing.Revision, ref, verified)
		}()
	}
	close(start)
	wg.Wait()
	close(results)
	winners := 0
	for err := range results {
		if err == nil {
			winners++
		}
	}
	if winners == 0 {
		t.Fatal("no completion committed")
	}
	events, err := client.ListingEvent.Query().Where(listingevent.ListingIDEQ(listing.ID), listingevent.EventTypeEQ(listingevent.EventTypeUpdated)).Count(ctx)
	if err != nil || events != 1 {
		t.Fatal("concurrent completion duplicated or lost audit")
	}
}
