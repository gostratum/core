package configx

const (
	// DefaultConfigPath is the default directory where config files are located.
	DefaultConfigPath = "./configs"

	// DefaultEnvPrefix is the default prefix for environment variables.
	DefaultEnvPrefix = "STRATUM"

	// EnvConfigPaths is the environment variable name for config paths.
	EnvConfigPaths = "CONFIG_PATHS"

	// EnvAppEnv is the environment variable name for application environment.
	EnvAppEnv = "APP_ENV"

	// EnvPrefix is the environment variable name to override the default prefix.
	// If set, this takes precedence over DefaultEnvPrefix but can be overridden by WithEnvPrefix() option.
	EnvPrefix = "ENV_PREFIX"

	// BaseConfigFile is the base configuration file name (without extension).
	BaseConfigFile = "base"
)
