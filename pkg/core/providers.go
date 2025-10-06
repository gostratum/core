package core

import (
	"github.com/gostratum/core/pkg/configx"
	"github.com/gostratum/core/pkg/logx"
)

// ProvideConfig is a Wire provider that loads configuration using configx.Load.
func ProvideConfig(paths []string, envPrefix string) (*configx.Config, error) {
	return configx.Load(paths, envPrefix)
}

// ProvideLogger is a Wire provider that constructs the core logger.
func ProvideLogger(cfg *configx.Config) *logx.Logger {
	if cfg == nil {
		return logx.New(false)
	}
	return logx.New(cfg.Observability.JSONLogs)
}
