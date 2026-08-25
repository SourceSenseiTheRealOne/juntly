package providers

import (
	"context"
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/reference"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/users"
	"github.com/google/uuid"
)

type ProviderAuthorizer interface {
	RequireProvider(context.Context, users.VerifiedIdentity) (users.InternalUser, error)
}

type ReferenceValidator interface {
	ValidateProfileReferences(context.Context, reference.ProfileReferences) error
}

type Service interface {
	Get(context.Context, users.VerifiedIdentity) (*Profile, error)
	Put(context.Context, users.VerifiedIdentity, ReplaceProfile) (Profile, error)
}

type service struct {
	authorizer ProviderAuthorizer
	repository Repository
	references ReferenceValidator
}

func NewService(authorizer ProviderAuthorizer, repository Repository, references ReferenceValidator) Service {
	return service{authorizer: authorizer, repository: repository, references: references}
}

func (s service) Get(ctx context.Context, identity users.VerifiedIdentity) (*Profile, error) {
	owner, err := s.authorize(ctx, identity)
	if err != nil {
		return nil, err
	}
	if s.repository == nil {
		return nil, ErrUnavailable
	}
	profile, err := s.repository.FindByOwner(ctx, owner.ID)
	if err != nil {
		return nil, ErrUnavailable
	}
	return cloneProfile(profile), nil
}

func (s service) Put(ctx context.Context, identity users.VerifiedIdentity, input ReplaceProfile) (Profile, error) {
	owner, err := s.authorize(ctx, identity)
	if err != nil {
		return Profile{}, err
	}
	input, valid := normalizeReplacement(input)
	if !valid {
		return Profile{}, ErrInvalidProfile
	}
	if s.references == nil || s.repository == nil {
		return Profile{}, ErrUnavailable
	}
	if err := s.references.ValidateProfileReferences(ctx, reference.ProfileReferences{
		PrimaryLocalityID:  input.PrimaryLocalityID,
		ServiceLocalityIDs: input.ServiceLocalityIDs,
		LanguageCodes:      input.LanguageCodes,
	}); err != nil {
		if errors.Is(err, reference.ErrInvalidReference) {
			return Profile{}, ErrInvalidProfile
		}
		return Profile{}, ErrUnavailable
	}
	profile, err := s.repository.Replace(ctx, owner.ID, input)
	if err != nil {
		return Profile{}, ErrUnavailable
	}
	return *cloneProfile(&profile), nil
}

func (s service) authorize(ctx context.Context, identity users.VerifiedIdentity) (users.InternalUser, error) {
	if s.authorizer == nil {
		return users.InternalUser{}, ErrUnavailable
	}
	return s.authorizer.RequireProvider(ctx, identity)
}

func normalizeReplacement(input ReplaceProfile) (ReplaceProfile, bool) {
	input.DisplayName = strings.TrimSpace(input.DisplayName)
	input.Bio = strings.TrimSpace(input.Bio)
	input.ServiceLocalityIDs = append([]uuid.UUID(nil), input.ServiceLocalityIDs...)
	input.LanguageCodes = append([]string(nil), input.LanguageCodes...)
	if utf8.RuneCountInString(input.DisplayName) < 2 || utf8.RuneCountInString(input.DisplayName) > 100 {
		return ReplaceProfile{}, false
	}
	switch input.ProviderType {
	case ProviderTypeIndividual, ProviderTypeProfessional, ProviderTypeBusiness:
	default:
		return ReplaceProfile{}, false
	}
	if utf8.RuneCountInString(input.Bio) > 1000 || input.PrimaryLocalityID == uuid.Nil {
		return ReplaceProfile{}, false
	}
	if input.MaxTravelDistanceKM < 0 || input.MaxTravelDistanceKM > 200 {
		return ReplaceProfile{}, false
	}
	if !input.TravelsToCustomer && !input.ReceivesCustomer && !input.RemoteServices {
		return ReplaceProfile{}, false
	}
	if input.MaxTravelDistanceKM == 0 && input.TravelsToCustomer && !input.ReceivesCustomer && !input.RemoteServices {
		return ReplaceProfile{}, false
	}
	if !validUUIDSet(input.ServiceLocalityIDs, 1, 20, input.PrimaryLocalityID) {
		return ReplaceProfile{}, false
	}
	if !validLanguageSet(input.LanguageCodes) {
		return ReplaceProfile{}, false
	}
	return input, true
}

func validUUIDSet(values []uuid.UUID, minimum, maximum int, required uuid.UUID) bool {
	if len(values) < minimum || len(values) > maximum {
		return false
	}
	seen := make(map[uuid.UUID]struct{}, len(values))
	hasRequired := false
	for _, value := range values {
		if value == uuid.Nil {
			return false
		}
		if _, exists := seen[value]; exists {
			return false
		}
		seen[value] = struct{}{}
		if value == required {
			hasRequired = true
		}
	}
	return hasRequired
}

func validLanguageSet(values []string) bool {
	if len(values) < 1 || len(values) > 10 {
		return false
	}
	seen := make(map[string]struct{}, len(values))
	for index, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || utf8.RuneCountInString(value) > 10 {
			return false
		}
		if _, exists := seen[value]; exists {
			return false
		}
		seen[value] = struct{}{}
		values[index] = value
	}
	return true
}

func cloneProfile(profile *Profile) *Profile {
	if profile == nil {
		return nil
	}
	cloned := *profile
	cloned.ServiceLocalityIDs = append([]uuid.UUID(nil), profile.ServiceLocalityIDs...)
	cloned.LanguageCodes = append([]string(nil), profile.LanguageCodes...)
	return &cloned
}
