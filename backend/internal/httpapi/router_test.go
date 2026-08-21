package httpapi_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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
	router := httpapi.NewRouter(healthService, verifier, reconcileService)

	healthResponse := httptest.NewRecorder()
	router.ServeHTTP(healthResponse, httptest.NewRequest(http.MethodGet, "/api/v1/health", nil))
	if healthResponse.Code != http.StatusOK {
		t.Fatalf("health status = %d, want %d", healthResponse.Code, http.StatusOK)
	}
	if verifier.calls != 0 {
		t.Fatalf("verifier calls after health = %d, want 0", verifier.calls)
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
}

type routerVerifier struct {
	identity users.VerifiedIdentity
	calls    int
}

func (v *routerVerifier) Verify(context.Context, string) (users.VerifiedIdentity, error) {
	v.calls++
	return v.identity, nil
}
