import { initCustomTheme } from 'picocrank/vue/composables/useCustomTheme.js'
import { useTheme } from 'picocrank/vue/composables/useTheme.js'

const COLOR_SCHEME_STORAGE_KEY = 'picocrank-theme'
const CUSTOM_THEME_STORAGE_KEY = 'picocrank-custom-theme'

export function initSiteThemes() {
  const { configure, discoverThemes } = initCustomTheme({
    includeSupplementalThemes: true,
  })
  configure({ includeSupplementalThemes: true })
  return discoverThemes()
}

export function applySiteThemes(theme = {}) {
  const customTheme = typeof theme.customTheme === 'string' ? theme.customTheme : ''
  const themeMode = typeof theme.themeMode === 'string' ? theme.themeMode : 'auto'

  const { setTheme: setCustomTheme } = initCustomTheme({ includeSupplementalThemes: true })
  const { setTheme: setColorScheme } = useTheme()

  setCustomTheme(customTheme)
  setColorScheme(['auto', 'light', 'dark'].includes(themeMode) ? themeMode : 'auto')

  if (typeof localStorage !== 'undefined') {
    localStorage.removeItem(COLOR_SCHEME_STORAGE_KEY)
    localStorage.removeItem(CUSTOM_THEME_STORAGE_KEY)
  }
}
