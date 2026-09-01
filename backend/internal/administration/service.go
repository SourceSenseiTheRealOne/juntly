package administration

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
	Metrics(context.Context, uuid.UUID) (Metrics, error)
	Queue(context.Context, uuid.UUID) (Queue, error)
	Moderate(context.Context, uuid.UUID, ModerationAction) error
}
type Service interface {
	Metrics(context.Context, users.VerifiedIdentity) (Metrics, error)
	Queue(context.Context, users.VerifiedIdentity) (Queue, error)
	Moderate(context.Context, users.VerifiedIdentity, ModerationAction) error
}
type service struct {
	identities IdentityReconciler
	store      Store
}

func NewService(i IdentityReconciler, s Store) Service { return service{identities: i, store: s} }
func (s service) Metrics(ctx context.Context, identity users.VerifiedIdentity) (Metrics, error) {
	actor, err := s.actor(ctx, identity)
	if err != nil {
		return Metrics{}, err
	}
	v, err := s.store.Metrics(ctx, actor)
	return v, normalize(err)
}
func (s service) Queue(ctx context.Context, identity users.VerifiedIdentity) (Queue, error) {
	actor, err := s.actor(ctx, identity)
	if err != nil {
		return Queue{}, err
	}
	v, err := s.store.Queue(ctx, actor)
	return v, normalize(err)
}
func (s service) Moderate(ctx context.Context, identity users.VerifiedIdentity, input ModerationAction) error {
	input.Kind = strings.TrimSpace(input.Kind)
	input.Reason = strings.TrimSpace(input.Reason)
	if input.TargetID == uuid.Nil || !(strings3{"hide_review", "publish_review", "resolve_report"}).contains(input.Kind) || utf8.RuneCountInString(input.Reason) < 5 || utf8.RuneCountInString(input.Reason) > 500 {
		return ErrInvalid
	}
	actor, err := s.actor(ctx, identity)
	if err != nil {
		return err
	}
	return normalize(s.store.Moderate(ctx, actor, input))
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
	if errors.Is(err, ErrInvalid) || errors.Is(err, ErrUnauthorized) || errors.Is(err, ErrForbidden) || errors.Is(err, ErrNotFound) {
		return err
	}
	return ErrUnavailable
}

type strings3 [3]string

func (v strings3) contains(candidate string) bool {
	for _, item := range v {
		if item == candidate {
			return true
		}
	}
	return false
}
