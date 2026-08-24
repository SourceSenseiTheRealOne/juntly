package listings

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidListing = errors.New("invalid listing")
	ErrConflict       = errors.New("listing revision or state conflict")
	ErrUnavailable    = errors.New("listing persistence unavailable")
)

type PriceType string

const (
	PriceTypeFixed      PriceType = "fixed"
	PriceTypeHourly     PriceType = "hourly"
	PriceTypeDaily      PriceType = "daily"
	PriceTypeQuote      PriceType = "quote"
	PriceTypeNegotiable PriceType = "negotiable"
)

type State string

const StateDraft State = "draft"

type CreateListing struct {
	CategoryID        uuid.UUID
	PrimaryLocalityID uuid.UUID
	Title             string
	Description       string
	PriceType         PriceType
	PriceMinor        *int
	Currency          string
	TravelsToCustomer bool
	ReceivesCustomer  bool
	RemoteServices    bool
}

type Listing struct {
	ID uuid.UUID
	CreateListing
	State     State
	Revision  int
	CreatedAt time.Time
	UpdatedAt time.Time
}
