package core

import (
	"fmt"

	"github.com/gostratum/core/pkg/configx"
	"github.com/gostratum/core/pkg/logx"
)

// App holds core services shared across the runtime lifecycle.
type App struct {
	Cfg *configx.Config
	Log *logx.Logger
}

// Bootstrap loads configuration, wires dependencies, and produces an App.
func Bootstrap(opts BuildOptions) (*App, error) {
	paths := opts.ConfigPaths
	envPrefix := opts.EnvPrefix
	if envPrefix == "" {
		envPrefix = "APP"
	}

	cfg, err := configx.Load(paths, envPrefix)
	if err != nil {
		return nil, fmt.Errorf("core: load config: %w", err)
	}

	if opts.Addr != "" {
		cfg.Server.Addr = opts.Addr
	}
	if opts.ReadHeaderTimeout > 0 {
		cfg.Server.ReadHeaderTimeout = opts.ReadHeaderTimeout
	}

	logger := logx.New(cfg.Observability.JSONLogs)

	return &App{Cfg: cfg, Log: logger}, nil
}
