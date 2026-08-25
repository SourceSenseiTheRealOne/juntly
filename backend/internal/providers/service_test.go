package providers

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/provideraccess"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/reference"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/users"
	"github.com/google/uuid"
)

func TestServiceGetIsScopedToAuthorizedOwner(t *testing.T) {
	t.Parallel()

	owner := providerOwner()
	authorizer := &recordingAuthorizer{owner: owner}
	repository := &recordingRepository{}

	profile, err := NewService(authorizer, repository, &recordingReferenceValidator{}).Get(
		context.Background(), users.VerifiedIdentity{Subject: "user_provider"},
	)

	if err != nil || profile != nil {
		t.Fatalf("profile = %#v, err = %v, want nil nil", profile, err)
	}
	if repository.owner != owner.ID {
		t.Fatalf("repository owner = %s, want %s", repository.owner, owner.ID)
	}
}

func TestServiceRejectsUnauthorizedAndDisabledProviderBeforeRepository(t *testing.T) {
	t.Parallel()

	for _, authorizationError := range []error{provideraccess.ErrUnauthorized, provideraccess.ErrForbidden, provideraccess.ErrUnavailable} {
		repository := &recordingRepository{}
		_, err := NewService(
			&recordingAuthorizer{err: authorizationError},
			repository,
			&recordingReferenceValidator{},
		).Get(context.Background(), users.VerifiedIdentity{})
		if !errors.Is(err, authorizationError) {
			t.Fatalf("error = %v, want %v", err, authorizationError)
		}
		if repository.calls != 0 {
			t.Fatalf("repository calls = %d, want 0", repository.calls)
		}
	}
}

func TestServiceValidatesFullReplacementBeforeReferencesAndRepository(t *testing.T) {
	t.Parallel()

	valid := validReplacement()
	cases := map[string]ReplaceProfile{
		"short display name":       replace(valid, func(value *ReplaceProfile) { value.DisplayName = "A" }),
		"invalid provider type":    replace(valid, func(value *ReplaceProfile) { value.ProviderType = "admin" }),
		"long biography":           replace(valid, func(value *ReplaceProfile) { value.Bio = strings.Repeat("x", 1001) }),
		"missing primary locality": replace(valid, func(value *ReplaceProfile) { value.PrimaryLocalityID = uuid.Nil }),
		"empty service localities": replace(valid, func(value *ReplaceProfile) { value.ServiceLocalityIDs = nil }),
		"duplicate localities": replace(valid, func(value *ReplaceProfile) {
			value.ServiceLocalityIDs = []uuid.UUID{value.PrimaryLocalityID, value.PrimaryLocalityID}
		}),
		"primary locality omitted": replace(valid, func(value *ReplaceProfile) { value.ServiceLocalityIDs = []uuid.UUID{uuid.New()} }),
		"negative radius":          replace(valid, func(value *ReplaceProfile) { value.MaxTravelDistanceKM = -1 }),
		"radius too large":         replace(valid, func(value *ReplaceProfile) { value.MaxTravelDistanceKM = 201 }),
		"no service mode": replace(valid, func(value *ReplaceProfile) {
			value.TravelsToCustomer = false
			value.ReceivesCustomer = false
			value.RemoteServices = false
		}),
		"zero travel only": replace(valid, func(value *ReplaceProfile) {
			value.MaxTravelDistanceKM = 0
			value.TravelsToCustomer = true
			value.ReceivesCustomer = false
			value.RemoteServices = false
		}),
		"empty languages":     replace(valid, func(value *ReplaceProfile) { value.LanguageCodes = nil }),
		"duplicate languages": replace(valid, func(value *ReplaceProfile) { value.LanguageCodes = []string{"pt-PT", "pt-PT"} }),
	}

	for name, input := range cases {
		name, input := name, input
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			repository := &recordingRepository{}
			references := &recordingReferenceValidator{}
			_, err := NewService(&recordingAuthorizer{owner: providerOwner()}, repository, references).Put(
				context.Background(), users.VerifiedIdentity{Subject: "user_provider"}, input,
			)
			if !errors.Is(err, ErrInvalidProfile) {
				t.Fatalf("error = %v, want ErrInvalidProfile", err)
			}
			if references.calls != 0 || repository.calls != 0 {
				t.Fatalf("calls = references:%d repository:%d, want none", references.calls, repository.calls)
			}
		})
	}
}

func TestServiceRejectsInactiveOrMissingReferences(t *testing.T) {
	t.Parallel()

	references := &recordingReferenceValidator{err: reference.ErrInvalidReference}
	repository := &recordingRepository{}
	_, err := NewService(&recordingAuthorizer{owner: providerOwner()}, repository, references).Put(
		context.Background(), users.VerifiedIdentity{Subject: "user_provider"}, validReplacement(),
	)

	if !errors.Is(err, ErrInvalidProfile) {
		t.Fatalf("error = %v, want ErrInvalidProfile", err)
	}
	if repository.calls != 0 {
		t.Fatalf("repository calls = %d, want 0", repository.calls)
	}
}

func TestServiceReplacesProfileForAuthorizedOwner(t *testing.T) {
	t.Parallel()

	owner := providerOwner()
	input := validReplacement()
	stored := profileFromReplacement(input)
	repository := &recordingRepository{replacement: stored}
	references := &recordingReferenceValidator{}

	profile, err := NewService(&recordingAuthorizer{owner: owner}, repository, references).Put(
		context.Background(), users.VerifiedIdentity{Subject: "user_provider"}, input,
	)

	if err != nil || !reflect.DeepEqual(profile, stored) {
		t.Fatalf("profile = %#v, err = %v, want %#v", profile, err, stored)
	}
	if repository.owner != owner.ID || !reflect.DeepEqual(repository.input, input) {
		t.Fatalf("repository owner/input = %s %#v", repository.owner, repository.input)
	}
	if references.calls != 1 {
		t.Fatalf("reference calls = %d, want 1", references.calls)
	}
}

func TestServiceMapsRepositoryFailureToUnavailable(t *testing.T) {
	t.Parallel()

	repository := &recordingRepository{err: errors.New("database private details")}
	_, err := NewService(&recordingAuthorizer{owner: providerOwner()}, repository, &recordingReferenceValidator{}).Get(
		context.Background(), users.VerifiedIdentity{Subject: "user_provider"},
	)
	if !errors.Is(err, ErrUnavailable) {
		t.Fatalf("error = %v, want ErrUnavailable", err)
	}
}

func providerOwner() users.InternalUser {
	return users.InternalUser{ID: uuid.MustParse("b00f5bf7-72a8-4a7d-bb23-2ef4c4daf3fb")}
}

func validReplacement() ReplaceProfile {
	primary := uuid.MustParse("9cd8c899-75ad-458d-9e40-a9f8ecdc7e48")
	return ReplaceProfile{
		DisplayName:         "Prestador local",
		ProviderType:        ProviderTypeIndividual,
		Bio:                 "Trabalho local de confiança.",
		PrimaryLocalityID:   primary,
		ServiceLocalityIDs:  []uuid.UUID{primary},
		MaxTravelDistanceKM: 25,
		TravelsToCustomer:   true,
		ReceivesCustomer:    false,
		RemoteServices:      false,
		LanguageCodes:       []string{"pt-PT"},
	}
}

func replace(value ReplaceProfile, mutate func(*ReplaceProfile)) ReplaceProfile {
	mutate(&value)
	return value
}

func profileFromReplacement(value ReplaceProfile) Profile {
	return Profile{
		DisplayName:         value.DisplayName,
		ProviderType:        value.ProviderType,
		Bio:                 value.Bio,
		PrimaryLocalityID:   value.PrimaryLocalityID,
		ServiceLocalityIDs:  value.ServiceLocalityIDs,
		MaxTravelDistanceKM: value.MaxTravelDistanceKM,
		TravelsToCustomer:   value.TravelsToCustomer,
		ReceivesCustomer:    value.ReceivesCustomer,
		RemoteServices:      value.RemoteServices,
		LanguageCodes:       value.LanguageCodes,
		CreatedAt:           time.Date(2026, 8, 23, 16, 0, 0, 0, time.UTC),
		UpdatedAt:           time.Date(2026, 8, 23, 16, 0, 0, 0, time.UTC),
	}
}

type recordingAuthorizer struct {
	owner users.InternalUser
	err   error
}

func (a *recordingAuthorizer) RequireProvider(context.Context, users.VerifiedIdentity) (users.InternalUser, error) {
	return a.owner, a.err
}

type recordingReferenceValidator struct {
	err   error
	calls int
}

func (v *recordingReferenceValidator) ValidateProfileReferences(context.Context, reference.ProfileReferences) error {
	v.calls++
	return v.err
}

type recordingRepository struct {
	profile     *Profile
	replacement Profile
	err         error
	calls       int
	owner       uuid.UUID
	input       ReplaceProfile
}

func (r *recordingRepository) FindByOwner(_ context.Context, owner uuid.UUID) (*Profile, error) {
	r.calls++
	r.owner = owner
	return r.profile, r.err
}

func (r *recordingRepository) Replace(_ context.Context, owner uuid.UUID, input ReplaceProfile) (Profile, error) {
	r.calls++
	r.owner = owner
	r.input = input
	return r.replacement, r.err
}
