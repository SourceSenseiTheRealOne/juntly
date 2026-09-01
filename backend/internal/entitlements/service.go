package entitlements

import (
	"context"
	"errors"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/users"
	"github.com/google/uuid"
)

type IdentityReconciler interface {
	Reconcile(context.Context, users.VerifiedIdentity) (users.InternalUser, bool, error)
}
type Store interface {
	Catalog(context.Context) (Catalog, error)
	RequestSubscription(context.Context, uuid.UUID, uuid.UUID) (Subscription, error)
	CurrentSubscription(context.Context, uuid.UUID) (*Subscription, error)
	RequestPromotion(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) (Promotion, error)
	ListPromotions(context.Context, uuid.UUID) ([]Promotion, error)
	Access(context.Context, uuid.UUID) (Access, error)
}
type Service interface {
	Catalog(context.Context) (Catalog, error)
	RequestSubscription(context.Context, users.VerifiedIdentity, uuid.UUID) (Subscription, error)
	CurrentSubscription(context.Context, users.VerifiedIdentity) (*Subscription, error)
	RequestPromotion(context.Context, users.VerifiedIdentity, uuid.UUID, uuid.UUID) (Promotion, error)
	ListPromotions(context.Context, users.VerifiedIdentity) ([]Promotion, error)
	Access(context.Context, users.VerifiedIdentity) (Access, error)
}
type service struct {
	identities IdentityReconciler
	store      Store
}

func NewService(i IdentityReconciler, s Store) Service { return service{identities: i, store: s} }
func (s service) Catalog(ctx context.Context) (Catalog, error) {
	if s.store == nil {
		return Catalog{}, ErrUnavailable
	}
	v, err := s.store.Catalog(ctx)
	return v, normalize(err)
}
func (s service) RequestSubscription(ctx context.Context, identity users.VerifiedIdentity, plan uuid.UUID) (Subscription, error) {
	if plan == uuid.Nil {
		return Subscription{}, ErrInvalid
	}
	actor, err := s.actor(ctx, identity)
	if err != nil {
		return Subscription{}, err
	}
	v, err := s.store.RequestSubscription(ctx, actor, plan)
	return v, normalize(err)
}
func (s service) CurrentSubscription(ctx context.Context, identity users.VerifiedIdentity) (*Subscription, error) {
	actor, err := s.actor(ctx, identity)
	if err != nil {
		return nil, err
	}
	v, err := s.store.CurrentSubscription(ctx, actor)
	return v, normalize(err)
}
func (s service) RequestPromotion(ctx context.Context, identity users.VerifiedIdentity, listing, period uuid.UUID) (Promotion, error) {
	if listing == uuid.Nil || period == uuid.Nil {
		return Promotion{}, ErrInvalid
	}
	actor, err := s.actor(ctx, identity)
	if err != nil {
		return Promotion{}, err
	}
	v, err := s.store.RequestPromotion(ctx, actor, listing, period)
	return v, normalize(err)
}
func (s service) ListPromotions(ctx context.Context, identity users.VerifiedIdentity) ([]Promotion, error) {
	actor, err := s.actor(ctx, identity)
	if err != nil {
		return nil, err
	}
	v, err := s.store.ListPromotions(ctx, actor)
	return v, normalize(err)
}
func (s service) Access(ctx context.Context, identity users.VerifiedIdentity) (Access, error) {
	actor, err := s.actor(ctx, identity)
	if err != nil {
		return Access{}, err
	}
	v, err := s.store.Access(ctx, actor)
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
	if errors.Is(err, ErrInvalid) || errors.Is(err, ErrUnauthorized) || errors.Is(err, ErrForbidden) || errors.Is(err, ErrConflict) {
		return err
	}
	return ErrUnavailable
}
