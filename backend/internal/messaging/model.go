package messaging

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalid      = errors.New("messaging invalid request")
	ErrUnauthorized = errors.New("messaging unauthorized")
	ErrForbidden    = errors.New("messaging forbidden")
	ErrNotFound     = errors.New("messaging not found")
	ErrUnavailable  = errors.New("messaging unavailable")
)

type Conversation struct {
	ID         uuid.UUID
	ListingID  *uuid.UUID
	CustomerID uuid.UUID
	ProviderID uuid.UUID
	BlockedBy  *uuid.UUID
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type Message struct {
	ID             uuid.UUID
	ConversationID uuid.UUID
	SenderID       uuid.UUID
	Body           string
	CreatedAt      time.Time
}

type NotificationPreferences struct {
	InAppEnabled bool
	EmailEnabled bool
}

type Notification struct {
	ID        uuid.UUID
	Kind      string
	ReadAt    *time.Time
	CreatedAt time.Time
}
