package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/authn"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/users"
	"github.com/google/uuid"
)

type ReconcileService interface {
	Reconcile(context.Context, users.VerifiedIdentity) (users.InternalUser, bool, error)
}

type reconcileHandler struct {
	service ReconcileService
}

type internalUserResponse struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt string    `json:"createdAt"`
}

type apiErrorResponse struct {
	Error apiErrorDetail `json:"error"`
}

type apiErrorDetail struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"requestId"`
}

func NewReconcileHandler(service ReconcileService) http.Handler {
	return reconcileHandler{service: service}
}

func (h reconcileHandler) ServeHTTP(w http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	requestID := requestIDFromHeader(request.Header.Get(RequestIDHeader))
	identity, ok := authn.IdentityFromContext(request.Context())
	if !ok || h.service == nil {
		writeAPIError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized", requestID)
		return
	}

	user, _, err := h.service.Reconcile(request.Context(), identity)
	if err != nil {
		writeAPIError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Service unavailable", requestID)
		return
	}

	writeJSON(w, http.StatusOK, internalUserResponse{
		ID:        user.ID,
		CreatedAt: user.CreatedAt.UTC().Format(time.RFC3339Nano),
	}, requestID)
}

func writeAPIError(w http.ResponseWriter, status int, code, message, requestID string) {
	writeJSON(w, status, apiErrorResponse{Error: apiErrorDetail{
		Code:      code,
		Message:   message,
		RequestID: requestID,
	}}, requestID)
}

func writeJSON(w http.ResponseWriter, status int, body any, requestID string) {
	w.Header().Set(RequestIDHeader, requestID)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
