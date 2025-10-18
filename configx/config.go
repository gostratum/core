package configx

import (
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/creasty/defaults"
	"github.com/go-playground/validator/v10"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

// Loader loads configuration into structs.
type Loader interface {
	Bind(Configurable) error
	BindEnv(key string, envVars ...string) error
}

// Configurable is implemented by config structs that can provide a prefix.
type Configurable interface {
	Prefix() string
}

type viperLoader struct {
	v *viper.Viper
}

// New creates a new viper-based Loader.
func New() Loader {
	v := viper.New()

	// 1) Config paths (multi)
	paths := strings.TrimSpace(os.Getenv("CONFIG_PATHS"))
	if paths == "" {
		paths = "./configs"
	}
	for p := range strings.SplitSeq(paths, ",") {
		if p = strings.TrimSpace(p); p != "" {
			v.AddConfigPath(p)
		}
	}

	// 2) Layering: base + env (APP_ENV=dev|staging|prod ...)
	v.SetConfigName("base")
	_ = v.MergeInConfig()

	if env := strings.TrimSpace(os.Getenv("APP_ENV")); env != "" {
		v.SetConfigName(env)
		_ = v.MergeInConfig()
	}

	// 3) ENV override (namespaced)
	v.SetEnvPrefix("STRATUM")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	v.AutomaticEnv()
	return &viperLoader{v: v}
}

// Bind loads configuration into the provided props based on its Prefix() and
// mapstructure tags. It also auto-binds environment variables for every leaf key
// discovered in the struct so that ENV-only values are respected.
func (l *viperLoader) Bind(props Configurable) error {
	if props == nil {
		return fmt.Errorf("props is nil")
	}
	prefix := strings.TrimSuffix(strings.TrimSpace(props.Prefix()), ".")
	if err := defaults.Set(props); err != nil {
		return err
	}

	// 1) Auto-bind env for every leaf field under this prefix.
	if err := walkFields(props, func(fullKey string, _ []string, _ reflect.StructField) error {
		// Bind with the Viper instance. This respects SetEnvPrefix and Replacer.
		// No-op if already bound.
		if err := l.v.BindEnv(fullKey); err != nil {
			return err
		}
		return nil
	}, prefix); err != nil {
		return err
	}

	// 2) Build a nested map[string]any by reading values via v.Get(fullKey).
	// Using Get() ensures ENV overrides are applied lazily by Viper.
	m := make(map[string]any)
	if err := walkFields(props, func(fullKey string, parts []string, _ reflect.StructField) error {
		val := l.v.Get(fullKey)
		// Only set when value is present (IsSet) OR defaults produced a non-zero.
		// We prefer IsSet to avoid polluting subtree with nils.
		if l.v.IsSet(fullKey) || val != nil {
			setNested(m, parts, val)
		}
		return nil
	}, prefix); err != nil {
		return err
	}

	// 3) Decode into struct with helpful hooks.
	decCfg := &mapstructure.DecoderConfig{
		TagName:          "mapstructure",
		WeaklyTypedInput: true,
		Result:           props,
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.StringToSliceHookFunc(","),
			strToRFC3339TimeHook,
		),
	}
	dec, err := mapstructure.NewDecoder(decCfg)
	if err != nil {
		return err
	}
	if err := dec.Decode(m); err != nil {
		return err
	}

	// 4) Validate struct (tags: `validate:"..."`). Fail-fast if required fields missing.
	return validator.New().Struct(props)
}

// BindEnv binds a viper key to one or more environment variable names (aliases).
// Example: BindEnv("db.dsn", "STRATUM_DB_DSN", "DATABASE_URL").
func (l *viperLoader) BindEnv(key string, envVars ...string) error {
	if len(envVars) == 0 {
		return l.v.BindEnv(key)
	}
	return l.v.BindEnv(append([]string{key}, envVars...)...)
}
