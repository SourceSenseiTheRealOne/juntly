package listingmedia

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/users"
	"github.com/google/uuid"
)

func TestReaderAuthorizesBeforeStorageAndChecksBytes(t *testing.T) {
	body := testPNG(t, 2, 3)
	image, err := SanitizeImage(body, imageRequest(body))
	if err != nil {
		t.Fatal(err)
	}
	id := uuid.New()
	entry := VerifiedMedia{ID: id, Ordinal: 1, ObjectReference: verifiedReference(id, image.ChecksumSHA256), ByteSize: int64(len(image.Bytes)), ChecksumSHA256: image.ChecksumSHA256, Width: image.Width, Height: image.Height}
	repo := &readTestRepository{media: []VerifiedMedia{entry}}
	storage := &readTestStorage{bytes: image.Bytes}
	identity := &readTestIdentity{user: users.InternalUser{ID: uuid.New()}}
	reader := NewReader(identity, repo, storage)
	listingID := uuid.New()
	photos, err := reader.List(context.Background(), users.VerifiedIdentity{}, PublicRead, listingID)
	if err != nil || len(photos) != 1 || photos[0].ID != id || identity.calls != 0 {
		t.Fatal("public safe projection failed")
	}
	data, err := reader.Get(context.Background(), users.VerifiedIdentity{Subject: "owner"}, OwnerRead, listingID, id)
	if err != nil || !bytes.Equal(data, image.Bytes) || repo.actor != identity.user.ID {
		t.Fatal("authorized image delivery failed")
	}
	storage.bytes = append(image.Bytes, 1)
	if _, err := reader.Get(context.Background(), users.VerifiedIdentity{}, PublicRead, listingID, id); !errors.Is(err, ErrUnavailable) {
		t.Fatal("corrupted object served")
	}
	repo.err = ErrNotFound
	before := storage.calls
	if _, err := reader.Get(context.Background(), users.VerifiedIdentity{}, PublicRead, listingID, id); !errors.Is(err, ErrNotFound) || storage.calls != before {
		t.Fatal("denied image reached storage")
	}
	identity.err = users.ErrInvalidIdentity
	before = repo.calls
	if _, err := reader.List(context.Background(), users.VerifiedIdentity{}, OwnerRead, listingID); err == nil || repo.calls != before {
		t.Fatal("missing identity reached repository")
	}
	if _, err := reader.List(context.Background(), users.VerifiedIdentity{}, ReadScope("administrator"), listingID); err == nil {
		t.Fatal("unknown scope accepted")
	}
}

type readTestRepository struct {
	media []VerifiedMedia
	err   error
	actor uuid.UUID
	calls int
}

func (r *readTestRepository) ListVerified(_ context.Context, scope ReadScope, actor, _ uuid.UUID) ([]VerifiedMedia, error) {
	r.calls++
	r.actor = actor
	return r.media, r.err
}

type readTestStorage struct {
	bytes []byte
	calls int
}

func (s *readTestStorage) DownloadVerified(context.Context, VerifiedMedia) ([]byte, error) {
	s.calls++
	return s.bytes, nil
}

type readTestIdentity struct {
	user  users.InternalUser
	calls int
	err   error
}

func (i *readTestIdentity) Reconcile(context.Context, users.VerifiedIdentity) (users.InternalUser, bool, error) {
	i.calls++
	return i.user, false, i.err
}
