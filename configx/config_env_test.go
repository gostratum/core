package configx

import (
	"testing"
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
	var cfg envOnlyConfig
	if err := loader.Bind(&cfg); err != nil {
		t.Fatalf("expected no error binding env-only config, got %v", err)
	}
	if cfg.Dur != "3m" {
		t.Fatalf("expected dur 3m from env, got %s", cfg.Dur)
	}
}

func TestBind_PrefixIsolation(t *testing.T) {
	t.Setenv("STRATUM_HOOK_DUR", "8m")
	t.Setenv("STRATUM_OTHER_DUR", "1m")

	loader := New()
	var h envOnlyConfig
	var o otherEnvConfig
	if err := loader.Bind(&h); err != nil {
		t.Fatalf("bind hook: %v", err)
	}
	if err := loader.Bind(&o); err != nil {
		t.Fatalf("bind other: %v", err)
	}

	if h.Dur != "8m" {
		t.Fatalf("expected hook dur 8m, got %s", h.Dur)
	}
	if o.Dur != "1m" {
		t.Fatalf("expected other dur 1m, got %s", o.Dur)
	}
}
