package httpapi_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/authn"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/httpapi"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/listingmedia"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/listings"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/users"
	"github.com/google/uuid"
)

func TestListingHandlerRequiresIdentityAndStrictlyCreatesOwnerDraft(t *testing.T) {
	t.Parallel()
	service := &recordingListingService{created: sampleListing()}
	handler := httpapi.NewListingHandler(service)
	unauthorized := httptest.NewRecorder()
	handler.ServeHTTP(unauthorized, httptest.NewRequest(http.MethodPost, "/api/v1/me/listings", strings.NewReader(validListingJSON())))
	if unauthorized.Code != http.StatusUnauthorized || service.calls != 0 {
		t.Fatalf("unauthorized status/calls = %d/%d", unauthorized.Code, service.calls)
	}

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/me/listings", strings.NewReader(validListingJSON()))
	request.Header.Set("Authorization", "Bearer synthetic-token")
	request.Header.Set(httpapi.RequestIDHeader, "req_listing_create")
	authn.RequireVerifiedIdentity(staticVerifier{identity: users.VerifiedIdentity{Subject: "provider"}}, handler).ServeHTTP(response, request)
	if response.Code != http.StatusOK || service.createCalls != 1 || strings.Contains(response.Body.String(), "internalUserId") {
		t.Fatalf("create status/body = %d/%s", response.Code, response.Body.String())
	}
}

func TestListingHandlerMapsConflictWithoutInternalDetails(t *testing.T) {
	t.Parallel()
	service := &recordingListingService{err: listings.ErrConflict}
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/me/listings/aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa/submit", strings.NewReader(`{"revision":1}`))
	request.Header.Set("Authorization", "Bearer synthetic-token")
	request.Header.Set(httpapi.RequestIDHeader, "req_listing_conflict")
	authn.RequireVerifiedIdentity(staticVerifier{identity: users.VerifiedIdentity{Subject: "provider"}}, httpapi.NewListingHandler(service)).ServeHTTP(response, request)
	if response.Code != http.StatusConflict || !strings.Contains(response.Body.String(), "CONFLICT") || strings.Contains(response.Body.String(), "private") {
		t.Fatalf("conflict status/body=%d/%s", response.Code, response.Body.String())
	}
}

func TestListingHandlerHandlesOwnerPauseArchiveAndUploadIntent(t *testing.T) {
	t.Parallel()
	service := &recordingListingService{created: sampleListing(), intent: listingmedia.UploadIntent{MediaID: uuid.MustParse("dddddddd-dddd-4ddd-8ddd-dddddddddddd"), Capability: listingmedia.UploadCapability{URL: "https://upload.example.invalid/capability", Method: "PUT", Headers: map[string]string{"Content-Type": "image/webp"}}}}
	handler := authn.RequireVerifiedIdentity(staticVerifier{identity: users.VerifiedIdentity{Subject: "provider"}}, httpapi.NewListingHandler(service))
	for path, body := range map[string]string{
		"/api/v1/me/listings/aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa/pause":                `{"revision":1}`,
		"/api/v1/me/listings/aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa/archive":              `{"revision":1,"state":"paused"}`,
		"/api/v1/me/listings/aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa/media/upload-intents": `{"ordinal":1,"contentType":"image/webp","byteSize":1024,"checksumSha256":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}`,
	} {
		response := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
		request.Header.Set("Authorization", "Bearer synthetic-token")
		request.Header.Set(httpapi.RequestIDHeader, "req_listing_owner_action")
		handler.ServeHTTP(response, request)
		if response.Code != http.StatusOK || strings.Contains(response.Body.String(), "objectReference") {
			t.Fatalf("%s status/body=%d/%s", path, response.Code, response.Body.String())
		}
	}
}

type recordingListingService struct {
	created     listings.Listing
	intent      listingmedia.UploadIntent
	err         error
	calls       int
	createCalls int
}

func (s *recordingListingService) Create(context.Context, users.VerifiedIdentity, listings.CreateListing) (listings.Listing, error) {
	s.calls++
	s.createCalls++
	return s.created, s.err
}
func (s *recordingListingService) ReplaceDraft(context.Context, users.VerifiedIdentity, uuid.UUID, int, listings.CreateListing) (listings.Listing, error) {
	s.calls++
	return s.created, s.err
}
func (s *recordingListingService) Get(context.Context, users.VerifiedIdentity, uuid.UUID) (*listings.Listing, error) {
	s.calls++
	return &s.created, s.err
}
func (s *recordingListingService) List(context.Context, users.VerifiedIdentity) ([]listings.Listing, error) {
	s.calls++
	return []listings.Listing{s.created}, s.err
}
func (s *recordingListingService) Submit(context.Context, users.VerifiedIdentity, uuid.UUID, int) (listings.Listing, error) {
	s.calls++
	return s.created, s.err
}
func (s *recordingListingService) Pause(context.Context, users.VerifiedIdentity, uuid.UUID, int) (listings.Listing, error) {
	s.calls++
	return s.created, s.err
}
func (s *recordingListingService) Archive(context.Context, users.VerifiedIdentity, uuid.UUID, listings.State, int) (listings.Listing, error) {
	s.calls++
	return s.created, s.err
}
func (s *recordingListingService) CreateUploadIntent(context.Context, users.VerifiedIdentity, uuid.UUID, listingmedia.UploadRequest) (listingmedia.UploadIntent, error) {
	s.calls++
	return s.intent, s.err
}
func sampleListing() listings.Listing {
	price := 5000
	return listings.Listing{ID: uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"), CreateListing: listings.CreateListing{CategoryID: uuid.MustParse("bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb"), PrimaryLocalityID: uuid.MustParse("cccccccc-cccc-4ccc-8ccc-cccccccccccc"), Title: "Listing test", Description: "Synthetic listing handler contract description.", PriceType: listings.PriceTypeFixed, PriceMinor: &price, Currency: "EUR", TravelsToCustomer: true}, State: listings.StateDraft, Revision: 1}
}
func validListingJSON() string {
	return `{"categoryId":"bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb","primaryLocalityId":"cccccccc-cccc-4ccc-8ccc-cccccccccccc","title":"Listing test","description":"Synthetic listing handler contract description.","priceType":"fixed","priceMinor":5000,"currency":"EUR","travelsToCustomer":true,"receivesCustomer":false,"remoteServices":false}`
}
