package logx

import (
	"testing"

	"go.uber.org/zap"
)

func TestSanitizeMap(t *testing.T) {
	in := map[string]any{
		"username": "alice",
		"password": "s3cr3t",
		"nested": map[string]interface{}{
			"api_key": "ak-123",
			"other":   "value",
		},
	}

	out := SanitizeMap(in)
	if out["password"] != "[redacted]" {
		t.Fatalf("password not redacted")
	}
	nested, ok := out["nested"].(map[string]any)
	if !ok {
		t.Fatalf("nested map shape unexpected")
	}
	if nested["api_key"] != "[redacted]" {
		t.Fatalf("nested api_key not redacted")
	}
	if nested["other"] != "value" {
		t.Fatalf("nested other value lost")
	}
}

func TestSensitiveField(t *testing.T) {
	f := Sensitive("secret", "value")
	// Sensitive returns a zap.Field with redacted value
	if f.Key != "secret" {
		t.Fatalf("unexpected key: %s", f.Key)
	}
	// Ensure that logging with the Sensitive field does not panic
	logger := zap.NewNop()
	sugar := logger.Sugar()
	sugar.Infow("test", f)
}
