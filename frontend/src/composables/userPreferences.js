import { client } from './client'

export function applyUserLanguage(localeRef, fallbackLocale, language) {
  if (!localeRef) {
    return
  }
  const next = (language || '').trim()
  localeRef.value = next || fallbackLocale || 'en'
}

export async function loadAndApplyUserPreferences({ localeRef, fallbackLocale } = {}) {
  try {
    const res = await client.getUserPreferences({})
    if (localeRef && res.language) {
      applyUserLanguage(localeRef, fallbackLocale, res.language)
    }
    return res
  } catch (e) {
    console.warn('Failed to load user preferences', e)
    return null
  }
}

export function languageLabel(code) {
  if (code === 'en') {
    return 'English'
  }
  return code
}
