<script setup>
import { computed, onMounted, ref } from 'vue'
import { RouterLink } from 'vue-router'
import Section from 'picocrank/vue/components/Section.vue'
import Table from 'picocrank/vue/components/Table.vue'
import { client } from '../composables/client'

const users = ref([])
const groups = ref([])
const permissions = ref([])
const groupTitle = ref('')
const grantPermission = ref('')
const error = ref('')

const userHeaders = [
  { key: 'id', label: 'ID', sortable: true, width: '4rem' },
  { key: 'username', label: 'Username', sortable: true },
  { key: 'groupTitle', label: 'Group', sortable: true },
  { key: 'actions', label: 'Actions', sortable: false },
]

const userRows = computed(() =>
  users.value.map((u) => ({ ...u, actions: '' })),
)

async function load() {
  try {
    const res = await client.listUsers({})
    users.value = res.users || []
    groups.value = res.groups || []
    permissions.value = res.permissions || []
  } catch (e) {
    error.value = e.message || String(e)
  }
}

async function createGroup() {
  await client.createGroup({ title: groupTitle.value })
  groupTitle.value = ''
  await load()
}

async function deleteGroup(id) {
  if (!confirm('Delete this group?')) return
  await client.deleteGroup({ id })
  await load()
}

async function grant(groupId) {
  await client.grantPermission({ groupId, permissionId: Number(grantPermission.value) })
  grantPermission.value = ''
  await load()
}

async function revoke(groupId, permissionId) {
  await client.revokePermission({ groupId, permissionId })
  await load()
}

onMounted(load)
</script>

<template>
  <Section title="Users" :padding="false">
    <p v-if="error" class="form-error padding">{{ error }}</p>
    <Table v-else :data="userRows" :headers="userHeaders" :show-pagination="true">
      <template #cell-actions="{ row }">
        <RouterLink :to="`/admin/users/${row.id}/edit`">Edit</RouterLink>
      </template>
    </Table>
  </Section>

  <Section title="Groups" :padding="true">
    <form class="form-stack" @submit.prevent="createGroup">
      <label>
        New group title
        <input v-model="groupTitle" required />
      </label>
      <button type="submit" class="button">Create group</button>
    </form>

    <div v-for="group in groups" :key="group.id" style="margin-top: 1.5rem">
      <h3>
        {{ group.title }}
        <button
          v-if="group.id > 2"
          type="button"
          class="button"
          @click="deleteGroup(group.id)"
        >Delete</button>
      </h3>
      <ul>
        <li v-for="perm in group.permissions" :key="perm.permissionId">
          {{ perm.key }} — {{ perm.description }}
          <button type="button" class="button" @click="revoke(group.id, perm.permissionId)">Revoke</button>
        </li>
      </ul>
      <form @submit.prevent="grant(group.id)">
        <select v-model="grantPermission" required>
          <option disabled value="">Select permission</option>
          <option v-for="p in permissions" :key="p.id" :value="p.id">{{ p.key }}</option>
        </select>
        <button type="submit" class="button">Grant</button>
      </form>
    </div>
  </Section>
</template>
