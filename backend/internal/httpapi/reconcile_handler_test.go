package httpapi_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/authn"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/httpapi"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/users"
	"github.com/google/uuid"
)

func TestReconcileHandlerReturnsUnauthorizedWithoutVerifiedIdentity(t *testing.T) {
	t.Parallel()

	service := &recordingReconcileService{}
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/reconcile", nil)
	request.Header.Set(httpapi.RequestIDHeader, "req_reconcile_unauthorized")

	httpapi.NewReconcileHandler(service).ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusUnauthorized)
	}
	if service.calls != 0 {
		t.Fatalf("service calls = %d, want 0", service.calls)
	}
	assertErrorResponse(t, response, "UNAUTHORIZED", "Unauthorized", "req_reconcile_unauthorized")
}

func TestReconcileHandlerReturnsOpaqueInternalUser(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, 8, 21, 12, 0, 0, 123456000, time.UTC)
	service := &recordingReconcileService{user: users.InternalUser{
		ID:        uuid.MustParse("7b7b7d7e-38f9-4f0c-8a10-0fce9cf6f82b"),
		CreatedAt: createdAt,
	}}
	handler := authn.RequireVerifiedIdentity(staticVerifier{identity: users.VerifiedIdentity{Subject: "user_synthetic"}}, httpapi.NewReconcileHandler(service))
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/reconcile", nil)
	request.Header.Set("Authorization", "Bearer synthetic-token")
	request.Header.Set(httpapi.RequestIDHeader, "req_reconcile_success")

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if response.Header().Get(httpapi.RequestIDHeader) != "req_reconcile_success" {
		t.Fatalf("request ID header = %q", response.Header().Get(httpapi.RequestIDHeader))
	}
	if service.identity.Subject != "user_synthetic" {
		t.Fatalf("service subject = %q, want user_synthetic", service.identity.Subject)
	}

	var body map[string]any
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(body) != 2 {
		t.Fatalf("response field count = %d, want 2 (%#v)", len(body), body)
	}
	if body["id"] != "7b7b7d7e-38f9-4f0c-8a10-0fce9cf6f82b" {
		t.Fatalf("id = %#v", body["id"])
	}
	if body["createdAt"] != "2026-08-21T12:00:00.123456Z" {
		t.Fatalf("createdAt = %#v", body["createdAt"])
	}
	serialized := mustJSON(t, body)
	if strings.Contains(serialized, "user_synthetic") || strings.Contains(serialized, "synthetic-token") {
		t.Fatalf("response leaks identity material: %s", serialized)
	}
}

func TestReconcileHandlerReturnsSafeUnavailableForDependencyFailure(t *testing.T) {
	t.Parallel()

	service := &recordingReconcileService{err: errors.New("database at internal-host rejected user_synthetic")}
	handler := authn.RequireVerifiedIdentity(staticVerifier{identity: users.VerifiedIdentity{Subject: "user_synthetic"}}, httpapi.NewReconcileHandler(service))
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/reconcile", nil)
	request.Header.Set("Authorization", "Bearer synthetic-token")
	request.Header.Set(httpapi.RequestIDHeader, "req_reconcile_unavailable")

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusServiceUnavailable)
	}
	assertErrorResponse(t, response, "SERVICE_UNAVAILABLE", "Service unavailable", "req_reconcile_unavailable")
}

func TestReconcileHandlerRejectsNonPOSTMethods(t *testing.T) {
	t.Parallel()

	service := &recordingReconcileService{}
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/auth/reconcile", nil)

	httpapi.NewReconcileHandler(service).ServeHTTP(response, request)

	if response.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusMethodNotAllowed)
	}
	if allow := response.Header().Get("Allow"); allow != http.MethodPost {
		t.Fatalf("Allow = %q, want %q", allow, http.MethodPost)
	}
	if service.calls != 0 {
		t.Fatalf("service calls = %d, want 0", service.calls)
	}
}

type staticVerifier struct {
	identity users.VerifiedIdentity
	err      error
}

func (v staticVerifier) Verify(context.Context, string) (users.VerifiedIdentity, error) {
	return v.identity, v.err
}

type recordingReconcileService struct {
	user     users.InternalUser
	err      error
	calls    int
	identity users.VerifiedIdentity
}

func (s *recordingReconcileService) Reconcile(_ context.Context, identity users.VerifiedIdentity) (users.InternalUser, bool, error) {
	s.calls++
	s.identity = identity
	return s.user, false, s.err
}

func assertErrorResponse(t *testing.T, response *httptest.ResponseRecorder, code, message, requestID string) {
	t.Helper()

	if response.Header().Get(httpapi.RequestIDHeader) != requestID {
		t.Fatalf("request ID header = %q, want %q", response.Header().Get(httpapi.RequestIDHeader), requestID)
	}
	var body map[string]any
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if mustJSON(t, body) != mustJSON(t, map[string]any{"error": map[string]any{
		"code":      code,
		"message":   message,
		"requestId": requestID,
	}}) {
		t.Fatalf("error body = %#v", body)
	}
}

func mustJSON(t *testing.T, value any) string {
	t.Helper()

	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal JSON: %v", err)
	}
	return string(encoded)
}
