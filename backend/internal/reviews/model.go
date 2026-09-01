package reviews

import (
	"errors"
	"github.com/google/uuid"
	"time"
)

var (
	ErrInvalid      = errors.New("review invalid request")
	ErrUnauthorized = errors.New("review unauthorized")
	ErrForbidden    = errors.New("review forbidden")
	ErrConflict     = errors.New("review conflict")
	ErrNotFound     = errors.New("review not found")
	ErrUnavailable  = errors.New("review unavailable")
)

type CreateReview struct {
	BookingID uuid.UUID
	Rating    int
	Body      string
}
type Review struct {
	ID               uuid.UUID
	BookingID        uuid.UUID
	CustomerID       uuid.UUID
	ProviderID       uuid.UUID
	Rating           int
	Body             string
	ProviderResponse string
	VerifiedBooking  bool
	State            string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}
type Aggregate struct {
	ProviderID    uuid.UUID
	AverageRating float64
	ReviewCount   int
}
