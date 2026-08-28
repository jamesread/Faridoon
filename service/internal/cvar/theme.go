package cvar

import (
	"fmt"
	"strings"
)

const (
	KeyThemeColorSchemeSwitcherEnabled = "theme_color_scheme_switcher_enabled"
	KeyThemeName                       = "theme_name"
	KeyThemeControl                    = "theme_control"

	ThemeControlSystem = "system"
	ThemeControlUser   = "user"

	CategoryTheme = "Theme"
)

var availableThemeNames = []string{
	"catppuccin-latte-frappe",
	"dracula-alucard",
	"gruvbox-dark-light",
	"waffles",
}

func AvailableThemeNames() []string {
	out := make([]string, len(availableThemeNames))
	copy(out, availableThemeNames)
	return out
}

func IsAvailableThemeName(name string) bool {
	if name == "" {
		return true
	}
	for _, n := range availableThemeNames {
		if n == name {
			return true
		}
	}
	return false
}

func IsValidCustomThemeID(id string) bool {
	if id == "" {
		return true
	}
	return len(id) <= 64 && !strings.ContainsFunc(id, func(r rune) bool {
		return !isCustomThemeChar(r)
	})
}

func isCustomThemeChar(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-'
}

func ValidateThemeName(value string) (string, error) {
	value = strings.TrimSpace(value)
	if !IsValidCustomThemeID(value) {
		return "", fmt.Errorf("invalid theme name")
	}
	if !IsAvailableThemeName(value) {
		return "", fmt.Errorf("unknown theme name")
	}
	return value, nil
}

func ValidateThemeControl(value string) (string, error) {
	value = strings.ToLower(strings.TrimSpace(value))
	switch value {
	case ThemeControlSystem, ThemeControlUser:
		return value, nil
	default:
		return "", fmt.Errorf("theme_control must be system or user")
	}
}

func NormalizeThemeControl(value string) string {
	normalized, err := ValidateThemeControl(value)
	if err != nil {
		return ThemeControlUser
	}
	return normalized
}
