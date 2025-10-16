package logx_test

import (
	"context"
	"testing"

	"github.com/gostratum/core/logx"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestAdapterForwardsAndWith(t *testing.T) {
	core, observed := observer.New(zapcore.InfoLevel)
	l := zap.New(core)
	adapter := logx.ProvideAdapter(l)

	adapter.Info("hello", logx.String("k", "v"))

	if observed.Len() != 1 {
		t.Fatalf("expected 1 log entry, got %d", observed.Len())
	}
	e := observed.All()[0]
	if e.Message != "hello" || e.Level != zapcore.InfoLevel {
		t.Fatalf("unexpected entry: %#v", e)
	}

	child := adapter.With(logx.String("component", "c"))
	child.Debug("x") // debug dropped by observer level
	child.Info("y")
	if observed.Len() != 2 {
		t.Fatalf("expected 2 log entries, got %d", observed.Len())
	}
}

func TestContextHelpers(t *testing.T) {
	core, observed := observer.New(zapcore.InfoLevel)
	l := zap.New(core)
	adapter := logx.ProvideAdapter(l)

	ctx := logx.WithContext(context.Background(), adapter)
	got := logx.FromContext(ctx)
	got.Info("fromctx", logx.String("k", "v"))

	if observed.Len() != 1 {
		t.Fatalf("expected 1 log entry, got %d", observed.Len())
	}
}
