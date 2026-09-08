package httpapi

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/authn"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/listingmedia"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/provideraccess"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/users"
	"github.com/google/uuid"
)

type MediaReader interface {
	List(context.Context, users.VerifiedIdentity, listingmedia.ReadScope, uuid.UUID) ([]listingmedia.Photo, error)
	Get(context.Context, users.VerifiedIdentity, listingmedia.ReadScope, uuid.UUID, uuid.UUID) ([]byte, error)
}
type MediaFinalizer interface {
	Finalize(context.Context, users.VerifiedIdentity, uuid.UUID, uuid.UUID) error
}

type mediaHandler struct {
	reader    MediaReader
	finalizer MediaFinalizer
	scope     listingmedia.ReadScope
}

// Specific media routes are composed around the existing API without changing
// unrelated routes. Scope comes from server registration, never query/body data.
func NewMediaRouter(fallback http.Handler, verifier authn.Verifier, reader MediaReader, finalizer MediaFinalizer) http.Handler {
	mux := http.NewServeMux()
	for segment, scope := range map[string]listingmedia.ReadScope{"public": listingmedia.PublicRead, "me": listingmedia.OwnerRead, "moderation": listingmedia.ModeratorRead} {
		var handler http.Handler = mediaHandler{reader: reader, finalizer: finalizer, scope: scope}
		if scope != listingmedia.PublicRead {
			handler = authn.RequireVerifiedIdentity(verifier, handler)
		}
		path := "/api/v1/" + segment + "/listings/{listingID}/media"
		mux.Handle("GET "+path, handler)
		mux.Handle("GET "+path+"/{mediaID}", handler)
		if scope == listingmedia.OwnerRead {
			mux.Handle("POST "+path+"/{mediaID}/finalize", handler)
		}
	}
	mux.Handle("/", fallback)
	return mux
}

func (h mediaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestID := requestIDFromHeader(r.Header.Get(RequestIDHeader))
	w.Header().Set(RequestIDHeader, requestID)
	w.Header().Set("Cache-Control", "private, no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Cross-Origin-Resource-Policy", "same-origin")
	identity, authenticated := authn.IdentityFromContext(r.Context())
	if h.scope != listingmedia.PublicRead && !authenticated {
		writeMediaError(w, provideraccess.ErrUnauthorized, requestID)
		return
	}
	listingID, valid := mediaPathID(r.PathValue("listingID"))
	if !valid || r.URL.RawQuery != "" || r.URL.RawPath != "" {
		writeMediaError(w, listingmedia.ErrInvalidUpload, requestID)
		return
	}
	mediaID := uuid.Nil
	if value := r.PathValue("mediaID"); value != "" {
		mediaID, valid = mediaPathID(value)
		if !valid {
			writeMediaError(w, listingmedia.ErrInvalidUpload, requestID)
			return
		}
	}
	if r.Method == http.MethodPost {
		if h.scope != listingmedia.OwnerRead || mediaID == uuid.Nil || h.finalizer == nil {
			writeMediaError(w, listingmedia.ErrUnavailable, requestID)
			return
		}
		body, err := io.ReadAll(io.LimitReader(r.Body, 1025))
		text := strings.TrimSpace(string(body))
		if err != nil || len(body) > 1024 || (text != "" && text != "{}") {
			writeMediaError(w, listingmedia.ErrInvalidUpload, requestID)
			return
		}
		if err := h.finalizer.Finalize(r.Context(), identity, listingID, mediaID); err != nil {
			writeMediaError(w, err, requestID)
			return
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if h.reader == nil {
		writeMediaError(w, listingmedia.ErrUnavailable, requestID)
		return
	}
	if mediaID == uuid.Nil {
		photos, err := h.reader.List(r.Context(), identity, h.scope, listingID)
		if err != nil {
			writeMediaError(w, err, requestID)
			return
		}
		if photos == nil {
			photos = []listingmedia.Photo{}
		}
		writeJSON(w, http.StatusOK, struct {
			Photos []listingmedia.Photo `json:"photos"`
		}{Photos: photos}, requestID)
		return
	}
	body, err := h.reader.Get(r.Context(), identity, h.scope, listingID, mediaID)
	if err != nil {
		writeMediaError(w, err, requestID)
		return
	}
	if len(body) == 0 || len(body) > 10485760 {
		writeMediaError(w, listingmedia.ErrUnavailable, requestID)
		return
	}
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Content-Length", strconv.Itoa(len(body)))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(body)
}

func mediaPathID(value string) (uuid.UUID, bool) {
	id, err := uuid.Parse(value)
	return id, err == nil && id != uuid.Nil && id.String() == value
}
func writeMediaError(w http.ResponseWriter, err error, id string) {
	switch {
	case errors.Is(err, provideraccess.ErrUnauthorized):
		writeAPIError(w, 401, "UNAUTHORIZED", "Unauthorized", id)
	case errors.Is(err, provideraccess.ErrForbidden):
		writeAPIError(w, 403, "FORBIDDEN", "Forbidden", id)
	case errors.Is(err, listingmedia.ErrNotFound):
		writeAPIError(w, 404, "NOT_FOUND", "Not found", id)
	case errors.Is(err, listingmedia.ErrConflict):
		writeAPIError(w, 409, "CONFLICT", "Conflict", id)
	case errors.Is(err, listingmedia.ErrInvalidUpload):
		writeAPIError(w, 400, "INVALID_REQUEST", "Invalid request", id)
	default:
		writeAPIError(w, 503, "SERVICE_UNAVAILABLE", "Service unavailable", id)
	}
}
