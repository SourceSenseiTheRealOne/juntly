package listings

import (
	"context"

	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/listingmedia"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/users"
	"github.com/google/uuid"
)

type OwnerService interface {
	Create(context.Context, users.VerifiedIdentity, CreateListing) (Listing, error)
	ReplaceDraft(context.Context, users.VerifiedIdentity, uuid.UUID, int, CreateListing) (Listing, error)
	Get(context.Context, users.VerifiedIdentity, uuid.UUID) (*Listing, error)
	List(context.Context, users.VerifiedIdentity) ([]Listing, error)
	Submit(context.Context, users.VerifiedIdentity, uuid.UUID, int) (Listing, error)
	Pause(context.Context, users.VerifiedIdentity, uuid.UUID, int) (Listing, error)
	Archive(context.Context, users.VerifiedIdentity, uuid.UUID, State, int) (Listing, error)
	CreateUploadIntent(context.Context, users.VerifiedIdentity, uuid.UUID, listingmedia.UploadRequest) (listingmedia.UploadIntent, error)
}
type ownerService struct {
	drafts    Service
	lifecycle LifecycleService
	media     listingmedia.Service
}

func NewOwnerService(drafts Service, lifecycle LifecycleService, media listingmedia.Service) OwnerService {
	return ownerService{drafts: drafts, lifecycle: lifecycle, media: media}
}
func (s ownerService) Create(ctx context.Context, i users.VerifiedIdentity, v CreateListing) (Listing, error) {
	if s.drafts == nil {
		return Listing{}, ErrUnavailable
	}
	return s.drafts.Create(ctx, i, v)
}
func (s ownerService) ReplaceDraft(ctx context.Context, i users.VerifiedIdentity, id uuid.UUID, r int, v CreateListing) (Listing, error) {
	if s.drafts == nil {
		return Listing{}, ErrUnavailable
	}
	return s.drafts.ReplaceDraft(ctx, i, id, r, v)
}
func (s ownerService) Get(ctx context.Context, i users.VerifiedIdentity, id uuid.UUID) (*Listing, error) {
	if s.drafts == nil {
		return nil, ErrUnavailable
	}
	return s.drafts.Get(ctx, i, id)
}
func (s ownerService) List(ctx context.Context, i users.VerifiedIdentity) ([]Listing, error) {
	if s.drafts == nil {
		return nil, ErrUnavailable
	}
	return s.drafts.List(ctx, i)
}
func (s ownerService) Submit(ctx context.Context, i users.VerifiedIdentity, id uuid.UUID, r int) (Listing, error) {
	if s.lifecycle == nil {
		return Listing{}, ErrUnavailable
	}
	return s.lifecycle.Submit(ctx, i, id, r)
}
func (s ownerService) Pause(ctx context.Context, i users.VerifiedIdentity, id uuid.UUID, r int) (Listing, error) {
	if s.lifecycle == nil {
		return Listing{}, ErrUnavailable
	}
	return s.lifecycle.Pause(ctx, i, id, r)
}
func (s ownerService) Archive(ctx context.Context, i users.VerifiedIdentity, id uuid.UUID, from State, r int) (Listing, error) {
	if s.lifecycle == nil {
		return Listing{}, ErrUnavailable
	}
	return s.lifecycle.Archive(ctx, i, id, from, r)
}
func (s ownerService) CreateUploadIntent(ctx context.Context, i users.VerifiedIdentity, id uuid.UUID, r listingmedia.UploadRequest) (listingmedia.UploadIntent, error) {
	if s.media == nil {
		return listingmedia.UploadIntent{}, listingmedia.ErrUnavailable
	}
	return s.media.CreateUploadIntent(ctx, i, id, r)
}
