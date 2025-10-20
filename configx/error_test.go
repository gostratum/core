package configx

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

type badTimeConfig struct {
	WhenAny time.Time `mapstructure:"when_any"`
}

func (b *badTimeConfig) Prefix() string { return "bad" }

func TestBind_BadTimeFormat(t *testing.T) {
	dir := t.TempDir()
	content := "bad:\n  when_any: not-a-time\n"
	if err := os.WriteFile(filepath.Join(dir, "base.yaml"), []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write base.yaml: %v", err)
	}
	loader := New(WithConfigPaths(dir))
	var cfg badTimeConfig
	if err := loader.Bind(&cfg); err == nil {
		t.Fatalf("expected error decoding bad time, got nil")
	}
}
