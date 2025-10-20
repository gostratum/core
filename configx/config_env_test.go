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

func TestBind_PicksUpEnvWithAutomaticEnv(t *testing.T) {
	t.Setenv("STRATUM_HOOK_DUR", "3m")

	loader := New()
	// Modules must explicitly bind env keys for sensitive/env-only fields
	if err := loader.BindEnv("hook.dur", "STRATUM_HOOK_DUR"); err != nil {
		t.Fatalf("BindEnv failed: %v", err)
	}

	var cfg envOnlyConfig
	if err := loader.Bind(&cfg); err != nil {
		t.Fatalf("Bind failed: %v", err)
	}

	if cfg.Dur != "3m" {
		t.Fatalf("expected dur 3m from env (STRATUM_HOOK_DUR), got %s", cfg.Dur)
	}
}

func TestBind_PrefixIsolation(t *testing.T) {
	t.Setenv("STRATUM_HOOK_DUR", "8m")
	t.Setenv("STRATUM_OTHER_DUR", "1m")

	loader := New()
	// Explicitly bind env keys for both prefixes
	if err := loader.BindEnv("hook.dur", "STRATUM_HOOK_DUR"); err != nil {
		t.Fatalf("BindEnv hook failed: %v", err)
	}
	if err := loader.BindEnv("other.dur", "STRATUM_OTHER_DUR"); err != nil {
		t.Fatalf("BindEnv other failed: %v", err)
	}

	var h envOnlyConfig
	if err := loader.Bind(&h); err != nil {
		t.Fatalf("Bind hook failed: %v", err)
	}

	var o otherEnvConfig
	if err := loader.Bind(&o); err != nil {
		t.Fatalf("Bind other failed: %v", err)
	}

	if h.Dur != "8m" {
		t.Fatalf("expected hook dur 8m, got %s", h.Dur)
	}
	if o.Dur != "1m" {
		t.Fatalf("expected other dur 1m, got %s", o.Dur)
	}
}

// TestBind_ExplicitBindEnvWithCustomEnvVarName verifies that explicit BindEnv
// can bind custom environment variable names (in addition to standard STRATUM_ prefix)
func TestBind_ExplicitBindEnvWithCustomEnvVarName(t *testing.T) {
	t.Setenv("CUSTOM_DUR_VAR", "5m")

	loader := New()
	// Bind a custom environment variable name
	if err := loader.BindEnv("hook.dur", "CUSTOM_DUR_VAR"); err != nil {
		t.Fatalf("BindEnv failed: %v", err)
	}

	var cfg envOnlyConfig
	if err := loader.Bind(&cfg); err != nil {
		t.Fatalf("Bind failed: %v", err)
	}

	if cfg.Dur != "5m" {
		t.Fatalf("expected dur 5m from custom env var, got %s", cfg.Dur)
	}
}
