package accounts

import (
	"context"
	"errors"

	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/users"
)

type IdentityReconciler interface {
	Reconcile(context.Context, users.VerifiedIdentity) (users.InternalUser, bool, error)
}

type Service interface {
	Get(context.Context, users.VerifiedIdentity) (Account, error)
	SetProviderEnabled(context.Context, users.VerifiedIdentity, bool) (Account, error)
}

type service struct {
	identities IdentityReconciler
	repository Repository
}

func NewService(identities IdentityReconciler, repository Repository) Service {
	return service{identities: identities, repository: repository}
}

func (s service) Get(ctx context.Context, identity users.VerifiedIdentity) (Account, error) {
	record, err := s.reconcileAccount(ctx, identity)
	if err != nil {
		return Account{}, err
	}
	return accountFromRecord(record), nil
}

func (s service) SetProviderEnabled(ctx context.Context, identity users.VerifiedIdentity, enabled bool) (Account, error) {
	record, err := s.reconcileAccount(ctx, identity)
	if err != nil {
		return Account{}, err
	}

	updated, err := s.repository.SetProviderEnabled(ctx, record.InternalUserID, enabled)
	if err != nil {
		return Account{}, ErrUnavailable
	}
	return accountFromRecord(updated), nil
}

func (s service) reconcileAccount(ctx context.Context, identity users.VerifiedIdentity) (Record, error) {
	if s.identities == nil || s.repository == nil {
		return Record{}, ErrUnavailable
	}

	internalUser, _, err := s.identities.Reconcile(ctx, identity)
	if err != nil {
		if errors.Is(err, users.ErrInvalidIdentity) {
			return Record{}, ErrInvalidIdentity
		}
		return Record{}, ErrUnavailable
	}

	existing, found, err := s.repository.FindByInternalUserID(ctx, internalUser.ID)
	if err != nil {
		return Record{}, ErrUnavailable
	}
	if found {
		return existing, nil
	}

	created, err := s.repository.Create(ctx, internalUser.ID)
	if err == nil {
		return created, nil
	}
	if !errors.Is(err, ErrAccountConflict) {
		return Record{}, ErrUnavailable
	}

	winner, found, err := s.repository.FindByInternalUserID(ctx, internalUser.ID)
	if err != nil || !found {
		return Record{}, ErrUnavailable
	}
	return winner, nil
}
