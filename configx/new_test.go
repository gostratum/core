package configx

import (
	"os"
	"path/filepath"
	"testing"
)

type envHookConfig struct {
	Dur string `mapstructure:"dur"`
}

func (e *envHookConfig) Prefix() string { return "hook" }

func TestNew_PicksUpEnvOverrides(t *testing.T) {
	dir := t.TempDir()
	content := "hook:\n  dur: 2m\n"
	if err := os.WriteFile(filepath.Join(dir, "base.yaml"), []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write base.yaml: %v", err)
	}

	loader := New(WithConfigPaths(dir))
	var cfg envHookConfig
	if err := loader.Bind(&cfg); err != nil {
		t.Fatalf("expected bind to succeed with file value, got %v", err)
	}
	if cfg.Dur != "2m" {
		t.Fatalf("expected dur 2m from file, got %s", cfg.Dur)
	}
}
