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

## Configuration (configx)

GoStratum's `configx` package provides **typed configuration loading** with automatic environment variable binding, file-based configuration, and struct tag defaults.

### Configuration Precedence

Configuration values are loaded with the following precedence (highest to lowest):

1. **Environment Variables** - `STRATUM_*` prefixed (customizable)
2. **Environment-Specific Config** - `{APP_ENV}.yaml` (e.g., `prod.yaml`, `dev.yaml`)  
3. **Base Config File** - `base.yaml`
4. **Struct Tag Defaults** - `default:"value"` tags

### Quick Start

```go
package main

import (
    "log"
    "github.com/gostratum/core/configx"
)

// 1. Define your configuration struct
type AppConfig struct {
    Port     int    `mapstructure:"port" default:"8080" validate:"required,min=1,max=65535"`
    Host     string `mapstructure:"host" default:"localhost"`
    LogLevel string `mapstructure:"log_level" default:"info"`
}

// 2. Implement Prefix() to specify config namespace
func (AppConfig) Prefix() string {
    return "app"
}

func main() {
    // 3. Create loader with optional configuration
    loader := configx.New(
        configx.WithConfigPaths("./config"),
        configx.WithEnvPrefix("MYAPP"),
    )

    // 4. Bind configuration
    config := &AppConfig{}
    if err := loader.Bind(config); err != nil {
        log.Fatal(err)
    }

    log.Printf("Server starting on %s:%d", config.Host, config.Port)
}
```

### Loader Options

The `configx.New()` function accepts functional options for customization:

#### WithConfigPaths(paths ...string)
Set custom configuration file directories. Default: `["./configs"]`

```go
loader := configx.New(
    configx.WithConfigPaths("./config", "/etc/myapp"),
)
```

#### WithEnvPrefix(prefix string)
Set custom environment variable prefix. Default: `"STRATUM"`

```go
loader := configx.New(
    configx.WithEnvPrefix("MYAPP"),
)
// Uses MYAPP_APP_PORT instead of STRATUM_APP_PORT
```

#### WithEnvKeyReplacer(replacer *strings.Replacer)
Customize how config keys are mapped to environment variables. Default: Replaces `"."` and `"-"` with `"_"`

```go
replacer := strings.NewReplacer(".", "_", "-", "_", "/", "_")
loader := configx.New(
    configx.WithEnvKeyReplacer(replacer),
)
```

#### WithDecodeHook(hook mapstructure.DecodeHookFunc)
Add custom type conversion during config decoding. Default hooks support `time.Duration`, `[]string`, and `time.Time`.

```go
customHook := mapstructure.ComposeDecodeHookFunc(
    mapstructure.StringToTimeDurationHookFunc(),
    myCustomHook,
)
loader := configx.New(
    configx.WithDecodeHook(customHook),
)
```

### Configuration Sources

#### 1. Environment Variables (Highest Priority)

Environment variables automatically override file-based configuration:

```bash
export STRATUM_APP_PORT=9000
export STRATUM_APP_HOST=0.0.0.0
export STRATUM_APP_LOG_LEVEL=debug
```

For nested configuration:
```bash
export STRATUM_DB_DATABASES_PRIMARY_DSN="postgres://localhost/mydb"
```

#### 2. Configuration Files

Place YAML files in your config directory (default: `./configs`):

```yaml
# configs/base.yaml - Base configuration for all environments
app:
  port: 8080
  host: localhost
  log_level: info

db:
  databases:
    primary:
      driver: postgres
      host: localhost
      port: 5432
```

```yaml
# configs/prod.yaml - Production overrides
app:
  port: 80
  host: 0.0.0.0
  log_level: warn

db:
  databases:
    primary:
      host: prod-db.example.com
```

Set `APP_ENV` to load environment-specific config:
```bash
export APP_ENV=prod  # Loads base.yaml, then prod.yaml
```

#### 3. Struct Tag Defaults (Lowest Priority)

Defaults are applied only if values aren't set by environment or config files:

```go
type Config struct {
    Port     int    `default:"8080"`
    Timeout  string `default:"30s"`
    LogLevel string `default:"info"`
}
```

### Binding Sensitive Environment-Only Values

Use `BindEnv()` for sensitive values that should only come from environment:

```go
loader := configx.New()

// Explicitly bind database DSN to environment variables
loader.BindEnv("db.databases.primary.dsn", "DATABASE_URL", "DB_DSN")
loader.BindEnv("api.secret_key", "API_SECRET")

config := &AppConfig{}
loader.Bind(config)
```

This ensures these values:
- Won't be accidentally committed in YAML files
- Can use multiple environment variable aliases
- Are checked even if not present in config files

### Validation

Configuration is automatically validated using struct tags:

```go
type Config struct {
    Port     int    `validate:"required,min=1,max=65535"`
    Email    string `validate:"required,email"`
    LogLevel string `validate:"oneof=debug info warn error"`
}
```

If validation fails, `Bind()` returns a descriptive error.

### Environment Variables

- `CONFIG_PATHS`: Comma-separated config directories (default: `./configs`)
- `APP_ENV`: Environment name for loading `{APP_ENV}.yaml`
- `STRATUM_*`: Configuration values (prefix customizable via `WithEnvPrefix`)

### Complete Example

```go
package main

import (
    "log"
    "time"
    "github.com/gostratum/core/configx"
)

type DatabaseConfig struct {
    Driver   string        `mapstructure:"driver" default:"postgres"`
    Host     string        `mapstructure:"host" default:"localhost" validate:"required"`
    Port     int           `mapstructure:"port" default:"5432" validate:"min=1,max=65535"`
    Timeout  time.Duration `mapstructure:"timeout" default:"30s"`
}

type AppConfig struct {
    Port     int              `mapstructure:"port" default:"8080"`
    Database DatabaseConfig   `mapstructure:"database"`
}

func (AppConfig) Prefix() string { return "app" }

func main() {
    loader := configx.New(
        configx.WithConfigPaths("./config"),
    )

    // Bind sensitive database DSN from environment only
    loader.BindEnv("app.database.dsn", "DATABASE_URL")

    config := &AppConfig{}
    if err := loader.Bind(config); err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }

    log.Printf("Server config loaded: %+v", config)
}
```

**Config file** (`config/base.yaml`):
```yaml
app:
  port: 8080
  database:
    driver: postgres
    host: localhost
    port: 5432
    timeout: 30s
```

**Environment override**:
```bash
export APP_ENV=prod
export STRATUM_APP_PORT=80
export STRATUM_APP_DATABASE_HOST=prod-db.example.com
export DATABASE_URL=postgres://user:pass@prod-db/myapp
```

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
