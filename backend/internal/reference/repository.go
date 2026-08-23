package reference

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Categories(context.Context, string) ([]Category, error)
	Languages(context.Context, string) ([]Language, error)
	Localities(context.Context, string) ([]Locality, error)
	NearbyLocalities(context.Context, uuid.UUID, int, string) ([]LocalityDistance, error)
	ValidateProfileReferences(context.Context, ProfileReferences) error
}
