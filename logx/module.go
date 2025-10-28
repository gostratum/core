package logx

import (
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
)

// Module provides the logx module for fx
func Module() fx.Option {
	return fx.Module(
		"logx",
		fx.Provide(
			NewLoggerConfig,
			NewLogger,
			ProvideAdapter,
			NewSugared,
		),
		fx.WithLogger(FxEventLogger),
	)
}

// FxEventLogger returns an fxevent.Logger for Fx lifecycle events
func FxEventLogger(l *zap.Logger) fxevent.Logger { return &fxevent.ZapLogger{Logger: l} }
