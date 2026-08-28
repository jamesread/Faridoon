<script setup>
import { computed, nextTick, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import Section from 'picocrank/vue/components/Section.vue'
import Navigation from 'picocrank/vue/components/Navigation.vue'
import NavigationGrid from 'picocrank/vue/components/NavigationGrid.vue'
import {
  Settings01Icon,
  UserShield01Icon,
  Link01Icon,
  WebhookIcon,
  FileSearchIcon,
  Activity01Icon,
} from '@hugeicons/core-free-icons'
import { initState, loadInit } from '../composables/useInit'
import {
  canAccessControlPanelFromStatus,
  canAccessIamFromStatus,
  canAccessSettingsFromStatus,
  statusFromUser,
} from '../utils/rbacAccess'

const router = useRouter()
const localNavigation = ref(null)
const refreshing = ref(false)
const canAccessControlPanel = computed(() =>
  canAccessControlPanelFromStatus(statusFromUser(initState.user)),
)

function populateHubTiles() {
  const nav = localNavigation.value
  if (!nav) return
  nav.clearNavigationLinks()
  const st = statusFromUser(initState.user)

  if (canAccessIamFromStatus(st)) {
    nav.addCallback('IAM', () => router.push({ name: 'iam' }), {
      icon: UserShield01Icon,
      name: 'iam',
      description: 'Users, groups, and roles',
    })
  }
  if (canAccessSettingsFromStatus(st)) {
    nav.addCallback('Settings', () => router.push({ name: 'settings' }), {
      icon: Settings01Icon,
      name: 'settings',
      description: 'Configure system settings',
    })
    nav.addCallback('Webhooks', () => router.push({ name: 'webhooks' }), {
      icon: WebhookIcon,
      name: 'webhooks',
      description: 'HTTP callbacks for site events',
    })
    nav.addCallback('Header links', () => router.push({ name: 'headerLinks' }), {
      icon: Link01Icon,
      name: 'header-links',
      description: 'Custom navigation links in the header',
    })
    nav.addCallback('Diagnostics', () => router.push({ name: 'diagnostics' }), {
      icon: Activity01Icon,
      name: 'diagnostics',
      description: 'Runtime and database statistics',
    })
    nav.addCallback('Audit logs', () => router.push({ name: 'logs' }), {
      icon: FileSearchIcon,
      name: 'logs',
      description: 'Review administrative actions',
    })
  }
}

async function refresh() {
  refreshing.value = true
  try {
    await loadInit()
    if (!canAccessControlPanelFromStatus(statusFromUser(initState.user))) {
      router.push('/')
      return
    }
    await nextTick()
    populateHubTiles()
  } finally {
    refreshing.value = false
  }
}

onMounted(refresh)
</script>

<template>
  <Section
    v-if="canAccessControlPanel"
    title="System Control Panel"
    subtitle="System and app settings"
    :padding="true"
  >
    <template #toolbar>
      <button type="button" class="neutral" :disabled="refreshing" @click="refresh">
        {{ refreshing ? 'Refreshing…' : 'Refresh' }}
      </button>
    </template>
    <Navigation ref="localNavigation">
      <NavigationGrid />
    </Navigation>
  </Section>
</template>
