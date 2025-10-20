package configx

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

type DBConfig struct {
	Dsn  string `mapstructure:"dsn"`
	Port int    `mapstructure:"port"`
	Host string `mapstructure:"host"`
}

func (DBConfig) Prefix() string { return "db" }

func TestBindEnv_BindsAliasesAndReadsEnv(t *testing.T) {
	// Create a temporary config directory with a base.yaml file.
	dir := t.TempDir()
	baseYaml := `db:
  port: 5432
  host: localhost
`
	if err := os.WriteFile(filepath.Join(dir, "base.yaml"), []byte(baseYaml), 0o644); err != nil {
		t.Fatalf("failed to write base.yaml: %v", err)
	}

	// Create loader pointing to the temp dir
	loader := New(WithConfigPaths(dir))

	// Bind key to env var alias and set the env var.
	if err := loader.BindEnv("db.dsn", "STRATUM_DB_DSN", "DATABASE_URL"); err != nil {
		t.Fatalf("BindEnv returned error: %v", err)
	}
	// Set environment and ensure loader picks it up via binding
	t.Setenv("STRATUM_DB_DSN", "postgres://user:pass@localhost/db")

	// Use package-level DBConfig type to bind into
	var cfg = &DBConfig{}

	if err := loader.Bind(cfg); err != nil {
		t.Fatalf("Bind failed: %v", err)
	}

	if cfg.Dsn != "postgres://user:pass@localhost/db" {
		t.Fatalf("expected dsn from env binding, got %q", cfg.Dsn)
	}
	if cfg.Port != 5432 {
		t.Fatalf("expected port from base.yaml, got %d", cfg.Port)
	}
	if cfg.Host != "localhost" {
		t.Fatalf("expected host from base.yaml, got %q", cfg.Host)
	}

	// Also verify that calling BindEnv on a fresh viper (without AutomaticEnv)
	// still returns nil (no-op) and does not panic.
	v := viper.New()
	l2 := &viperLoader{v: v}
	if err := l2.BindEnv("a.b", "SOME_ENV"); err != nil {
		t.Fatalf("BindEnv on fresh viper returned error: %v", err)
	}
}

func TestSetNested_OverwriteNonMap(t *testing.T) {
	mp := map[string]any{"a": "x"}
	// Overwrite existing non-map with nested map
	setNestedValue(mp, []string{"a", "b", "c"}, 123)

	// Since intermediate value 'a' is not a map, setNestedValue should
	// leave it unchanged (no overwrite).
	if got := mp["a"]; got != "x" {
		t.Fatalf("expected mp['a'] to remain 'x', got %#v", got)
	}
}
