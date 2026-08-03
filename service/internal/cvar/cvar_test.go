package cvar

import "testing"

func TestDefaultsHaveMetadata(t *testing.T) {
	defs := Defaults("Faridoon")
	if len(defs) == 0 {
		t.Fatal("expected defaults")
	}
	seen := map[string]bool{}
	for _, def := range defs {
		assertDefMetadata(t, def, seen)
	}
	if !seen[KeySiteTitle] {
		t.Fatal("missing site_title")
	}
}

func assertDefMetadata(t *testing.T, def Def, seen map[string]bool) {
	t.Helper()
	requireNonEmpty(t, def.Key, "key")
	if seen[def.Key] {
		t.Fatalf("duplicate key %s", def.Key)
	}
	seen[def.Key] = true
	requireNonEmpty(t, def.Title, def.Key+": title")
	requireNonEmpty(t, def.Description, def.Key+": description")
	requireNonEmpty(t, def.Category, def.Key+": category")
	requireNonEmpty(t, def.MainType, def.Key+": main type")
	if def.Ordinal <= 0 {
		t.Fatalf("%s: ordinal must be positive", def.Key)
	}
}

func requireNonEmpty(t *testing.T, value, label string) {
	t.Helper()
	if value == "" {
		t.Fatalf("%s: missing", label)
	}
}
