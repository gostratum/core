package logx

import (
	"errors"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
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
	// Test the FxEventLogger adapter
	// Create a zap nop logger for adapter consumers
	z := zap.NewNop()
	ev := FxEventLogger(z)
	if ev == nil {
		t.Fatalf("expected non-nil fx event logger")
	}

	// Test sugared logger directly
	sug := z.Sugar()
	if sug == nil {
		t.Fatalf("expected non-nil sugared logger")
	}
}

func TestErrHelper(t *testing.T) {
	sample := errors.New("sample error")
	f := Err(sample)
	if f.Key != "error" {
		t.Fatalf("expected field key 'error', got %s", f.Key)
	}
	// The Interface is unexported in zap.Field, but when created by zap.Error
	// the field's Integer/Type/Interface combination should result in the
	// reflected error being stored. We can check the string form by using
	// f.String() which is stable for testing.
	if f.Type != zapcore.ErrorType {
		t.Fatalf("expected field Type to be ErrorType, got %v", f.Type)
	}
}
