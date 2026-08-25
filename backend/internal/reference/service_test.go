package reference

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
)

func TestServiceRejectsUnsupportedLocaleBeforeRepositoryAccess(t *testing.T) {
	t.Parallel()

	repository := &recordingRepository{}
	service := NewService(repository)
	for _, locale := range []string{"", "fr", "pt", "pt-PT-extra"} {
		if _, err := service.Categories(context.Background(), locale); !errors.Is(err, ErrInvalidRequest) {
			t.Fatalf("locale %q error = %v, want ErrInvalidRequest", locale, err)
		}
	}
	if repository.calls != 0 {
		t.Fatalf("repository calls = %d, want 0", repository.calls)
	}
}

func TestServiceReturnsReferenceCatalogs(t *testing.T) {
	t.Parallel()

	category := Category{ID: uuid.New(), Slug: "cleaning", Name: "Limpeza", SortOrder: 10}
	language := Language{Code: "pt-PT", Name: "Português", SortOrder: 10}
	locality := Locality{ID: uuid.New(), Slug: "zebreira", Name: "Zebreira", ParishName: "Zebreira e Segura", MunicipalityName: "Idanha-a-Nova", DistrictName: "Castelo Branco"}
	repository := &recordingRepository{
		categories: []Category{category},
		languages:  []Language{language},
		localities: []Locality{locality},
	}
	service := NewService(repository)
	ctx := context.Background()

	categories, err := service.Categories(ctx, "pt-PT")
	if err != nil || len(categories) != 1 || categories[0] != category {
		t.Fatalf("categories = %#v, err = %v", categories, err)
	}
	languages, err := service.Languages(ctx, "pt-PT")
	if err != nil || len(languages) != 1 || languages[0] != language {
		t.Fatalf("languages = %#v, err = %v", languages, err)
	}
	localities, err := service.Localities(ctx, "pt-PT")
	if err != nil || len(localities) != 1 || localities[0] != locality {
		t.Fatalf("localities = %#v, err = %v", localities, err)
	}
}

func TestServiceValidatesNearbyLocalityRequest(t *testing.T) {
	t.Parallel()

	repository := &recordingRepository{}
	service := NewService(repository)
	ctx := context.Background()
	validID := uuid.New()

	for _, test := range []struct {
		origin uuid.UUID
		radius int
		locale string
	}{
		{uuid.Nil, 10, "pt-PT"},
		{validID, 0, "pt-PT"},
		{validID, 201, "pt-PT"},
		{validID, 10, "fr"},
	} {
		if _, err := service.NearbyLocalities(ctx, test.origin, test.radius, test.locale); !errors.Is(err, ErrInvalidRequest) {
			t.Fatalf("request %#v error = %v, want ErrInvalidRequest", test, err)
		}
	}
	if repository.calls != 0 {
		t.Fatalf("repository calls = %d, want 0", repository.calls)
	}
}

func TestServiceMapsRepositoryFailures(t *testing.T) {
	t.Parallel()

	repository := &recordingRepository{err: errors.New("database internal details")}
	service := NewService(repository)
	ctx := context.Background()

	if _, err := service.Categories(ctx, "en"); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("categories error = %v, want ErrUnavailable", err)
	}
	if _, err := service.NearbyLocalities(ctx, uuid.New(), 10, "es"); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("nearby error = %v, want ErrUnavailable", err)
	}
}

type recordingRepository struct {
	categories []Category
	languages  []Language
	localities []Locality
	nearby     []LocalityDistance
	err        error
	calls      int
}

func (r *recordingRepository) Categories(context.Context, string) ([]Category, error) {
	r.calls++
	return r.categories, r.err
}

func (r *recordingRepository) Languages(context.Context, string) ([]Language, error) {
	r.calls++
	return r.languages, r.err
}

func (r *recordingRepository) Localities(context.Context, string) ([]Locality, error) {
	r.calls++
	return r.localities, r.err
}

func (r *recordingRepository) NearbyLocalities(context.Context, uuid.UUID, int, string) ([]LocalityDistance, error) {
	r.calls++
	return r.nearby, r.err
}

func (r *recordingRepository) ValidateProfileReferences(context.Context, ProfileReferences) error {
	r.calls++
	return r.err
}
