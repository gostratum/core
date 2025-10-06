package core

import (
	"net/http"
	"time"

	"github.com/gostratum/core/pkg/configx"
)

// BuildOptions configures the application bootstrap process.
type BuildOptions struct {
	ConfigPaths       []string
	EnvPrefix         string
	Addr              string
	ReadHeaderTimeout time.Duration
}

// Deps aggregates dependencies exposed to caller-provided HTTP factories.
type Deps struct {
	Config *configx.Config
}

// HTTPHandlerFactory builds the net/http handler used by the runtime.
type HTTPHandlerFactory func(d Deps) http.Handler
