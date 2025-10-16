# gostratum/core

Minimal core for Gostratum cloud applications.

## Overview

This package provides the foundational building blocks for Gostratum applications. It includes a dedicated `configx` package for typed configuration loading and validation.

- **Application lifecycle & dependency injection** via [Uber FX](https://uber-go.github.io/fx/)
- **Configuration management** via [Viper](https://github.com/spf13/viper)
- **Logging** via [Zap](https://github.com/uber-go/zap)
- **Health registry** for liveness and readiness checks

## Installation

```bash
go get github.com/gostratum/core
```

## Quick Start

```go
package main

import (
	"context"
	"fmt"

	"github.com/gostratum/core"
	"github.com/gostratum/core/configx"
	"github.com/spf13/viper"
	"go.uber.org/fx"
)

type MyConfig struct {
	Port int `mapstructure:"port"`
}

func (MyConfig) Prefix() string { return "app" }

type PingCheck struct{}

func (PingCheck) Name() string            { return "ping" }
func (PingCheck) Kind() core.Kind         { return core.Readiness }
func (PingCheck) Check(ctx context.Context) error { return nil }

func main() {
	// Use the typed loader from `configx` for binding/validation.
	app := core.New(
		fx.Invoke(func(l configx.Loader, h core.Registry) {
			var cfg MyConfig
			_ = l.Bind(&cfg) // MyConfig must implement Prefix() string if using sub-keys
			fmt.Println("Loaded config:", cfg)
			h.Register(PingCheck{})
		}),
	)
	app.Run()

	// Option B: use the new typed loader in `configx` which supports defaults and validation.
	loader := configx.New()
	var cfg MyConfig
	_ = loader.Bind(&cfg) // cfg must implement Prefix() string to locate its sub-key
}
```

## Configuration

The `NewViper()` function loads configuration from:

1. **Config files**: By default looks in `./configs` directory
   - `base.yaml` - Base configuration
   - `{APP_ENV}.yaml` - Environment-specific overrides (e.g., `dev.yaml`, `prod.yaml`)
2. **Environment variables**: Prefixed with `STRATUM_` (e.g., `STRATUM_SERVER_PORT`)
   - Use `_` to represent `.` or `-` in config keys

### Environment Variables

- `CONFIG_PATHS`: Comma-separated list of config directories (default: `./configs`)
- `APP_ENV`: Environment name for config file selection (e.g., `dev`, `prod`)
- `STRATUM_*`: Configuration overrides (e.g., `STRATUM_SERVER_PORT=8080`)

## Logging

The logger automatically selects development or production mode based on `APP_ENV` and is configurable via `core.logger` config (see `configx`):

- `APP_ENV=dev`: Development mode (human-readable console output)
- Other values: Production mode (JSON structured logs)

Recent fix: the logger now preserves the development encoder defaults provided by `zap.NewDevelopmentConfig()` and only applies a custom time format when needed. This avoids unintentionally overwriting development-friendly settings.

## Health Checks

The health registry supports two types of checks:

- **Liveness**: Is the service alive?
- **Readiness**: Is the service ready to serve traffic?

```go
type CustomCheck struct{}

func (c CustomCheck) Name() string { return "custom" }
func (c CustomCheck) Kind() core.Kind { return core.Liveness }
func (c CustomCheck) Check(ctx context.Context) error {
	// Your health check logic
	return nil
}

// Register in your Fx app
fx.Invoke(func(h core.Registry) {
	h.Register(CustomCheck{})
})
```

### Health Check Timeout

Control health check timeout via environment variable:

- `STRATUM_HEALTH_TIMEOUT_MS`: Timeout in milliseconds (default: 300ms)

Note: the health registry supports programmatic `Set` calls to update check status; however `Aggregate` currently queries registered checks and runs them with the configured timeout. Consider reviewing `Set` vs `Aggregate` semantics if you need stored status to be included in aggregation results.

## Dependencies

This package depends on a small set of well-known libraries:

- `go.uber.org/fx` - Application lifecycle and dependency injection
- `go.uber.org/zap` - Structured logging
- `github.com/spf13/viper` - Configuration management
- `github.com/creasty/defaults` - Struct defaulting
- `github.com/go-playground/validator/v10` - Validation for config structs

Run `go mod tidy` to ensure the `go.mod` is clean after making cross-module changes.

## License

MIT
