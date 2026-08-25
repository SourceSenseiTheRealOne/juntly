package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/authn"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/provideraccess"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/providers"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/users"
	"github.com/google/uuid"
)

const maxProviderProfileRequestBytes = 16 * 1024

type ProviderProfileService interface {
	Get(context.Context, users.VerifiedIdentity) (*providers.Profile, error)
	Put(context.Context, users.VerifiedIdentity, providers.ReplaceProfile) (providers.Profile, error)
}

type providerProfileHandler struct{ service ProviderProfileService }

type providerProfileEnvelope struct {
	Profile *providerProfileResponse `json:"profile"`
}

type providerProfileResponse struct {
	DisplayName         string                 `json:"displayName"`
	ProviderType        providers.ProviderType `json:"providerType"`
	Bio                 string                 `json:"bio"`
	PrimaryLocalityID   uuid.UUID              `json:"primaryLocalityId"`
	ServiceLocalityIDs  []uuid.UUID            `json:"serviceLocalityIds"`
	MaxTravelDistanceKM int                    `json:"maxTravelDistanceKm"`
	TravelsToCustomer   bool                   `json:"travelsToCustomer"`
	ReceivesCustomer    bool                   `json:"receivesCustomer"`
	RemoteServices      bool                   `json:"remoteServices"`
	LanguageCodes       []string               `json:"languageCodes"`
	CreatedAt           string                 `json:"createdAt"`
	UpdatedAt           string                 `json:"updatedAt"`
}

type replaceProviderProfileRequest struct {
	DisplayName         *string      `json:"displayName"`
	ProviderType        *string      `json:"providerType"`
	Bio                 *string      `json:"bio"`
	PrimaryLocalityID   *uuid.UUID   `json:"primaryLocalityId"`
	ServiceLocalityIDs  *[]uuid.UUID `json:"serviceLocalityIds"`
	MaxTravelDistanceKM *int         `json:"maxTravelDistanceKm"`
	TravelsToCustomer   *bool        `json:"travelsToCustomer"`
	ReceivesCustomer    *bool        `json:"receivesCustomer"`
	RemoteServices      *bool        `json:"remoteServices"`
	LanguageCodes       *[]string    `json:"languageCodes"`
}

func NewProviderProfileHandler(service ProviderProfileService) http.Handler {
	return providerProfileHandler{service: service}
}

func (h providerProfileHandler) ServeHTTP(w http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet && request.Method != http.MethodPut {
		w.Header().Set("Allow", "GET, PUT")
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	requestID := requestIDFromHeader(request.Header.Get(RequestIDHeader))
	identity, ok := authn.IdentityFromContext(request.Context())
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized", requestID)
		return
	}
	if h.service == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Service unavailable", requestID)
		return
	}

	if request.Method == http.MethodGet {
		profile, err := h.service.Get(request.Context(), identity)
		if err != nil {
			writeProviderProfileError(w, err, requestID)
			return
		}
		var response *providerProfileResponse
		if profile != nil {
			value := providerProfileResponseFromDomain(*profile)
			response = &value
		}
		writeJSON(w, http.StatusOK, providerProfileEnvelope{Profile: response}, requestID)
		return
	}

	input, valid := decodeProviderProfileRequest(request.Body)
	if !valid {
		writeAPIError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request", requestID)
		return
	}
	profile, err := h.service.Put(request.Context(), identity, input)
	if err != nil {
		writeProviderProfileError(w, err, requestID)
		return
	}
	response := providerProfileResponseFromDomain(profile)
	writeJSON(w, http.StatusOK, providerProfileEnvelope{Profile: &response}, requestID)
}

func decodeProviderProfileRequest(body io.Reader) (providers.ReplaceProfile, bool) {
	decoder := json.NewDecoder(io.LimitReader(body, maxProviderProfileRequestBytes+1))
	decoder.DisallowUnknownFields()
	var value replaceProviderProfileRequest
	if err := decoder.Decode(&value); err != nil || !completeProviderProfileRequest(value) {
		return providers.ReplaceProfile{}, false
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return providers.ReplaceProfile{}, false
	}
	return providers.ReplaceProfile{
		DisplayName:         *value.DisplayName,
		ProviderType:        providers.ProviderType(*value.ProviderType),
		Bio:                 *value.Bio,
		PrimaryLocalityID:   *value.PrimaryLocalityID,
		ServiceLocalityIDs:  append([]uuid.UUID(nil), (*value.ServiceLocalityIDs)...),
		MaxTravelDistanceKM: *value.MaxTravelDistanceKM,
		TravelsToCustomer:   *value.TravelsToCustomer,
		ReceivesCustomer:    *value.ReceivesCustomer,
		RemoteServices:      *value.RemoteServices,
		LanguageCodes:       append([]string(nil), (*value.LanguageCodes)...),
	}, true
}

func completeProviderProfileRequest(value replaceProviderProfileRequest) bool {
	return value.DisplayName != nil && value.ProviderType != nil && value.Bio != nil &&
		value.PrimaryLocalityID != nil && value.ServiceLocalityIDs != nil &&
		value.MaxTravelDistanceKM != nil && value.TravelsToCustomer != nil &&
		value.ReceivesCustomer != nil && value.RemoteServices != nil && value.LanguageCodes != nil
}

func providerProfileResponseFromDomain(profile providers.Profile) providerProfileResponse {
	return providerProfileResponse{
		DisplayName:         profile.DisplayName,
		ProviderType:        profile.ProviderType,
		Bio:                 profile.Bio,
		PrimaryLocalityID:   profile.PrimaryLocalityID,
		ServiceLocalityIDs:  append([]uuid.UUID(nil), profile.ServiceLocalityIDs...),
		MaxTravelDistanceKM: profile.MaxTravelDistanceKM,
		TravelsToCustomer:   profile.TravelsToCustomer,
		ReceivesCustomer:    profile.ReceivesCustomer,
		RemoteServices:      profile.RemoteServices,
		LanguageCodes:       append([]string(nil), profile.LanguageCodes...),
		CreatedAt:           profile.CreatedAt.UTC().Format(time.RFC3339Nano),
		UpdatedAt:           profile.UpdatedAt.UTC().Format(time.RFC3339Nano),
	}
}

func writeProviderProfileError(w http.ResponseWriter, err error, requestID string) {
	switch {
	case errors.Is(err, provideraccess.ErrUnauthorized):
		writeAPIError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized", requestID)
	case errors.Is(err, provideraccess.ErrForbidden):
		writeAPIError(w, http.StatusForbidden, "FORBIDDEN", "Forbidden", requestID)
	case errors.Is(err, providers.ErrInvalidProfile):
		writeAPIError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request", requestID)
	default:
		writeAPIError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Service unavailable", requestID)
	}
}
