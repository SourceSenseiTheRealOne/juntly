package httpapi_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/authn"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/httpapi"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/listings"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/users"
	"github.com/google/uuid"
)

func TestModerationListingHandlerRequiresIdentityAndRoutesReview(t *testing.T) {
	t.Parallel()
	service := &recordingModerationReview{listing: sampleListing()}
	handler := httpapi.NewModerationListingHandler(service)
	unauthorized := httptest.NewRecorder()
	handler.ServeHTTP(unauthorized, httptest.NewRequest(http.MethodGet, "/api/v1/moderation/listings", nil))
	if unauthorized.Code != http.StatusUnauthorized || service.calls != 0 {
		t.Fatalf("unauthorized=%d/%d", unauthorized.Code, service.calls)
	}

	for path, body := range map[string]string{"/api/v1/moderation/listings": "", "/api/v1/moderation/listings/aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa/approve": `{"revision":1}`, "/api/v1/moderation/listings/aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa/reject": `{"revision":1,"reason":"Needs clearer scope"}`} {
		response := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, path, strings.NewReader(body))
		if body != "" {
			request.Method = http.MethodPost
		}
		request.Header.Set("Authorization", "Bearer synthetic-token")
		request.Header.Set(httpapi.RequestIDHeader, "req_moderation")
		authn.RequireVerifiedIdentity(staticVerifier{identity: users.VerifiedIdentity{Subject: "moderator"}}, handler).ServeHTTP(response, request)
		if response.Code != http.StatusOK || strings.Contains(response.Body.String(), "internalUserId") {
			t.Fatalf("%s=%d/%s", path, response.Code, response.Body.String())
		}
	}
}

type recordingModerationReview struct {
	listing listings.Listing
	calls   int
}

func (s *recordingModerationReview) ListPending(context.Context, users.VerifiedIdentity) ([]listings.Listing, error) {
	s.calls++
	return []listings.Listing{s.listing}, nil
}
func (s *recordingModerationReview) Approve(context.Context, users.VerifiedIdentity, uuid.UUID, int) (listings.Listing, error) {
	s.calls++
	return s.listing, nil
}
func (s *recordingModerationReview) Reject(context.Context, users.VerifiedIdentity, uuid.UUID, int, string) (listings.Listing, error) {
	s.calls++
	return s.listing, nil
}
