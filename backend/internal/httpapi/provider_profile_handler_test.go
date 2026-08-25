package httpapi_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/authn"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/httpapi"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/provideraccess"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/providers"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/users"
	"github.com/google/uuid"
)

func TestProviderProfileHandlerRequiresVerifiedIdentity(t *testing.T) {
	t.Parallel()

	service := &recordingProviderProfileService{}
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/me/provider-profile", nil)
	request.Header.Set(httpapi.RequestIDHeader, "req_provider_unauthorized")

	httpapi.NewProviderProfileHandler(service).ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized || service.calls != 0 {
		t.Fatalf("status/calls = %d/%d, want 401/0", response.Code, service.calls)
	}
	assertErrorResponse(t, response, "UNAUTHORIZED", "Unauthorized", "req_provider_unauthorized")
}

func TestProviderProfileHandlerReturnsNullableAndClosedOwnerProfile(t *testing.T) {
	t.Parallel()

	for name, profile := range map[string]*providers.Profile{
		"missing": nil,
		"existing": {
			DisplayName:         "Prestador local",
			ProviderType:        providers.ProviderTypeProfessional,
			Bio:                 "Serviço de confiança.",
			PrimaryLocalityID:   uuid.MustParse("11111111-1111-4111-8111-111111111111"),
			ServiceLocalityIDs:  []uuid.UUID{uuid.MustParse("11111111-1111-4111-8111-111111111111")},
			MaxTravelDistanceKM: 25,
			TravelsToCustomer:   true,
			LanguageCodes:       []string{"pt-PT"},
			CreatedAt:           time.Date(2026, 8, 23, 16, 0, 0, 0, time.UTC),
			UpdatedAt:           time.Date(2026, 8, 23, 16, 0, 0, 0, time.UTC),
		},
	} {
		name, profile := name, profile
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			service := &recordingProviderProfileService{profile: profile}
			handler := authn.RequireVerifiedIdentity(staticVerifier{identity: users.VerifiedIdentity{Subject: "user_provider"}}, httpapi.NewProviderProfileHandler(service))
			response := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, "/api/v1/me/provider-profile", nil)
			request.Header.Set("Authorization", "Bearer synthetic-token")
			request.Header.Set(httpapi.RequestIDHeader, "req_provider_get")

			handler.ServeHTTP(response, request)

			if response.Code != http.StatusOK {
				t.Fatalf("status = %d", response.Code)
			}
			var body map[string]any
			if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
				t.Fatalf("decode: %v", err)
			}
			if len(body) != 1 {
				t.Fatalf("body fields = %#v", body)
			}
			serialized := mustJSON(t, body)
			for _, prohibited := range []string{"user_provider", "synthetic-token", "internalUserId", "clerk", "phone", "address"} {
				if strings.Contains(serialized, prohibited) {
					t.Fatalf("response exposes %q: %s", prohibited, serialized)
				}
			}
		})
	}
}

func TestProviderProfileHandlerPUTStrictlyDecodesReplacement(t *testing.T) {
	t.Parallel()

	service := &recordingProviderProfileService{replacement: providers.Profile{DisplayName: "Prestador local", ProviderType: providers.ProviderTypeIndividual, Bio: "", PrimaryLocalityID: uuid.MustParse("11111111-1111-4111-8111-111111111111"), ServiceLocalityIDs: []uuid.UUID{uuid.MustParse("11111111-1111-4111-8111-111111111111")}, MaxTravelDistanceKM: 10, TravelsToCustomer: true, LanguageCodes: []string{"pt-PT"}, CreatedAt: time.Now(), UpdatedAt: time.Now()}}
	handler := authn.RequireVerifiedIdentity(staticVerifier{identity: users.VerifiedIdentity{Subject: "user_provider"}}, httpapi.NewProviderProfileHandler(service))
	valid := `{"displayName":"Prestador local","providerType":"individual","bio":"","primaryLocalityId":"11111111-1111-4111-8111-111111111111","serviceLocalityIds":["11111111-1111-4111-8111-111111111111"],"maxTravelDistanceKm":10,"travelsToCustomer":true,"receivesCustomer":false,"remoteServices":false,"languageCodes":["pt-PT"]}`
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPut, "/api/v1/me/provider-profile", strings.NewReader(valid))
	request.Header.Set("Authorization", "Bearer synthetic-token")
	request.Header.Set(httpapi.RequestIDHeader, "req_provider_put")

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK || service.putCalls != 1 || service.input.DisplayName != "Prestador local" {
		t.Fatalf("status/calls/input = %d/%d/%#v", response.Code, service.putCalls, service.input)
	}

	for _, body := range []string{`{}`, `null`, valid + `{}`, strings.Replace(valid, `}`, `,"admin":true}`, 1)} {
		service := &recordingProviderProfileService{}
		handler := authn.RequireVerifiedIdentity(staticVerifier{identity: users.VerifiedIdentity{Subject: "user_provider"}}, httpapi.NewProviderProfileHandler(service))
		response := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodPut, "/api/v1/me/provider-profile", strings.NewReader(body))
		request.Header.Set("Authorization", "Bearer synthetic-token")
		request.Header.Set(httpapi.RequestIDHeader, "req_provider_invalid")
		handler.ServeHTTP(response, request)
		if response.Code != http.StatusBadRequest || service.putCalls != 0 {
			t.Fatalf("invalid body status/calls = %d/%d", response.Code, service.putCalls)
		}
	}
}

func TestProviderProfileHandlerMapsForbiddenAndUnavailable(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		err    error
		status int
		code   string
	}{
		{provideraccess.ErrForbidden, http.StatusForbidden, "FORBIDDEN"},
		{provideraccess.ErrUnavailable, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE"},
	} {
		handler := authn.RequireVerifiedIdentity(
			staticVerifier{identity: users.VerifiedIdentity{Subject: "user_provider"}},
			httpapi.NewProviderProfileHandler(&recordingProviderProfileService{err: test.err}),
		)
		response := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, "/api/v1/me/provider-profile", nil)
		request.Header.Set("Authorization", "Bearer synthetic-token")
		handler.ServeHTTP(response, request)
		if response.Code != test.status || !strings.Contains(response.Body.String(), test.code) {
			t.Fatalf("error %v status/body = %d/%s", test.err, response.Code, response.Body.String())
		}
	}
}

type recordingProviderProfileService struct {
	profile     *providers.Profile
	replacement providers.Profile
	err         error
	calls       int
	putCalls    int
	input       providers.ReplaceProfile
}

func (s *recordingProviderProfileService) Get(context.Context, users.VerifiedIdentity) (*providers.Profile, error) {
	s.calls++
	return s.profile, s.err
}
func (s *recordingProviderProfileService) Put(_ context.Context, _ users.VerifiedIdentity, input providers.ReplaceProfile) (providers.Profile, error) {
	s.putCalls++
	s.input = input
	return s.replacement, s.err
}
