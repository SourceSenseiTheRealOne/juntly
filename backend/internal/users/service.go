package users

import (
	"context"
	"errors"
	"strings"
	"unicode/utf8"
)

var ErrInvalidIdentity = errors.New("invalid verified identity")

type Service interface {
	Reconcile(context.Context, VerifiedIdentity) (InternalUser, bool, error)
}

type service struct {
	repository Repository
}

func NewService(repository Repository) Service {
	return service{repository: repository}
}

func (s service) Reconcile(ctx context.Context, identity VerifiedIdentity) (InternalUser, bool, error) {
	if !validSubject(identity.Subject) {
		return InternalUser{}, false, ErrInvalidIdentity
	}
	if s.repository == nil {
		return InternalUser{}, false, ErrUnavailable
	}

	existing, found, err := s.repository.FindBySubject(ctx, identity.Subject)
	if err != nil {
		return InternalUser{}, false, ErrUnavailable
	}
	if found {
		return existing, false, nil
	}

	created, err := s.repository.Create(ctx, identity.Subject)
	if err == nil {
		return created, true, nil
	}
	if !errors.Is(err, ErrSubjectConflict) {
		return InternalUser{}, false, ErrUnavailable
	}

	winner, found, err := s.repository.FindBySubject(ctx, identity.Subject)
	if err != nil || !found {
		return InternalUser{}, false, ErrUnavailable
	}
	return winner, false, nil
}

func validSubject(subject string) bool {
	return strings.TrimSpace(subject) != "" && utf8.RuneCountInString(subject) <= maxSubjectLength
}
