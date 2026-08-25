package discovery

import (
	"context"
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
)

type Service interface {
	Search(context.Context, Request) ([]Listing, error)
	Get(context.Context, string, string) (*Listing, error)
}

type service struct{ repository Repository }

func NewService(repository Repository) Service { return service{repository: repository} }

func (s service) Search(ctx context.Context, request Request) ([]Listing, error) {
	request, ok := normalizeRequest(request)
	if !ok {
		return nil, ErrInvalidRequest
	}
	if s.repository == nil {
		return nil, ErrUnavailable
	}
	values, err := s.repository.Search(ctx, request)
	if err != nil {
		return nil, ErrUnavailable
	}
	return append([]Listing(nil), values...), nil
}

func (s service) Get(ctx context.Context, rawID, locale string) (*Listing, error) {
	if !validLocale(locale) {
		return nil, ErrInvalidRequest
	}
	if s.repository == nil {
		return nil, ErrUnavailable
	}
	id, err := uuid.Parse(rawID)
	if err != nil {
		return nil, ErrNotFound
	}
	value, err := s.repository.Get(ctx, id, locale)
	switch {
	case err == nil && value != nil:
		cloned := *value
		return &cloned, nil
	case errors.Is(err, ErrNotFound):
		return nil, ErrNotFound
	default:
		return nil, ErrUnavailable
	}
}

func normalizeRequest(request Request) (Request, bool) {
	if !validLocale(request.Locale) {
		return Request{}, false
	}
	request.Query = strings.Join(strings.Fields(request.Query), " ")
	if (request.NearLocalityID == uuid.Nil) != (request.RadiusKM == 0) || request.RadiusKM < 0 || request.RadiusKM > 200 {
		return Request{}, false
	}
	if request.RadiusKM > 0 && request.RadiusKM < 1 {
		return Request{}, false
	}
	if request.Query != "" && (utf8.RuneCountInString(request.Query) < 2 || utf8.RuneCountInString(request.Query) > 80) {
		return Request{}, false
	}
	if request.PriceType != "" && !validPriceType(request.PriceType) {
		return Request{}, false
	}
	if request.ServiceMode != "" && !validServiceMode(request.ServiceMode) {
		return Request{}, false
	}
	return request, true
}

func validLocale(locale string) bool {
	switch locale {
	case "pt-PT", "en", "es":
		return true
	default:
		return false
	}
}

func validPriceType(value PriceType) bool {
	switch value {
	case PriceTypeFixed, PriceTypeHourly, PriceTypeDaily, PriceTypeQuote, PriceTypeNegotiable:
		return true
	default:
		return false
	}
}

func validServiceMode(value ServiceMode) bool {
	switch value {
	case ServiceModeTravelsToCustomer, ServiceModeReceivesCustomer, ServiceModeRemoteServices:
		return true
	default:
		return false
	}
}
