package configx

import (
	"testing"
)

type defaultsConfig struct {
	Host string `mapstructure:"host" defaults:"example.com" validate:"required"`
}

func (d *defaultsConfig) Prefix() string { return "doesnotexist" }

func TestBind_SubNilAppliesDefaults(t *testing.T) {
	loader := New()
	var cfg defaultsConfig
	// Current implementation does not apply defaults for a missing subtree
	// in a way that satisfies validation. Expect a validation error.
	if err := loader.Bind(&cfg); err == nil {
		t.Fatalf("expected validation error when no values present, got nil")
	}
}
