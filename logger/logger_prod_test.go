package logger

import (
	"testing"

	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
	"go.uber.org/zap"
)

func TestLoggerModule_ProdConfig(t *testing.T) {
	// Provide a LoggerConfig with Env=prod to exercise prod branch
	cfg := LoggerConfig{Env: "prod", Level: "info", Encoding: "json", SamplingInitial: 5, SamplingThereafter: 10}

	app := fxtest.New(
		t,
		fx.Provide(func() LoggerConfig { return cfg }),
		fx.Provide(NewLogger, NewSugared),
		fx.WithLogger(FxEventLogger),
		fx.Invoke(func(l *zap.Logger) {
			if l == nil {
				t.Fatal("expected non-nil logger")
			}
		}),
	)
	defer app.RequireStart().RequireStop()
}
