package core

import (
	"testing"

	"github.com/gostratum/core/configx"
	"github.com/gostratum/core/logger"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

func TestAppNewBuilds(t *testing.T) {
	app := fxtest.New(t, fx.Provide(configx.New), logger.Module(), fx.Provide(NewHealthRegistry))
	defer app.RequireStart().RequireStop()
}
