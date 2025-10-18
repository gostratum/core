package configx

import (
	"testing"

	"github.com/mitchellh/mapstructure"
)

// envHookConfig is defined in new_test.go; redefine a small type here for isolation
type envOnlyConfig struct {
	Dur string `mapstructure:"dur"`
}

func (e *envOnlyConfig) Prefix() string { return "hook" }

type otherEnvConfig struct {
	Dur string `mapstructure:"dur"`
}

func (o *otherEnvConfig) Prefix() string { return "other" }

func TestBind_PicksUpEnvWhenNoFile(t *testing.T) {
	t.Setenv("STRATUM_HOOK_DUR", "3m")

	loader := New()
	// Explicitly bind the viper key to the environment variable since
	// automatic reflection-based binding was removed.
	if err := loader.BindEnv("hook.dur", "STRATUM_HOOK_DUR"); err != nil {
		t.Fatalf("BindEnv failed: %v", err)
	}

	// Decode directly from viper's settings map: Bind no longer merges
	// env-only values into UnmarshalKey, so tests must read from the
	// viper instance directly.
	vl := loader.(*viperLoader)
	raw := vl.v.AllSettings()["hook"]
	var cfg envOnlyConfig
	if err := mapstructure.Decode(raw, &cfg); err != nil {
		t.Fatalf("mapstructure.Decode failed: %v", err)
	}
	if cfg.Dur != "3m" {
		t.Fatalf("expected dur 3m from env, got %s", cfg.Dur)
	}
}

func TestBind_PrefixIsolation(t *testing.T) {
	t.Setenv("STRATUM_HOOK_DUR", "8m")
	t.Setenv("STRATUM_OTHER_DUR", "1m")

	loader := New()
	// bind env vars for each leaf key
	if err := loader.BindEnv("hook.dur", "STRATUM_HOOK_DUR"); err != nil {
		t.Fatalf("BindEnv hook failed: %v", err)
	}
	if err := loader.BindEnv("other.dur", "STRATUM_OTHER_DUR"); err != nil {
		t.Fatalf("BindEnv other failed: %v", err)
	}

	vl := loader.(*viperLoader)
	rawH := vl.v.AllSettings()["hook"]
	var h envOnlyConfig
	if err := mapstructure.Decode(rawH, &h); err != nil {
		t.Fatalf("mapstructure.Decode hook failed: %v", err)
	}
	rawO := vl.v.AllSettings()["other"]
	var o otherEnvConfig
	if err := mapstructure.Decode(rawO, &o); err != nil {
		t.Fatalf("mapstructure.Decode other failed: %v", err)
	}

	if h.Dur != "8m" {
		t.Fatalf("expected hook dur 8m, got %s", h.Dur)
	}
	if o.Dur != "1m" {
		t.Fatalf("expected other dur 1m, got %s", o.Dur)
	}
}
