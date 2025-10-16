package configx

import (
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/creasty/defaults"
	"github.com/go-playground/validator/v10"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

// Loader loads configuration into structs.
type Loader interface {
	Bind(Configurable) error
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
	paths := os.Getenv("CONFIG_PATHS")
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

func (l *viperLoader) Bind(props Configurable) error {
	prefix := props.Prefix()
	sub := l.v.Sub(prefix)
	if sub == nil {
		sub = viper.New()
	}

	if err := defaults.Set(props); err != nil {
		return err
	}

	// Unmarshal into a map first, then decode with mapstructure to set decoder options.
	var m map[string]any
	if err := sub.Unmarshal(&m); err != nil {
		return err
	}

	decCfg := &mapstructure.DecoderConfig{
		TagName:          "mapstructure",
		WeaklyTypedInput: true,
		Result:           props,
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.StringToSliceHookFunc(","),
			func(from, to reflect.Type, data any) (any, error) {
				if from.Kind() == reflect.String && to == reflect.TypeOf(time.Time{}) {
					s := data.(string)
					if s == "" {
						return time.Time{}, nil
					}
					t, err := time.Parse(time.RFC3339, s)
					return t, err
				}
				return data, nil
			},
		),
	}
	dec, err := mapstructure.NewDecoder(decCfg)
	if err != nil {
		return err
	}
	if err := dec.Decode(m); err != nil {
		return err
	}

	return validator.New().Struct(props)
}
