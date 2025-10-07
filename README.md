# gostratum/core

Minimal core for Gostratum cloud applications.

## Overview

This package provides the foundational building blocks for Gostratum applications:

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
	"github.com/spf13/viper"
	"go.uber.org/fx"
)

type MyConfig struct {
	Port int `mapstructure:"port"`
}

type PingCheck struct{}

func (PingCheck) Name() string            { return "ping" }
func (PingCheck) Kind() core.Kind         { return core.Readiness }
func (PingCheck) Check(ctx context.Context) error { return nil }

func main() {
	app := core.New(
		fx.Invoke(func(v *viper.Viper, h core.Registry) {
			cfg, _ := core.LoadConfig[MyConfig](v, "app")
			fmt.Println("Loaded config:", cfg)
			h.Register(PingCheck{})
		}),
	)
	app.Run()
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

The logger automatically selects development or production mode based on `APP_ENV`:

- `APP_ENV=dev`: Development mode (human-readable console output)
- Other values: Production mode (JSON structured logs)

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

## Dependencies

This package has minimal dependencies:

- `go.uber.org/fx` - Application lifecycle and dependency injection
- `go.uber.org/zap` - Structured logging
- `github.com/spf13/viper` - Configuration management
- `github.com/spf13/pflag` - Command-line flag parsing

## License

MIT
