package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/accounts"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/authn"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/users"
)

const maxAccountRequestBytes = 1024

type AccountService interface {
	Get(context.Context, users.VerifiedIdentity) (accounts.Account, error)
	SetProviderEnabled(context.Context, users.VerifiedIdentity, bool) (accounts.Account, error)
}

type accountHandler struct {
	service AccountService
}

type accountCapabilitiesResponse struct {
	CustomerEnabled       bool   `json:"customerEnabled"`
	ProviderEnabled       bool   `json:"providerEnabled"`
	OnboardingCompletedAt string `json:"onboardingCompletedAt"`
}

type updateAccountCapabilitiesRequest struct {
	ProviderEnabled *bool `json:"providerEnabled"`
}

func NewAccountHandler(service AccountService) http.Handler {
	return accountHandler{service: service}
}

func (h accountHandler) ServeHTTP(w http.ResponseWriter, request *http.Request) {
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

	var (
		account accounts.Account
		err     error
	)
	switch request.Method {
	case http.MethodGet:
		account, err = h.service.Get(request.Context(), identity)
	case http.MethodPut:
		enabled, valid := decodeProviderEnabled(request.Body)
		if !valid {
			writeAPIError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request", requestID)
			return
		}
		account, err = h.service.SetProviderEnabled(request.Context(), identity, enabled)
	}

	if err != nil {
		if errors.Is(err, accounts.ErrInvalidIdentity) {
			writeAPIError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized", requestID)
			return
		}
		writeAPIError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Service unavailable", requestID)
		return
	}

	writeJSON(w, http.StatusOK, accountCapabilitiesResponse{
		CustomerEnabled:       account.CustomerEnabled,
		ProviderEnabled:       account.ProviderEnabled,
		OnboardingCompletedAt: account.OnboardingCompletedAt.UTC().Format(time.RFC3339Nano),
	}, requestID)
}

func decodeProviderEnabled(body io.Reader) (bool, bool) {
	decoder := json.NewDecoder(io.LimitReader(body, maxAccountRequestBytes+1))
	decoder.DisallowUnknownFields()

	var payload updateAccountCapabilitiesRequest
	if err := decoder.Decode(&payload); err != nil || payload.ProviderEnabled == nil {
		return false, false
	}

	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return false, false
	}
	return *payload.ProviderEnabled, true
}
