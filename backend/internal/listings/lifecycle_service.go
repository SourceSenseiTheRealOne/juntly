package listings

import (
	"context"
	"strings"
	"unicode/utf8"

	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/users"
	"github.com/google/uuid"
)

type ModeratorAuthorizer interface {
	RequireModerator(context.Context, users.VerifiedIdentity) (users.InternalUser, error)
}
type LifecycleRepository interface {
	TransitionOwned(context.Context, uuid.UUID, uuid.UUID, State, State, int, *string) (Listing, error)
	TransitionModerated(context.Context, uuid.UUID, uuid.UUID, State, State, int, *string) (Listing, error)
}
type LifecycleService interface {
	Submit(context.Context, users.VerifiedIdentity, uuid.UUID, int) (Listing, error)
	Approve(context.Context, users.VerifiedIdentity, uuid.UUID, int) (Listing, error)
	Reject(context.Context, users.VerifiedIdentity, uuid.UUID, int, string) (Listing, error)
	Pause(context.Context, users.VerifiedIdentity, uuid.UUID, int) (Listing, error)
	Archive(context.Context, users.VerifiedIdentity, uuid.UUID, State, int) (Listing, error)
}
type lifecycleService struct {
	providers  ProviderAuthorizer
	moderators ModeratorAuthorizer
	repository LifecycleRepository
}

func NewLifecycleService(providers ProviderAuthorizer, moderators ModeratorAuthorizer, repository LifecycleRepository) LifecycleService {
	return lifecycleService{providers: providers, moderators: moderators, repository: repository}
}
func (s lifecycleService) Submit(ctx context.Context, identity users.VerifiedIdentity, id uuid.UUID, revision int) (Listing, error) {
	if s.providers == nil || s.repository == nil || id == uuid.Nil || revision < 1 {
		return Listing{}, ErrInvalidListing
	}
	owner, err := s.providers.RequireProvider(ctx, identity)
	if err != nil {
		return Listing{}, err
	}
	return s.repository.TransitionOwned(ctx, owner.ID, id, StateDraft, StatePendingReview, revision, nil)
}
func (s lifecycleService) Approve(ctx context.Context, identity users.VerifiedIdentity, id uuid.UUID, revision int) (Listing, error) {
	if s.moderators == nil || s.repository == nil || id == uuid.Nil || revision < 1 {
		return Listing{}, ErrInvalidListing
	}
	moderator, err := s.moderators.RequireModerator(ctx, identity)
	if err != nil {
		return Listing{}, err
	}
	return s.repository.TransitionModerated(ctx, moderator.ID, id, StatePendingReview, StateActive, revision, nil)
}

func (s lifecycleService) Reject(ctx context.Context, identity users.VerifiedIdentity, id uuid.UUID, revision int, reason string) (Listing, error) {
	if s.moderators == nil || s.repository == nil || id == uuid.Nil || revision < 1 {
		return Listing{}, ErrInvalidListing
	}
	reason = strings.TrimSpace(reason)
	if utf8.RuneCountInString(reason) < 1 || utf8.RuneCountInString(reason) > 500 {
		return Listing{}, ErrInvalidListing
	}
	moderator, err := s.moderators.RequireModerator(ctx, identity)
	if err != nil {
		return Listing{}, err
	}
	return s.repository.TransitionModerated(ctx, moderator.ID, id, StatePendingReview, StateRejected, revision, &reason)
}

func (s lifecycleService) Pause(ctx context.Context, identity users.VerifiedIdentity, id uuid.UUID, revision int) (Listing, error) {
	if s.providers == nil || s.repository == nil || id == uuid.Nil || revision < 1 {
		return Listing{}, ErrInvalidListing
	}
	owner, err := s.providers.RequireProvider(ctx, identity)
	if err != nil {
		return Listing{}, err
	}
	return s.repository.TransitionOwned(ctx, owner.ID, id, StateActive, StatePaused, revision, nil)
}

func (s lifecycleService) Archive(ctx context.Context, identity users.VerifiedIdentity, id uuid.UUID, from State, revision int) (Listing, error) {
	if s.providers == nil || s.repository == nil || id == uuid.Nil || revision < 1 || !archivableState(from) {
		return Listing{}, ErrInvalidListing
	}
	owner, err := s.providers.RequireProvider(ctx, identity)
	if err != nil {
		return Listing{}, err
	}
	return s.repository.TransitionOwned(ctx, owner.ID, id, from, StateArchived, revision, nil)
}

func archivableState(state State) bool {
	switch state {
	case StateDraft, StateRejected, StateActive, StatePaused:
		return true
	default:
		return false
	}
}
