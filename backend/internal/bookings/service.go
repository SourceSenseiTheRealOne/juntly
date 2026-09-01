package bookings

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/users"
	"github.com/google/uuid"
)

var idempotencyPattern = regexp.MustCompile(`^[A-Za-z0-9._:-]{8,128}$`)

type IdentityReconciler interface {
	Reconcile(context.Context, users.VerifiedIdentity) (users.InternalUser, bool, error)
}
type Store interface {
	Create(context.Context, uuid.UUID, CreateBooking) (Booking, error)
	List(context.Context, uuid.UUID) ([]Booking, error)
	Get(context.Context, uuid.UUID, uuid.UUID) (Booking, error)
	Transition(context.Context, uuid.UUID, uuid.UUID, Transition) (Booking, error)
}
type Service interface {
	Create(context.Context, users.VerifiedIdentity, CreateBooking) (Booking, error)
	List(context.Context, users.VerifiedIdentity) ([]Booking, error)
	Get(context.Context, users.VerifiedIdentity, uuid.UUID) (Booking, error)
	Transition(context.Context, users.VerifiedIdentity, uuid.UUID, Transition) (Booking, error)
}
type service struct {
	identities IdentityReconciler
	store      Store
}

func NewService(identities IdentityReconciler, store Store) Service {
	return service{identities: identities, store: store}
}
func (s service) Create(ctx context.Context, identity users.VerifiedIdentity, input CreateBooking) (Booking, error) {
	input.IdempotencyKey = strings.TrimSpace(input.IdempotencyKey)
	input.PrivateLocation = strings.TrimSpace(input.PrivateLocation)
	if !validCreate(input) {
		return Booking{}, ErrInvalid
	}
	actor, err := s.actor(ctx, identity)
	if err != nil {
		return Booking{}, err
	}
	v, err := s.store.Create(ctx, actor, input)
	return v, normalize(err)
}
func (s service) List(ctx context.Context, identity users.VerifiedIdentity) ([]Booking, error) {
	actor, err := s.actor(ctx, identity)
	if err != nil {
		return nil, err
	}
	v, err := s.store.List(ctx, actor)
	return v, normalize(err)
}
func (s service) Get(ctx context.Context, identity users.VerifiedIdentity, id uuid.UUID) (Booking, error) {
	if id == uuid.Nil {
		return Booking{}, ErrInvalid
	}
	actor, err := s.actor(ctx, identity)
	if err != nil {
		return Booking{}, err
	}
	v, err := s.store.Get(ctx, actor, id)
	return v, normalize(err)
}
func (s service) Transition(ctx context.Context, identity users.VerifiedIdentity, id uuid.UUID, input Transition) (Booking, error) {
	if id == uuid.Nil || input.Revision < 1 || !validTransition(input.ExpectedState, input.TargetState) || input.Reason != nil && (utf8.RuneCountInString(strings.TrimSpace(*input.Reason)) < 3 || utf8.RuneCountInString(*input.Reason) > 500) {
		return Booking{}, ErrInvalid
	}
	actor, err := s.actor(ctx, identity)
	if err != nil {
		return Booking{}, err
	}
	v, err := s.store.Transition(ctx, actor, id, input)
	return v, normalize(err)
}
func validCreate(v CreateBooking) bool {
	if !idempotencyPattern.MatchString(v.IdempotencyKey) || v.ScheduledAt.IsZero() || utf8.RuneCountInString(v.PrivateLocation) < 5 || utf8.RuneCountInString(v.PrivateLocation) > 500 {
		return false
	}
	switch v.SourceType {
	case SourceProposal:
		return v.SourceID != nil && *v.SourceID != uuid.Nil && v.ProviderID == nil
	case SourceListing:
		return v.SourceID != nil && *v.SourceID != uuid.Nil && v.ProviderID == nil
	case SourceDirect:
		return v.SourceID == nil && v.ProviderID != nil && *v.ProviderID != uuid.Nil && v.AgreedPriceMinor != nil && *v.AgreedPriceMinor > 0
	default:
		return false
	}
}
func validTransition(from, to State) bool {
	switch from {
	case StatePendingProviderConfirmation:
		return to == StateConfirmed || to == StateCancelled
	case StateConfirmed:
		return to == StateScheduled || to == StateCancelled || to == StateDisputed
	case StateScheduled:
		return to == StateInProgress || to == StateCancelled || to == StateDisputed
	case StateInProgress:
		return to == StateCompleted || to == StateDisputed
	case StateCompleted:
		return to == StateDisputed
	case StateDisputed:
		return to == StateRefunded
	default:
		return false
	}
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
	if errors.Is(err, ErrInvalid) || errors.Is(err, ErrUnauthorized) || errors.Is(err, ErrForbidden) || errors.Is(err, ErrNotFound) || errors.Is(err, ErrConflict) {
		return err
	}
	return ErrUnavailable
}
