package reference

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

type Service interface {
	Categories(context.Context, string) ([]Category, error)
	Languages(context.Context, string) ([]Language, error)
	Localities(context.Context, string) ([]Locality, error)
	NearbyLocalities(context.Context, uuid.UUID, int, string) ([]LocalityDistance, error)
}

type service struct{ repository Repository }

func NewService(repository Repository) Service {
	return service{repository: repository}
}

func (s service) Categories(ctx context.Context, locale string) ([]Category, error) {
	if !validLocale(locale) {
		return nil, ErrInvalidRequest
	}
	if s.repository == nil {
		return nil, ErrUnavailable
	}
	values, err := s.repository.Categories(ctx, locale)
	return values, publicRepositoryError(err)
}

func (s service) Languages(ctx context.Context, locale string) ([]Language, error) {
	if !validLocale(locale) {
		return nil, ErrInvalidRequest
	}
	if s.repository == nil {
		return nil, ErrUnavailable
	}
	values, err := s.repository.Languages(ctx, locale)
	return values, publicRepositoryError(err)
}

func (s service) Localities(ctx context.Context, locale string) ([]Locality, error) {
	if !validLocale(locale) {
		return nil, ErrInvalidRequest
	}
	if s.repository == nil {
		return nil, ErrUnavailable
	}
	values, err := s.repository.Localities(ctx, locale)
	return values, publicRepositoryError(err)
}

func (s service) NearbyLocalities(ctx context.Context, origin uuid.UUID, radiusKM int, locale string) ([]LocalityDistance, error) {
	if origin == uuid.Nil || radiusKM < 1 || radiusKM > 200 || !validLocale(locale) {
		return nil, ErrInvalidRequest
	}
	if s.repository == nil {
		return nil, ErrUnavailable
	}
	values, err := s.repository.NearbyLocalities(ctx, origin, radiusKM, locale)
	return values, publicRepositoryError(err)
}

func validLocale(locale string) bool {
	switch locale {
	case "pt-PT", "en", "es":
		return true
	default:
		return false
	}
}

func publicRepositoryError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, ErrNotFound):
		return ErrNotFound
	case errors.Is(err, ErrInvalidReference):
		return ErrInvalidReference
	default:
		return ErrUnavailable
	}
}
