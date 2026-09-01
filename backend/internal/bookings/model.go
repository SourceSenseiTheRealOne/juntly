package bookings

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalid      = errors.New("booking invalid request")
	ErrUnauthorized = errors.New("booking unauthorized")
	ErrForbidden    = errors.New("booking forbidden")
	ErrNotFound     = errors.New("booking not found")
	ErrConflict     = errors.New("booking conflict")
	ErrUnavailable  = errors.New("booking unavailable")
)

type State string

const (
	StateDraft                       State = "draft"
	StatePendingProviderConfirmation State = "pending_provider_confirmation"
	StateConfirmed                   State = "confirmed"
	StateScheduled                   State = "scheduled"
	StateInProgress                  State = "in_progress"
	StateCompleted                   State = "completed"
	StateCancelled                   State = "cancelled"
	StateDisputed                    State = "disputed"
	StateRefunded                    State = "refunded"
)

type SourceType string

const (
	SourceProposal SourceType = "proposal"
	SourceListing  SourceType = "listing"
	SourceDirect   SourceType = "direct"
)

type CreateBooking struct {
	SourceType       SourceType
	SourceID         *uuid.UUID
	ProviderID       *uuid.UUID
	IdempotencyKey   string
	ScheduledAt      time.Time
	PrivateLocation  string
	AgreedPriceMinor *int
}
type Transition struct {
	ExpectedState State
	TargetState   State
	Revision      int
	Reason        *string
}
type Booking struct {
	ID               uuid.UUID
	CustomerID       uuid.UUID
	ProviderID       uuid.UUID
	SourceType       SourceType
	SourceID         *uuid.UUID
	State            State
	Revision         int
	ScheduledAt      time.Time
	PrivateLocation  string
	AgreedPriceMinor int
	Currency         string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}
