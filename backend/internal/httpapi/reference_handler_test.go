package httpapi_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/httpapi"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/reference"
	"github.com/google/uuid"
)

func TestReferenceHandlersReturnClosedPublicCatalogs(t *testing.T) {
	t.Parallel()

	parentID := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	categoryID := uuid.MustParse("22222222-2222-4222-8222-222222222222")
	localityID := uuid.MustParse("33333333-3333-4333-8333-333333333333")
	service := &recordingReferenceService{
		categories: []reference.Category{{ID: categoryID, ParentID: &parentID, Slug: "cleaning", Name: "Limpeza"}},
		languages:  []reference.Language{{Code: "pt-PT", Name: "Português"}},
		localities: []reference.Locality{{ID: localityID, Slug: "zebreira", Name: "Zebreira", ParishName: "Zebreira e Segura", MunicipalityName: "Idanha-a-Nova", DistrictName: "Castelo Branco"}},
	}

	for path, handler := range map[string]http.Handler{
		"/api/v1/catalog/categories?locale=pt-PT":   httpapi.NewCategoriesHandler(service),
		"/api/v1/reference/languages?locale=pt-PT":  httpapi.NewLanguagesHandler(service),
		"/api/v1/reference/localities?locale=pt-PT": httpapi.NewLocalitiesHandler(service),
	} {
		response := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, path, nil)
		request.Header.Set(httpapi.RequestIDHeader, "req_reference_public")
		handler.ServeHTTP(response, request)
		if response.Code != http.StatusOK {
			t.Fatalf("%s status = %d", path, response.Code)
		}
		if response.Header().Get(httpapi.RequestIDHeader) != "req_reference_public" {
			t.Fatalf("%s request ID mismatch", path)
		}
		body := response.Body.String()
		for _, prohibited := range []string{"latitude", "longitude", "internalUser", "clerk", "phone", "address"} {
			if strings.Contains(strings.ToLower(body), strings.ToLower(prohibited)) {
				t.Fatalf("%s response exposes %q: %s", path, prohibited, body)
			}
		}
	}
}

func TestLocalitiesHandlerSupportsPairedRadiusQuery(t *testing.T) {
	t.Parallel()

	origin := uuid.MustParse("33333333-3333-4333-8333-333333333333")
	service := &recordingReferenceService{nearby: []reference.LocalityDistance{{Locality: reference.Locality{ID: origin, Slug: "zebreira", Name: "Zebreira", ParishName: "Zebreira e Segura", MunicipalityName: "Idanha-a-Nova", DistrictName: "Castelo Branco"}, DistanceMeters: 0}}}
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/reference/localities?locale=pt-PT&nearLocalityId="+origin.String()+"&radiusKm=25", nil)
	request.Header.Set(httpapi.RequestIDHeader, "req_reference_radius")

	httpapi.NewLocalitiesHandler(service).ServeHTTP(response, request)

	if response.Code != http.StatusOK || service.nearbyCalls != 1 || service.origin != origin || service.radius != 25 {
		t.Fatalf("status/calls/origin/radius = %d %d %s %d", response.Code, service.nearbyCalls, service.origin, service.radius)
	}
	if !strings.Contains(response.Body.String(), `"distanceMeters":0`) {
		t.Fatalf("radius response = %s", response.Body.String())
	}
}

func TestReferenceHandlersRejectUnknownOrUnpairedQueriesBeforeService(t *testing.T) {
	t.Parallel()

	cases := []struct {
		path    string
		handler http.Handler
	}{
		{"/api/v1/catalog/categories?locale=pt-PT&admin=true", httpapi.NewCategoriesHandler(&recordingReferenceService{})},
		{"/api/v1/reference/languages?locale=fr", httpapi.NewLanguagesHandler(&recordingReferenceService{})},
		{"/api/v1/reference/localities?locale=pt-PT&radiusKm=25", httpapi.NewLocalitiesHandler(&recordingReferenceService{})},
		{"/api/v1/reference/localities?locale=pt-PT&nearLocalityId=not-a-uuid&radiusKm=25", httpapi.NewLocalitiesHandler(&recordingReferenceService{})},
	}
	for _, item := range cases {
		response := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, item.path, nil)
		request.Header.Set(httpapi.RequestIDHeader, "req_reference_invalid")
		item.handler.ServeHTTP(response, request)
		if response.Code != http.StatusBadRequest {
			t.Fatalf("%s status = %d, want 400", item.path, response.Code)
		}
		assertErrorResponse(t, response, "INVALID_REQUEST", "Invalid request", "req_reference_invalid")
	}
}

func TestReferenceHandlerReturnsSafeUnavailable(t *testing.T) {
	t.Parallel()

	service := &recordingReferenceService{err: reference.ErrUnavailable}
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/catalog/categories?locale=en", nil)
	request.Header.Set(httpapi.RequestIDHeader, "req_reference_unavailable")

	httpapi.NewCategoriesHandler(service).ServeHTTP(response, request)

	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", response.Code)
	}
	assertErrorResponse(t, response, "SERVICE_UNAVAILABLE", "Service unavailable", "req_reference_unavailable")
}

type recordingReferenceService struct {
	categories  []reference.Category
	languages   []reference.Language
	localities  []reference.Locality
	nearby      []reference.LocalityDistance
	err         error
	nearbyCalls int
	origin      uuid.UUID
	radius      int
}

func (s *recordingReferenceService) Categories(context.Context, string) ([]reference.Category, error) {
	return s.categories, s.err
}
func (s *recordingReferenceService) Languages(context.Context, string) ([]reference.Language, error) {
	return s.languages, s.err
}
func (s *recordingReferenceService) Localities(context.Context, string) ([]reference.Locality, error) {
	return s.localities, s.err
}
func (s *recordingReferenceService) NearbyLocalities(_ context.Context, origin uuid.UUID, radius int, _ string) ([]reference.LocalityDistance, error) {
	s.nearbyCalls++
	s.origin = origin
	s.radius = radius
	return s.nearby, s.err
}
