package listings

import (
	"context"
	"errors"
	"testing"

	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/moderation"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/provideraccess"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/users"
	"github.com/google/uuid"
)

func TestLifecycleServiceScopesOwnerAndModeratorTransitions(t *testing.T) {
	t.Parallel()
	owner := users.InternalUser{ID: uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa")}
	moderator := users.InternalUser{ID: uuid.MustParse("bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb")}
	listingID := uuid.MustParse("cccccccc-cccc-4ccc-8ccc-cccccccccccc")
	repository := &recordingLifecycleRepository{}
	service := NewLifecycleService(&recordingAuthorizer{owner: owner}, &recordingModeratorAuthorizer{user: moderator}, repository)

	if _, err := service.Submit(context.Background(), users.VerifiedIdentity{Subject: "provider"}, listingID, 1); err != nil {
		t.Fatalf("submit: %v", err)
	}
	if repository.actor != owner.ID || repository.from != StateDraft || repository.to != StatePendingReview || repository.revision != 1 {
		t.Fatalf("submit transition = %#v", repository)
	}
	if _, err := service.Approve(context.Background(), users.VerifiedIdentity{Subject: "moderator"}, listingID, 2); err != nil {
		t.Fatalf("approve: %v", err)
	}
	if repository.actor != moderator.ID || repository.from != StatePendingReview || repository.to != StateActive || repository.revision != 2 {
		t.Fatalf("approve transition = %#v", repository)
	}
}

func TestLifecycleServiceShortCircuitsUnauthorizedOwnersAndModerators(t *testing.T) {
	t.Parallel()
	listingID := uuid.New()
	for _, test := range []struct {
		providerErr, moderatorErr, want error
		action                          string
	}{
		{providerErr: provideraccess.ErrForbidden, want: provideraccess.ErrForbidden, action: "submit"},
		{moderatorErr: moderation.ErrForbidden, want: moderation.ErrForbidden, action: "approve"},
	} {
		repository := &recordingLifecycleRepository{}
		service := NewLifecycleService(&recordingAuthorizer{err: test.providerErr}, &recordingModeratorAuthorizer{err: test.moderatorErr}, repository)
		var err error
		if test.action == "submit" {
			_, err = service.Submit(context.Background(), users.VerifiedIdentity{}, listingID, 1)
		} else {
			_, err = service.Approve(context.Background(), users.VerifiedIdentity{}, listingID, 1)
		}
		if !errors.Is(err, test.want) || repository.calls != 0 {
			t.Fatalf("error/calls = %v/%d", err, repository.calls)
		}
	}
}

func TestLifecycleServiceRejectsPendingReviewWithBoundedModeratorReason(t *testing.T) {
	t.Parallel()
	moderator := users.InternalUser{ID: uuid.MustParse("dddddddd-dddd-4ddd-8ddd-dddddddddddd")}
	repository := &recordingLifecycleRepository{}
	service := NewLifecycleService(&recordingAuthorizer{}, &recordingModeratorAuthorizer{user: moderator}, repository)
	listingID := uuid.MustParse("eeeeeeee-eeee-4eee-8eee-eeeeeeeeeeee")

	if _, err := service.Reject(context.Background(), users.VerifiedIdentity{Subject: "moderator"}, listingID, 3, "Needs clearer scope"); err != nil {
		t.Fatalf("reject: %v", err)
	}
	if repository.actor != moderator.ID || repository.from != StatePendingReview || repository.to != StateRejected || repository.revision != 3 {
		t.Fatalf("reject transition = %#v", repository)
	}
}

func TestLifecycleServicePausesAndArchivesOnlyThroughOwnerScope(t *testing.T) {
	t.Parallel()
	owner := users.InternalUser{ID: uuid.MustParse("ffffffff-ffff-4fff-8fff-ffffffffffff")}
	repository := &recordingLifecycleRepository{}
	service := NewLifecycleService(&recordingAuthorizer{owner: owner}, &recordingModeratorAuthorizer{}, repository)
	listingID := uuid.MustParse("99999999-9999-4999-8999-999999999999")
	if _, err := service.Pause(context.Background(), users.VerifiedIdentity{Subject: "provider"}, listingID, 4); err != nil {
		t.Fatalf("pause: %v", err)
	}
	if repository.actor != owner.ID || repository.from != StateActive || repository.to != StatePaused {
		t.Fatalf("pause transition = %#v", repository)
	}
	if _, err := service.Archive(context.Background(), users.VerifiedIdentity{Subject: "provider"}, listingID, StatePaused, 5); err != nil {
		t.Fatalf("archive: %v", err)
	}
	if repository.actor != owner.ID || repository.from != StatePaused || repository.to != StateArchived {
		t.Fatalf("archive transition = %#v", repository)
	}
}

type recordingModeratorAuthorizer struct {
	user users.InternalUser
	err  error
}

func (a *recordingModeratorAuthorizer) RequireModerator(context.Context, users.VerifiedIdentity) (users.InternalUser, error) {
	return a.user, a.err
}

type recordingLifecycleRepository struct {
	actor    uuid.UUID
	from     State
	to       State
	revision int
	calls    int
}

func (r *recordingLifecycleRepository) TransitionOwned(_ context.Context, actor, id uuid.UUID, from, to State, revision int, reason *string) (Listing, error) {
	r.calls++
	r.actor, r.from, r.to, r.revision = actor, from, to, revision
	return Listing{ID: id, State: to, Revision: revision + 1}, nil
}
func (r *recordingLifecycleRepository) TransitionModerated(_ context.Context, actor, id uuid.UUID, from, to State, revision int, reason *string) (Listing, error) {
	r.calls++
	r.actor, r.from, r.to, r.revision = actor, from, to, revision
	return Listing{ID: id, State: to, Revision: revision + 1}, nil
}
