<script setup>
import { nextTick, onMounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import Section from 'picocrank/vue/components/Section.vue'
import Navigation from 'picocrank/vue/components/Navigation.vue'
import NavigationGrid from 'picocrank/vue/components/NavigationGrid.vue'
import {
  Settings01Icon,
  SecurityValidationIcon,
  WebSecurityIcon,
  DashboardSquareSettingIcon,
} from '@hugeicons/core-free-icons'
import { client } from '../composables/client'
import { initState, loadInit, setUser } from '../composables/useInit'
import { canAccessControlPanelFromStatus, statusFromUser } from '../utils/rbacAccess'
import { formatDateTimeFull } from '../utils/formatDateTime'

const router = useRouter()
const nav = ref(null)
const refreshing = ref(false)

function refreshUserData() {
  return loadInit()
}

async function onRefresh() {
  refreshing.value = true
  try {
    await refreshUserData()
  } finally {
    refreshing.value = false
  }
}

function rebuildNav() {
  const n = nav.value
  if (!n) {
    return
  }
  n.clearNavigationLinks()

  n.addCallback('User Preferences', () => router.push({ name: 'userPreferences' }), {
    icon: Settings01Icon,
    name: 'user-preferences',
    description: 'Language and personal settings',
  })
  n.addCallback('Change Password', () => router.push({ name: 'changePassword' }), {
    icon: SecurityValidationIcon,
    name: 'change-password',
    description: 'Change your account password',
  })
  n.addCallback('My Permissions', () => router.push({ name: 'myPermissions' }), {
    icon: WebSecurityIcon,
    name: 'my-permissions',
    description: 'Review your group membership and effective permissions',
  })

  if (canAccessControlPanelFromStatus(statusFromUser(initState.user))) {
    n.addCallback('System Control Panel', () => router.push({ name: 'controlPanel' }), {
      icon: DashboardSquareSettingIcon,
      name: 'system-control-panel',
      description: 'System and app settings',
    })
  }
}

async function signOut() {
  await client.logout({})
  setUser(null)
  await loadInit()
  router.push('/')
  window.location.reload()
}

onMounted(async () => {
  await refreshUserData()
  rebuildNav()
})

watch(
  () => initState.user,
  async () => {
    await nextTick()
    rebuildNav()
  },
  { deep: true },
)
</script>

<template>
  <Section title="Identity" :padding="true">
    <template #toolbar>
      <button type="button" class="neutral" :disabled="refreshing" @click="onRefresh">
        {{ refreshing ? 'Refreshing…' : 'Refresh' }}
      </button>
    </template>

    <template v-if="initState.user">
      <h2>{{ initState.user.username }}</h2>
      <dl class="account-info">
        <dt>Username</dt>
        <dd>{{ initState.user.username }}</dd>
        <dt>Email</dt>
        <dd>{{ initState.user.email || 'Not provided' }}</dd>
        <dt>Group</dt>
        <dd>{{ initState.user.groupTitle || '—' }}</dd>
        <dt>Account created</dt>
        <dd>{{ formatDateTimeFull(initState.user.createdAt) }}</dd>
        <dt>Last login</dt>
        <dd>{{ formatDateTimeFull(initState.user.lastLoginAt) }}</dd>
      </dl>
    </template>
  </Section>

  <Section title="Quick actions" :padding="true">
    <Navigation ref="nav">
      <NavigationGrid />
    </Navigation>
  </Section>

  <Section title="Session" :padding="true">
    <button type="button" class="bad" @click="signOut">
      Sign Out
    </button>
  </Section>
</template>

<style scoped>
.account-info {
  display: grid;
  grid-template-columns: 200px 1fr;
  column-gap: 1em;
  row-gap: 0.25em;
  margin: 0;
}

.account-info dt {
  font-weight: bold;
}

.account-info dd {
  margin: 0;
}
</style>
