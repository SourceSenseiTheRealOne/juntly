package listingmedia

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

type unavailableStorage struct{}

func NewUnavailableStorage() Storage { return unavailableStorage{} }
func (unavailableStorage) CreateUploadReservation(context.Context, uuid.UUID, UploadRequest) (StorageReservation, error) {
	return StorageReservation{}, errors.New("listing media storage unavailable")
}
