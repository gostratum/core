package core

import (
	"context"
	"os"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

// NewLogger creates a new zap logger based on the APP_ENV environment variable.
func NewLogger(lc fx.Lifecycle) (*zap.Logger, error) {
	var cfg zap.Config
	if os.Getenv("APP_ENV") == "dev" {
		cfg = zap.NewDevelopmentConfig()
	} else {
		cfg = zap.NewProductionConfig()
	}

	logger, err := cfg.Build()
	if err != nil {
		return nil, err
	}

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			_ = logger.Sync()
			return nil
		},
	})
	return logger, nil
}
