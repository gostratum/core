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
	envPrefix  string
}

// New creates a new Loader with optional configuration.
//
// Configuration Precedence (highest to lowest):
//  1. Environment variables (customizable prefix, see below)
//  2. Environment-specific config file ({APP_ENV}.yaml)
//  3. Base config file (base.yaml)
//  4. Struct tag defaults (default:"value")
//
// Environment Prefix Precedence (highest to lowest):
//  1. WithEnvPrefix() option (in code)
//  2. ENV_PREFIX environment variable (global default)
//  3. "STRATUM" (hardcoded default)
//
// Configuration paths can be set via:
//   - CONFIG_PATHS environment variable (comma-separated)
//   - WithConfigPaths() option
//
// Example:
//
//	// Option 1: Use ENV_PREFIX environment variable
//	// export ENV_PREFIX=MYAPP
//	loader := configx.New()
//	// Uses MYAPP_* environment variables
//
//	// Option 2: Use WithEnvPrefix() option (highest priority)
//	loader := configx.New(
//	    configx.WithConfigPaths("./config"),
//	    configx.WithEnvPrefix("MYAPP"),
//	)
//	// Uses MYAPP_* environment variables
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

	// Check ENV_PREFIX env var for global prefix override
	if envPrefix := strings.TrimSpace(os.Getenv(EnvPrefix)); envPrefix != "" {
		cfg.EnvPrefix = envPrefix
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
		envPrefix:  cfg.EnvPrefix,
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

	// Build env var names using a replacer that mirrors how we convert viper keys
	// into env var names (dot and dash become underscores, then upper-cased).
	replacer := strings.NewReplacer(".", "_", "-", "_")
	makeEnv := func(k string) string {
		return strings.ToUpper(replacer.Replace(k))
	}

	// Unprefixed env var name (e.g. db.databases.primary.dsn -> DB_DATABASES_PRIMARY_DSN)
	unprefixed := makeEnv(normalizedKey)

	// Prefixed env var name using the loader's resolved prefix (if any)
	var prefixed string
	if p := strings.TrimSpace(l.envPrefix); p != "" {
		prefixed = strings.ToUpper(p) + "_" + unprefixed
	}

	// Build final argument list for viper.BindEnv: first the key, then one or more
	// explicit env var names. If the caller provided envVars, ensure the prefixed
	// alias is also present (so callers don't need to hardcode the prefix).
	args := []string{normalizedKey}
	if len(envVars) == 0 {
		// No explicit aliases: bind both unprefixed and prefixed (if available).
		args = append(args, unprefixed)
		if prefixed != "" {
			args = append(args, prefixed)
		}
	} else {
		// Use caller-provided aliases but add the computed prefixed alias if missing.
		args = append(args, envVars...)
		if prefixed != "" {
			found := false
			for _, v := range envVars {
				if strings.EqualFold(v, prefixed) {
					found = true
					break
				}
			}
			if !found {
				args = append(args, prefixed)
			}
		}
	}

	if err := l.v.BindEnv(args...); err != nil {
		return fmt.Errorf("failed to bind env for key '%s': %w", normalizedKey, err)
	}

	// Track bound keys for resolution during Bind()
	if l.boundKeys != nil {
		l.boundKeys[normalizedKey] = true
	}

	return nil
}
