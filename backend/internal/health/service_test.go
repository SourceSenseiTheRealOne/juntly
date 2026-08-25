package health_test

import (
	"testing"
	"time"

	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/health"
)

func TestServiceCheckReturnsVersionedHealth(t *testing.T) {
	now := time.Date(2026, 8, 20, 9, 30, 0, 0, time.UTC)
	service := health.NewService("0.1.0", func() time.Time {
		return now
	})

	result := service.Check("req_test_123")

	if result.Status != "ok" {
		t.Fatalf("Status = %q, want ok", result.Status)
	}
	if result.Service != "juntly-api" {
		t.Fatalf("Service = %q, want juntly-api", result.Service)
	}
	if result.Version != "0.1.0" {
		t.Fatalf("Version = %q, want 0.1.0", result.Version)
	}
	if !result.CheckedAt.Equal(now) {
		t.Fatalf("CheckedAt = %s, want %s", result.CheckedAt, now)
	}
	if result.RequestID != "req_test_123" {
		t.Fatalf("RequestID = %q, want req_test_123", result.RequestID)
	}
}
