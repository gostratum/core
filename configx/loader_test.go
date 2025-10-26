package configx

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type emptyPrefixConfig struct{}

func (emptyPrefixConfig) Prefix() string { return "" }

type invalidConfig struct {
	Port int `mapstructure:"port" validate:"required,min=1,max=65535"`
}

func (invalidConfig) Prefix() string { return "invalid" }

type testConfig struct {
	Value string `mapstructure:"value"`
}

func (testConfig) Prefix() string { return "test" }

type multiBindConfig struct {
	Port int `mapstructure:"port" default:"8080"`
}

func (multiBindConfig) Prefix() string { return "app" }

func TestBind_ErrorCases(t *testing.T) {
	loader := New()

	t.Run("nil props", func(t *testing.T) {
		err := loader.Bind(nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "props is nil")
	})

	t.Run("empty prefix", func(t *testing.T) {
		cfg := &emptyPrefixConfig{}
		err := loader.Bind(cfg)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be empty")
	})

	t.Run("validation failure", func(t *testing.T) {
		cfg := &invalidConfig{Port: 999999}
		err := loader.Bind(cfg)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "validation failed")
	})
}

func TestBindEnv_ErrorCases(t *testing.T) {
	loader := New()

	t.Run("empty key", func(t *testing.T) {
		err := loader.BindEnv("")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot bind empty key")
	})

	t.Run("whitespace only key", func(t *testing.T) {
		err := loader.BindEnv("   ")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot bind empty key")
	})

	t.Run("valid key with aliases", func(t *testing.T) {
		err := loader.BindEnv("test.key", "TEST_ALIAS1", "TEST_ALIAS2")
		require.NoError(t, err)
	})
}

func TestLoader_Concurrency(t *testing.T) {
	loader := New()

	// Test concurrent Bind calls
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			cfg := &testConfig{}
			_ = loader.Bind(cfg)
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestLoader_MultipleBindSameConfig(t *testing.T) {
	loader := New(WithConfigPaths("./testdata"))

	// Bind multiple times should work
	cfg1 := &multiBindConfig{}
	err := loader.Bind(cfg1)
	require.NoError(t, err)

	cfg2 := &multiBindConfig{}
	err = loader.Bind(cfg2)
	require.NoError(t, err)

	assert.Equal(t, cfg1.Port, cfg2.Port)
}

func TestNew_EnvPrefixPrecedence(t *testing.T) {
	t.Run("default STRATUM prefix", func(t *testing.T) {
		t.Setenv("STRATUM_TEST_VALUE", "from_stratum")

		loader := New()
		require.NoError(t, loader.BindEnv("test.value"))

		cfg := &testConfig{}
		err := loader.Bind(cfg)
		require.NoError(t, err)
		assert.Equal(t, "from_stratum", cfg.Value)
	})

	t.Run("ENV_PREFIX overrides default", func(t *testing.T) {
		t.Setenv("ENV_PREFIX", "MYAPP")
		t.Setenv("MYAPP_TEST_VALUE", "from_myapp")
		t.Setenv("STRATUM_TEST_VALUE", "from_stratum")

		loader := New()
		require.NoError(t, loader.BindEnv("test.value"))

		cfg := &testConfig{}
		err := loader.Bind(cfg)
		require.NoError(t, err)
		assert.Equal(t, "from_myapp", cfg.Value)
	})

	t.Run("WithEnvPrefix option overrides ENV_PREFIX", func(t *testing.T) {
		t.Setenv("ENV_PREFIX", "MYAPP")
		t.Setenv("CUSTOM_TEST_VALUE", "from_custom")
		t.Setenv("MYAPP_TEST_VALUE", "from_myapp")
		t.Setenv("STRATUM_TEST_VALUE", "from_stratum")

		loader := New(WithEnvPrefix("CUSTOM"))
		require.NoError(t, loader.BindEnv("test.value"))

		cfg := &testConfig{}
		err := loader.Bind(cfg)
		require.NoError(t, err)
		assert.Equal(t, "from_custom", cfg.Value)
	})

	t.Run("WithEnvPrefix with empty ENV_PREFIX", func(t *testing.T) {
		t.Setenv("CUSTOM_TEST_VALUE", "from_custom")

		loader := New(WithEnvPrefix("CUSTOM"))
		require.NoError(t, loader.BindEnv("test.value"))

		cfg := &testConfig{}
		err := loader.Bind(cfg)
		require.NoError(t, err)
		assert.Equal(t, "from_custom", cfg.Value)
	})
}
