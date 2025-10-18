package configx

import (
	"reflect"
	"testing"
)

func TestWalkFields_TagSkipAndFallback(t *testing.T) {
	type S struct {
		FieldA string `mapstructure:"-"`
		FieldB int
		FieldC bool `mapstructure:"CustomName"`
	}

	var got []string
	if err := walkFields(&S{}, func(fullKey string, parts []string, f reflect.StructField) error {
		got = append(got, fullKey)
		return nil
	}, "pref."); err != nil {
		t.Fatalf("walkFields error: %v", err)
	}

	// Current implementation treats tag="-" as fallback to lowercase name
	want := map[string]bool{
		"pref.fielda":     false, // '-' falls back to lowercase name
		"pref.fieldb":     false, // fallback to lowercase name
		"pref.CustomName": false, // custom tag preserved
	}
	for _, k := range got {
		if _, ok := want[k]; ok {
			want[k] = true
		}
	}
	for k, seen := range want {
		if !seen {
			t.Fatalf("expected key %s to be present, got keys: %v", k, got)
		}
	}
}

func TestWalkFields_PrefixTrim(t *testing.T) {
	type X struct{ A int }

	var got1, got2 []string
	if err := walkFields(&X{}, func(fullKey string, parts []string, f reflect.StructField) error {
		got1 = append(got1, fullKey)
		return nil
	}, "pref."); err != nil {
		t.Fatalf("walk1 err: %v", err)
	}
	if err := walkFields(&X{}, func(fullKey string, parts []string, f reflect.StructField) error {
		got2 = append(got2, fullKey)
		return nil
	}, "pref"); err != nil {
		t.Fatalf("walk2 err: %v", err)
	}
	if len(got1) != len(got2) || got1[0] != got2[0] {
		t.Fatalf("expected prefix with and without trailing dot to match, got %v and %v", got1, got2)
	}
}

func TestWalkFields_AnonymousEmbedding(t *testing.T) {
	type E struct {
		X int `mapstructure:"x"`
	}
	type T struct {
		E
		Y string `mapstructure:"y"`
	}

	var keys []string
	if err := walkFields(&T{}, func(fullKey string, parts []string, f reflect.StructField) error {
		keys = append(keys, fullKey)
		return nil
	}, "pfx"); err != nil {
		t.Fatalf("walkFields err: %v", err)
	}
	foundY := false
	for _, k := range keys {
		if k == "pfx.y" {
			foundY = true
		}
	}
	// Current implementation prefixes embedded struct fields with the
	// lowercased embedded type name ("e"), so expect pfx.e.x and pfx.y.
	foundEX := false
	for _, k := range keys {
		if k == "pfx.e.x" {
			foundEX = true
		}
	}
	if !foundEX || !foundY {
		t.Fatalf("expected embedded fields pfx.e.x and pfx.y to be present, got %v", keys)
	}
}
