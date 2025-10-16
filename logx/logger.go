package logx

import (
	"context"
	"runtime"

	"github.com/gostratum/core/configx"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type LoggerConfig struct {
	// "dev" | "prod"
	Env string `mapstructure:"env" default:"dev" validate:"oneof=dev prod"`
	// "info","debug","warn","error"
	Level string `mapstructure:"level" default:"info"`
	// "json" | "console"
	Encoding string `mapstructure:"encoding" default:"json"`
	// true to include caller
	Caller bool `mapstructure:"caller" default:"true"`
	// true to include stack on Error+
	Stacktrace bool `mapstructure:"stacktrace" default:"false"`
	// Optional sampling in prod
	SamplingInitial    int `mapstructure:"sampling_initial" default:"100"`
	SamplingThereafter int `mapstructure:"sampling_thereafter" default:"100"`
}

// Prefix enables configx.Bind
func (LoggerConfig) Prefix() string { return "core.logger" }

func NewLoggerConfig(loader configx.Loader) (LoggerConfig, error) {
	var c LoggerConfig
	return c, loader.Bind(&c)
}

func NewLogger(lc fx.Lifecycle, c LoggerConfig) (*zap.Logger, error) {
	level := zapcore.InfoLevel
	_ = level.Set(c.Level)

	var cfg zap.Config
	if c.Env == "dev" {
		cfg = zap.NewDevelopmentConfig()
		cfg.Level = zap.NewAtomicLevelAt(level)
		cfg.Encoding = ifEmpty(c.Encoding, "console")
		// NewDevelopmentConfig already sets a human-friendly time encoder; only
		// set a custom layout if a non-empty encoding is requested that
		// necessitates overriding the default.
		if cfg.EncoderConfig.EncodeTime == nil {
			cfg.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("15:04:05.000")
		}
	} else {
		cfg = zap.NewProductionConfig()
		cfg.Level = zap.NewAtomicLevelAt(level)
		cfg.Encoding = ifEmpty(c.Encoding, "json")
		cfg.Sampling = &zap.SamplingConfig{
			Initial:    max(1, c.SamplingInitial),
			Thereafter: max(1, c.SamplingThereafter),
		}
	}

	cfg.EncoderConfig.TimeKey = "ts"
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg.EncoderConfig.StacktraceKey = "stack"
	cfg.EncoderConfig.CallerKey = "caller"

	cfg.DisableCaller = !c.Caller
	cfg.DisableStacktrace = !c.Stacktrace

	// write to stdout by default
	if len(cfg.OutputPaths) == 0 {
		cfg.OutputPaths = []string{"stdout"}
		cfg.ErrorOutputPaths = []string{"stderr"}
	}

	logger, err := cfg.Build()
	if err != nil {
		return nil, err
	}

	// Optionally set as global for libraries that use zap.L()
	zap.ReplaceGlobals(logger)

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			// Avoid noisy sync error on Windows console
			if runtime.GOOS == "windows" && isStdStream(cfg.OutputPaths) {
				return nil
			}
			_ = logger.Sync()
			return nil
		},
	})
	return logger, nil
}

func NewSugared(l *zap.Logger) *zap.SugaredLogger { return l.Sugar() }

func FxEventLogger(l *zap.Logger) fxevent.Logger { return &fxevent.ZapLogger{Logger: l} }

func Module() fx.Option {
	return fx.Module(
		"logx",
		fx.Provide(
			NewLoggerConfig,
			NewLogger,
			ProvideAdapter,
			NewSugared,
		),
		fx.WithLogger(FxEventLogger),
	)
}
func ifEmpty(s, d string) string {
	if s == "" {
		return d
	}
	return s
}
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
func isStdStream(paths []string) bool {
	for _, p := range paths {
		if p == "stdout" || p == "stderr" {
			return true
		}
	}
	return false
}
