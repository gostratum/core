package configx

import "time"

type Server struct {
	Addr              string        `koanf:"addr"`
	ReadHeaderTimeout time.Duration `koanf:"read_header_timeout"`
}

type Security struct {
	EnableCORS     bool     `koanf:"enable_cors"`
	AllowedOrigins []string `koanf:"allowed_origins"`
}

type Observability struct {
	JSONLogs bool `koanf:"json_logs"`
}

type Config struct {
	Server        Server        `koanf:"server"`
	Security      Security      `koanf:"security"`
	Observability Observability `koanf:"observability"`
}
