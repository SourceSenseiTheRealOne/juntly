package listingmedia

import (
	"context"
	"errors"
	"testing"

	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/provideraccess"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/users"
	"github.com/google/uuid"
)

func TestFinalizeVerifiesBeforePublishing(t *testing.T) {
	body := testPNG(t, 2, 3)
	repo := &finalizationTestRepo{media: MediaObject{ID: uuid.New(), Request: imageRequest(body), ObjectReference: "pending/test", Revision: 1}}
	storage := &finalizationTestStorage{body: body}
	finalizer := NewFinalizer(&recordingAuthorizer{owner: users.InternalUser{ID: uuid.New()}}, repo, storage)
	if err := finalizer.Finalize(context.Background(), users.VerifiedIdentity{Subject: "provider"}, uuid.New(), repo.media.ID); err != nil {
		t.Fatal(err)
	}
	if repo.completed != 1 || storage.writes != 1 || repo.verified.ContentType != "image/png" || repo.reference != "verified/test.png" {
		t.Fatal("verified publication not persisted")
	}
}

func TestFinalizeRejectsCorruptionAndUnauthorizedReads(t *testing.T) {
	for _, unauthorized := range []bool{false, true} {
		body := testPNG(t, 2, 3)
		repo := &finalizationTestRepo{media: MediaObject{ID: uuid.New(), Request: imageRequest(body), ObjectReference: "pending/test", Revision: 1}}
		storage := &finalizationTestStorage{body: append(body, 1)}
		if unauthorized {
			repo.err = provideraccess.ErrForbidden
		}
		f := NewFinalizer(&recordingAuthorizer{owner: users.InternalUser{ID: uuid.New()}}, repo, storage)
		err := f.Finalize(context.Background(), users.VerifiedIdentity{Subject: "provider"}, uuid.New(), repo.media.ID)
		if err == nil || storage.writes != 0 || repo.completed != 0 || (unauthorized && storage.reads != 0) {
			t.Fatal("unsafe finalization crossed publication boundary")
		}
		if unauthorized && !errors.Is(err, provideraccess.ErrForbidden) {
			t.Fatal("lost authorization error")
		}
	}
}

func TestFinalizeReadyReplaySkipsStorage(t *testing.T) {
	repo := &finalizationTestRepo{media: MediaObject{ID: uuid.New(), Ready: true}}
	storage := &finalizationTestStorage{}
	f := NewFinalizer(&recordingAuthorizer{owner: users.InternalUser{ID: uuid.New()}}, repo, storage)
	if err := f.Finalize(context.Background(), users.VerifiedIdentity{Subject: "provider"}, uuid.New(), repo.media.ID); err != nil {
		t.Fatal(err)
	}
	if storage.reads != 0 || storage.writes != 0 || repo.completed != 0 {
		t.Fatal("ready replay performed storage mutation")
	}
}

type finalizationTestRepo struct {
	media     MediaObject
	err       error
	completed int
	reference string
	verified  SanitizedImage
}

func (r *finalizationTestRepo) GetEditable(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) (MediaObject, error) {
	return r.media, r.err
}
func (r *finalizationTestRepo) Complete(_ context.Context, _ uuid.UUID, _ uuid.UUID, _ uuid.UUID, _ int, ref string, img SanitizedImage) error {
	r.completed++
	r.reference = ref
	r.verified = img
	return nil
}

type finalizationTestStorage struct {
	body          []byte
	reads, writes int
}

func (s *finalizationTestStorage) DownloadPending(context.Context, string) ([]byte, error) {
	s.reads++
	return s.body, nil
}
func (s *finalizationTestStorage) WriteVerified(context.Context, uuid.UUID, SanitizedImage) (string, error) {
	s.writes++
	return "verified/test.png", nil
}
