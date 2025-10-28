package configx

import (
	"strings"

	"github.com/mitchellh/mapstructure"
)

// Option configures the Loader during initialization.
type Option func(*LoaderConfig)

// LoaderConfig holds configuration for creating a Loader.
type LoaderConfig struct {
	ConfigPaths       []string
	EnvPrefix         string
	OverrideEnvPrefix string
	EnvReplacer       *strings.Replacer
	DecodeHooks       mapstructure.DecodeHookFunc
}

// WithConfigPaths sets the configuration paths for the Loader.
// Default: ["./configs"]
//
// Example:
//
//	loader := configx.New(
//	    configx.WithConfigPaths("./config", "/etc/myapp"),
//	)
func WithConfigPaths(paths ...string) Option {
	return func(cfg *LoaderConfig) {
		if len(paths) > 0 {
			cfg.ConfigPaths = paths
		}
	}
}

// WithEnvPrefix sets the environment variable prefix for the Loader.
// Default: "STRATUM"
//
// Example:
//
//	loader := configx.New(
//	    configx.WithEnvPrefix("MYAPP"),
//	)
//	// Will use MYAPP_* environment variables
func WithEnvPrefix(prefix string) Option {
	return func(cfg *LoaderConfig) {
		if prefix != "" {
			cfg.OverrideEnvPrefix = prefix
		}
	}
}

// WithEnvKeyReplacer sets the string replacer for environment variable key transformation.
// Default: Replaces "." and "-" with "_"
//
// Example:
//
//	replacer := strings.NewReplacer(".", "_", "-", "_", "/", "_")
//	loader := configx.New(
//	    configx.WithEnvKeyReplacer(replacer),
//	)
func WithEnvKeyReplacer(replacer *strings.Replacer) Option {
	return func(cfg *LoaderConfig) {
		if replacer != nil {
			cfg.EnvReplacer = replacer
		}
	}
}

// WithDecodeHook sets custom decode hooks for type conversions during unmarshaling.
// Default: Supports time.Duration, []string (comma-separated), and time.Time (RFC3339)
//
// Example:
//
//	customHook := mapstructure.StringToTimeDurationHookFunc()
//	loader := configx.New(
//	    configx.WithDecodeHook(customHook),
//	)
func WithDecodeHook(hook mapstructure.DecodeHookFunc) Option {
	return func(cfg *LoaderConfig) {
		if hook != nil {
			cfg.DecodeHooks = hook
		}
	}
}
