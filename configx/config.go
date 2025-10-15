package configx

import (
	"strings"

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
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
	v.SetConfigName("config")
	v.AddConfigPath("./config")
	_ = v.ReadInConfig()
	return &viperLoader{v: v}
}

func (l *viperLoader) Bind(props Configurable) error {
	prefix := props.Prefix()
	sub := l.v.Sub(prefix)
	if sub == nil {
		sub = viper.New()
	}
	_ = defaults.Set(props)

	// Unmarshal into a map first, then decode with mapstructure to set decoder options.
	var m map[string]interface{}
	if err := sub.Unmarshal(&m); err != nil {
		return err
	}

	decCfg := &mapstructure.DecoderConfig{
		TagName:          "mapstructure",
		WeaklyTypedInput: true,
		Result:           props,
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
