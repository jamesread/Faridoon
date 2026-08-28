<script setup>
import { computed, onMounted, ref } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import Section from 'picocrank/vue/components/Section.vue'
import Table from 'picocrank/vue/components/Table.vue'
import DangerZone from '../components/DangerZone.vue'
import { client } from '../composables/client'

const props = defineProps({ id: { type: [String, Number], required: true } })
const router = useRouter()
const group = ref(null)
const members = ref([])
const permissions = ref([])
const grantPermission = ref('')
const error = ref('')

const memberHeaders = [
  { key: 'id', label: 'ID', sortable: true, width: '4rem' },
  { key: 'username', label: 'Username', sortable: true },
  { key: 'actions', label: 'Actions', sortable: false, width: '6rem' },
]

const memberRows = computed(() =>
  members.value.map((u) => ({ id: u.id, username: u.username, actions: '' })),
)

async function load() {
  try {
    const res = await client.getGroup({ id: Number(props.id) })
    group.value = res.group
    members.value = res.members || []
    permissions.value = res.permissions || []
    error.value = ''
  } catch (e) {
    error.value = e.message || String(e)
  }
}

async function grant() {
  await client.grantPermission({
    groupId: Number(props.id),
    permissionId: Number(grantPermission.value),
  })
  grantPermission.value = ''
  await load()
}

async function revoke(permissionId) {
  await client.revokePermission({ groupId: Number(props.id), permissionId })
  await load()
}

async function destroy() {
  if (!confirm('Delete this group?')) return
  await client.deleteGroup({ id: Number(props.id) })
  router.push({ name: 'iamUsers' })
}

onMounted(load)
</script>

<template>
  <Section :title="group ? group.title : 'Group'" :padding="true">
    <template #toolbar>
      <RouterLink :to="{ name: 'iamUsers' }" class="button">Back</RouterLink>
    </template>

    <p v-if="error" class="form-error">{{ error }}</p>
    <template v-else-if="group">
      <dl class="detail-list">
        <dt>ID</dt>
        <dd>{{ group.id }}</dd>

        <dt>Title</dt>
        <dd>{{ group.title }}</dd>

        <dt>Members</dt>
        <dd>{{ members.length }}</dd>
      </dl>
    </template>
  </Section>

  <Section v-if="group" title="Permissions" :padding="true">
    <ul v-if="group.permissions?.length" class="detail-list-inline">
      <li v-for="perm in group.permissions" :key="perm.permissionId">
        <code>{{ perm.key }}</code>
        <span class="subtle"> — {{ perm.description }}</span>
        <button type="button" class="button" @click="revoke(perm.permissionId)">Revoke</button>
      </li>
    </ul>
    <p v-else class="subtle">No permissions granted.</p>

    <form class="form-stack" style="margin-top: 1rem" @submit.prevent="grant">
      <label>
        Grant permission
        <select v-model="grantPermission" required>
          <option disabled value="">Select permission</option>
          <option v-for="p in permissions" :key="p.id" :value="p.id">{{ p.key }}</option>
        </select>
      </label>
      <button type="submit" class="button">Grant</button>
    </form>
  </Section>

  <Section v-if="group" title="Members" :padding="false">
    <Table v-if="members.length > 0" :data="memberRows" :headers="memberHeaders" :show-pagination="members.length > 10">
      <template #cell-username="{ row, value }">
        <RouterLink :to="{ name: 'iamUser', params: { id: String(row.id) } }">{{ value }}</RouterLink>
      </template>
      <template #cell-actions="{ row }">
        <RouterLink :to="{ name: 'iamUser', params: { id: String(row.id) } }">View</RouterLink>
      </template>
    </Table>
    <p v-else class="padding subtle">No members in this group.</p>
  </Section>

  <DangerZone v-if="group && group.id > 2" :title="`Delete ${group.title}`">
    <p>Permanently remove this group. Users should be moved to another group first.</p>
    <button type="button" class="button bad" @click="destroy">Delete</button>
  </DangerZone>
</template>
