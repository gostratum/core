package configx

import (
	"fmt"
	"os"
	"strings"

	"github.com/creasty/defaults"
	"github.com/go-playground/validator/v10"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

// viperLoader implements the Loader interface using Viper.
type viperLoader struct {
	v          *viper.Viper
	decodeHook mapstructure.DecodeHookFunc
	boundKeys  map[string]bool
}

// New creates a new Loader with optional configuration.
//
// Configuration Precedence (highest to lowest):
//  1. Environment variables (STRATUM_* by default)
//  2. Environment-specific config file ({APP_ENV}.yaml)
//  3. Base config file (base.yaml)
//  4. Struct tag defaults (default:"value")
//
// Configuration paths can be set via:
//   - CONFIG_PATHS environment variable (comma-separated)
//   - WithConfigPaths() option
//
// Example:
//
//	loader := configx.New(
//	    configx.WithConfigPaths("./config"),
//	    configx.WithEnvPrefix("MYAPP"),
//	)
func New(opts ...Option) Loader {
	cfg := &LoaderConfig{
		ConfigPaths: []string{DefaultConfigPath},
		EnvPrefix:   DefaultEnvPrefix,
		EnvReplacer: strings.NewReplacer(".", "_", "-", "_"),
		DecodeHooks: mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.StringToSliceHookFunc(","),
			strToRFC3339TimeHook,
		),
	}

	// Check CONFIG_PATHS env var for backward compatibility
	if paths := strings.TrimSpace(os.Getenv(EnvConfigPaths)); paths != "" {
		pathList := []string{}
		for p := range strings.SplitSeq(paths, ",") {
			if p = strings.TrimSpace(p); p != "" {
				pathList = append(pathList, p)
			}
		}
		if len(pathList) > 0 {
			cfg.ConfigPaths = pathList
		}
	}

	// Apply options (these override env var if provided)
	for _, opt := range opts {
		opt(cfg)
	}

	v := viper.New()

	// Add config paths
	for _, path := range cfg.ConfigPaths {
		if p := strings.TrimSpace(path); p != "" {
			v.AddConfigPath(p)
		}
	}

	// Layering: base + environment-specific config
	v.SetConfigName(BaseConfigFile)
	_ = v.MergeInConfig()

	if env := strings.TrimSpace(os.Getenv(EnvAppEnv)); env != "" {
		v.SetConfigName(env)
		_ = v.MergeInConfig()
	}

	// Environment variable override
	v.SetEnvPrefix(cfg.EnvPrefix)
	v.SetEnvKeyReplacer(cfg.EnvReplacer)
	v.AutomaticEnv()

	return &viperLoader{
		v:          v,
		decodeHook: cfg.DecodeHooks,
		boundKeys:  make(map[string]bool),
	}
}

// Bind loads configuration into the provided struct.
// Configuration precedence: ENV > YAML > Defaults
func (l *viperLoader) Bind(props Configurable) error {
	if props == nil {
		return fmt.Errorf("props is nil")
	}

	prefix := normalizeKey(props.Prefix())
	if prefix == "" {
		return fmt.Errorf("props.Prefix() cannot be empty")
	}

	rebuildSettings := make(map[string]any)

	// Iterate all keys from Viper (includes YAML + env vars)
	for _, fullKey := range l.v.AllKeys() {
		if strings.HasPrefix(fullKey, prefix+".") {
			value := l.v.Get(fullKey)
			if value != nil {
				keyWithoutPrefix := strings.TrimPrefix(fullKey, prefix+".")
				setNestedValue(rebuildSettings, strings.Split(keyWithoutPrefix, "."), value)
			}
		}
	}

	// Check explicitly bound keys (for env-only sensitive values)
	for boundKey := range l.boundKeys {
		keyWithoutPrefix := strings.TrimPrefix(boundKey, prefix+".")
		if keyWithoutPrefix != boundKey && keyWithoutPrefix != "" {
			value := l.v.Get(boundKey)
			if value != nil {
				setNestedValue(rebuildSettings, strings.Split(keyWithoutPrefix, "."), value)
			}
		}
	}

	// Decode into struct
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:           props,
		DecodeHook:       l.decodeHook,
		WeaklyTypedInput: true,
	})
	if err != nil {
		return fmt.Errorf("failed to create decoder: %w", err)
	}

	if err := decoder.Decode(rebuildSettings); err != nil {
		return fmt.Errorf("failed to decode config for prefix '%s': %w", prefix, err)
	}

	// Apply struct tag defaults
	if err := defaults.Set(props); err != nil {
		return fmt.Errorf("failed to set defaults: %w", err)
	}

	// Validate configuration
	if err := validator.New().Struct(props); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	return nil
}

// BindEnv explicitly binds configuration keys to environment variables.
// Use this for sensitive values that should only come from environment.
//
// Example:
//
//	loader.BindEnv("db.dsn", "DATABASE_URL")
//	loader.BindEnv("api.key")
func (l *viperLoader) BindEnv(key string, envVars ...string) error {
	normalizedKey := normalizeKey(key)
	if normalizedKey == "" {
		return fmt.Errorf("cannot bind empty key")
	}

	var err error
	if len(envVars) == 0 {
		err = l.v.BindEnv(normalizedKey)
	} else {
		err = l.v.BindEnv(append([]string{normalizedKey}, envVars...)...)
	}

	if err != nil {
		return fmt.Errorf("failed to bind env for key '%s': %w", normalizedKey, err)
	}

	// Track bound keys for resolution during Bind()
	if l.boundKeys != nil {
		l.boundKeys[normalizedKey] = true
	}

	return nil
}
