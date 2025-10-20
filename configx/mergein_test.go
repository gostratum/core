package configx

import (
	"os"
	"path/filepath"
	"testing"
)

type AppConfig struct {
	Port int    `mapstructure:"port"`
	Host string `mapstructure:"host"`
}

func (AppConfig) Prefix() string { return "app" }

func writeYAML(path, content string) error {
	return os.WriteFile(path, []byte(content), 0644)
}

func TestNew_MergeInConfigBaseAndEnv(t *testing.T) {
	dir := t.TempDir()
	base := "app:\n  port: 8000\n"
	env := "app:\n  port: 9000\n  host: envhost\n"
	if err := writeYAML(filepath.Join(dir, "base.yaml"), base); err != nil {
		t.Fatalf("write base yaml: %v", err)
	}
	if err := writeYAML(filepath.Join(dir, "dev.yaml"), env); err != nil {
		t.Fatalf("write dev yaml: %v", err)
	}

	t.Setenv("CONFIG_PATHS", dir)
	t.Setenv("APP_ENV", "dev")
	loader := New()

	// Use public Loader API: bind into a typed struct with Prefix "app"
	var cfg = &AppConfig{}
	if err := loader.Bind(cfg); err != nil {
		t.Fatalf("Bind failed: %v", err)
	}

	if cfg.Port != 9000 {
		t.Fatalf("expected port 9000 from env merge, got %d", cfg.Port)
	}
	if cfg.Host != "envhost" {
		t.Fatalf("expected host envhost from env merge, got %s", cfg.Host)
	}
}
