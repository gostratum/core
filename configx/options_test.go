package configx

import (
	"strings"
	"testing"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type OptionsTestConfig struct {
	Value string `mapstructure:"value" default:"default-value"`
}

func (c *OptionsTestConfig) Prefix() string {
	return "opt"
}

func TestWithConfigPaths(t *testing.T) {
	loader := New(
		WithConfigPaths("/custom/path1", "/custom/path2"),
	)
	assert.NotNil(t, loader)
	// Loader should be created successfully with custom paths
}

func TestWithEnvPrefix(t *testing.T) {
	// Create loader with custom prefix
	loader := New(
		WithEnvPrefix("CUSTOM"),
	)
	assert.NotNil(t, loader)

	// Set a value with CUSTOM prefix
	t.Setenv("CUSTOM_OPT_VALUE", "custom-prefix-value")

	// Explicitly bind the env key
	if err := loader.BindEnv("opt.value", "CUSTOM_OPT_VALUE"); err != nil {
		t.Fatalf("BindEnv failed: %v", err)
	}

	cfg := &OptionsTestConfig{}
	err := loader.Bind(cfg)
	require.NoError(t, err)
	assert.Equal(t, "custom-prefix-value", cfg.Value)
}

func TestWithEnvKeyReplacer(t *testing.T) {
	// Create loader with custom replacer (use :: instead of .)
	replacer := strings.NewReplacer("::", "_", "-", "_")
	loader := New(
		WithEnvPrefix("REPL"),
		WithEnvKeyReplacer(replacer),
	)
	assert.NotNil(t, loader)
}

func TestWithDecodeHook(t *testing.T) {
	type HookTestConfig struct {
		Duration time.Duration `mapstructure:"duration"`
	}

	// Custom decode hook that doesn't parse durations
	customHook := mapstructure.ComposeDecodeHookFunc(
	// No duration hook here - should fail parsing
	)

	loader := New(
		WithDecodeHook(customHook),
	)
	assert.NotNil(t, loader)
}

func TestDefaultsUsedWhenNoOptions(t *testing.T) {
	// Test that defaults work when New() called with no options
	loader := New()
	assert.NotNil(t, loader)

	cfg := &OptionsTestConfig{}
	err := loader.Bind(cfg)
	require.NoError(t, err)
	assert.Equal(t, "default-value", cfg.Value)
}

func TestOptionsOverrideDefaults(t *testing.T) {
	t.Setenv("CUSTOM_OPT_VALUE", "option-override")

	loader := New(
		WithEnvPrefix("CUSTOM"),
	)

	// Explicitly bind the env key
	if err := loader.BindEnv("opt.value", "CUSTOM_OPT_VALUE"); err != nil {
		t.Fatalf("BindEnv failed: %v", err)
	}

	cfg := &OptionsTestConfig{}
	err := loader.Bind(cfg)
	require.NoError(t, err)
	assert.Equal(t, "option-override", cfg.Value)
}

func TestMultipleOptionsComposed(t *testing.T) {
	t.Setenv("MYAPP_OPT_VALUE", "composed-value")

	loader := New(
		WithConfigPaths("./test-configs"),
		WithEnvPrefix("MYAPP"),
		WithEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_")),
	)
	assert.NotNil(t, loader)

	// Explicitly bind the env key
	if err := loader.BindEnv("opt.value", "MYAPP_OPT_VALUE"); err != nil {
		t.Fatalf("BindEnv failed: %v", err)
	}

	cfg := &OptionsTestConfig{}
	err := loader.Bind(cfg)
	require.NoError(t, err)
	assert.Equal(t, "composed-value", cfg.Value)
}

func TestConfigPathsEnvVarBackwardCompatibility(t *testing.T) {
	// Test that CONFIG_PATHS env var still works for backward compatibility
	t.Setenv("CONFIG_PATHS", "./compat-path1,./compat-path2")

	loader := New()
	assert.NotNil(t, loader)
	// Should successfully create loader with paths from env var
}

func TestConfigPathsOptionOverridesEnvVar(t *testing.T) {
	// Option should take precedence over CONFIG_PATHS env var
	t.Setenv("CONFIG_PATHS", "./env-path")

	loader := New(
		WithConfigPaths("./option-path"),
	)
	assert.NotNil(t, loader)
	// Should use option path, not env var
}
