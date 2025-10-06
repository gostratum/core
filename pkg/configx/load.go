package configx

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

const (
	defaultBaseConfigPath  = "config/base.yaml"
	defaultLocalConfigPath = "config/local.yaml"
)

// Load reads configuration from the provided YAML files and environment variables.
// Paths fall back to config/base.yaml and config/local.yaml when empty, and the
// environment overlay defaults to the APP__ prefix.
func Load(paths []string, envPrefix string) (*Config, error) {
	if len(paths) == 0 {
		paths = []string{defaultBaseConfigPath, defaultLocalConfigPath}
	}
	if envPrefix == "" {
		envPrefix = "APP"
	}

	k := koanf.New(".")

	for _, path := range paths {
		if path == "" {
			continue
		}
		info, err := os.Stat(path)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return nil, fmt.Errorf("configx: stat %s: %w", path, err)
		}
		if info.IsDir() {
			continue
		}
		if err := k.Load(file.Provider(path), yaml.Parser()); err != nil {
			return nil, fmt.Errorf("configx: load %s: %w", path, err)
		}
	}

	envProvider := env.Provider(envPrefix+"__", "__", func(key string) string {
		trimmed := strings.TrimPrefix(key, envPrefix+"__")
		trimmed = strings.ReplaceAll(trimmed, "__", ".")
		return strings.ToLower(trimmed)
	})

	if err := k.Load(envProvider, nil); err != nil {
		return nil, fmt.Errorf("configx: load env: %w", err)
	}

	cfg := &Config{
		Server: Server{
			Addr:              ":8080",
			ReadHeaderTimeout: 5 * time.Second,
		},
	}

	if err := k.UnmarshalWithConf("", cfg, koanf.UnmarshalConf{Tag: "koanf"}); err != nil {
		return nil, fmt.Errorf("configx: unmarshal: %w", err)
	}

	if cfg.Server.Addr == "" {
		cfg.Server.Addr = ":8080"
	}
	if cfg.Server.ReadHeaderTimeout == 0 {
		cfg.Server.ReadHeaderTimeout = 5 * time.Second
	}

	return cfg, nil
}
