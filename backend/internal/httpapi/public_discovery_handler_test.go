package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/discovery"
	"github.com/google/uuid"
)

func TestPublicDiscoveryHandlerServesClosedActiveProjectionWithoutIdentity(t *testing.T) {
	listingID := uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa")
	service := &recordingDiscoveryService{values: []discovery.Listing{{
		ID: listingID, Title: "Public plumbing", Description: "Public plumbing listing with enough descriptive text.",
		CategoryID: uuid.MustParse("bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb"), CategorySlug: "plumbing", CategoryName: "Canalização",
		PrimaryLocalityID: uuid.MustParse("cccccccc-cccc-4ccc-8ccc-cccccccccccc"), LocalitySlug: "zebreira", LocalityName: "Zebreira",
		PriceType: discovery.PriceTypeFixed, Currency: "EUR", TravelsToCustomer: true,
		ProviderDisplayName: "Public provider", ProviderType: "professional", UpdatedAt: time.Date(2026, 8, 24, 12, 0, 0, 0, time.UTC),
	}}}
	handler := NewPublicDiscoveryHandler(service)
	request := httptest.NewRequest(http.MethodGet, "/api/v1/discovery/listings?locale=pt-PT&q=plumbing", nil)
	request.Header.Set(RequestIDHeader, "req_discovery_public")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK || service.searches != 1 || service.request.Locale != "pt-PT" || service.request.Query != "plumbing" {
		t.Fatalf("status/service = %d/%#v", response.Code, service)
	}
	var body map[string]any
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	encoded, _ := json.Marshal(body)
	for _, forbidden := range []string{"internalUserId", "clerkSubject", "objectReference", "latitude", "longitude", "bio", "reason"} {
		if string(encoded) != "" && containsJSONField(body, forbidden) {
			t.Fatalf("public response contains %q: %s", forbidden, encoded)
		}
	}
}

func TestPublicDiscoveryHandlerRejectsUnknownOrUnpairedQueries(t *testing.T) {
	service := &recordingDiscoveryService{}
	handler := NewPublicDiscoveryHandler(service)
	for _, target := range []string{
		"/api/v1/discovery/listings?locale=pt-PT&admin=true",
		"/api/v1/discovery/listings?locale=pt-PT&radiusKm=25",
		"/api/v1/discovery/listings?locale=fr",
	} {
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, target, nil))
		if response.Code != http.StatusBadRequest || service.searches != 0 {
			t.Fatalf("target/status/calls = %s/%d/%d", target, response.Code, service.searches)
		}
	}
}

func TestPublicListingHandlerReturnsActiveLocalizedDetail(t *testing.T) {
	listingID := uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa")
	service := &detailDiscoveryService{value: &discovery.Listing{
		ID: listingID, Title: "Public plumbing", Description: "Public plumbing listing with enough descriptive text.",
		CategoryID: uuid.MustParse("bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb"), CategorySlug: "plumbing", CategoryName: "Canalização",
		PrimaryLocalityID: uuid.MustParse("cccccccc-cccc-4ccc-8ccc-cccccccccccc"), LocalitySlug: "zebreira", LocalityName: "Zebreira",
		PriceType: discovery.PriceTypeFixed, Currency: "EUR", TravelsToCustomer: true,
		ProviderDisplayName: "Public provider", ProviderType: "professional", UpdatedAt: time.Date(2026, 8, 24, 12, 0, 0, 0, time.UTC),
	}}
	handler := NewPublicListingHandler(service)
	request := httptest.NewRequest(http.MethodGet, "/api/v1/public/listings/aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa?locale=pt-PT", nil)
	request.Header.Set(RequestIDHeader, "req_public_detail")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK || service.gets != 1 || service.locale != "pt-PT" {
		t.Fatalf("status/service = %d/%#v", response.Code, service)
	}
}

type detailDiscoveryService struct {
	value  *discovery.Listing
	gets   int
	locale string
}

func (s *detailDiscoveryService) Search(context.Context, discovery.Request) ([]discovery.Listing, error) {
	return nil, nil
}
func (s *detailDiscoveryService) Get(_ context.Context, _ string, locale string) (*discovery.Listing, error) {
	s.gets++
	s.locale = locale
	return s.value, nil
}

type recordingDiscoveryService struct {
	request  discovery.Request
	values   []discovery.Listing
	searches int
}

func (s *recordingDiscoveryService) Search(_ context.Context, request discovery.Request) ([]discovery.Listing, error) {
	s.searches++
	s.request = request
	return s.values, nil
}
func (s *recordingDiscoveryService) Get(context.Context, string, string) (*discovery.Listing, error) {
	return nil, discovery.ErrNotFound
}

func containsJSONField(value any, field string) bool {
	object, ok := value.(map[string]any)
	if !ok {
		return false
	}
	if _, exists := object[field]; exists {
		return true
	}
	for _, nested := range object {
		if containsJSONField(nested, field) {
			return true
		}
	}
	return false
}
