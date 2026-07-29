import { reactive, readonly } from 'vue'
import { client } from './client'

const state = reactive({
  ready: false,
  version: 'development',
  siteTitle: 'Faridoon',
  features: {
    votingEnabled: true,
    registrationEnabled: true,
    guestAddEnabled: true,
    syntaxHighlightingEnabled: false,
  },
  user: null,
  pendingApprovals: 0,
  webhookEvents: ['approval.requested'],
  error: null,
})

export async function loadInit() {
  try {
    const res = await client.init({})
    state.version = res.version || 'development'
    state.siteTitle = res.siteTitle || 'Faridoon'
    state.features = {
      votingEnabled: !!res.features?.votingEnabled,
      registrationEnabled: !!res.features?.registrationEnabled,
      guestAddEnabled: !!res.features?.guestAddEnabled,
      syntaxHighlightingEnabled: !!res.features?.syntaxHighlightingEnabled,
    }
    state.user = res.user || null
    state.pendingApprovals = res.pendingApprovals || 0
    state.webhookEvents = res.webhookEvents?.length ? [...res.webhookEvents] : ['approval.requested']
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
