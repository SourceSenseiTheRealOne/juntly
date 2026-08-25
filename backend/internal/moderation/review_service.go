package moderation

import (
	"context"

	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/listings"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/users"
	"github.com/google/uuid"
)

type ListingQueue interface {
	ListPending(context.Context, users.VerifiedIdentity) ([]listings.Listing, error)
}
type ModeratedLifecycle interface {
	Approve(context.Context, users.VerifiedIdentity, uuid.UUID, int) (listings.Listing, error)
	Reject(context.Context, users.VerifiedIdentity, uuid.UUID, int, string) (listings.Listing, error)
}
type ReviewService interface {
	ListPending(context.Context, users.VerifiedIdentity) ([]listings.Listing, error)
	Approve(context.Context, users.VerifiedIdentity, uuid.UUID, int) (listings.Listing, error)
	Reject(context.Context, users.VerifiedIdentity, uuid.UUID, int, string) (listings.Listing, error)
}
type reviewService struct {
	queue     ListingQueue
	lifecycle ModeratedLifecycle
}

func NewReviewService(queue ListingQueue, lifecycle ModeratedLifecycle) ReviewService {
	return reviewService{queue: queue, lifecycle: lifecycle}
}
func (s reviewService) ListPending(ctx context.Context, i users.VerifiedIdentity) ([]listings.Listing, error) {
	if s.queue == nil {
		return nil, ErrUnavailable
	}
	return s.queue.ListPending(ctx, i)
}
func (s reviewService) Approve(ctx context.Context, i users.VerifiedIdentity, id uuid.UUID, r int) (listings.Listing, error) {
	if s.lifecycle == nil {
		return listings.Listing{}, ErrUnavailable
	}
	return s.lifecycle.Approve(ctx, i, id, r)
}
func (s reviewService) Reject(ctx context.Context, i users.VerifiedIdentity, id uuid.UUID, r int, reason string) (listings.Listing, error) {
	if s.lifecycle == nil {
		return listings.Listing{}, ErrUnavailable
	}
	return s.lifecycle.Reject(ctx, i, id, r, reason)
}
