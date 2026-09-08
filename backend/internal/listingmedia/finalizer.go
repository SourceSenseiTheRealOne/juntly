package listingmedia

import (
	"context"

	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/users"
	"github.com/google/uuid"
)

// MediaObject is a private repository projection, never a public response.
type MediaObject struct {
	ID              uuid.UUID
	Request         UploadRequest
	ObjectReference string
	Revision        int
	Ready           bool
}

type FinalizationRepository interface {
	GetEditable(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) (MediaObject, error)
	Complete(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, int, string, SanitizedImage) error
}

type FinalizationStorage interface {
	DownloadPending(context.Context, string) ([]byte, error)
	WriteVerified(context.Context, uuid.UUID, SanitizedImage) (string, error)
}

type Finalizer struct {
	authorizer ProviderAuthorizer
	repository FinalizationRepository
	storage    FinalizationStorage
	slots      chan struct{}
}

func NewFinalizer(authorizer ProviderAuthorizer, repository FinalizationRepository, storage FinalizationStorage) *Finalizer {
	return &Finalizer{authorizer: authorizer, repository: repository, storage: storage, slots: make(chan struct{}, 2)}
}

func (f *Finalizer) Finalize(ctx context.Context, identity users.VerifiedIdentity, listingID, mediaID uuid.UUID) error {
	if f == nil || f.authorizer == nil || f.repository == nil || f.storage == nil {
		return ErrUnavailable
	}
	if listingID == uuid.Nil || mediaID == uuid.Nil {
		return ErrInvalidUpload
	}
	owner, err := f.authorizer.RequireProvider(ctx, identity)
	if err != nil {
		return err
	}
	media, err := f.repository.GetEditable(ctx, owner.ID, listingID, mediaID)
	if err != nil {
		return err
	}
	if media.Ready {
		return nil
	}
	select {
	case f.slots <- struct{}{}:
		defer func() { <-f.slots }()
	default:
		return ErrUnavailable
	}
	body, err := f.storage.DownloadPending(ctx, media.ObjectReference)
	if err != nil {
		return ErrUnavailable
	}
	verified, err := SanitizeImage(body, media.Request)
	if err != nil {
		return err
	}
	reference, err := f.storage.WriteVerified(ctx, mediaID, verified)
	if err != nil || reference == "" {
		return ErrUnavailable
	}
	return f.repository.Complete(ctx, owner.ID, listingID, mediaID, media.Revision, reference, verified)
}
