package moderation

import (
	"context"
	"errors"
	"testing"

	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/listings"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/users"
)

func TestQueueServiceRequiresModeratorBeforePendingListingRead(t *testing.T) {
	t.Parallel()
	authorizer := &recordingModeratorAuthorizer{user: users.InternalUser{}}
	repository := &recordingPendingRepository{listings: []listings.Listing{{State: listings.StatePendingReview}}}
	values, err := NewQueueService(authorizer, repository).ListPending(context.Background(), users.VerifiedIdentity{Subject: "moderator"})
	if err != nil || len(values) != 1 || repository.calls != 1 {
		t.Fatalf("values/err/calls=%#v/%v/%d", values, err, repository.calls)
	}
	_, err = NewQueueService(&recordingModeratorAuthorizer{err: ErrForbidden}, repository).ListPending(context.Background(), users.VerifiedIdentity{})
	if !errors.Is(err, ErrForbidden) || repository.calls != 1 {
		t.Fatalf("forbidden/calls=%v/%d", err, repository.calls)
	}
}

type recordingModeratorAuthorizer struct {
	user users.InternalUser
	err  error
}

func (a *recordingModeratorAuthorizer) RequireModerator(context.Context, users.VerifiedIdentity) (users.InternalUser, error) {
	return a.user, a.err
}

type recordingPendingRepository struct {
	listings []listings.Listing
	calls    int
	err      error
}

func (r *recordingPendingRepository) ListPending(context.Context) ([]listings.Listing, error) {
	r.calls++
	return r.listings, r.err
}
