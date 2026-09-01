package readiness

import (
	"context"
	"errors"
	"testing"
)

func TestServiceReportsReadyWhenDatabaseResponds(t *testing.T) {
	t.Parallel()
	service := NewService(staticPinger{})
	value := service.Check(context.Background())
	if !value.Ready || value.Database != "ready" {
		t.Fatalf("Check() = %#v", value)
	}
}
func TestServiceReportsUnavailableWithoutLeakingDependencyError(t *testing.T) {
	t.Parallel()
	service := NewService(staticPinger{err: errors.New("secret connection details")})
	value := service.Check(context.Background())
	if value.Ready || value.Database != "unavailable" {
		t.Fatalf("Check() = %#v", value)
	}
}

type staticPinger struct{ err error }

func (p staticPinger) PingContext(context.Context) error { return p.err }
