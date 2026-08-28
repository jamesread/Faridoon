import { useCustomTheme } from 'picocrank/vue/composables/useCustomTheme.js'

export function themeControlFromFeatures(features = {}) {
  const raw = (features.themeControl || features.theme_control || 'user').trim()
  return raw === 'system' ? 'system' : 'user'
}

export function applyAppTheming(features = {}) {
  const { themePreference, setTheme, clearTheme } = useCustomTheme()

  const control = themeControlFromFeatures(features)
  const systemTheme = (features.themeName || features.theme_name || '').trim()

  if (control === 'system') {
    if (systemTheme) {
      setTheme(systemTheme)
    } else {
      clearTheme()
    }
    return { control, appliedTheme: systemTheme }
  }

  const stored = (themePreference.value || '').trim()
  if (stored) {
    setTheme(stored)
    return { control, appliedTheme: stored }
  }
  if (systemTheme) {
    setTheme(systemTheme)
    return { control, appliedTheme: systemTheme }
  }
  clearTheme()
  return { control, appliedTheme: '' }
}

export function themeColorSchemeSwitcherEnabledFromFeatures(features = {}) {
  return features.themeColorSchemeSwitcherEnabled === true
    || features.theme_color_scheme_switcher_enabled === true
}
