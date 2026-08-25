package provideraccess

import (
	"context"
	"errors"

	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/accounts"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/users"
)

var (
	ErrUnauthorized = errors.New("provider access unauthorized")
	ErrForbidden    = errors.New("provider access forbidden")
	ErrUnavailable  = errors.New("provider access unavailable")
)

type InternalUserReconciler interface {
	Reconcile(context.Context, users.VerifiedIdentity) (users.InternalUser, bool, error)
}

type CapabilityReader interface {
	Get(context.Context, users.VerifiedIdentity) (accounts.Account, error)
}

type Service interface {
	RequireProvider(context.Context, users.VerifiedIdentity) (users.InternalUser, error)
}

type service struct {
	identities   InternalUserReconciler
	capabilities CapabilityReader
}

func NewService(identities InternalUserReconciler, capabilities CapabilityReader) Service {
	return service{identities: identities, capabilities: capabilities}
}

func (s service) RequireProvider(ctx context.Context, identity users.VerifiedIdentity) (users.InternalUser, error) {
	if s.identities == nil || s.capabilities == nil {
		return users.InternalUser{}, ErrUnavailable
	}
	owner, _, err := s.identities.Reconcile(ctx, identity)
	if err != nil {
		if errors.Is(err, users.ErrInvalidIdentity) {
			return users.InternalUser{}, ErrUnauthorized
		}
		return users.InternalUser{}, ErrUnavailable
	}
	account, err := s.capabilities.Get(ctx, identity)
	if err != nil {
		if errors.Is(err, accounts.ErrInvalidIdentity) {
			return users.InternalUser{}, ErrUnauthorized
		}
		return users.InternalUser{}, ErrUnavailable
	}
	if !account.ProviderEnabled {
		return users.InternalUser{}, ErrForbidden
	}
	return owner, nil
}
