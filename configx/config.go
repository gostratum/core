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

	// 2) Unmarshal the subtree; root viper handles precedence (ENV > file > defaults)
	if err := l.v.UnmarshalKey(
		prefix,
		props,
		viper.DecodeHook(
			mapstructure.ComposeDecodeHookFunc(
				mapstructure.StringToTimeDurationHookFunc(),
				mapstructure.StringToSliceHookFunc(","),
				strToRFC3339TimeHook,
			),
		),
	); err != nil {
		return err
	}

	// 3) Validate
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
