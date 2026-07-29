<script setup>
import { computed, onMounted, ref } from 'vue'
import { RouterLink } from 'vue-router'
import Section from 'picocrank/vue/components/Section.vue'
import Table from 'picocrank/vue/components/Table.vue'
import { HugeiconsIcon } from '@hugeicons/vue'
import { PlusSignIcon } from '@hugeicons/core-free-icons'
import { client } from '../composables/client'

const users = ref([])
const groups = ref([])
const error = ref('')

const userHeaders = [
  { key: 'id', label: 'ID', sortable: true, width: '4rem' },
  { key: 'username', label: 'Username', sortable: true },
  { key: 'groupTitle', label: 'Group', sortable: true },
  { key: 'actions', label: 'Actions', sortable: false, width: '8rem' },
]

const groupHeaders = [
  { key: 'id', label: 'ID', sortable: true, width: '4rem' },
  { key: 'title', label: 'Title', sortable: true },
  { key: 'permissionCount', label: 'Permissions', sortable: true, width: '8rem' },
  { key: 'actions', label: 'Actions', sortable: false, width: '6rem' },
]

const userRows = computed(() =>
  users.value.map((u) => ({ ...u, actions: '' })),
)

const groupRows = computed(() =>
  groups.value.map((g) => ({
    id: g.id,
    title: g.title,
    permissionCount: g.permissions?.length || 0,
    actions: '',
  })),
)

async function load() {
  try {
    const res = await client.listUsers({})
    users.value = res.users || []
    groups.value = res.groups || []
    error.value = ''
  } catch (e) {
    error.value = e.message || String(e)
  }
}

onMounted(load)
</script>

<template>
  <Section title="Users" :padding="false">
    <p v-if="error" class="form-error padding">{{ error }}</p>
    <Table v-else :data="userRows" :headers="userHeaders" :show-pagination="true">
      <template #cell-username="{ row, value }">
        <RouterLink :to="`/admin/users/${row.id}`">{{ value }}</RouterLink>
      </template>
      <template #cell-groupTitle="{ row, value }">
        <RouterLink :to="`/admin/groups/${row.groupId}`">{{ value }}</RouterLink>
      </template>
      <template #cell-actions="{ row }">
        <RouterLink :to="`/admin/users/${row.id}`">View</RouterLink>
        ·
        <RouterLink :to="`/admin/users/${row.id}/edit`">Edit</RouterLink>
      </template>
    </Table>
  </Section>

  <Section title="Groups" :padding="false">
    <template #toolbar>
      <RouterLink
        to="/admin/groups/create"
        class="button"
        title="Create group"
        aria-label="Create group"
      >
        <HugeiconsIcon :icon="PlusSignIcon" width="1em" height="1em" />
      </RouterLink>
    </template>

    <Table v-if="groups.length > 0" :data="groupRows" :headers="groupHeaders" :show-pagination="groups.length > 10">
      <template #cell-title="{ row, value }">
        <RouterLink :to="`/admin/groups/${row.id}`">{{ value }}</RouterLink>
      </template>
      <template #cell-actions="{ row }">
        <RouterLink :to="`/admin/groups/${row.id}`">View</RouterLink>
      </template>
    </Table>
    <p v-else class="padding subtle">No groups yet.</p>
  </Section>
</template>
