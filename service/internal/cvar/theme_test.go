package cvar

import "testing"

func TestIsValidThemeMode(t *testing.T) {
	for _, mode := range []string{"auto", "light", "dark", " AUTO ", "Light"} {
		if !IsValidThemeMode(mode) {
			t.Fatalf("expected valid theme mode %q", mode)
		}
	}
	if IsValidThemeMode("sepia") {
		t.Fatal("expected invalid theme mode")
	}
}

func TestNormalizeThemeMode(t *testing.T) {
	if got := NormalizeThemeMode("dark"); got != "dark" {
		t.Fatalf("got %q", got)
	}
	if got := NormalizeThemeMode("invalid"); got != DefaultThemeMode {
		t.Fatalf("got %q", got)
	}
}

func TestIsValidCustomThemeID(t *testing.T) {
	if !IsValidCustomThemeID("") {
		t.Fatal("expected empty theme id to be valid")
	}
	if !IsValidCustomThemeID("dracula-alucard") {
		t.Fatal("expected supplemental theme id to be valid")
	}
	if IsValidCustomThemeID("../evil") {
		t.Fatal("expected path-like theme id to be invalid")
	}
}
