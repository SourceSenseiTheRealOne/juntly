package listingmedia

import (
	"context"
	"strings"

	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/users"
	"github.com/google/uuid"
)

type Service interface {
	CreateUploadIntent(context.Context, users.VerifiedIdentity, uuid.UUID, UploadRequest) (UploadIntent, error)
}
type service struct {
	authorizer ProviderAuthorizer
	repository Repository
	storage    Storage
}

func NewService(authorizer ProviderAuthorizer, repository Repository, storage Storage) Service {
	return service{authorizer: authorizer, repository: repository, storage: storage}
}
func (s service) CreateUploadIntent(ctx context.Context, identity users.VerifiedIdentity, listingID uuid.UUID, request UploadRequest) (UploadIntent, error) {
	if s.authorizer == nil || s.repository == nil || s.storage == nil || listingID == uuid.Nil || !validUploadRequest(request) {
		return UploadIntent{}, ErrInvalidUpload
	}
	owner, err := s.authorizer.RequireProvider(ctx, identity)
	if err != nil {
		return UploadIntent{}, err
	}
	mediaID := uuid.New()
	reservation, err := s.storage.CreateUploadReservation(ctx, mediaID, request)
	if err != nil {
		return UploadIntent{}, ErrUnavailable
	}
	if reservation.ObjectReference == "" || reservation.Capability.URL == "" || reservation.Capability.Method == "" {
		return UploadIntent{}, ErrUnavailable
	}
	if err := s.repository.ReservePending(ctx, owner.ID, listingID, mediaID, request, reservation.ObjectReference); err != nil {
		return UploadIntent{}, ErrUnavailable
	}
	return UploadIntent{MediaID: mediaID, Capability: reservation.Capability}, nil
}
func validUploadRequest(request UploadRequest) bool {
	if request.Ordinal < 1 || request.Ordinal > 10 || request.ByteSize < 1 || request.ByteSize > 10485760 || len(request.ChecksumSHA256) != 64 {
		return false
	}
	for _, r := range request.ChecksumSHA256 {
		if !(r >= '0' && r <= '9' || r >= 'a' && r <= 'f') {
			return false
		}
	}
	return strings.EqualFold(request.ContentType, "image/jpeg") || strings.EqualFold(request.ContentType, "image/png") || strings.EqualFold(request.ContentType, "image/webp")
}
