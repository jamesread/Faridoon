package i18n

import "slices"

// AvailableLanguageCodes returns BCP 47 language tags supported by the UI.
// Faridoon is English-only today; extend when vue-i18n locales are added.
func AvailableLanguageCodes() []string {
	return []string{"en"}
}

func IsSupportedLanguage(code string) bool {
	if code == "" {
		return true
	}
	return slices.Contains(AvailableLanguageCodes(), code)
}
