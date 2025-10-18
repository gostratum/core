package configx

import (
	"testing"

	"github.com/spf13/viper"
)

func TestBindEnv_BindsAliasesAndReadsEnv(t *testing.T) {
	// Use New() to ensure AutomaticEnv and replacer are configured.
	loader := New()
	vl := loader.(*viperLoader)

	// Bind key to env var alias and set the env var.
	if err := loader.BindEnv("db.dsn", "STRATUM_DB_DSN", "DATABASE_URL"); err != nil {
		t.Fatalf("BindEnv returned error: %v", err)
	}
	// Set environment and ensure viper picks it up via binding
	t.Setenv("STRATUM_DB_DSN", "postgres://user:pass@localhost/db")

	if got := vl.v.GetString("db.dsn"); got != "postgres://user:pass@localhost/db" {
		t.Fatalf("expected db.dsn from env, got %q", got)
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
	setNested(mp, []string{"a", "b", "c"}, 123)

	a, ok := mp["a"].(map[string]any)
	if !ok {
		t.Fatalf("expected mp['a'] to be map after overwrite, got %T", mp["a"])
	}
	b, ok := a["b"].(map[string]any)
	if !ok {
		t.Fatalf("expected nested map at ['a']['b'], got %T", a["b"])
	}
	if got := b["c"]; got != 123 {
		t.Fatalf("expected value 123 at ['a']['b']['c'], got %#v", got)
	}
}

