package core

import (
	"github.com/gostratum/core/configx"
	"github.com/gostratum/core/logger"
	"go.uber.org/fx"
)

// New builds the Fx app with default Gostratum core modules.
func New(opts ...fx.Option) *fx.App {
	return fx.New(
		fx.Provide(configx.New, logger.NewLogger, NewHealthRegistry),
		fx.Options(opts...),
	)
}

// Run starts the Fx application and blocks until shutdown.
func Run(app *fx.App) {
	app.Run()
}
