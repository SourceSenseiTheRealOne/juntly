package listings

import (
	"context"
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/users"
	"github.com/google/uuid"
)

type ProviderAuthorizer interface {
	RequireProvider(context.Context, users.VerifiedIdentity) (users.InternalUser, error)
}

type Service interface {
	Create(context.Context, users.VerifiedIdentity, CreateListing) (Listing, error)
	ReplaceDraft(context.Context, users.VerifiedIdentity, uuid.UUID, int, CreateListing) (Listing, error)
	Get(context.Context, users.VerifiedIdentity, uuid.UUID) (*Listing, error)
	List(context.Context, users.VerifiedIdentity) ([]Listing, error)
}

type service struct {
	authorizer ProviderAuthorizer
	repository Repository
}

func NewService(authorizer ProviderAuthorizer, repository Repository) Service {
	return service{authorizer: authorizer, repository: repository}
}

func (s service) Create(ctx context.Context, identity users.VerifiedIdentity, input CreateListing) (Listing, error) {
	owner, err := s.requireOwner(ctx, identity)
	if err != nil {
		return Listing{}, err
	}
	input, valid := normalizeCreate(input)
	if !valid {
		return Listing{}, ErrInvalidListing
	}
	listing, err := s.repository.Create(ctx, owner.ID, input)
	if err != nil {
		return Listing{}, publicRepositoryError(err)
	}
	return listing, nil
}

func (s service) ReplaceDraft(ctx context.Context, identity users.VerifiedIdentity, id uuid.UUID, revision int, input CreateListing) (Listing, error) {
	owner, err := s.requireOwner(ctx, identity)
	if err != nil {
		return Listing{}, err
	}
	input, valid := normalizeCreate(input)
	if id == uuid.Nil || revision < 1 || !valid {
		return Listing{}, ErrInvalidListing
	}
	updated, err := s.repository.ReplaceDraft(ctx, owner.ID, id, revision, input)
	if err != nil {
		return Listing{}, publicRepositoryError(err)
	}
	return updated, nil
}

func (s service) Get(ctx context.Context, identity users.VerifiedIdentity, id uuid.UUID) (*Listing, error) {
	owner, err := s.requireOwner(ctx, identity)
	if err != nil {
		return nil, err
	}
	if id == uuid.Nil {
		return nil, ErrInvalidListing
	}
	listing, err := s.repository.FindByOwner(ctx, owner.ID, id)
	if err != nil {
		return nil, ErrUnavailable
	}
	return listing, nil
}

func (s service) List(ctx context.Context, identity users.VerifiedIdentity) ([]Listing, error) {
	owner, err := s.requireOwner(ctx, identity)
	if err != nil {
		return nil, err
	}
	listings, err := s.repository.ListByOwner(ctx, owner.ID)
	if err != nil {
		return nil, ErrUnavailable
	}
	return listings, nil
}

func (s service) requireOwner(ctx context.Context, identity users.VerifiedIdentity) (users.InternalUser, error) {
	if s.authorizer == nil || s.repository == nil {
		return users.InternalUser{}, ErrUnavailable
	}
	owner, err := s.authorizer.RequireProvider(ctx, identity)
	if err != nil {
		return users.InternalUser{}, err
	}
	return owner, nil
}

func normalizeCreate(input CreateListing) (CreateListing, bool) {
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)
	input.Currency = strings.TrimSpace(input.Currency)
	if input.CategoryID == uuid.Nil || input.PrimaryLocalityID == uuid.Nil || utf8.RuneCountInString(input.Title) < 2 || utf8.RuneCountInString(input.Title) > 140 || utf8.RuneCountInString(input.Description) < 20 || utf8.RuneCountInString(input.Description) > 4000 || input.Currency != "EUR" || (!input.TravelsToCustomer && !input.ReceivesCustomer && !input.RemoteServices) {
		return CreateListing{}, false
	}
	switch input.PriceType {
	case PriceTypeFixed, PriceTypeHourly, PriceTypeDaily:
		if input.PriceMinor == nil || *input.PriceMinor <= 0 {
			return CreateListing{}, false
		}
	case PriceTypeQuote, PriceTypeNegotiable:
		if input.PriceMinor != nil {
			return CreateListing{}, false
		}
	default:
		return CreateListing{}, false
	}
	return input, true
}

func publicRepositoryError(err error) error {
	switch {
	case errors.Is(err, ErrInvalidListing):
		return ErrInvalidListing
	case errors.Is(err, ErrConflict):
		return ErrConflict
	default:
		return ErrUnavailable
	}
}
