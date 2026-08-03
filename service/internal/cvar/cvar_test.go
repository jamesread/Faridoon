package cvar

import "testing"

func TestDefaultsHaveMetadata(t *testing.T) {
	defs := Defaults("Faridoon")
	if len(defs) == 0 {
		t.Fatal("expected defaults")
	}
	seen := map[string]bool{}
	for _, def := range defs {
		if def.Key == "" {
			t.Fatal("empty key")
		}
		if seen[def.Key] {
			t.Fatalf("duplicate key %s", def.Key)
		}
		seen[def.Key] = true
		if def.Title == "" {
			t.Fatalf("%s: missing title", def.Key)
		}
		if def.Description == "" {
			t.Fatalf("%s: missing description", def.Key)
		}
		if def.Category == "" {
			t.Fatalf("%s: missing category", def.Key)
		}
		if def.Ordinal <= 0 {
			t.Fatalf("%s: ordinal must be positive", def.Key)
		}
		if def.MainType == "" {
			t.Fatalf("%s: missing main type", def.Key)
		}
	}
	if !seen[KeySiteTitle] {
		t.Fatal("missing site_title")
	}
}
