package accounts

import (
	"context"
	"errors"
	"time"

	"github.com/SourceSenseiTheRealOne/juntly/backend/ent"
	"github.com/SourceSenseiTheRealOne/juntly/backend/ent/useraccount"
	"github.com/google/uuid"
)

type entRepository struct {
	client *ent.Client
}

func NewEntRepository(client *ent.Client) Repository {
	return entRepository{client: client}
}

func (r entRepository) FindByInternalUserID(ctx context.Context, internalUserID uuid.UUID) (Record, bool, error) {
	if r.client == nil {
		return Record{}, false, errors.New("Ent client is nil")
	}

	entity, err := r.client.UserAccount.Query().Where(useraccount.IDEQ(internalUserID)).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return Record{}, false, nil
		}
		return Record{}, false, err
	}
	return recordFromEnt(entity), true, nil
}

func (r entRepository) Create(ctx context.Context, internalUserID uuid.UUID) (Record, error) {
	if r.client == nil {
		return Record{}, errors.New("Ent client is nil")
	}

	entity, err := r.client.UserAccount.Create().SetID(internalUserID).Save(ctx)
	if err != nil {
		if ent.IsConstraintError(err) {
			return Record{}, ErrAccountConflict
		}
		return Record{}, err
	}
	return recordFromEnt(entity), nil
}

func (r entRepository) SetProviderEnabled(ctx context.Context, internalUserID uuid.UUID, enabled bool) (Record, error) {
	if r.client == nil {
		return Record{}, errors.New("Ent client is nil")
	}

	entity, err := r.client.UserAccount.UpdateOneID(internalUserID).SetProviderEnabled(enabled).Save(ctx)
	if err != nil {
		return Record{}, err
	}
	return recordFromEnt(entity), nil
}

func recordFromEnt(entity *ent.UserAccount) Record {
	return Record{
		InternalUserID:        entity.ID,
		ProviderEnabled:       entity.ProviderEnabled,
		OnboardingCompletedAt: normalizeDatabaseTime(entity.OnboardingCompletedAt),
		CreatedAt:             normalizeDatabaseTime(entity.CreatedAt),
		UpdatedAt:             normalizeDatabaseTime(entity.UpdatedAt),
	}
}

func normalizeDatabaseTime(value time.Time) time.Time {
	return value.UTC().Truncate(time.Microsecond)
}
