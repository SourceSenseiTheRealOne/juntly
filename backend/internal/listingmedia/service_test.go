package listingmedia

import (
	"context"
	"errors"
	"testing"

	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/provideraccess"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/users"
	"github.com/google/uuid"
)

func TestServiceCreatesOwnerUploadIntentWithoutStorageReference(t *testing.T) {
	t.Parallel()
	owner := users.InternalUser{ID: uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa")}
	listingID := uuid.MustParse("bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb")
	storage := &recordingStorage{reservation: StorageReservation{ObjectReference: "private-object-reference", Capability: UploadCapability{URL: "https://upload.example.invalid/capability", Method: "PUT", Headers: map[string]string{"Content-Type": "image/webp"}}}}
	repository := &recordingRepository{}
	intent, err := NewService(&recordingAuthorizer{owner: owner}, repository, storage).CreateUploadIntent(context.Background(), users.VerifiedIdentity{Subject: "provider"}, listingID, UploadRequest{Ordinal: 1, ContentType: "image/webp", ByteSize: 1024, ChecksumSHA256: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"})
	if err != nil || intent.MediaID == uuid.Nil || intent.Capability.URL == "" || repository.owner != owner.ID || repository.objectReference == "" || intent.Capability.URL == repository.objectReference {
		t.Fatalf("intent/err/repository = %#v/%v/%#v", intent, err, repository)
	}
}

func TestServiceRejectsInvalidUploadOrUnauthorizedProviderBeforeStorage(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		err     error
		request UploadRequest
		want    error
	}{
		{err: provideraccess.ErrForbidden, request: validRequest(), want: provideraccess.ErrForbidden},
		{request: UploadRequest{Ordinal: 11}, want: ErrInvalidUpload},
	} {
		repository := &recordingRepository{}
		storage := &recordingStorage{}
		_, err := NewService(&recordingAuthorizer{err: test.err}, repository, storage).CreateUploadIntent(context.Background(), users.VerifiedIdentity{}, uuid.New(), test.request)
		if !errors.Is(err, test.want) || repository.calls != 0 || storage.calls != 0 {
			t.Fatalf("error/calls = %v/%d/%d", err, repository.calls, storage.calls)
		}
	}
}

func validRequest() UploadRequest {
	return UploadRequest{Ordinal: 1, ContentType: "image/webp", ByteSize: 1024, ChecksumSHA256: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}
}

type recordingAuthorizer struct {
	owner users.InternalUser
	err   error
}

func (a *recordingAuthorizer) RequireProvider(context.Context, users.VerifiedIdentity) (users.InternalUser, error) {
	return a.owner, a.err
}

type recordingRepository struct {
	owner           uuid.UUID
	objectReference string
	calls           int
}

func (r *recordingRepository) ReservePending(_ context.Context, owner, listingID, mediaID uuid.UUID, request UploadRequest, objectReference string) error {
	r.calls++
	r.owner = owner
	r.objectReference = objectReference
	return nil
}

type recordingStorage struct {
	reservation StorageReservation
	calls       int
}

func (s *recordingStorage) CreateUploadReservation(context.Context, uuid.UUID, UploadRequest) (StorageReservation, error) {
	s.calls++
	return s.reservation, nil
}
