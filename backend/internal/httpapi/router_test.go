package httpapi_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/accounts"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/contactreveal"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/discovery"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/health"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/httpapi"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/users"
	"github.com/google/uuid"
)

func TestRouterLeavesHealthPublicAndProtectsReconciliation(t *testing.T) {
	t.Parallel()

	verifier := &routerVerifier{identity: users.VerifiedIdentity{Subject: "user_synthetic"}}
	reconcileService := &recordingReconcileService{user: users.InternalUser{
		ID:        uuid.MustParse("7b7b7d7e-38f9-4f0c-8a10-0fce9cf6f82b"),
		CreatedAt: time.Date(2026, 8, 21, 12, 0, 0, 0, time.UTC),
	}}
	healthService := health.NewService("0.1.0", func() time.Time {
		return time.Date(2026, 8, 21, 12, 0, 0, 0, time.UTC)
	})
	accountService := &recordingAccountService{account: accounts.Account{
		CustomerEnabled:       true,
		ProviderEnabled:       false,
		OnboardingCompletedAt: time.Date(2026, 8, 23, 12, 5, 0, 0, time.UTC),
	}}
	referenceService := &recordingReferenceService{}
	providerProfileService := &recordingProviderProfileService{}
	discoveryService := &recordingPublicDiscoveryService{}
	contactChannelService := &recordingRouterContactChannelService{}
	contactRevealService := &recordingRouterContactRevealService{}
	router := httpapi.NewRouter(healthService, nil, verifier, reconcileService, accountService, referenceService, providerProfileService, &recordingListingService{created: sampleListing()}, &recordingModerationReview{listing: sampleListing()}, discoveryService, contactChannelService, contactRevealService, nil, nil, nil, nil, nil, nil)

	healthResponse := httptest.NewRecorder()
	router.ServeHTTP(healthResponse, httptest.NewRequest(http.MethodGet, "/api/v1/health", nil))
	if healthResponse.Code != http.StatusOK {
		t.Fatalf("health status = %d, want %d", healthResponse.Code, http.StatusOK)
	}
	if verifier.calls != 0 {
		t.Fatalf("verifier calls after health = %d, want 0", verifier.calls)
	}

	publicDiscoveryResponse := httptest.NewRecorder()
	router.ServeHTTP(publicDiscoveryResponse, httptest.NewRequest(http.MethodGet, "/api/v1/discovery/listings?locale=pt-PT", nil))
	if publicDiscoveryResponse.Code != http.StatusOK || verifier.calls != 0 || discoveryService.calls != 1 {
		t.Fatalf("public discovery status/verifier/service = %d/%d/%d", publicDiscoveryResponse.Code, verifier.calls, discoveryService.calls)
	}

	unauthorizedResponse := httptest.NewRecorder()
	router.ServeHTTP(unauthorizedResponse, httptest.NewRequest(http.MethodPost, "/api/v1/auth/reconcile", nil))
	if unauthorizedResponse.Code != http.StatusUnauthorized {
		t.Fatalf("unauthorized reconciliation status = %d, want %d", unauthorizedResponse.Code, http.StatusUnauthorized)
	}
	if verifier.calls != 0 {
		t.Fatalf("verifier calls after missing bearer = %d, want 0", verifier.calls)
	}

	authorizedRequest := httptest.NewRequest(http.MethodPost, "/api/v1/auth/reconcile", nil)
	authorizedRequest.Header.Set("Authorization", "Bearer synthetic-token")
	authorizedResponse := httptest.NewRecorder()
	router.ServeHTTP(authorizedResponse, authorizedRequest)
	if authorizedResponse.Code != http.StatusOK {
		t.Fatalf("authorized reconciliation status = %d, want %d", authorizedResponse.Code, http.StatusOK)
	}
	if verifier.calls != 1 {
		t.Fatalf("verifier calls after valid bearer = %d, want 1", verifier.calls)
	}

	unauthorizedAccountResponse := httptest.NewRecorder()
	router.ServeHTTP(unauthorizedAccountResponse, httptest.NewRequest(http.MethodGet, "/api/v1/me/account", nil))
	if unauthorizedAccountResponse.Code != http.StatusUnauthorized {
		t.Fatalf("unauthorized account status = %d, want %d", unauthorizedAccountResponse.Code, http.StatusUnauthorized)
	}
	if accountService.calls != 0 {
		t.Fatalf("account service calls after missing bearer = %d, want 0", accountService.calls)
	}

	authorizedAccountRequest := httptest.NewRequest(http.MethodGet, "/api/v1/me/account", nil)
	authorizedAccountRequest.Header.Set("Authorization", "Bearer synthetic-token")
	authorizedAccountResponse := httptest.NewRecorder()
	router.ServeHTTP(authorizedAccountResponse, authorizedAccountRequest)
	if authorizedAccountResponse.Code != http.StatusOK {
		t.Fatalf("authorized account status = %d, want %d", authorizedAccountResponse.Code, http.StatusOK)
	}
	if verifier.calls != 2 {
		t.Fatalf("verifier calls after account bearer = %d, want 2", verifier.calls)
	}
	if accountService.calls != 1 {
		t.Fatalf("account service calls = %d, want 1", accountService.calls)
	}

	unauthorizedContactResponse := httptest.NewRecorder()
	router.ServeHTTP(unauthorizedContactResponse, httptest.NewRequest(http.MethodGet, "/api/v1/me/contact-channels", nil))
	if unauthorizedContactResponse.Code != http.StatusUnauthorized || contactChannelService.calls != 0 {
		t.Fatalf("contact status/calls = %d/%d", unauthorizedContactResponse.Code, contactChannelService.calls)
	}

	unauthorizedRevealResponse := httptest.NewRecorder()
	router.ServeHTTP(unauthorizedRevealResponse, httptest.NewRequest(http.MethodPost, "/api/v1/listings/aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa/contact-reveals", nil))
	if unauthorizedRevealResponse.Code != http.StatusUnauthorized || contactRevealService.calls != 0 {
		t.Fatalf("reveal status/calls = %d/%d", unauthorizedRevealResponse.Code, contactRevealService.calls)
	}
}

type recordingRouterContactChannelService struct{ calls int }

func (s *recordingRouterContactChannelService) Get(context.Context, users.VerifiedIdentity) ([]contactreveal.ChannelStatus, error) {
	s.calls++
	return nil, nil
}
func (s *recordingRouterContactChannelService) Put(context.Context, users.VerifiedIdentity, contactreveal.ReplaceChannel) (contactreveal.ChannelStatus, error) {
	s.calls++
	return contactreveal.ChannelStatus{}, nil
}

type recordingRouterContactRevealService struct{ calls int }

func (s *recordingRouterContactRevealService) Reveal(context.Context, users.VerifiedIdentity, uuid.UUID, contactreveal.Channel) (contactreveal.RevealedContact, error) {
	s.calls++
	return contactreveal.RevealedContact{}, nil
}

type recordingPublicDiscoveryService struct{ calls int }

func (s *recordingPublicDiscoveryService) Search(context.Context, discovery.Request) ([]discovery.Listing, error) {
	s.calls++
	return nil, nil
}

func (*recordingPublicDiscoveryService) Get(context.Context, string, string) (*discovery.Listing, error) {
	return nil, discovery.ErrNotFound
}

type routerVerifier struct {
	identity users.VerifiedIdentity
	calls    int
}

func (v *routerVerifier) Verify(context.Context, string) (users.VerifiedIdentity, error) {
	v.calls++
	return v.identity, nil
}
