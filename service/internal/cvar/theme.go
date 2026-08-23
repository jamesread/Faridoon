package cvar

import (
	"fmt"
	"strings"
)

var validThemeModes = map[string]bool{
	"auto":  true,
	"light": true,
	"dark":  true,
}

func IsValidThemeMode(mode string) bool {
	return validThemeModes[strings.ToLower(strings.TrimSpace(mode))]
}

func NormalizeThemeMode(mode string) string {
	normalized := strings.ToLower(strings.TrimSpace(mode))
	if IsValidThemeMode(normalized) {
		return normalized
	}
	return DefaultThemeMode
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

func ValidateCustomTheme(value string) (string, error) {
	if !IsValidCustomThemeID(value) {
		return "", fmt.Errorf("invalid custom theme")
	}
	return value, nil
}

func ValidateThemeMode(value string) (string, error) {
	if !IsValidThemeMode(value) {
		return "", fmt.Errorf("theme_mode must be auto, light, or dark")
	}
	return NormalizeThemeMode(value), nil
}
