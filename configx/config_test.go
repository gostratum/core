package configx

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// sample config that implements Configurable
type sampleConfig struct {
	Host string `mapstructure:"host" defaults:"localhost" validate:"required"`
	Port int    `mapstructure:"port" defaults:"8080" validate:"required"`
}

func (s *sampleConfig) Prefix() string { return "test" }

func TestBind_Success(t *testing.T) {
	dir := t.TempDir()
	content := "test:\n  host: example.com\n  port: 9000\n"
	if err := os.WriteFile(filepath.Join(dir, "base.yaml"), []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write base.yaml: %v", err)
	}

	loader := New(WithConfigPaths(dir))
	var cfg sampleConfig
	if err := loader.Bind(&cfg); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if cfg.Host != "example.com" {
		t.Fatalf("expected host example.com, got %s", cfg.Host)
	}
	if cfg.Port != 9000 {
		t.Fatalf("expected port 9000, got %d", cfg.Port)
	}
}

func TestBind_ValidationFail(t *testing.T) {
	// no values set -> validation should fail for required field
	loader := New()
	var c cfg2
	if err := loader.Bind(&c); err == nil {
		t.Fatalf("expected validation error, got nil")
	}
}

// cfg2 defined at package scope for method receiver
type cfg2 struct {
	Name string `mapstructure:"name" validate:"required"`
}

func (c *cfg2) Prefix() string { return "missing" }

func TestNewConfig(t *testing.T) {
	loader := New()
	config, err := NewConfig(loader)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	// Default should be empty
	if config.EnvPrefix != "" {
		t.Fatalf("expected empty EnvPrefix, got %s", config.EnvPrefix)
	}
}

func TestNewConfig_WithEnv(t *testing.T) {
	loader := New()
	require.NoError(t, loader.BindEnv("core.config.env_prefix", "STRATUM_CORE_CONFIG_ENV_PREFIX"))
	t.Setenv("STRATUM_CORE_CONFIG_ENV_PREFIX", "TESTPREFIX")
	config, err := NewConfig(loader)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if config.EnvPrefix != "TESTPREFIX" {
		t.Fatalf("expected EnvPrefix TESTPREFIX, got %s", config.EnvPrefix)
	}
}

func TestNewConfig_WithConfigFile(t *testing.T) {
	dir := t.TempDir()
	content := "core:\n  config:\n    env_prefix: FILEPREFIX\n"
	if err := os.WriteFile(filepath.Join(dir, "base.yaml"), []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write base.yaml: %v", err)
	}

	loader := New(WithConfigPaths(dir))
	config, err := NewConfig(loader)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if config.EnvPrefix != "FILEPREFIX" {
		t.Fatalf("expected EnvPrefix FILEPREFIX, got %s", config.EnvPrefix)
	}
}
