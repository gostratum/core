// Package logx adapter and helper wrappers around zap.
package logx

import (
	"context"
	"reflect"

	"go.uber.org/zap"
)

// Field alias so consumers don't need to import zap directly.
type Field = zap.Field

// Logger is the small logging interface consumers should use.
type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	With(fields ...Field) Logger
}

// zapAdapter implements Logger by forwarding to *zap.Logger
type zapAdapter struct{ l *zap.Logger }

func (z *zapAdapter) Debug(msg string, fields ...Field) { z.l.Debug(msg, fields...) }
func (z *zapAdapter) Info(msg string, fields ...Field)  { z.l.Info(msg, fields...) }
func (z *zapAdapter) Warn(msg string, fields ...Field)  { z.l.Warn(msg, fields...) }
func (z *zapAdapter) Error(msg string, fields ...Field) { z.l.Error(msg, fields...) }
func (z *zapAdapter) With(fields ...Field) Logger       { return &zapAdapter{l: z.l.With(fields...)} }

// ProvideAdapter allows Fx consumers to get the Logger interface.
func ProvideAdapter(l *zap.Logger) Logger { return &zapAdapter{l: l} }

// NewNoopLogger returns a Logger that is a nop implementation (adapter over zap.NewNop()).
func NewNoopLogger() Logger { return ProvideAdapter(zap.NewNop()) }

// Context helpers: attach and retrieve a logger from context
type ctxKeyType struct{}

var ctxKey = ctxKeyType{}

// WithContext returns a new context with the provided logger attached.
func WithContext(ctx context.Context, l Logger) context.Context {
	return context.WithValue(ctx, ctxKey, l)
}

// FromContext extracts a Logger from context or returns a nop logger adapted from zap.NewNop().
func FromContext(ctx context.Context) Logger {
	if ctx == nil {
		return &zapAdapter{l: zap.NewNop()}
	}
	if v := ctx.Value(ctxKey); v != nil {
		if l, ok := v.(Logger); ok {
			return l
		}
	}
	return &zapAdapter{l: zap.NewNop()}
}

// Sanitizable is optionally implemented by configuration structs with sensitive fields.
// The logger will automatically sanitize types implementing this interface when
// logging with logx.Any(), preventing accidental exposure of secrets.
//
// Example:
//
//	type DBConfig struct {
//	    Host     string `mapstructure:"host"`
//	    Password string `mapstructure:"password"`
//	}
//
//	func (c *DBConfig) Sanitize() any {
//	    safe := *c
//	    safe.Password = "[redacted]"
//	    return &safe
//	}
//
// When logged, secrets are automatically redacted:
//
//	logger.Info("Config loaded", logx.Any("db", dbConfig))
//	// Output: {"db": {"host": "localhost", "password": "[redacted]"}}
type Sanitizable interface {
	Sanitize() any
}

// Err wraps an error into a zap.Field so callers can use logx.Err(err)
func Err(err error) zap.Field { return zap.Error(err) }

// Convenience wrappers so consumers don't need to import zap just to create
// common fields. These mirror the zap field constructors.
func String(key, val string) zap.Field          { return zap.String(key, val) }
func Bool(key string, val bool) zap.Field       { return zap.Bool(key, val) }
func Int(key string, val int) zap.Field         { return zap.Int(key, val) }
func Int64(key string, val int64) zap.Field     { return zap.Int64(key, val) }
func Float64(key string, val float64) zap.Field { return zap.Float64(key, val) }

// Any creates a field for arbitrary values. If the value implements Sanitizable,
// it will be automatically sanitized before logging to prevent accidental exposure
// of secrets (passwords, API keys, tokens, DSNs, etc.).
//
// This provides defense-in-depth security: developers don't need to remember to
// sanitize configs manually - it happens automatically.
//
// To bypass sanitization (e.g., for debugging), use zap.Any() directly:
//
//	import "go.uber.org/zap"
//	logger.(*zapAdapter).l.Info("Debug", zap.Any("config", rawConfig))
func Any(key string, val any) zap.Field {
	// Check for nil interface
	if val == nil {
		return zap.Any(key, nil)
	}

	// Type assert to Sanitizable
	if s, ok := val.(Sanitizable); ok {
		// Use reflection to check if it's a nil pointer
		rv := reflect.ValueOf(val)
		if rv.Kind() == reflect.Ptr && rv.IsNil() {
			return zap.Any(key, nil)
		}

		// Safe to call Sanitize()
		return zap.Any(key, s.Sanitize())
	}
	return zap.Any(key, val)
}

func Duration(key string, val any) zap.Field { return zap.Any(key, val) }
