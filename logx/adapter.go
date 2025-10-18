// Package logx adapter and helper wrappers around zap.
package logx

import (
	"context"

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

// Err wraps an error into a zap.Field so callers can use logx.Err(err)
func Err(err error) zap.Field { return zap.Error(err) }

// Convenience wrappers so consumers don't need to import zap just to create
// common fields. These mirror the zap field constructors.
func String(key, val string) zap.Field          { return zap.String(key, val) }
func Bool(key string, val bool) zap.Field       { return zap.Bool(key, val) }
func Int(key string, val int) zap.Field         { return zap.Int(key, val) }
func Int64(key string, val int64) zap.Field     { return zap.Int64(key, val) }
func Float64(key string, val float64) zap.Field { return zap.Float64(key, val) }
func Any(key string, val any) zap.Field         { return zap.Any(key, val) }
func Duration(key string, val any) zap.Field    { return zap.Any(key, val) }
