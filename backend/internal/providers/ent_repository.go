package providers

import (
	"bytes"
	"context"
	"errors"
	"slices"
	"time"

	"github.com/SourceSenseiTheRealOne/juntly/backend/ent"
	"github.com/SourceSenseiTheRealOne/juntly/backend/ent/providerprofile"
	"github.com/google/uuid"
)

type entRepository struct{ client *ent.Client }

func NewEntRepository(client *ent.Client) Repository {
	return entRepository{client: client}
}

func (r entRepository) FindByOwner(ctx context.Context, owner uuid.UUID) (*Profile, error) {
	if r.client == nil {
		return nil, errors.New("Ent client is nil")
	}
	entity, err := queryProviderProfile(ctx, r.client, owner)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	profile := profileFromEnt(entity)
	return &profile, nil
}

func (r entRepository) Replace(ctx context.Context, owner uuid.UUID, input ReplaceProfile) (Profile, error) {
	if r.client == nil {
		return Profile{}, errors.New("Ent client is nil")
	}
	return r.replace(ctx, owner, canonicalReplacement(input), true)
}

func (r entRepository) replace(ctx context.Context, owner uuid.UUID, input ReplaceProfile, retryConflict bool) (Profile, error) {
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return Profile{}, err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()
	client := tx.Client()

	existing, err := queryProviderProfile(ctx, client, owner)
	switch {
	case ent.IsNotFound(err):
		_, err = client.ProviderProfile.Create().
			SetID(owner).
			SetDisplayName(input.DisplayName).
			SetProviderType(string(input.ProviderType)).
			SetBio(input.Bio).
			SetPrimaryLocalityID(input.PrimaryLocalityID).
			SetMaxTravelDistanceKm(input.MaxTravelDistanceKM).
			SetTravelsToCustomer(input.TravelsToCustomer).
			SetReceivesCustomer(input.ReceivesCustomer).
			SetRemoteServices(input.RemoteServices).
			AddServiceLocalityIDs(input.ServiceLocalityIDs...).
			AddSpokenLanguageIDs(input.LanguageCodes...).
			Save(ctx)
		if err != nil {
			_ = tx.Rollback()
			committed = true
			if retryConflict && ent.IsConstraintError(err) {
				return r.replace(ctx, owner, input, false)
			}
			return Profile{}, err
		}
	case err != nil:
		return Profile{}, err
	case profileMatchesReplacement(profileFromEnt(existing), input):
		profile := profileFromEnt(existing)
		if err := tx.Commit(); err != nil {
			return Profile{}, err
		}
		committed = true
		return profile, nil
	default:
		_, err = client.ProviderProfile.UpdateOneID(owner).
			SetDisplayName(input.DisplayName).
			SetProviderType(string(input.ProviderType)).
			SetBio(input.Bio).
			SetPrimaryLocalityID(input.PrimaryLocalityID).
			SetMaxTravelDistanceKm(input.MaxTravelDistanceKM).
			SetTravelsToCustomer(input.TravelsToCustomer).
			SetReceivesCustomer(input.ReceivesCustomer).
			SetRemoteServices(input.RemoteServices).
			ClearServiceLocalities().
			AddServiceLocalityIDs(input.ServiceLocalityIDs...).
			ClearSpokenLanguages().
			AddSpokenLanguageIDs(input.LanguageCodes...).
			Save(ctx)
		if err != nil {
			return Profile{}, err
		}
	}

	canonical, err := queryProviderProfile(ctx, client, owner)
	if err != nil {
		return Profile{}, err
	}
	profile := profileFromEnt(canonical)
	if err := tx.Commit(); err != nil {
		return Profile{}, err
	}
	committed = true
	return profile, nil
}

func queryProviderProfile(ctx context.Context, client *ent.Client, owner uuid.UUID) (*ent.ProviderProfile, error) {
	return client.ProviderProfile.Query().
		Where(providerprofile.IDEQ(owner)).
		WithServiceLocalities().
		WithSpokenLanguages().
		Only(ctx)
}

func profileFromEnt(entity *ent.ProviderProfile) Profile {
	localityIDs := make([]uuid.UUID, 0, len(entity.Edges.ServiceLocalities))
	for _, locality := range entity.Edges.ServiceLocalities {
		localityIDs = append(localityIDs, locality.ID)
	}
	languageCodes := make([]string, 0, len(entity.Edges.SpokenLanguages))
	for _, language := range entity.Edges.SpokenLanguages {
		languageCodes = append(languageCodes, language.ID)
	}
	sortUUIDs(localityIDs)
	slices.Sort(languageCodes)
	return Profile{
		DisplayName:         entity.DisplayName,
		ProviderType:        ProviderType(entity.ProviderType),
		Bio:                 entity.Bio,
		PrimaryLocalityID:   entity.PrimaryLocalityID,
		ServiceLocalityIDs:  localityIDs,
		MaxTravelDistanceKM: entity.MaxTravelDistanceKm,
		TravelsToCustomer:   entity.TravelsToCustomer,
		ReceivesCustomer:    entity.ReceivesCustomer,
		RemoteServices:      entity.RemoteServices,
		LanguageCodes:       languageCodes,
		CreatedAt:           normalizeTime(entity.CreatedAt),
		UpdatedAt:           normalizeTime(entity.UpdatedAt),
	}
}

func canonicalReplacement(input ReplaceProfile) ReplaceProfile {
	input.ServiceLocalityIDs = append([]uuid.UUID(nil), input.ServiceLocalityIDs...)
	input.LanguageCodes = append([]string(nil), input.LanguageCodes...)
	sortUUIDs(input.ServiceLocalityIDs)
	slices.Sort(input.LanguageCodes)
	return input
}

func profileMatchesReplacement(profile Profile, input ReplaceProfile) bool {
	return profile.DisplayName == input.DisplayName &&
		profile.ProviderType == input.ProviderType &&
		profile.Bio == input.Bio &&
		profile.PrimaryLocalityID == input.PrimaryLocalityID &&
		profile.MaxTravelDistanceKM == input.MaxTravelDistanceKM &&
		profile.TravelsToCustomer == input.TravelsToCustomer &&
		profile.ReceivesCustomer == input.ReceivesCustomer &&
		profile.RemoteServices == input.RemoteServices &&
		slices.Equal(profile.ServiceLocalityIDs, input.ServiceLocalityIDs) &&
		slices.Equal(profile.LanguageCodes, input.LanguageCodes)
}

func sortUUIDs(values []uuid.UUID) {
	slices.SortFunc(values, func(left, right uuid.UUID) int {
		return bytes.Compare(left[:], right[:])
	})
}

func normalizeTime(value time.Time) time.Time {
	return value.UTC().Truncate(time.Microsecond)
}
