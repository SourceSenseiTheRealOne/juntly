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

	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/accounts"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/authn"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/httpapi"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/users"
)

func TestAccountHandlerReturnsUnauthorizedWithoutVerifiedIdentity(t *testing.T) {
	t.Parallel()

	for _, method := range []string{http.MethodGet, http.MethodPut} {
		method := method
		t.Run(method, func(t *testing.T) {
			t.Parallel()
			service := &recordingAccountService{}
			response := httptest.NewRecorder()
			request := httptest.NewRequest(method, "/api/v1/me/account", strings.NewReader(`{"providerEnabled":true}`))
			request.Header.Set(httpapi.RequestIDHeader, "req_account_unauthorized")

			httpapi.NewAccountHandler(service).ServeHTTP(response, request)

			if response.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, want %d", response.Code, http.StatusUnauthorized)
			}
			if service.calls != 0 {
				t.Fatalf("service calls = %d, want 0", service.calls)
			}
			assertErrorResponse(t, response, "UNAUTHORIZED", "Unauthorized", "req_account_unauthorized")
		})
	}
}

func TestAccountHandlerGETReturnsClosedCapabilities(t *testing.T) {
	t.Parallel()

	completedAt := time.Date(2026, 8, 23, 12, 5, 0, 123456000, time.UTC)
	service := &recordingAccountService{account: accounts.Account{
		CustomerEnabled:       true,
		ProviderEnabled:       false,
		OnboardingCompletedAt: completedAt,
	}}
	handler := authn.RequireVerifiedIdentity(
		staticVerifier{identity: users.VerifiedIdentity{Subject: "user_synthetic"}},
		httpapi.NewAccountHandler(service),
	)
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/me/account", nil)
	request.Header.Set("Authorization", "Bearer synthetic-token")
	request.Header.Set(httpapi.RequestIDHeader, "req_account_get")

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if response.Header().Get(httpapi.RequestIDHeader) != "req_account_get" {
		t.Fatalf("request ID header = %q", response.Header().Get(httpapi.RequestIDHeader))
	}
	body := decodeJSONMap(t, response)
	want := map[string]any{
		"customerEnabled":       true,
		"providerEnabled":       false,
		"onboardingCompletedAt": "2026-08-23T12:05:00.123456Z",
	}
	if mustJSON(t, body) != mustJSON(t, want) {
		t.Fatalf("body = %#v, want %#v", body, want)
	}
	serialized := mustJSON(t, body)
	for _, prohibited := range []string{"user_synthetic", "synthetic-token", "internalUserId", "clerk"} {
		if strings.Contains(serialized, prohibited) {
			t.Fatalf("response leaks %q: %s", prohibited, serialized)
		}
	}
}

func TestAccountHandlerPUTUpdatesProviderCapability(t *testing.T) {
	t.Parallel()

	for _, enabled := range []bool{true, false} {
		enabled := enabled
		t.Run(map[bool]string{true: "enable", false: "disable"}[enabled], func(t *testing.T) {
			t.Parallel()
			service := &recordingAccountService{account: accounts.Account{
				CustomerEnabled:       true,
				ProviderEnabled:       enabled,
				OnboardingCompletedAt: time.Date(2026, 8, 23, 12, 5, 0, 0, time.UTC),
			}}
			handler := authn.RequireVerifiedIdentity(
				staticVerifier{identity: users.VerifiedIdentity{Subject: "user_synthetic"}},
				httpapi.NewAccountHandler(service),
			)
			response := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodPut, "/api/v1/me/account", strings.NewReader(map[bool]string{true: `{"providerEnabled":true}`, false: `{"providerEnabled":false}`}[enabled]))
			request.Header.Set("Authorization", "Bearer synthetic-token")
			request.Header.Set(httpapi.RequestIDHeader, "req_account_put")

			handler.ServeHTTP(response, request)

			if response.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
			}
			if service.calls != 1 || service.providerEnabled == nil || *service.providerEnabled != enabled {
				t.Fatalf("service calls=%d providerEnabled=%v, want 1 %t", service.calls, service.providerEnabled, enabled)
			}
			if service.identity.Subject != "user_synthetic" {
				t.Fatalf("service subject = %q", service.identity.Subject)
			}
		})
	}
}

func TestAccountHandlerPUTRejectsNonExactJSONBeforeService(t *testing.T) {
	t.Parallel()

	cases := map[string]string{
		"empty body":       "",
		"empty object":     `{}`,
		"null body":        `null`,
		"unknown property": `{"providerEnabled":true,"admin":true}`,
		"non boolean":      `{"providerEnabled":"true"}`,
		"null capability":  `{"providerEnabled":null}`,
		"trailing bytes":   `{"providerEnabled":true} trailing`,
		"second value":     `{"providerEnabled":true}{}`,
	}
	for name, body := range cases {
		name, body := name, body
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			service := &recordingAccountService{}
			handler := authn.RequireVerifiedIdentity(
				staticVerifier{identity: users.VerifiedIdentity{Subject: "user_synthetic"}},
				httpapi.NewAccountHandler(service),
			)
			response := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodPut, "/api/v1/me/account", strings.NewReader(body))
			request.Header.Set("Authorization", "Bearer synthetic-token")
			request.Header.Set(httpapi.RequestIDHeader, "req_account_invalid")

			handler.ServeHTTP(response, request)

			if response.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
			}
			if service.calls != 0 {
				t.Fatalf("service calls = %d, want 0", service.calls)
			}
			assertErrorResponse(t, response, "INVALID_REQUEST", "Invalid request", "req_account_invalid")
		})
	}
}

func TestAccountHandlerReturnsSafeUnavailable(t *testing.T) {
	t.Parallel()

	service := &recordingAccountService{err: errors.New("database at internal-host failed for user_synthetic")}
	handler := authn.RequireVerifiedIdentity(
		staticVerifier{identity: users.VerifiedIdentity{Subject: "user_synthetic"}},
		httpapi.NewAccountHandler(service),
	)
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/me/account", nil)
	request.Header.Set("Authorization", "Bearer synthetic-token")
	request.Header.Set(httpapi.RequestIDHeader, "req_account_unavailable")

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusServiceUnavailable)
	}
	assertErrorResponse(t, response, "SERVICE_UNAVAILABLE", "Service unavailable", "req_account_unavailable")
}

func TestAccountHandlerRejectsOtherMethods(t *testing.T) {
	t.Parallel()

	service := &recordingAccountService{}
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/me/account", nil)

	httpapi.NewAccountHandler(service).ServeHTTP(response, request)

	if response.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusMethodNotAllowed)
	}
	if allow := response.Header().Get("Allow"); allow != "GET, PUT" {
		t.Fatalf("Allow = %q, want %q", allow, "GET, PUT")
	}
	if service.calls != 0 {
		t.Fatalf("service calls = %d, want 0", service.calls)
	}
}

type recordingAccountService struct {
	account         accounts.Account
	err             error
	calls           int
	identity        users.VerifiedIdentity
	providerEnabled *bool
}

func (s *recordingAccountService) Get(_ context.Context, identity users.VerifiedIdentity) (accounts.Account, error) {
	s.calls++
	s.identity = identity
	return s.account, s.err
}

func (s *recordingAccountService) SetProviderEnabled(_ context.Context, identity users.VerifiedIdentity, enabled bool) (accounts.Account, error) {
	s.calls++
	s.identity = identity
	s.providerEnabled = &enabled
	return s.account, s.err
}

func decodeJSONMap(t *testing.T, response *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var body map[string]any
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return body
}
