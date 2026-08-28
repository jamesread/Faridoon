import { initCustomTheme } from 'picocrank/vue/composables/useCustomTheme.js'

export function initSiteThemes() {
  const { configure, discoverThemes } = initCustomTheme({
    storageKey: 'faridoon-custom-theme',
    includeSupplementalThemes: true,
  })
  configure({ includeSupplementalThemes: true })
  return discoverThemes()
}
