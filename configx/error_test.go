package configx

import (
	"testing"
	"time"

	"github.com/spf13/viper"
)

type badTimeConfig struct {
	WhenAny time.Time `mapstructure:"when_any"`
}

func (b *badTimeConfig) Prefix() string { return "bad" }

func TestBind_BadTimeFormat(t *testing.T) {
	v := viper.New()
	v.Set("bad.when_any", "not-a-time")
	loader := &viperLoader{v: v}
	var cfg badTimeConfig
	if err := loader.Bind(&cfg); err == nil {
		t.Fatalf("expected error decoding bad time, got nil")
	}
}
