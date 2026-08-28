package cvar

import "testing"

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

func TestIsAvailableThemeName(t *testing.T) {
	if !IsAvailableThemeName("") {
		t.Fatal("expected empty theme name to be valid")
	}
	if !IsAvailableThemeName("waffles") {
		t.Fatal("expected waffles to be available")
	}
	if IsAvailableThemeName("not-a-theme") {
		t.Fatal("expected unknown theme to be invalid")
	}
}

func TestValidateThemeControl(t *testing.T) {
	for _, value := range []string{"system", "user", " SYSTEM ", "User"} {
		if _, err := ValidateThemeControl(value); err != nil {
			t.Fatalf("expected valid theme control %q: %v", value, err)
		}
	}
	if _, err := ValidateThemeControl("forced"); err == nil {
		t.Fatal("expected invalid theme control")
	}
}

func TestNormalizeThemeControl(t *testing.T) {
	if got := NormalizeThemeControl("system"); got != ThemeControlSystem {
		t.Fatalf("got %q", got)
	}
	if got := NormalizeThemeControl("invalid"); got != ThemeControlUser {
		t.Fatalf("got %q", got)
	}
}
