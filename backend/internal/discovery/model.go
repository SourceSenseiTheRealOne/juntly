package discovery

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidRequest = errors.New("invalid public discovery request")
	ErrNotFound       = errors.New("public listing not found")
	ErrUnavailable    = errors.New("public discovery unavailable")
)

type PriceType string

const (
	PriceTypeFixed      PriceType = "fixed"
	PriceTypeHourly     PriceType = "hourly"
	PriceTypeDaily      PriceType = "daily"
	PriceTypeQuote      PriceType = "quote"
	PriceTypeNegotiable PriceType = "negotiable"
)

type ServiceMode string

const (
	ServiceModeTravelsToCustomer ServiceMode = "travels_to_customer"
	ServiceModeReceivesCustomer  ServiceMode = "receives_customer"
	ServiceModeRemoteServices    ServiceMode = "remote_services"
)

type Request struct {
	Locale         string
	CategoryID     uuid.UUID
	Query          string
	NearLocalityID uuid.UUID
	RadiusKM       int
	PriceType      PriceType
	ServiceMode    ServiceMode
}

type Listing struct {
	ID                  uuid.UUID
	Title               string
	Description         string
	CategoryID          uuid.UUID
	CategorySlug        string
	CategoryName        string
	PrimaryLocalityID   uuid.UUID
	LocalitySlug        string
	LocalityName        string
	PriceType           PriceType
	PriceMinor          *int
	Currency            string
	TravelsToCustomer   bool
	ReceivesCustomer    bool
	RemoteServices      bool
	ProviderDisplayName string
	ProviderType        string
	Promoted            bool
	UpdatedAt           time.Time
}

type Repository interface {
	Search(context.Context, Request) ([]Listing, error)
	Get(context.Context, uuid.UUID, string) (*Listing, error)
}
