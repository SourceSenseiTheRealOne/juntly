package providers

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	FindByOwner(context.Context, uuid.UUID) (*Profile, error)
	Replace(context.Context, uuid.UUID, ReplaceProfile) (Profile, error)
}
