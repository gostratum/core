package logx

import (
	"strings"

	"go.uber.org/zap"
)

// Sensitive returns a zap.Field that represents a sensitive value. Use this when
// you need to mark a value as secret at the call site.
func Sensitive(key string, _ any) zap.Field {
	return zap.String(key, "[redacted]")
}

// SanitizeMap returns a shallow copy of the input map where keys that look like
// secrets are redacted. Keys are tested case-insensitively for substrings like
// password, secret, token, key, api_key, apikey, private, pem, hmac.
func SanitizeMap(in map[string]any) map[string]any {
	out := make(map[string]any, len(in))
	for k, v := range in {
		lk := strings.ToLower(k)
		if isSecretKey(lk) {
			out[k] = "[redacted]"
			continue
		}
		// If value itself is a map[string]any, sanitize nested maps shallowly.
		switch m := v.(type) {
		case map[string]interface{}:
			mm := make(map[string]any, len(m))
			for kk, vv := range m {
				mm[kk] = vv
			}
			out[k] = SanitizeMap(mm)
		default:
			out[k] = v
		}
	}
	return out
}

func isSecretKey(k string) bool {
	secrets := []string{"password", "passwd", "secret", "token", "key", "api_key", "apikey", "private", "pem", "hmac"}
	for _, s := range secrets {
		if strings.Contains(k, s) {
			return true
		}
	}
	return false
}
