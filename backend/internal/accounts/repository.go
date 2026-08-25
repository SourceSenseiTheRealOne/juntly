package accounts

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

var (
	ErrAccountConflict = errors.New("user account conflict")
	ErrUnavailable     = errors.New("user account persistence unavailable")
	ErrInvalidIdentity = errors.New("invalid verified identity")
)

type Repository interface {
	FindByInternalUserID(context.Context, uuid.UUID) (Record, bool, error)
	Create(context.Context, uuid.UUID) (Record, error)
	SetProviderEnabled(context.Context, uuid.UUID, bool) (Record, error)
}
