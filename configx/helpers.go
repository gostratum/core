package configx

import (
	"reflect"
	"time"
)

func setNested(mp map[string]any, parts []string, val any) {
	if len(parts) == 0 {
		return
	}
	if len(parts) == 1 {
		mp[parts[0]] = val
		return
	}
	head := parts[0]
	rest := parts[1:]
	child, ok := mp[head]
	if !ok {
		child = map[string]any{}
		mp[head] = child
	}
	cmap, ok := child.(map[string]any)
	if !ok {
		cmap = map[string]any{}
		mp[head] = cmap
	}
	setNested(cmap, rest, val)
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
