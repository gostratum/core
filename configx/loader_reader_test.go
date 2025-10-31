package configx

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type readerTestConfig struct {
	Port    int           `mapstructure:"port" default:"3000"`
	Host    string        `mapstructure:"host" default:"localhost"`
	Timeout time.Duration `mapstructure:"timeout"`
	Debug   bool          `mapstructure:"debug"`
}

func (readerTestConfig) Prefix() string { return "server" }

type validatedReaderConfig struct {
	Port int `mapstructure:"port" validate:"required,min=1,max=65535"`
}

func (validatedReaderConfig) Prefix() string { return "app" }

type nestedReaderConfig struct {
	Database struct {
		Host     string `mapstructure:"host"`
		Port     int    `mapstructure:"port"`
		Username string `mapstructure:"username"`
	} `mapstructure:"database"`
	Cache struct {
		TTL time.Duration `mapstructure:"ttl"`
	} `mapstructure:"cache"`
}

func (nestedReaderConfig) Prefix() string { return "service" }

func TestNewWithReader_BasicYAML(t *testing.T) {
	yaml := `
server:
  port: 8080
  host: example.com
  timeout: 30s
  debug: true
`
	loader, err := NewWithReader(strings.NewReader(yaml))
	require.NoError(t, err)
	require.NotNil(t, loader)

	var cfg readerTestConfig
	err = loader.Bind(&cfg)
	require.NoError(t, err)

	assert.Equal(t, 8080, cfg.Port)
	assert.Equal(t, "example.com", cfg.Host)
	assert.Equal(t, 30*time.Second, cfg.Timeout)
	assert.True(t, cfg.Debug)
}

func TestNewWithReader_Defaults(t *testing.T) {
	yaml := `
server:
  timeout: 5s
`
	loader, err := NewWithReader(strings.NewReader(yaml))
	require.NoError(t, err)

	var cfg readerTestConfig
	err = loader.Bind(&cfg)
	require.NoError(t, err)

	// YAML values
	assert.Equal(t, 5*time.Second, cfg.Timeout)

	// Default tag values (not in YAML)
	assert.Equal(t, 3000, cfg.Port)
	assert.Equal(t, "localhost", cfg.Host)
	assert.False(t, cfg.Debug) // bool zero value
}

func TestNewWithReader_NestedConfig(t *testing.T) {
	yaml := `
service:
  database:
    host: db.example.com
    port: 5432
    username: dbuser
  cache:
    ttl: 1h
`
	loader, err := NewWithReader(strings.NewReader(yaml))
	require.NoError(t, err)

	var cfg nestedReaderConfig
	err = loader.Bind(&cfg)
	require.NoError(t, err)

	assert.Equal(t, "db.example.com", cfg.Database.Host)
	assert.Equal(t, 5432, cfg.Database.Port)
	assert.Equal(t, "dbuser", cfg.Database.Username)
	assert.Equal(t, 1*time.Hour, cfg.Cache.TTL)
}

func TestNewWithReader_EnvironmentOverride(t *testing.T) {
	yaml := `
server:
  port: 8080
  host: yaml.example.com
`
	// Set environment variable (should override YAML)
	t.Setenv("STRATUM_SERVER_HOST", "env.example.com")
	t.Setenv("STRATUM_SERVER_PORT", "9090")

	loader, err := NewWithReader(strings.NewReader(yaml))
	require.NoError(t, err)

	var cfg readerTestConfig
	err = loader.Bind(&cfg)
	require.NoError(t, err)

	// Environment variables should override YAML
	assert.Equal(t, 9090, cfg.Port)
	assert.Equal(t, "env.example.com", cfg.Host)
}

func TestNewWithReader_WithEnvPrefix(t *testing.T) {
	yaml := `
server:
  port: 8080
`
	// Set env var with custom prefix
	t.Setenv("MYAPP_SERVER_PORT", "7070")

	loader, err := NewWithReader(strings.NewReader(yaml), WithEnvPrefix("MYAPP"))
	require.NoError(t, err)

	var cfg readerTestConfig
	err = loader.Bind(&cfg)
	require.NoError(t, err)

	// Custom prefix env var should override YAML
	assert.Equal(t, 7070, cfg.Port)
}

func TestNewWithReader_Validation(t *testing.T) {
	t.Run("valid port", func(t *testing.T) {
		yaml := `
app:
  port: 8080
`
		loader, err := NewWithReader(strings.NewReader(yaml))
		require.NoError(t, err)

		var cfg validatedReaderConfig
		err = loader.Bind(&cfg)
		require.NoError(t, err)
		assert.Equal(t, 8080, cfg.Port)
	})

	t.Run("invalid port - too high", func(t *testing.T) {
		yaml := `
app:
  port: 99999
`
		loader, err := NewWithReader(strings.NewReader(yaml))
		require.NoError(t, err)

		var cfg validatedReaderConfig
		err = loader.Bind(&cfg)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "validation failed")
	})

	t.Run("invalid port - missing required", func(t *testing.T) {
		yaml := `
app:
  other: value
`
		loader, err := NewWithReader(strings.NewReader(yaml))
		require.NoError(t, err)

		var cfg validatedReaderConfig
		err = loader.Bind(&cfg)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "validation failed")
	})
}

func TestNewWithReader_EmptyYAML(t *testing.T) {
	yaml := ``
	loader, err := NewWithReader(strings.NewReader(yaml))
	require.NoError(t, err)

	var cfg readerTestConfig
	err = loader.Bind(&cfg)
	require.NoError(t, err)

	// Should use default values
	assert.Equal(t, 3000, cfg.Port)
	assert.Equal(t, "localhost", cfg.Host)
}

func TestNewWithReader_MalformedYAML(t *testing.T) {
	yaml := `
server:
  port: not-a-number
  invalid yaml structure
`
	_, err := NewWithReader(strings.NewReader(yaml))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read config from reader")
}

func TestNewWithReader_DurationDecodeHook(t *testing.T) {
	yaml := `
server:
  timeout: 2h30m15s
`
	loader, err := NewWithReader(strings.NewReader(yaml))
	require.NoError(t, err)

	var cfg readerTestConfig
	err = loader.Bind(&cfg)
	require.NoError(t, err)

	expected := 2*time.Hour + 30*time.Minute + 15*time.Second
	assert.Equal(t, expected, cfg.Timeout)
}

func TestNewWithReader_BindEnv(t *testing.T) {
	yaml := `
server:
  port: 8080
`
	t.Setenv("CUSTOM_PORT", "9999")

	loader, err := NewWithReader(strings.NewReader(yaml))
	require.NoError(t, err)

	// Explicitly bind custom env var
	err = loader.BindEnv("server.port", "CUSTOM_PORT")
	require.NoError(t, err)

	var cfg readerTestConfig
	err = loader.Bind(&cfg)
	require.NoError(t, err)

	// Should use explicitly bound env var
	assert.Equal(t, 9999, cfg.Port)
}

func TestNewWithReader_MultipleBind(t *testing.T) {
	yaml := `
server:
  port: 8080
  host: example.com
`
	loader, err := NewWithReader(strings.NewReader(yaml))
	require.NoError(t, err)

	// Bind multiple times should work
	var cfg1 readerTestConfig
	err = loader.Bind(&cfg1)
	require.NoError(t, err)

	var cfg2 readerTestConfig
	err = loader.Bind(&cfg2)
	require.NoError(t, err)

	assert.Equal(t, cfg1.Port, cfg2.Port)
	assert.Equal(t, cfg1.Host, cfg2.Host)
}

func TestNewWithReader_CaseInsensitiveKeys(t *testing.T) {
	yaml := `
SERVER:
  PORT: 8080
  HOST: example.com
`
	loader, err := NewWithReader(strings.NewReader(yaml))
	require.NoError(t, err)

	var cfg readerTestConfig
	err = loader.Bind(&cfg)
	require.NoError(t, err)

	assert.Equal(t, 8080, cfg.Port)
	assert.Equal(t, "example.com", cfg.Host)
}

func TestNewWithReader_WithOptions(t *testing.T) {
	t.Run("custom env prefix", func(t *testing.T) {
		yaml := `
server:
  port: 8080
`
		t.Setenv("CUSTOM_SERVER_PORT", "7777")

		loader, err := NewWithReader(
			strings.NewReader(yaml),
			WithEnvPrefix("CUSTOM"),
		)
		require.NoError(t, err)

		var cfg readerTestConfig
		err = loader.Bind(&cfg)
		require.NoError(t, err)

		assert.Equal(t, 7777, cfg.Port)
	})

	t.Run("multiple options", func(t *testing.T) {
		yaml := `
server:
  port: 8080
`
		t.Setenv("APP_SERVER_PORT", "6666")

		loader, err := NewWithReader(
			strings.NewReader(yaml),
			WithEnvPrefix("APP"),
			WithConfigPaths("./test"), // Ignored for reader-based loader
		)
		require.NoError(t, err)

		var cfg readerTestConfig
		err = loader.Bind(&cfg)
		require.NoError(t, err)

		assert.Equal(t, 6666, cfg.Port)
	})
}

// Benchmark tests
func BenchmarkNewWithReader(b *testing.B) {
	yaml := `
server:
  port: 8080
  host: example.com
  timeout: 30s
  debug: true
`

	for b.Loop() {
		loader, err := NewWithReader(strings.NewReader(yaml))
		if err != nil {
			b.Fatal(err)
		}

		var cfg readerTestConfig
		if err := loader.Bind(&cfg); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkNewWithReader_Nested(b *testing.B) {
	yaml := `
service:
  database:
    host: db.example.com
    port: 5432
    username: dbuser
  cache:
    ttl: 1h
`

	for b.Loop() {
		loader, err := NewWithReader(strings.NewReader(yaml))
		if err != nil {
			b.Fatal(err)
		}

		var cfg nestedReaderConfig
		if err := loader.Bind(&cfg); err != nil {
			b.Fatal(err)
		}
	}
}
