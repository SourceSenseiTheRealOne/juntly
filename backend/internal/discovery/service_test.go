package discovery

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
)

func TestServiceRejectsUnpairedOrInvalidDiscoveryFilters(t *testing.T) {
	t.Parallel()
	repository := &recordingRepository{}
	for _, request := range []Request{
		{Locale: "pt-PT", NearLocalityID: uuid.New()},
		{Locale: "pt-PT", RadiusKM: 25},
		{Locale: "pt-PT", Query: "x"},
		{Locale: "pt-PT", RadiusKM: 201},
		{Locale: "fr"},
	} {
		_, err := NewService(repository).Search(context.Background(), request)
		if !errors.Is(err, ErrInvalidRequest) || repository.calls != 0 {
			t.Fatalf("err/calls=%v/%d", err, repository.calls)
		}
	}
}
func TestServiceReturnsActivePublicProjectionOnlyThroughRepository(t *testing.T) {
	t.Parallel()
	repository := &recordingRepository{results: []Listing{{ID: uuid.New(), Title: "Active listing", CategorySlug: "plumbing", LocalitySlug: "zebreira", PriceType: "fixed", Currency: "EUR"}}}
	request := Request{Locale: "en", Query: "plumbing"}
	values, err := NewService(repository).Search(context.Background(), request)
	if err != nil || len(values) != 1 || repository.calls != 1 || repository.request.Query != "plumbing" {
		t.Fatalf("values/err/repository=%#v/%v/%#v", values, err, repository)
	}
}

func TestServiceHidesMalformedOrMissingPublicListing(t *testing.T) {
	t.Parallel()
	repository := &recordingRepository{err: ErrNotFound}
	service := NewService(repository)
	value, err := service.Get(context.Background(), "not-a-uuid", "pt-PT")
	if value != nil || !errors.Is(err, ErrNotFound) || repository.getCalls != 0 {
		t.Fatalf("malformed value/error/calls = %#v/%v/%d", value, err, repository.getCalls)
	}
	value, err = service.Get(context.Background(), uuid.NewString(), "pt-PT")
	if value != nil || !errors.Is(err, ErrNotFound) || repository.getCalls != 1 {
		t.Fatalf("missing value/error/calls = %#v/%v/%d", value, err, repository.getCalls)
	}
}

func TestServiceRejectsUnsupportedDetailLocale(t *testing.T) {
	t.Parallel()
	value, err := NewService(&recordingRepository{}).Get(context.Background(), uuid.NewString(), "fr")
	if value != nil || !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("value/error = %#v/%v", value, err)
	}
}

type recordingRepository struct {
	request  Request
	results  []Listing
	calls    int
	getCalls int
	err      error
}

func (r *recordingRepository) Search(_ context.Context, q Request) ([]Listing, error) {
	r.calls++
	r.request = q
	return r.results, r.err
}
func (r *recordingRepository) Get(context.Context, uuid.UUID, string) (*Listing, error) {
	r.getCalls++
	return nil, r.err
}
