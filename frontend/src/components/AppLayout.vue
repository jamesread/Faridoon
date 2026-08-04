<script setup>
import Header from 'picocrank/vue/components/Header.vue'
import QuickSearch from 'picocrank/vue/components/QuickSearch.vue'
import { computed, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import {
  Clock01Icon,
  ShuffleIcon,
  ChampionIcon,
  PlusSignIcon,
  Tick02Icon,
  Link01Icon,
  QuoteDownIcon,
} from '@hugeicons/core-free-icons'
import { initState, loadInit } from '../composables/useInit'
import { client } from '../composables/client'

const router = useRouter()
const siteTitle = computed(() => initState.siteTitle || 'Faridoon')
const appVersion = computed(() => initState.version || 'development')
const auth = computed(() => initState.user)
const features = computed(() => initState.features)
const pendingApprovals = computed(() => initState.pendingApprovals || 0)
const headerLinks = computed(() => initState.headerLinks || [])

watch(
  siteTitle,
  (title) => {
    document.title = title
  },
  { immediate: true },
)

const navigationLinks = ref([])

function pushQuoteOrder(order) {
  router.push({ name: 'quotes', query: { order } })
}

function openCustomLink(link) {
  if (link.url.startsWith('/')) {
    if (link.openInNewTab) {
      window.open(link.url, '_blank', 'noopener,noreferrer')
      return
    }
    router.push(link.url)
    return
  }
  if (link.openInNewTab) {
    window.open(link.url, '_blank', 'noopener,noreferrer')
    return
  }
  window.location.assign(link.url)
}

function appendFeatureLinks(links) {
  if (features.value.votingEnabled) {
    links.push({
      name: 'rank',
      title: 'Highest voted',
      type: 'callback',
      icon: ChampionIcon,
      callback: () => pushQuoteOrder('rank'),
    })
  }
  if (auth.value?.canApproveQuotes) {
    links.push({
      name: 'approvals',
      title: `Approvals (${pendingApprovals.value})`,
      type: 'callback',
      icon: Tick02Icon,
      callback: () => router.push({ name: 'approvals' }),
    })
  }
  if (features.value.guestAddEnabled || auth.value) {
    links.push({
      name: 'add',
      title: 'Add',
      type: 'callback',
      icon: PlusSignIcon,
      callback: () => router.push({ name: 'quote-create' }),
    })
  }
}

function appendCustomLinks(links) {
  if (!headerLinks.value.length) {
    return
  }
  links.push({ name: 'custom-sep', type: 'separator' })
  for (const link of headerLinks.value) {
    links.push({
      name: `custom-${link.id}`,
      title: link.title,
      type: 'callback',
      icon: Link01Icon,
      callback: () => openCustomLink(link),
    })
  }
}

function rebuildNavigation() {
  const links = [
    {
      name: 'latest',
      title: 'Latest',
      type: 'callback',
      icon: Clock01Icon,
      callback: () => pushQuoteOrder('latest'),
    },
    {
      name: 'random',
      title: 'Random',
      type: 'callback',
      icon: ShuffleIcon,
      callback: () => pushQuoteOrder('random'),
    },
  ]
  appendFeatureLinks(links)
  appendCustomLinks(links)
  navigationLinks.value = links
}

watch([auth, features, pendingApprovals, headerLinks], rebuildNavigation, { immediate: true, deep: true })

const navigation = computed(() => ({
  navigationLinks: navigationLinks.value,
  isActive: () => false,
}))

function goHome() {
  router.push({ name: 'quotes' })
}

function truncateQuote(content, max = 120) {
  const text = String(content || '').replace(/\s+/g, ' ').trim()
  if (text.length <= max) {
    return text
  }
  return `${text.slice(0, max)}…`
}

async function fetchQuoteResults(query, { signal } = {}) {
  const q = String(query || '').trim()
  if (!q) {
    return []
  }
  const res = await client.listQuotes(
    {
      query: q,
      page: 1,
      order: 'latest',
    },
    signal ? { signal } : undefined,
  )
  return (res.quotes || []).map((quote) => ({
    id: `quote-${quote.id}`,
    title: `#${quote.id}`,
    description: truncateQuote(quote.content),
    category: 'Quote',
    icon: QuoteDownIcon,
    type: 'callback',
    callback: () => router.push({ name: 'quote-show', params: { id: String(quote.id) } }),
  }))
}

defineExpose({ loadInit })
</script>

<template>
  <Header
    :title="siteTitle"
    logo-url="/faridoon.png"
    :sidebar-enabled="false"
    :top-bar-enabled="true"
    :theme-toggle-enabled="true"
    :show-branding="true"
    :navigation="navigation"
    :username="auth?.username || ''"
    @logo-click="goHome"
  >
    <template #toolbar>
      <QuickSearch
        placeholder="Search quotes..."
        :auto-import-routes="false"
        :max-results="15"
        :fetch-results="fetchQuoteResults"
      />
    </template>
    <template #user-info>
      <div class="user-info">
        <router-link v-if="auth" to="/account">{{ auth.username }}</router-link>
        <router-link v-else to="/login">Login</router-link>
      </div>
    </template>
  </Header>

  <div id="layout">
    <div id="content">
      <main>
        <slot />
      </main>

      <footer>
        <span class="subtle">
          Powered by
          <a href="https://github.com/jamesread/Faridoon" target="_blank" rel="noopener noreferrer">Faridoon</a>
          {{ appVersion }}
        </span>
      </footer>
    </div>
  </div>
</template>
