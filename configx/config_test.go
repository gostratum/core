package configx

import (
	"testing"

	"github.com/spf13/viper"
)

// sample config that implements Configurable
type sampleConfig struct {
	Host string `mapstructure:"host" defaults:"localhost" validate:"required"`
	Port int    `mapstructure:"port" defaults:"8080" validate:"required"`
}

func (s *sampleConfig) Prefix() string { return "test" }

func TestBind_Success(t *testing.T) {
	v := viper.New()
	v.Set("test.host", "example.com")
	v.Set("test.port", 9000)

	loader := &viperLoader{v: v}
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
	v := viper.New()
	// no values set -> validation should fail for required field
	loader := &viperLoader{v: v}
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
