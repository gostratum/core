package configx

import (
	"os"
 	"path/filepath"
 	"testing"
)

// TestBind_FileOnly verifies that values present in base.yaml are loaded when
// no environment variable is set.
func TestBind_FileOnly(t *testing.T) {
	dir := t.TempDir()
	content := "db:\n  dsn: file-dsn\n"
 	if err := os.WriteFile(filepath.Join(dir, "base.yaml"), []byte(content), 0o644); err != nil {
 		t.Fatalf("failed to write base.yaml: %v", err)
 	}

	loader := New(WithConfigPaths(dir))
	var cfg DBConfig
	if err := loader.Bind(&cfg); err != nil {
		t.Fatalf("Bind failed: %v", err)
	}

	if cfg.Dsn != "file-dsn" {
		t.Fatalf("expected dsn from file, got %q", cfg.Dsn)
	}
}

// TestBind_EnvOnly verifies that an environment variable bound via BindEnv
// will provide the value when the file does not contain the key.
func TestBind_EnvOnly(t *testing.T) {
	dir := t.TempDir()
 	// write base.yaml without dsn
 	content := "db:\n  port: 5432\n"
 	if err := os.WriteFile(filepath.Join(dir, "base.yaml"), []byte(content), 0o644); err != nil {
 		t.Fatalf("failed to write base.yaml: %v", err)
 	}

 	loader := New(WithConfigPaths(dir))
 	if err := loader.BindEnv("db.dsn", "STRATUM_DB_DSN", "DATABASE_URL"); err != nil {
 		t.Fatalf("BindEnv returned error: %v", err)
 	}
 	t.Setenv("STRATUM_DB_DSN", "env-only-dsn")

 	var cfg DBConfig
 	if err := loader.Bind(&cfg); err != nil {
 		t.Fatalf("Bind failed: %v", err)
 	}

 	if cfg.Dsn != "env-only-dsn" {
 		t.Fatalf("expected dsn from env, got %q", cfg.Dsn)
 	}
}

// TestBindEnv_AliasPrecedence verifies that when multiple env aliases are
// bound, the loader uses the expected alias precedence. We assert that the
// first alias provided takes precedence when multiple aliases are set.
func TestBindEnv_AliasPrecedence(t *testing.T) {
 	dir := t.TempDir()
 	// empty base
 	if err := os.WriteFile(filepath.Join(dir, "base.yaml"), []byte("db:\n"), 0o644); err != nil {
 		t.Fatalf("failed to write base.yaml: %v", err)
 	}

 	loader := New(WithConfigPaths(dir))
 	// bind aliases: first is STRATUM_DB_DSN, then DATABASE_URL
 	if err := loader.BindEnv("db.dsn", "STRATUM_DB_DSN", "DATABASE_URL"); err != nil {
 		t.Fatalf("BindEnv returned error: %v", err)
 	}

 	// Set both env vars to different values
 	t.Setenv("STRATUM_DB_DSN", "first-alias-dsn")
 	t.Setenv("DATABASE_URL", "second-alias-dsn")

 	var cfg DBConfig
 	if err := loader.Bind(&cfg); err != nil {
 		t.Fatalf("Bind failed: %v", err)
 	}

 	// Expect the first alias to win
 	if cfg.Dsn != "first-alias-dsn" {
 		t.Fatalf("expected dsn from first alias, got %q", cfg.Dsn)
 	}
}
