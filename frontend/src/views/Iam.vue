<script setup>
import { computed, nextTick, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { HugeiconsIcon } from '@hugeicons/vue'
import { ArrowLeft01Icon, UserMultiple02Icon } from '@hugeicons/core-free-icons'
import Section from 'picocrank/vue/components/Section.vue'
import Navigation from 'picocrank/vue/components/Navigation.vue'
import NavigationGrid from 'picocrank/vue/components/NavigationGrid.vue'
import { initState, loadInit } from '../composables/useInit'
import { canAccessIamFromStatus, statusFromUser } from '../utils/rbacAccess'

const router = useRouter()
const navRef = ref(null)
const canAccessIam = computed(() => canAccessIamFromStatus(statusFromUser(initState.user)))

onMounted(async () => {
  await loadInit()
  if (!canAccessIamFromStatus(statusFromUser(initState.user))) {
    router.push('/')
    return
  }
  await nextTick()
  const nav = navRef.value
  if (!nav) return
  nav.clearNavigationLinks()
  nav.addCallback('Users & groups', () => router.push({ name: 'iamUsers' }), {
    icon: UserMultiple02Icon,
    name: 'iam-users',
    description: 'Manage accounts and group membership',
  })
})
</script>

<template>
  <Section v-if="canAccessIam" title="IAM" subtitle="Users, groups, and roles" :padding="true">
    <template #toolbar>
      <router-link :to="{ name: 'controlPanel' }" class="button inline-icon neutral">
        <HugeiconsIcon :icon="ArrowLeft01Icon" width="1em" height="1em" aria-hidden="true" />
        <span>System Control Panel</span>
      </router-link>
    </template>
    <Navigation ref="navRef">
      <NavigationGrid />
    </Navigation>
  </Section>
</template>
