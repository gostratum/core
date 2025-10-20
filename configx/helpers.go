package configx

import (
	"reflect"
	"strings"
	"time"
)

// setNestedValue sets a value in a nested map structure, creating intermediate maps as needed.
// For example, setNestedValue(m, []string{"a", "b", "c"}, value) sets m["a"]["b"]["c"] = value
func setNestedValue(m map[string]any, keys []string, value any) {
	if len(keys) == 0 {
		return
	}

	// Navigate/create nested maps for all but the last key
	current := m
	for _, key := range keys[:len(keys)-1] {
		if _, exists := current[key]; !exists {
			current[key] = make(map[string]any)
		}
		// Type assert and navigate deeper
		if next, ok := current[key].(map[string]any); ok {
			current = next
		} else {
			// If the intermediate value is not a map, we can't go deeper
			// This shouldn't happen with properly formed configs
			return
		}
	}

	// Set the final value
	current[keys[len(keys)-1]] = value
}

func strToRFC3339TimeHook(from, to reflect.Type, data any) (any, error) {
	if from.Kind() == reflect.String && to == reflect.TypeOf(time.Time{}) {
		s := data.(string)
		if s == "" {
			return time.Time{}, nil
		}
		return time.Parse(time.RFC3339, s)
	}
	return data, nil
}

// normalizeKey converts a key to lowercase for case-insensitive handling.
// This ensures consistent key handling across config files and env vars.
func normalizeKey(key string) string {
	return strings.ToLower(strings.TrimSpace(key))
}
