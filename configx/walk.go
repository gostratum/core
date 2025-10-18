package configx

import (
	"fmt"
	"reflect"
	"strings"
	"time"
)

// walkFields traverses exported fields of a struct recursively,
// calling fn for each leaf field (non-struct, except time.Time).
func walkFields(props any, fn func(fullKey string, parts []string, f reflect.StructField) error, prefix string) error {
	val := reflect.ValueOf(props)
	if val.Kind() != reflect.Ptr {
		return fmt.Errorf("props must be pointer")
	}
	typ := val.Elem().Type()

	var walk func(t reflect.Type, path []string) error
	walk = func(t reflect.Type, path []string) error {
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			if f.PkgPath != "" { // unexported
				continue
			}
			tag := f.Tag.Get("mapstructure")
			if tag == "" || tag == "-" {
				tag = strings.ToLower(f.Name)
			}

			// dive into nested structs (except time.Time)
			if f.Type.Kind() == reflect.Struct && f.Type != reflect.TypeOf(time.Time{}) {
				if err := walk(f.Type, append(path, tag)); err != nil {
					return err
				}
				continue
			}

			parts := append(path, tag)
			fullKey := strings.Join(append([]string{strings.TrimSuffix(prefix, ".")}, parts...), ".")
			if err := fn(fullKey, parts, f); err != nil {
				return err
			}
		}
		return nil
	}
	return walk(typ, nil)
}
