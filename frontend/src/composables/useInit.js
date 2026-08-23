import { reactive, readonly } from 'vue'
import { client } from './client'
import { applySiteThemes } from './useSiteTheme.js'

const defaultTheme = {
  themeMode: 'auto',
  customTheme: '',
}

const state = reactive({
  ready: false,
  version: 'development',
  siteTitle: 'Faridoon',
  theme: { ...defaultTheme },
  features: {
    votingEnabled: false,
    registrationEnabled: true,
    guestAddEnabled: true,
    syntaxHighlightingEnabled: false,
    markdownEnabled: false,
    showPwaPrompt: false,
  },
  user: null,
  pendingApprovals: 0,
  webhookEvents: ['approval.requested'],
  headerLinks: [],
  error: null,
})

export async function loadInit() {
  try {
    const res = await client.init({})
    state.version = res.version || 'development'
    state.siteTitle = res.siteTitle || 'Faridoon'
    state.theme = {
      themeMode: res.theme?.themeMode || defaultTheme.themeMode,
      customTheme: res.theme?.customTheme || defaultTheme.customTheme,
    }
    applySiteThemes(state.theme)
    state.features = {
      votingEnabled: !!res.features?.votingEnabled,
      registrationEnabled: !!res.features?.registrationEnabled,
      guestAddEnabled: !!res.features?.guestAddEnabled,
      syntaxHighlightingEnabled: !!res.features?.syntaxHighlightingEnabled,
      markdownEnabled: !!res.features?.markdownEnabled,
      showPwaPrompt: !!res.features?.showPwaPrompt,
    }
    state.user = res.user || null
    state.pendingApprovals = res.pendingApprovals || 0
    state.webhookEvents = res.webhookEvents?.length ? [...res.webhookEvents] : ['approval.requested']
    state.headerLinks = res.headerLinks?.length ? res.headerLinks.map((l) => ({
      id: l.id,
      title: l.title,
      url: l.url,
      sortOrder: l.sortOrder || 0,
      openInNewTab: !!l.openInNewTab,
    })) : []
    state.error = null
  } catch (e) {
    state.error = e.message || String(e)
  } finally {
    state.ready = true
  }
}

export function setUser(user) {
  state.user = user
}

export function useInit() {
  return readonly(state)
}

export { state as initState }
