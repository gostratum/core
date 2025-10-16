package logx_test

import (
	"testing"

	"github.com/gostratum/core/configx"
	"github.com/gostratum/core/logx"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
	"go.uber.org/zap"
)

// TestLoggerModule verifies that logger.Module provides a *zap.Logger and
// that components depending on it receive a non-nil logger.
func TestLoggerModule(t *testing.T) {
	called := false

	// function that depends on *zap.Logger and is invoked on start
	ctor := func(l *zap.Logger) {
		if l == nil {
			t.Fatal("expected non-nil logger")
		}
		called = true
	}

	app := fxtest.New(
		t,
		fx.Provide(configx.New),
		logx.Module(),
		fx.Invoke(ctor),
	)
	defer app.RequireStart().RequireStop()

	if !called {
		t.Fatal("expected constructor to be invoked")
	}
}
