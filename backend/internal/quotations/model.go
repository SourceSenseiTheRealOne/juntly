package quotations

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalid      = errors.New("quotation invalid request")
	ErrUnauthorized = errors.New("quotation unauthorized")
	ErrForbidden    = errors.New("quotation forbidden")
	ErrNotFound     = errors.New("quotation not found")
	ErrConflict     = errors.New("quotation conflict")
	ErrUnavailable  = errors.New("quotation unavailable")
)

type RequestState string

const (
	RequestOpen     RequestState = "open"
	RequestAccepted RequestState = "accepted"
	RequestClosed   RequestState = "closed"
)

type ProposalState string

const (
	ProposalSubmitted ProposalState = "submitted"
	ProposalAccepted  ProposalState = "accepted"
	ProposalRejected  ProposalState = "rejected"
	ProposalExpired   ProposalState = "expired"
)

type CreateRequest struct {
	Title            string
	Description      string
	CategoryID       uuid.UUID
	LocalityID       uuid.UUID
	BudgetMinor      *int
	ProposalDeadline time.Time
}
type Request struct {
	ID               uuid.UUID
	CustomerID       uuid.UUID
	Title            string
	Description      string
	CategoryID       uuid.UUID
	LocalityID       uuid.UUID
	BudgetMinor      *int
	ProposalDeadline time.Time
	State            RequestState
	CreatedAt        time.Time
	UpdatedAt        time.Time
}
type SubmitProposal struct {
	PriceMinor       int
	Message          string
	AvailableAt      time.Time
	EstimatedMinutes *int
	ExpiresAt        *time.Time
}
type Proposal struct {
	ID               uuid.UUID
	RequestID        uuid.UUID
	ProviderID       uuid.UUID
	PriceMinor       int
	Message          string
	AvailableAt      time.Time
	EstimatedMinutes *int
	ExpiresAt        *time.Time
	State            ProposalState
	CreatedAt        time.Time
	UpdatedAt        time.Time
}
