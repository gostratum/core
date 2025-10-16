package core

import (
	"testing"

	"github.com/gostratum/core/configx"
	"github.com/gostratum/core/logx"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

func TestAppNewBuilds(t *testing.T) {
	app := fxtest.New(t, fx.Provide(configx.New), logx.Module(), fx.Provide(NewHealthRegistry))
	defer app.RequireStart().RequireStop()
}
