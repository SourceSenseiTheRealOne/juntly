package moderation

import (
	"context"
	"errors"

	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/users"
	"github.com/google/uuid"
)

var (
	ErrUnauthorized = errors.New("moderator access unauthorized")
	ErrForbidden    = errors.New("moderator access forbidden")
	ErrUnavailable  = errors.New("moderator access unavailable")
)

type InternalUserReconciler interface {
	Reconcile(context.Context, users.VerifiedIdentity) (users.InternalUser, bool, error)
}
type Repository interface {
	HasModeratorGrant(context.Context, uuid.UUID) (bool, error)
}
type Service interface {
	RequireModerator(context.Context, users.VerifiedIdentity) (users.InternalUser, error)
}
type service struct {
	identities InternalUserReconciler
	repository Repository
}

func NewService(identities InternalUserReconciler, repository Repository) Service {
	return service{identities: identities, repository: repository}
}
func (s service) RequireModerator(ctx context.Context, identity users.VerifiedIdentity) (users.InternalUser, error) {
	if s.identities == nil || s.repository == nil {
		return users.InternalUser{}, ErrUnavailable
	}
	user, _, err := s.identities.Reconcile(ctx, identity)
	if err != nil {
		if errors.Is(err, users.ErrInvalidIdentity) {
			return users.InternalUser{}, ErrUnauthorized
		}
		return users.InternalUser{}, ErrUnavailable
	}
	granted, err := s.repository.HasModeratorGrant(ctx, user.ID)
	if err != nil {
		return users.InternalUser{}, ErrUnavailable
	}
	if !granted {
		return users.InternalUser{}, ErrForbidden
	}
	return user, nil
}
