package users

import (
	"context"
	"errors"
	"time"

	"github.com/SourceSenseiTheRealOne/juntly/backend/ent"
	"github.com/SourceSenseiTheRealOne/juntly/backend/ent/internaluser"
)

type entRepository struct {
	client *ent.Client
}

func NewEntRepository(client *ent.Client) Repository {
	return entRepository{client: client}
}

func (r entRepository) FindBySubject(ctx context.Context, subject string) (InternalUser, bool, error) {
	if r.client == nil {
		return InternalUser{}, false, errors.New("Ent client is nil")
	}

	entity, err := r.client.InternalUser.Query().Where(internaluser.ClerkSubjectEQ(subject)).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return InternalUser{}, false, nil
		}
		return InternalUser{}, false, err
	}
	return internalUserFromEnt(entity), true, nil
}

func (r entRepository) Create(ctx context.Context, subject string) (InternalUser, error) {
	if r.client == nil {
		return InternalUser{}, errors.New("Ent client is nil")
	}

	entity, err := r.client.InternalUser.Create().SetClerkSubject(subject).Save(ctx)
	if err != nil {
		if ent.IsConstraintError(err) {
			return InternalUser{}, ErrSubjectConflict
		}
		return InternalUser{}, err
	}
	return internalUserFromEnt(entity), nil
}

func internalUserFromEnt(entity *ent.InternalUser) InternalUser {
	return InternalUser{
		ID:        entity.ID,
		CreatedAt: entity.CreatedAt.UTC().Truncate(time.Microsecond),
	}
}
