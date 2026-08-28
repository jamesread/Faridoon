<script setup>
import { computed } from 'vue'
import { RouterLink } from 'vue-router'
import Section from 'picocrank/vue/components/Section.vue'
import { initState } from '../composables/useInit'

const user = computed(() => initState.user)
const privileges = computed(() => user.value?.privileges || [])
</script>

<template>
  <Section title="My Permissions" :padding="true">
    <template #toolbar>
      <RouterLink :to="{ name: 'userControlPanel' }" class="button">Back</RouterLink>
    </template>

    <template v-if="user">
      <dl class="account-info">
        <dt>Username</dt>
        <dd>{{ user.username }}</dd>
        <dt>Group</dt>
        <dd>{{ user.groupTitle || '—' }}</dd>
      </dl>

      <h3>Effective permissions</h3>
      <ul v-if="privileges.length">
        <li v-for="priv in privileges" :key="priv">{{ priv }}</li>
      </ul>
      <p v-else class="subtle">No permissions assigned.</p>
    </template>
  </Section>
</template>

<style scoped>
.account-info {
  display: grid;
  grid-template-columns: 200px 1fr;
  column-gap: 1em;
  row-gap: 0.25em;
  margin: 0 0 1.5em;
}

.account-info dt {
  font-weight: bold;
}

.account-info dd {
  margin: 0;
}
</style>
