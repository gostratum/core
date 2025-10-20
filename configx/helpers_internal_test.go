package configx

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetNestedValue(t *testing.T) {
	t.Run("simple key", func(t *testing.T) {
		m := make(map[string]any)
		setNestedValue(m, []string{"key"}, "value")
		assert.Equal(t, "value", m["key"])
	})

	t.Run("nested keys", func(t *testing.T) {
		m := make(map[string]any)
		setNestedValue(m, []string{"a", "b", "c"}, "value")
		assert.Equal(t, "value", m["a"].(map[string]any)["b"].(map[string]any)["c"])
	})

	t.Run("empty keys slice", func(t *testing.T) {
		m := make(map[string]any)
		setNestedValue(m, []string{}, "value")
		assert.Empty(t, m)
	})

	t.Run("existing intermediate maps", func(t *testing.T) {
		m := map[string]any{
			"a": map[string]any{
				"b": "old",
			},
		}
		setNestedValue(m, []string{"a", "b"}, "new")
		assert.Equal(t, "new", m["a"].(map[string]any)["b"])
	})

	t.Run("non-map intermediate value", func(t *testing.T) {
		m := map[string]any{
			"a": "string-value",
		}
		// Should not panic or modify when intermediate is not a map
		setNestedValue(m, []string{"a", "b", "c"}, "value")
		assert.Equal(t, "string-value", m["a"])
	})

	t.Run("deep nesting", func(t *testing.T) {
		m := make(map[string]any)
		keys := []string{"level1", "level2", "level3", "level4", "level5"}
		setNestedValue(m, keys, "deep-value")

		current := m["level1"].(map[string]any)
		current = current["level2"].(map[string]any)
		current = current["level3"].(map[string]any)
		current = current["level4"].(map[string]any)
		assert.Equal(t, "deep-value", current["level5"])
	})
}

func TestNormalizeKey(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"lowercase", "key", "key"},
		{"uppercase", "KEY", "key"},
		{"mixed case", "MyKey", "mykey"},
		{"with spaces", "  key  ", "key"},
		{"dotted", "my.key.name", "my.key.name"},
		{"empty", "", ""},
		{"whitespace only", "   ", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeKey(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
