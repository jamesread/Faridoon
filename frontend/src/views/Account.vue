<script setup>
import { ref, watch, nextTick, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import Section from 'picocrank/vue/components/Section.vue'
import Navigation from 'picocrank/vue/components/Navigation.vue'
import NavigationGrid from 'picocrank/vue/components/NavigationGrid.vue'
import {
  Login01Icon,
  Logout01Icon,
  Tick02Icon,
  UserMultiple02Icon,
  WebhookIcon,
  PlusSignIcon,
  File01Icon,
  Link01Icon,
  Settings01Icon,
  Activity01Icon,
} from '@hugeicons/core-free-icons'
import { client } from '../composables/client'
import { initState, loadInit, setUser } from '../composables/useInit'

const router = useRouter()
const nav = ref(null)

function rebuildNav() {
  const n = nav.value
  if (!n) {
    return
  }
  n.clearNavigationLinks()

  if (!initState.user) {
    n.addCallback('Login', () => router.push('/login'), {
      name: 'login',
      icon: Login01Icon,
      description: 'Sign in to your account',
    })
    return
  }

  if (initState.user.canApproveQuotes) {
    n.addCallback('Approvals', () => router.push('/approvals'), {
      name: 'approvals',
      icon: Tick02Icon,
      description: 'Review pending quotes',
      count: initState.pendingApprovals || null,
    })
  }

  if (initState.user.isAdmin) {
    n.addCallback('Users', () => router.push('/admin/users'), {
      name: 'users',
      icon: UserMultiple02Icon,
      description: 'Manage users and groups',
    })
    n.addCallback('Webhooks', () => router.push('/admin/webhooks'), {
      name: 'webhooks',
      icon: WebhookIcon,
      description: 'Configure event webhook endpoints',
    })
    n.addCallback('Header links', () => router.push('/admin/header-links'), {
      name: 'header-links',
      icon: Link01Icon,
      description: 'Manage custom links in the site header',
    })
    n.addCallback('Settings', () => router.push('/admin/settings'), {
      name: 'settings',
      icon: Settings01Icon,
      description: 'Edit configuration variables',
    })
    n.addCallback('Diagnostics', () => router.push('/admin/diagnostics'), {
      name: 'diagnostics',
      icon: Activity01Icon,
      description: 'View runtime and database diagnostics',
    })
    n.addCallback('Audit logs', () => router.push('/admin/logs'), {
      name: 'logs',
      icon: File01Icon,
      description: 'View administrative audit trail',
    })
  }

  n.addCallback('Add quote', () => router.push('/quotes/create'), {
    name: 'add-quote',
    icon: PlusSignIcon,
    description: 'Submit a new quote',
  })

  n.addCallback('Logout', async () => {
    await client.logout({})
    setUser(null)
    await loadInit()
    router.push('/quotes')
  }, {
    name: 'logout',
    icon: Logout01Icon,
    description: 'End your session',
  })
}

onMounted(() => {
  rebuildNav()
})

watch(
  () => [initState.user, initState.pendingApprovals, initState.ready],
  async () => {
    await nextTick()
    rebuildNav()
  },
  { deep: true },
)
</script>

<template>
  <Section title="Account" :padding="true">
    <p v-if="initState.user">Signed in as <strong>{{ initState.user.username }}</strong>.</p>
    <p v-else>You are not signed in.</p>

    <Navigation ref="nav">
      <NavigationGrid />
    </Navigation>
  </Section>
</template>
