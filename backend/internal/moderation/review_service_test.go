package moderation

import (
	"context"
	"testing"

	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/listings"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/users"
	"github.com/google/uuid"
)

func TestReviewServiceDelegatesQueueAndModeratedTransitions(t *testing.T) {
	t.Parallel()
	queue := &recordingQueue{listings: []listings.Listing{{State: listings.StatePendingReview}}}
	lifecycle := &recordingReviewLifecycle{listing: listings.Listing{State: listings.StateActive}}
	service := NewReviewService(queue, lifecycle)
	if values, err := service.ListPending(context.Background(), users.VerifiedIdentity{Subject: "moderator"}); err != nil || len(values) != 1 {
		t.Fatalf("queue=%#v err=%v", values, err)
	}
	if _, err := service.Approve(context.Background(), users.VerifiedIdentity{Subject: "moderator"}, uuid.New(), 1); err != nil {
		t.Fatalf("approve=%v", err)
	}
	if _, err := service.Reject(context.Background(), users.VerifiedIdentity{Subject: "moderator"}, uuid.New(), 2, "Needs scope"); err != nil {
		t.Fatalf("reject=%v", err)
	}
	if lifecycle.approvals != 1 || lifecycle.rejections != 1 {
		t.Fatalf("lifecycle=%#v", lifecycle)
	}
}

type recordingQueue struct{ listings []listings.Listing }

func (q *recordingQueue) ListPending(context.Context, users.VerifiedIdentity) ([]listings.Listing, error) {
	return q.listings, nil
}

type recordingReviewLifecycle struct {
	listing               listings.Listing
	approvals, rejections int
}

func (s *recordingReviewLifecycle) Approve(context.Context, users.VerifiedIdentity, uuid.UUID, int) (listings.Listing, error) {
	s.approvals++
	return s.listing, nil
}
func (s *recordingReviewLifecycle) Reject(context.Context, users.VerifiedIdentity, uuid.UUID, int, string) (listings.Listing, error) {
	s.rejections++
	return s.listing, nil
}
