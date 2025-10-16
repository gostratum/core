package core

import (
	"context"
	"testing"
	"time"
)

type slowCheck struct{ name string }

func (s *slowCheck) Name() string { return s.name }
func (s *slowCheck) Kind() Kind   { return Readiness }
func (s *slowCheck) Check(ctx context.Context) error {
	// exceed typical timeout
	select {
	case <-time.After(500 * time.Millisecond):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func TestRegistrySetAndAggregate(t *testing.T) {
	r := NewHealthRegistry()
	// Set an explicit failure status
	r.Set(Readiness, "db", ErrHealthCheckFailed)
	// Current implementation stores status separately; Aggregate only
	// considers registered checks. With no checks registered, result
	// should be OK and have no details.
	res := r.Aggregate(context.Background(), Readiness)
	if !res.OK {
		t.Fatalf("expected overall OK when only Set used and no checks")
	}
	if len(res.Details) != 0 {
		t.Fatalf("expected no details when only Set used and no checks, got %#v", res.Details)
	}
}

func TestRegistryTimeoutBehavior(t *testing.T) {
	r := NewHealthRegistry()
	r.Register(&slowCheck{name: "slow"})
	// set a low env timeout to trigger
	t.Setenv("STRATUM_HEALTH_TIMEOUT_MS", "50")
	res := r.Aggregate(context.Background(), Readiness)
	if res.OK {
		t.Fatalf("expected overall not OK due to timeout")
	}
	if _, ok := res.Details["slow"]; !ok {
		t.Fatalf("expected slow check detail present")
	}
}

func TestRegistryNoChecks(t *testing.T) {
	r := NewHealthRegistry()
	res := r.Aggregate(context.Background(), Readiness)
	if !res.OK {
		t.Fatalf("expected OK when no checks registered")
	}
	if len(res.Details) != 0 {
		t.Fatalf("expected no details when no checks, got %d", len(res.Details))
	}
}
