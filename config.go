package core

import (
	"os"
	"strings"

	"github.com/spf13/viper"
)

// NewViper loads configuration using Viper.
func NewViper() (*viper.Viper, error) {
	v := viper.New()

	paths := os.Getenv("CONFIG_PATHS")
	if paths == "" {
		paths = "./configs"
	}
	for _, p := range strings.Split(paths, ",") {
		v.AddConfigPath(strings.TrimSpace(p))
	}

	v.SetConfigName("base")
	_ = v.MergeInConfig()

	if env := os.Getenv("APP_ENV"); env != "" {
		v.SetConfigName(env)
		_ = v.MergeInConfig()
	}

	v.SetEnvPrefix("STRATUM")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	v.AutomaticEnv()

	return v, nil
}

// LoadConfig unmarshals a key into T.
func LoadConfig[T any](v *viper.Viper, key string) (T, error) {
	var cfg T
	if err := v.UnmarshalKey(key, &cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}
