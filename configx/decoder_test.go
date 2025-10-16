package configx

import (
	"testing"
	"time"

	"github.com/spf13/viper"
)

type hookConfig struct {
	Dur  time.Duration `mapstructure:"dur"`
	Tags []string      `mapstructure:"tags"`
	When time.Time     `mapstructure:"when"`
}

func (h *hookConfig) Prefix() string { return "hook" }

func TestBind_DurationSliceAndTime(t *testing.T) {
	v := viper.New()
	v.Set("hook.dur", "1m")
	v.Set("hook.tags", "a,b,c")
	v.Set("hook.when", "2020-01-02T15:04:05Z")

	loader := &viperLoader{v: v}
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
	v := viper.New()
	v.Set("hook.when", "")
	loader := &viperLoader{v: v}
	var cfg hookConfig
	if err := loader.Bind(&cfg); err != nil {
		t.Fatalf("expected no error for empty time, got %v", err)
	}
	if !cfg.When.IsZero() {
		t.Fatalf("expected zero time, got %v", cfg.When)
	}
}
