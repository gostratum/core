package configx

import (
	"testing"

	"github.com/spf13/viper"
)

type envHookConfig struct {
	Dur string `mapstructure:"dur"`
}

func (e *envHookConfig) Prefix() string { return "hook" }

func TestNew_PicksUpEnvOverrides(t *testing.T) {
	v := viper.New()
	// Simulate config file value directly so Sub() will find it.
	v.Set("hook.dur", "2m")

	loader := &viperLoader{v: v}
	var cfg envHookConfig
	if err := loader.Bind(&cfg); err != nil {
		t.Fatalf("expected bind to succeed with env override, got %v", err)
	}
	if cfg.Dur != "2m" {
		t.Fatalf("expected dur 2m from env, got %s", cfg.Dur)
	}
}
