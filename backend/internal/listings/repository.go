package listings

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Create(context.Context, uuid.UUID, CreateListing) (Listing, error)
	FindByOwner(context.Context, uuid.UUID, uuid.UUID) (*Listing, error)
	ListByOwner(context.Context, uuid.UUID) ([]Listing, error)
	ReplaceDraft(context.Context, uuid.UUID, uuid.UUID, int, CreateListing) (Listing, error)
	TransitionOwned(context.Context, uuid.UUID, uuid.UUID, State, State, int, *string) (Listing, error)
	TransitionModerated(context.Context, uuid.UUID, uuid.UUID, State, State, int, *string) (Listing, error)
}
