package moderation

import (
	"context"

	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/listings"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/users"
)

type PendingListingReader interface {
	ListPending(context.Context) ([]listings.Listing, error)
}
type QueueService interface {
	ListPending(context.Context, users.VerifiedIdentity) ([]listings.Listing, error)
}
type queueService struct {
	moderators Service
	repository PendingListingReader
}

func NewQueueService(moderators Service, repository PendingListingReader) QueueService {
	return queueService{moderators: moderators, repository: repository}
}
func (s queueService) ListPending(ctx context.Context, identity users.VerifiedIdentity) ([]listings.Listing, error) {
	if s.moderators == nil || s.repository == nil {
		return nil, ErrUnavailable
	}
	if _, err := s.moderators.RequireModerator(ctx, identity); err != nil {
		return nil, err
	}
	values, err := s.repository.ListPending(ctx)
	if err != nil {
		return nil, ErrUnavailable
	}
	return values, nil
}
