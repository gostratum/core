package logx

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

// Test config with secrets (implements Sanitizable)
type testConfigWithSecrets struct {
	Host     string
	Port     int
	Password string
	APIKey   string
}

func (c *testConfigWithSecrets) Sanitize() any {
	safe := *c
	safe.Password = "[redacted]"
	safe.APIKey = "[redacted]"
	return &safe
}

// Test config without secrets (does NOT implement Sanitizable)
type testConfigNoSecrets struct {
	Host string
	Port int
}

func TestAny_AutoSanitization(t *testing.T) {
	// Create an observed logger to capture output
	core, logs := observer.New(zapcore.InfoLevel)
	logger := zap.New(core)

	t.Run("sanitizes config implementing Sanitizable", func(t *testing.T) {
		cfg := &testConfigWithSecrets{
			Host:     "localhost",
			Port:     5432,
			Password: "super-secret-password",
			APIKey:   "api-key-12345",
		}

		// Log using logx.Any() - should auto-sanitize
		logger.Info("Config loaded", Any("config", cfg))

		// Verify log was created
		require.Equal(t, 1, logs.Len())
		entry := logs.All()[0]

		// Verify message
		assert.Equal(t, "Config loaded", entry.Message)

		// Verify the config field was sanitized
		require.Len(t, entry.Context, 1)
		configField := entry.Context[0]
		assert.Equal(t, "config", configField.Key)

		// Extract the actual value
		sanitized, ok := configField.Interface.(*testConfigWithSecrets)
		require.True(t, ok, "expected *testConfigWithSecrets")

		// Verify secrets are redacted
		assert.Equal(t, "localhost", sanitized.Host)
		assert.Equal(t, 5432, sanitized.Port)
		assert.Equal(t, "[redacted]", sanitized.Password)
		assert.Equal(t, "[redacted]", sanitized.APIKey)
	})

	t.Run("does not sanitize config without Sanitizable", func(t *testing.T) {
		logs.TakeAll() // Clear previous logs

		cfg := &testConfigNoSecrets{
			Host: "localhost",
			Port: 8080,
		}

		// Log using logx.Any() - should pass through as-is
		logger.Info("Config loaded", Any("config", cfg))

		// Verify log was created
		require.Equal(t, 1, logs.Len())
		entry := logs.All()[0]

		// Verify the config field was NOT modified
		require.Len(t, entry.Context, 1)
		configField := entry.Context[0]

		actual, ok := configField.Interface.(*testConfigNoSecrets)
		require.True(t, ok)
		assert.Equal(t, "localhost", actual.Host)
		assert.Equal(t, 8080, actual.Port)
	})

	t.Run("sanitizes value types implementing Sanitizable", func(t *testing.T) {
		logs.TakeAll() // Clear previous logs

		// Use the value type defined at the bottom of the file
		cfg := valueConfigWithSecret{
			Public: "visible",
			Secret: "my-secret",
		}

		logger.Info("Value config", Any("cfg", cfg))

		require.Equal(t, 1, logs.Len())
		entry := logs.All()[0]
		// Verify it was sanitized
		sanitized, ok := entry.Context[0].Interface.(valueConfigWithSecret)
		require.True(t, ok)
		assert.Equal(t, "[redacted]", sanitized.Secret)
	})

	t.Run("handles nil sanitizable gracefully", func(t *testing.T) {
		logs.TakeAll() // Clear previous logs

		var cfg *testConfigWithSecrets // nil pointer

		// Should not panic
		assert.NotPanics(t, func() {
			logger.Info("Nil config", Any("config", cfg))
		})

		require.Equal(t, 1, logs.Len())
		entry := logs.All()[0]
		require.Len(t, entry.Context, 1)

		// Verify nil was logged
		assert.Nil(t, entry.Context[0].Interface)
	})
}

func TestAny_NonSanitizableTypes(t *testing.T) {
	core, logs := observer.New(zapcore.InfoLevel)
	logger := zap.New(core)

	tests := []struct {
		name  string
		key   string
		value any
	}{
		{"string", "key", "value"},
		{"int", "count", 42},
		{"bool", "enabled", true},
		{"map", "data", map[string]int{"x": 1}},
		{"struct", "point", struct{ X, Y int }{1, 2}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logs.TakeAll() // Clear logs

			logger.Info("Test", Any(tt.key, tt.value))

			require.Equal(t, 1, logs.Len())
			entry := logs.All()[0]
			require.Len(t, entry.Context, 1)

			// Verify the key is correct
			assert.Equal(t, tt.key, entry.Context[0].Key)
			// Value is logged (exact format depends on zap's encoding, so we just verify it's present)
			assert.NotNil(t, entry.Context[0])
		})
	}
}

// Benchmark to verify negligible performance impact
func BenchmarkAny_WithSanitizable(b *testing.B) {
	logger := zap.NewNop()
	cfg := &testConfigWithSecrets{
		Host:     "localhost",
		Password: "secret",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("Test", Any("config", cfg))
	}
}

func BenchmarkAny_WithoutSanitizable(b *testing.B) {
	logger := zap.NewNop()
	cfg := &testConfigNoSecrets{
		Host: "localhost",
		Port: 8080,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("Test", Any("config", cfg))
	}
}

func BenchmarkAny_Baseline(b *testing.B) {
	logger := zap.NewNop()
	value := "simple string"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("Test", Any("value", value))
	}
}

// Test that value types work with Sanitizable
type valueConfigWithSecret struct {
	Public string
	Secret string
}

func (c valueConfigWithSecret) Sanitize() any {
	safe := c
	safe.Secret = "[redacted]"
	return safe
}

func TestSanitizable_ValueReceiver(t *testing.T) {
	core, logs := observer.New(zapcore.InfoLevel)
	logger := zap.New(core)

	cfg := valueConfigWithSecret{
		Public: "visible",
		Secret: "hidden",
	}

	logger.Info("Test", Any("config", cfg))

	require.Equal(t, 1, logs.Len())
	entry := logs.All()[0]

	// Verify sanitization happened
	sanitized, ok := entry.Context[0].Interface.(valueConfigWithSecret)
	require.True(t, ok)
	assert.Equal(t, "visible", sanitized.Public)
	assert.Equal(t, "[redacted]", sanitized.Secret)
}
