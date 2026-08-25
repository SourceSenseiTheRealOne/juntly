package listings

import (
	"context"
	"testing"

	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/listingmedia"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/users"
	"github.com/google/uuid"
)

func TestOwnerServiceDelegatesOnlyToServerSideDomains(t *testing.T) {
	t.Parallel()
	listingID := uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa")
	identity := users.VerifiedIdentity{Subject: "provider"}
	draft := &recordingDraftService{listing: sampleOwnerListing()}
	lifecycle := &recordingLifecycleService{listing: sampleOwnerListing()}
	media := &recordingMediaService{intent: listingmedia.UploadIntent{MediaID: uuid.MustParse("bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb")}}
	service := NewOwnerService(draft, lifecycle, media)
	if _, err := service.Submit(context.Background(), identity, listingID, 1); err != nil {
		t.Fatalf("submit: %v", err)
	}
	if _, err := service.Pause(context.Background(), identity, listingID, 2); err != nil {
		t.Fatalf("pause: %v", err)
	}
	if _, err := service.Archive(context.Background(), identity, listingID, StatePaused, 3); err != nil {
		t.Fatalf("archive: %v", err)
	}
	if _, err := service.CreateUploadIntent(context.Background(), identity, listingID, listingmedia.UploadRequest{Ordinal: 1, ContentType: "image/webp", ByteSize: 1, ChecksumSHA256: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}); err != nil {
		t.Fatalf("media: %v", err)
	}
	if lifecycle.calls != 3 || media.calls != 1 {
		t.Fatalf("calls lifecycle/media=%d/%d", lifecycle.calls, media.calls)
	}
}

type recordingDraftService struct{ listing Listing }

func (s *recordingDraftService) Create(context.Context, users.VerifiedIdentity, CreateListing) (Listing, error) {
	return s.listing, nil
}
func (s *recordingDraftService) ReplaceDraft(context.Context, users.VerifiedIdentity, uuid.UUID, int, CreateListing) (Listing, error) {
	return s.listing, nil
}
func (s *recordingDraftService) Get(context.Context, users.VerifiedIdentity, uuid.UUID) (*Listing, error) {
	return &s.listing, nil
}
func (s *recordingDraftService) List(context.Context, users.VerifiedIdentity) ([]Listing, error) {
	return []Listing{s.listing}, nil
}

type recordingLifecycleService struct {
	listing Listing
	calls   int
}

func (s *recordingLifecycleService) Submit(context.Context, users.VerifiedIdentity, uuid.UUID, int) (Listing, error) {
	s.calls++
	return s.listing, nil
}
func (s *recordingLifecycleService) Approve(context.Context, users.VerifiedIdentity, uuid.UUID, int) (Listing, error) {
	return s.listing, nil
}
func (s *recordingLifecycleService) Reject(context.Context, users.VerifiedIdentity, uuid.UUID, int, string) (Listing, error) {
	return s.listing, nil
}
func (s *recordingLifecycleService) Pause(context.Context, users.VerifiedIdentity, uuid.UUID, int) (Listing, error) {
	s.calls++
	return s.listing, nil
}
func (s *recordingLifecycleService) Archive(context.Context, users.VerifiedIdentity, uuid.UUID, State, int) (Listing, error) {
	s.calls++
	return s.listing, nil
}

type recordingMediaService struct {
	intent listingmedia.UploadIntent
	calls  int
}

func (s *recordingMediaService) CreateUploadIntent(context.Context, users.VerifiedIdentity, uuid.UUID, listingmedia.UploadRequest) (listingmedia.UploadIntent, error) {
	s.calls++
	return s.intent, nil
}
func sampleOwnerListing() Listing {
	return Listing{ID: uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"), State: StateDraft, Revision: 1}
}
