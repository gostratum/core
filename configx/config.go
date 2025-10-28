package configx

// Loader loads configuration into structs with validation.
type Loader interface {
	// Bind loads configuration into a struct implementing Configurable.
	// Configuration precedence: ENV > YAML > Defaults
	Bind(Configurable) error

	// BindEnv explicitly binds a key to environment variables.
	// Use for sensitive values that should only come from environment.
	BindEnv(key string, envVars ...string) error
}

// Configurable must be implemented by configuration structs.
// The Prefix() method returns the configuration key prefix.
//
// Example:
//
//	type DBConfig struct {
//	    Host string `mapstructure:"host" default:"localhost"`
//	    Port int    `mapstructure:"port" default:"5432"`
//	}
//
//	func (DBConfig) Prefix() string { return "db" }
type Configurable interface {
	Prefix() string
}

type Config struct {
	EnvPrefix string `mapstructure:"env_prefix"`
}

func (Config) Prefix() string {
	return "core.config"
}

func NewConfig(loader Loader) (Config, error) {
	var c Config
	return c, loader.Bind(&c)
}
