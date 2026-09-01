package reviews

import (
	"context"
	"errors"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/users"
	"github.com/google/uuid"
	"strings"
	"unicode/utf8"
)

type IdentityReconciler interface {
	Reconcile(context.Context, users.VerifiedIdentity) (users.InternalUser, bool, error)
}
type Store interface {
	Create(context.Context, uuid.UUID, CreateReview) (Review, error)
	ListForProvider(context.Context, uuid.UUID) ([]Review, error)
	Respond(context.Context, uuid.UUID, uuid.UUID, string) (Review, error)
	Aggregate(context.Context, uuid.UUID) (Aggregate, error)
}
type Service interface {
	Create(context.Context, users.VerifiedIdentity, CreateReview) (Review, error)
	ListForProvider(context.Context, users.VerifiedIdentity) ([]Review, error)
	Respond(context.Context, users.VerifiedIdentity, uuid.UUID, string) (Review, error)
	Aggregate(context.Context, uuid.UUID) (Aggregate, error)
}
type service struct {
	identities IdentityReconciler
	store      Store
}

func NewService(i IdentityReconciler, s Store) Service { return service{identities: i, store: s} }
func (s service) Create(ctx context.Context, identity users.VerifiedIdentity, input CreateReview) (Review, error) {
	input.Body = strings.TrimSpace(input.Body)
	if input.BookingID == uuid.Nil || input.Rating < 1 || input.Rating > 5 || utf8.RuneCountInString(input.Body) < 10 || utf8.RuneCountInString(input.Body) > 2000 {
		return Review{}, ErrInvalid
	}
	actor, err := s.actor(ctx, identity)
	if err != nil {
		return Review{}, err
	}
	v, err := s.store.Create(ctx, actor, input)
	return v, normalize(err)
}
func (s service) ListForProvider(ctx context.Context, identity users.VerifiedIdentity) ([]Review, error) {
	actor, err := s.actor(ctx, identity)
	if err != nil {
		return nil, err
	}
	v, err := s.store.ListForProvider(ctx, actor)
	return v, normalize(err)
}
func (s service) Respond(ctx context.Context, identity users.VerifiedIdentity, id uuid.UUID, response string) (Review, error) {
	response = strings.TrimSpace(response)
	if id == uuid.Nil || utf8.RuneCountInString(response) < 3 || utf8.RuneCountInString(response) > 1000 {
		return Review{}, ErrInvalid
	}
	actor, err := s.actor(ctx, identity)
	if err != nil {
		return Review{}, err
	}
	v, err := s.store.Respond(ctx, actor, id, response)
	return v, normalize(err)
}
func (s service) Aggregate(ctx context.Context, provider uuid.UUID) (Aggregate, error) {
	if s.store == nil || provider == uuid.Nil {
		return Aggregate{}, ErrInvalid
	}
	v, err := s.store.Aggregate(ctx, provider)
	return v, normalize(err)
}
func (s service) actor(ctx context.Context, identity users.VerifiedIdentity) (uuid.UUID, error) {
	if s.identities == nil || s.store == nil {
		return uuid.Nil, ErrUnavailable
	}
	u, _, err := s.identities.Reconcile(ctx, identity)
	if err != nil {
		if errors.Is(err, users.ErrInvalidIdentity) {
			return uuid.Nil, ErrUnauthorized
		}
		return uuid.Nil, ErrUnavailable
	}
	if u.ID == uuid.Nil {
		return uuid.Nil, ErrUnauthorized
	}
	return u.ID, nil
}
func normalize(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, ErrInvalid) || errors.Is(err, ErrUnauthorized) || errors.Is(err, ErrForbidden) || errors.Is(err, ErrConflict) || errors.Is(err, ErrNotFound) {
		return err
	}
	return ErrUnavailable
}
