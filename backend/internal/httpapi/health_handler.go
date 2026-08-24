package httpapi

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"

	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/authn"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/health"
)

const RequestIDHeader = "X-Request-ID"

type HealthHandler struct {
	service health.Service
}

func NewHealthHandler(service health.Service) HealthHandler {
	return HealthHandler{service: service}
}

func (h HealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	requestID := requestIDFromHeader(r.Header.Get(RequestIDHeader))
	w.Header().Set(RequestIDHeader, requestID)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	_ = json.NewEncoder(w).Encode(h.service.Check(requestID))
}

func NewRouter(service health.Service, verifier authn.Verifier, reconcileService ReconcileService, accountService AccountService, referenceService ReferenceService, providerProfileService ProviderProfileService, listingService ListingService, moderationListingService ModerationListingService) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/api/v1/health", NewHealthHandler(service))
	mux.Handle("/api/v1/catalog/categories", NewCategoriesHandler(referenceService))
	mux.Handle("/api/v1/reference/languages", NewLanguagesHandler(referenceService))
	mux.Handle("/api/v1/reference/localities", NewLocalitiesHandler(referenceService))
	mux.Handle("/api/v1/auth/reconcile", authn.RequireVerifiedIdentity(verifier, NewReconcileHandler(reconcileService)))
	mux.Handle("/api/v1/me/account", authn.RequireVerifiedIdentity(verifier, NewAccountHandler(accountService)))
	mux.Handle("/api/v1/me/provider-profile", authn.RequireVerifiedIdentity(verifier, NewProviderProfileHandler(providerProfileService)))
	mux.Handle("/api/v1/me/listings", authn.RequireVerifiedIdentity(verifier, NewListingHandler(listingService)))
	mux.Handle("/api/v1/me/listings/", authn.RequireVerifiedIdentity(verifier, NewListingHandler(listingService)))
	mux.Handle("/api/v1/moderation/listings", authn.RequireVerifiedIdentity(verifier, NewModerationListingHandler(moderationListingService)))
	mux.Handle("/api/v1/moderation/listings/", authn.RequireVerifiedIdentity(verifier, NewModerationListingHandler(moderationListingService)))
	return mux
}

func requestIDFromHeader(value string) string {
	if validRequestID(value) {
		return value
	}

	var randomBytes [16]byte
	if _, err := rand.Read(randomBytes[:]); err != nil {
		return "req_unavailable"
	}

	return "req_" + hex.EncodeToString(randomBytes[:])
}

func validRequestID(value string) bool {
	if len(value) < 8 || len(value) > 128 {
		return false
	}

	for _, char := range value {
		switch {
		case char >= 'A' && char <= 'Z':
		case char >= 'a' && char <= 'z':
		case char >= '0' && char <= '9':
		case char == '.', char == '_', char == ':', char == '-':
		default:
			return false
		}
	}

	return true
}
