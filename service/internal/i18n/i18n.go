package i18n

// AvailableLanguageCodes returns BCP 47 language tags supported by the UI.
// Faridoon is English-only today; extend when vue-i18n locales are added.
func AvailableLanguageCodes() []string {
	return []string{"en"}
}

func IsSupportedLanguage(code string) bool {
	if code == "" {
		return true
	}
	for _, available := range AvailableLanguageCodes() {
		if available == code {
			return true
		}
	}
	return false
}
