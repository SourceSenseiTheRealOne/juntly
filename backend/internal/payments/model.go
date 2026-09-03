package payments

import (
	"errors"
	"time"
)

var (
	ErrInvalid      = errors.New("payment invalid request")
	ErrUnauthorized = errors.New("payment unauthorized")
	ErrForbidden    = errors.New("payment forbidden")
	ErrNotFound     = errors.New("payment not found")
	ErrConflict     = errors.New("payment conflict")
	ErrUnavailable  = errors.New("payment unavailable")
)

type State string

const (
	StatePendingCheckout State = "pending_checkout"
	StateCheckoutCreated State = "checkout_created"
	StateProcessing      State = "processing"
	StatePaid            State = "paid"
	StateFailed          State = "failed"
	StateRefundPending   State = "refund_pending"
	StateRefunded        State = "refunded"
	StateDisputed        State = "disputed"
	StateDisputeWon      State = "dispute_won"
	StateDisputeLost     State = "dispute_lost"
	StateCancelled       State = "cancelled"
)

type EventKind string

const (
	EventProcessing    EventKind = "processing"
	EventPaid          EventKind = "paid"
	EventFailed        EventKind = "failed"
	EventRefunded      EventKind = "refunded"
	EventDisputeOpened EventKind = "dispute_opened"
	EventDisputeWon    EventKind = "dispute_won"
	EventDisputeLost   EventKind = "dispute_lost"
	EventAccountUpdate EventKind = "account_updated"
)

type Order struct {
	ID                string    `json:"id"`
	BookingID         string    `json:"bookingId"`
	CustomerID        string    `json:"customerId"`
	ProviderID        string    `json:"providerId"`
	State             State     `json:"state"`
	GrossMinor        int64     `json:"grossMinor"`
	PlatformFeeMinor  int64     `json:"platformFeeMinor"`
	ProviderNetMinor  int64     `json:"providerNetMinor"`
	Currency          string    `json:"currency"`
	CheckoutSessionID string    `json:"-"`
	PaymentIntentID   string    `json:"-"`
	InvoiceID         string    `json:"-"`
	RefundID          string    `json:"-"`
	CreatedAt         time.Time `json:"createdAt"`
	UpdatedAt         time.Time `json:"updatedAt"`
}

type ProviderAccount struct {
	InternalUserID   string    `json:"-"`
	StripeAccountID  string    `json:"-"`
	DetailsSubmitted bool      `json:"detailsSubmitted"`
	ChargesEnabled   bool      `json:"chargesEnabled"`
	PayoutsEnabled   bool      `json:"payoutsEnabled"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

type CheckoutRequest struct {
	OrderID            string
	BookingID          string
	ConnectedAccountID string
	GrossMinor         int64
	FeeMinor           int64
	Locale             string
	IdempotencyKey     string
}

type CheckoutSession struct {
	ID  string
	URL string
}

type ConnectedAccount struct {
	ID               string
	DetailsSubmitted bool
	ChargesEnabled   bool
	PayoutsEnabled   bool
}

type AccountLink struct{ URL string }

type RefundResult struct{ ID string }

type ProviderEvent struct {
	ID               string
	Kind             EventKind
	ProviderObjectID string
	OrderID          string
	PaymentIntentID  string
	InvoiceID        string
	RefundID         string
	AccountID        string
	ChargeID         string
	DisputeID        string
	DisputeState     string
	DisputeReason    string
	AmountMinor      int64
	Currency         string
	OccurredAt       time.Time
	DetailsSubmitted bool
	ChargesEnabled   bool
	PayoutsEnabled   bool
}
