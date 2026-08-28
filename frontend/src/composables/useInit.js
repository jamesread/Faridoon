import { reactive, readonly } from 'vue'
import { client } from './client'
import { applyAppTheming } from './applyAppTheming.js'
import { loadAndApplyUserPreferences } from './userPreferences.js'

const state = reactive({
  ready: false,
  version: 'development',
  siteTitle: 'Faridoon',
  features: {
    votingEnabled: false,
    registrationEnabled: true,
    guestAddEnabled: true,
    syntaxHighlightingEnabled: false,
    markdownEnabled: false,
    showPwaPrompt: false,
    themeColorSchemeSwitcherEnabled: false,
    themeName: '',
    themeControl: 'user',
    availableThemes: [],
  },
  user: null,
  pendingApprovals: 0,
  webhookEvents: ['approval.requested'],
  headerLinks: [],
  error: null,
})

function syncFeatures(res) {
  state.features = {
    votingEnabled: !!res.features?.votingEnabled,
    registrationEnabled: !!res.features?.registrationEnabled,
    guestAddEnabled: !!res.features?.guestAddEnabled,
    syntaxHighlightingEnabled: !!res.features?.syntaxHighlightingEnabled,
    markdownEnabled: !!res.features?.markdownEnabled,
    showPwaPrompt: !!res.features?.showPwaPrompt,
    themeColorSchemeSwitcherEnabled: !!res.features?.themeColorSchemeSwitcherEnabled,
    themeName: res.features?.themeName || '',
    themeControl: res.features?.themeControl || 'user',
    availableThemes: res.features?.availableThemes?.length ? [...res.features.availableThemes] : [],
  }
}

export async function loadInit() {
  try {
    const res = await client.init({})
    state.version = res.version || 'development'
    state.siteTitle = res.siteTitle || 'Faridoon'
    syncFeatures(res)
    applyAppTheming(state.features)
    state.user = res.user || null
    state.pendingApprovals = res.pendingApprovals || 0
    if (state.user) {
      await loadAndApplyUserPreferences()
    }
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
