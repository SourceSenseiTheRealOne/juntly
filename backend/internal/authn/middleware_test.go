package authn

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/users"
)

func TestRequireVerifiedIdentityRejectsMissingBearerBeforeVerifier(t *testing.T) {
	t.Parallel()

	verifier := &recordingVerifier{}
	nextCalled := false
	handler := RequireVerifiedIdentity(verifier, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		nextCalled = true
	}))
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/reconcile", nil)
	request.Header.Set("X-Request-ID", "req_missing_bearer")

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusUnauthorized)
	}
	if verifier.calls != 0 {
		t.Fatalf("verifier calls = %d, want 0", verifier.calls)
	}
	if nextCalled {
		t.Fatal("downstream handler was called")
	}
	assertUnauthorizedResponse(t, response, "req_missing_bearer")
}

func TestRequireVerifiedIdentityRejectsMalformedAuthorizationBeforeVerifier(t *testing.T) {
	t.Parallel()

	verifier := &recordingVerifier{}
	nextCalled := false
	handler := RequireVerifiedIdentity(verifier, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		nextCalled = true
	}))
	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/reconcile", nil)
	request.Header.Set("Authorization", "Basic not-a-bearer-token")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusUnauthorized)
	}
	if verifier.calls != 0 {
		t.Fatalf("verifier calls = %d, want 0", verifier.calls)
	}
	if nextCalled {
		t.Fatal("downstream handler was called")
	}
}

func TestRequireVerifiedIdentityRejectsVerifierFailure(t *testing.T) {
	t.Parallel()

	verifier := &recordingVerifier{err: errors.New("verification failed")}
	nextCalled := false
	handler := RequireVerifiedIdentity(verifier, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		nextCalled = true
	}))
	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/reconcile", nil)
	request.Header.Set("Authorization", "Bearer synthetic-token")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusUnauthorized)
	}
	if verifier.calls != 1 {
		t.Fatalf("verifier calls = %d, want 1", verifier.calls)
	}
	if nextCalled {
		t.Fatal("downstream handler was called")
	}
}

func TestRequireVerifiedIdentityRejectsEmptyVerifiedSubject(t *testing.T) {
	t.Parallel()

	verifier := &recordingVerifier{identity: users.VerifiedIdentity{}}
	nextCalled := false
	handler := RequireVerifiedIdentity(verifier, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		nextCalled = true
	}))
	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/reconcile", nil)
	request.Header.Set("Authorization", "Bearer synthetic-token")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusUnauthorized)
	}
	if verifier.calls != 1 {
		t.Fatalf("verifier calls = %d, want 1", verifier.calls)
	}
	if nextCalled {
		t.Fatal("downstream handler was called")
	}
}

func TestRequireVerifiedIdentityPropagatesVerifiedSubjectToDownstreamHandler(t *testing.T) {
	t.Parallel()

	identity := users.VerifiedIdentity{Subject: "user_synthetic"}
	verifier := &recordingVerifier{identity: identity}
	handler := RequireVerifiedIdentity(verifier, http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		actual, ok := IdentityFromContext(request.Context())
		if !ok {
			t.Fatal("verified identity missing from context")
		}
		if actual != identity {
			t.Fatalf("identity = %#v, want %#v", actual, identity)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/reconcile", nil)
	request.Header.Set("Authorization", "Bearer synthetic-token")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNoContent)
	}
	if verifier.token != "synthetic-token" {
		t.Fatalf("verifier token = %q, want synthetic-token", verifier.token)
	}
}

func assertUnauthorizedResponse(t *testing.T, response *httptest.ResponseRecorder, requestID string) {
	t.Helper()

	if got := response.Header().Get("X-Request-ID"); got != requestID {
		t.Fatalf("X-Request-ID = %q, want %q", got, requestID)
	}
	if got := response.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", got)
	}

	var body map[string]any
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	encoded, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal response: %v", err)
	}
	want, err := json.Marshal(map[string]any{"error": map[string]any{
		"code":      "UNAUTHORIZED",
		"message":   "Unauthorized",
		"requestId": requestID,
	}})
	if err != nil {
		t.Fatalf("marshal expected response: %v", err)
	}
	if string(encoded) != string(want) {
		t.Fatalf("error body = %s, want %s", encoded, want)
	}
}

type recordingVerifier struct {
	identity users.VerifiedIdentity
	err      error
	calls    int
	token    string
}

func (v *recordingVerifier) Verify(_ context.Context, token string) (users.VerifiedIdentity, error) {
	v.calls++
	v.token = token
	return v.identity, v.err
}
