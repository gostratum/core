package core_test

import (
	"context"
	"errors"
	"testing"

	"github.com/gostratum/core"
	"github.com/gostratum/core/configx"
	"github.com/gostratum/core/logx"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

// TestAppStartStop verifies that the Fx app starts and stops cleanly.
func TestAppStartStop(t *testing.T) {
	app := fxtest.New(
		t,
		// logger.Module is an fx.Option (module) and should be passed directly
		// to fx.New / fxtest.New. Provide core constructors separately.
		fx.Provide(configx.New, logx.Module, core.NewHealthRegistry),
	)
	defer app.RequireStart().RequireStop()
}

// TestNewProvidesCoreConfig verifies that core.New() provides the core config and runs without issues.
func TestNewProvidesCoreConfig(t *testing.T) {
	app := core.New()
	// Just test that it builds and can start/stop without the config causing issues
	go func() {
		app.Run()
	}()
	app.Stop(context.Background())
}

// TestHealthRegistry verifies basic health check registration and aggregation.
func TestHealthRegistry(t *testing.T) {
	registry := core.NewHealthRegistry()

	successCheck := &testCheck{
		name: "success",
		kind: core.Readiness,
		err:  nil,
	}
	failureCheck := &testCheck{
		name: "failure",
		kind: core.Readiness,
		err:  errors.New("check failed"),
	}

	registry.Register(successCheck)
	registry.Register(failureCheck)

	result := registry.Aggregate(context.Background(), core.Readiness)

	if result.OK {
		t.Error("Expected result.OK to be false when one check fails")
	}

	if len(result.Details) != 2 {
		t.Errorf("Expected 2 details, got %d", len(result.Details))
	}

	if detail, ok := result.Details["success"]; !ok || !detail.OK {
		t.Error("Expected 'success' check to pass")
	}

	if detail, ok := result.Details["failure"]; !ok || detail.OK {
		t.Error("Expected 'failure' check to fail")
	}
}

// TestHealthRegistryLivenessReadiness verifies liveness and readiness are separate.
func TestHealthRegistryLivenessReadiness(t *testing.T) {
	registry := core.NewHealthRegistry()

	livenessCheck := &testCheck{
		name: "live",
		kind: core.Liveness,
		err:  nil,
	}
	readinessCheck := &testCheck{
		name: "ready",
		kind: core.Readiness,
		err:  nil,
	}

	registry.Register(livenessCheck)
	registry.Register(readinessCheck)

	livenessResult := registry.Aggregate(context.Background(), core.Liveness)
	if !livenessResult.OK || len(livenessResult.Details) != 1 {
		t.Error("Liveness check should only include liveness checks")
	}

	readinessResult := registry.Aggregate(context.Background(), core.Readiness)
	if !readinessResult.OK || len(readinessResult.Details) != 1 {
		t.Error("Readiness check should only include readiness checks")
	}
}

// testCheck is a simple implementation of Check for testing.
type testCheck struct {
	name string
	kind core.Kind
	err  error
}

func (c *testCheck) Name() string {
	return c.name
}

func (c *testCheck) Kind() core.Kind {
	return c.kind
}

func (c *testCheck) Check(ctx context.Context) error {
	return c.err
}
