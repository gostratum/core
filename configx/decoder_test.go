package configx

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/mitchellh/mapstructure"
)

type hookConfig struct {
	Dur  time.Duration `mapstructure:"dur"`
	Tags []string      `mapstructure:"tags"`
	When time.Time     `mapstructure:"when"`
}

func (h *hookConfig) Prefix() string { return "hook" }

func TestBind_DurationSliceAndTime(t *testing.T) {
	dir := t.TempDir()
	content := "hook:\n  dur: 1m\n  tags: a,b,c\n  when: 2020-01-02T15:04:05Z\n"
	if err := os.WriteFile(filepath.Join(dir, "base.yaml"), []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write base.yaml: %v", err)
	}

	// Create loader with decode hooks
	decodeHook := mapstructure.ComposeDecodeHookFunc(
		mapstructure.StringToTimeDurationHookFunc(),
		mapstructure.StringToSliceHookFunc(","),
		strToRFC3339TimeHook,
	)
	loader := New(WithConfigPaths(dir), WithDecodeHook(decodeHook))
	var cfg hookConfig
	if err := loader.Bind(&cfg); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if cfg.Dur != time.Minute {
		t.Fatalf("expected dur 1m, got %v", cfg.Dur)
	}
	if len(cfg.Tags) != 3 || cfg.Tags[0] != "a" || cfg.Tags[2] != "c" {
		t.Fatalf("unexpected tags: %#v", cfg.Tags)
	}
	if cfg.When.Year() != 2020 || cfg.When.Month() != 1 || cfg.When.Day() != 2 {
		t.Fatalf("unexpected time: %v", cfg.When)
	}
}

func TestBind_EmptyTime(t *testing.T) {
	dir := t.TempDir()
	content := "hook:\n  when: \n"
	if err := os.WriteFile(filepath.Join(dir, "base.yaml"), []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write base.yaml: %v", err)
	}

	// Create loader with decode hooks
	decodeHook := mapstructure.ComposeDecodeHookFunc(
		mapstructure.StringToTimeDurationHookFunc(),
		mapstructure.StringToSliceHookFunc(","),
		strToRFC3339TimeHook,
	)
	loader := New(WithConfigPaths(dir), WithDecodeHook(decodeHook))
	var cfg hookConfig
	if err := loader.Bind(&cfg); err != nil {
		t.Fatalf("expected no error for empty time, got %v", err)
	}
	if !cfg.When.IsZero() {
		t.Fatalf("expected zero time, got %v", cfg.When)
	}
}
