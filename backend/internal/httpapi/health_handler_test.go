package httpapi_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/health"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/httpapi"
)

func TestHealthHandlerReturnsCorrelationHeaderAndBody(t *testing.T) {
	service := health.NewService("0.1.0", func() time.Time {
		return time.Date(2026, 8, 20, 9, 30, 0, 0, time.UTC)
	})
	handler := httpapi.NewHealthHandler(service)
	request := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	request.Header.Set("X-Request-ID", "req_from_client")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if got := response.Header().Get("X-Request-ID"); got != "req_from_client" {
		t.Fatalf("X-Request-ID = %q, want req_from_client", got)
	}
	if got := response.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", got)
	}

	var body struct {
		Status    string `json:"status"`
		Service   string `json:"service"`
		Version   string `json:"version"`
		CheckedAt string `json:"checkedAt"`
		RequestID string `json:"requestId"`
	}
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.RequestID != "req_from_client" {
		t.Fatalf("body requestId = %q, want req_from_client", body.RequestID)
	}
	if body.Status != "ok" || body.Service != "juntly-api" || body.Version != "0.1.0" {
		t.Fatalf("unexpected health body: %#v", body)
	}
}

func TestHealthHandlerGeneratesCorrelationIDWhenMissing(t *testing.T) {
	service := health.NewService("0.1.0", func() time.Time {
		return time.Date(2026, 8, 20, 9, 30, 0, 0, time.UTC)
	})
	handler := httpapi.NewHealthHandler(service)
	request := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	headerRequestID := response.Header().Get("X-Request-ID")
	if headerRequestID == "" {
		t.Fatal("X-Request-ID header is empty")
	}

	var body struct {
		RequestID string `json:"requestId"`
	}
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.RequestID != headerRequestID {
		t.Fatalf("body requestId = %q, want header %q", body.RequestID, headerRequestID)
	}
}
