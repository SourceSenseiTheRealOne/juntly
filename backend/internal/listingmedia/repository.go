package listingmedia

import (
	"context"

	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/users"
	"github.com/google/uuid"
)

type ProviderAuthorizer interface {
	RequireProvider(context.Context, users.VerifiedIdentity) (users.InternalUser, error)
}
type Repository interface {
	ReservePending(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, UploadRequest, string) error
}
type Storage interface {
	CreateUploadReservation(context.Context, uuid.UUID, UploadRequest) (StorageReservation, error)
}
