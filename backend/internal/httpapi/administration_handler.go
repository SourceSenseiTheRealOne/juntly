package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/administration"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/authn"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/users"
	"github.com/google/uuid"
)

type AdministrationService interface {
	Metrics(context.Context, users.VerifiedIdentity) (administration.Metrics, error)
	Queue(context.Context, users.VerifiedIdentity) (administration.Queue, error)
	Moderate(context.Context, users.VerifiedIdentity, administration.ModerationAction) error
}
type administrationHandler struct{ service AdministrationService }

func NewAdministrationHandler(s AdministrationService) http.Handler {
	return administrationHandler{service: s}
}

type moderationRequest struct {
	Kind     *string    `json:"kind"`
	TargetID *uuid.UUID `json:"targetId"`
	Reason   *string    `json:"reason"`
}

func (h administrationHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	id := requestIDFromHeader(r.Header.Get(RequestIDHeader))
	if h.service == nil {
		writeAPIError(w, 503, "SERVICE_UNAVAILABLE", "Service unavailable", id)
		return
	}
	identity, ok := authn.IdentityFromContext(r.Context())
	if !ok {
		writeAPIError(w, 401, "UNAUTHORIZED", "Unauthorized", id)
		return
	}
	switch r.URL.Path {
	case "/api/v1/admin/dashboard":
		if r.Method != http.MethodGet {
			http.Error(w, http.StatusText(405), 405)
			return
		}
		metrics, err := h.service.Metrics(r.Context(), identity)
		if err != nil {
			writeAdministrationError(w, err, id)
			return
		}
		queue, err := h.service.Queue(r.Context(), identity)
		if err != nil {
			writeAdministrationError(w, err, id)
			return
		}
		writeJSON(w, 200, map[string]any{"metrics": metrics, "queue": queue}, id)
	case "/api/v1/admin/moderation":
		if r.Method != http.MethodPost {
			http.Error(w, http.StatusText(405), 405)
			return
		}
		var body moderationRequest
		if !decodeAdministration(r.Body, &body) || body.Kind == nil || body.TargetID == nil || body.Reason == nil {
			writeAPIError(w, 400, "INVALID_REQUEST", "Invalid request", id)
			return
		}
		if err := h.service.Moderate(r.Context(), identity, administration.ModerationAction{Kind: *body.Kind, TargetID: *body.TargetID, Reason: *body.Reason}); err != nil {
			writeAdministrationError(w, err, id)
			return
		}
		w.Header().Set(RequestIDHeader, id)
		w.WriteHeader(204)
	default:
		writeAPIError(w, 404, "NOT_FOUND", "Not found", id)
	}
}
func decodeAdministration(r io.Reader, target any) bool {
	d := json.NewDecoder(io.LimitReader(r, 4097))
	d.DisallowUnknownFields()
	if d.Decode(target) != nil {
		return false
	}
	var extra any
	return errors.Is(d.Decode(&extra), io.EOF)
}
func writeAdministrationError(w http.ResponseWriter, err error, id string) {
	switch {
	case errors.Is(err, administration.ErrInvalid):
		writeAPIError(w, 400, "INVALID_REQUEST", "Invalid request", id)
	case errors.Is(err, administration.ErrUnauthorized):
		writeAPIError(w, 401, "UNAUTHORIZED", "Unauthorized", id)
	case errors.Is(err, administration.ErrForbidden):
		writeAPIError(w, 403, "FORBIDDEN", "Forbidden", id)
	case errors.Is(err, administration.ErrNotFound):
		writeAPIError(w, 404, "NOT_FOUND", "Not found", id)
	default:
		writeAPIError(w, 503, "SERVICE_UNAVAILABLE", "Service unavailable", id)
	}
}
