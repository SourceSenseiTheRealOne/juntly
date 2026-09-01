package httpapi

import (
	"context"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/readiness"
	"net/http"
	"time"
)

type ReadinessService interface {
	Check(context.Context) readiness.Result
}
type readinessHandler struct{ service ReadinessService }

func NewReadinessHandler(s ReadinessService) http.Handler { return readinessHandler{service: s} }
func (h readinessHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	id := requestIDFromHeader(r.Header.Get(RequestIDHeader))
	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(405), 405)
		return
	}
	if h.service == nil {
		writeJSON(w, 503, readiness.Result{Database: "unavailable"}, id)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	result := h.service.Check(ctx)
	status := 200
	if !result.Ready {
		status = 503
	}
	writeJSON(w, status, result, id)
}
