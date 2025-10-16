package logger

import (
	"testing"

	"go.uber.org/zap"
)

func TestIfEmptyAndMax(t *testing.T) {
	if s := ifEmpty("", "def"); s != "def" {
		t.Fatalf("expected default, got %s", s)
	}
	if s := ifEmpty("val", "def"); s != "val" {
		t.Fatalf("expected val, got %s", s)
	}
	if max(1, 2) != 2 || max(5, 3) != 5 {
		t.Fatalf("max function incorrect")
	}
}

func TestIsStdStream(t *testing.T) {
	if !isStdStream([]string{"stdout"}) {
		t.Fatalf("expected stdout to be std stream")
	}
	if isStdStream([]string{"file"}) {
		t.Fatalf("expected file not to be std stream")
	}
}

func TestNewLogger_SmallConfigs(t *testing.T) {
	// Instead of invoking NewLogger (which appends hooks into an fx.Lifecycle),
	// test the small helpers and conversions that don't rely on fx lifecycle.
	// Create a nop logger and ensure NewSugared and FxEventLogger adapt it.
	nop := zap.NewNop()
	sug := NewSugared(nop)
	if sug == nil {
		t.Fatalf("expected non-nil sugared logger")
	}
	ev := FxEventLogger(nop)
	if ev == nil {
		t.Fatalf("expected non-nil fx event logger")
	}
}
