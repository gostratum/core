package core

import (
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
)

// New builds the Fx app with default Gostratum core modules.
func New(opts ...fx.Option) *fx.App {
	return fx.New(
		fx.Provide(NewLogger, NewViper, NewHealthRegistry),
		fx.WithLogger(func(l *zap.Logger) fxevent.Logger {
			return &fxevent.ZapLogger{Logger: l}
		}),
		fx.Options(opts...),
	)
}

// Run starts the Fx application and blocks until shutdown.
func Run(app *fx.App) {
	app.Run()
}
