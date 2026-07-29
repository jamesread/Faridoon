<script setup>
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { RouterLink } from 'vue-router'
import Section from 'picocrank/vue/components/Section.vue'
import Table from 'picocrank/vue/components/Table.vue'
import DangerZone from '../components/DangerZone.vue'
import { HugeiconsIcon } from '@hugeicons/vue'
import { PlusSignIcon } from '@hugeicons/core-free-icons'
import { client } from '../composables/client'

const webhooks = ref([])
const events = ref(['approval.requested'])
const edits = reactive({})
const editingId = ref(null)
const error = ref('')

const listHeaders = [
  { key: 'id', label: 'ID', sortable: true, width: '4rem' },
  { key: 'url', label: 'URL', sortable: true },
  { key: 'event', label: 'Event', sortable: true },
  { key: 'status', label: 'Status', sortable: true, width: '6rem' },
  { key: 'actions', label: 'Actions', sortable: false, width: '6rem' },
]

const listRows = computed(() =>
  webhooks.value.map((wh) => ({
    id: wh.id,
    url: wh.url,
    event: wh.event,
    status: wh.enabled ? 'Enabled' : 'Disabled',
    actions: '',
  })),
)

function syncEdits() {
  for (const key of Object.keys(edits)) delete edits[key]
  for (const wh of webhooks.value) {
    edits[wh.id] = { url: wh.url, secret: '', event: wh.event, enabled: !!wh.enabled }
  }
  if (editingId.value && !edits[editingId.value]) editingId.value = null
}

async function load() {
  try {
    const res = await client.listWebhooks({})
    webhooks.value = res.webhooks || []
    events.value = res.events?.length ? res.events : ['approval.requested']
    syncEdits()
  } catch (e) {
    error.value = e.message || String(e)
  }
}

const editingWebhook = computed(() => webhooks.value.find((wh) => wh.id === editingId.value) || null)

async function update(id) {
  const e = edits[id]
  await client.updateWebhook({
    id,
    url: e.url,
    secret: e.secret || '',
    event: e.event,
    enabled: e.enabled,
  })
  editingId.value = null
  await load()
}

async function destroy(id) {
  if (!confirm('Delete this webhook?')) return
  await client.deleteWebhook({ id })
  editingId.value = null
  await load()
}

onMounted(load)
watch(webhooks, syncEdits, { deep: true })
</script>

<template>
  <Section title="Webhooks" subtitle="HTTP callbacks for Faridoon events" :padding="false">
    <template #toolbar>
      <RouterLink
        to="/admin/webhooks/create"
        class="button"
        title="Add webhook"
        aria-label="Add webhook"
      >
        <HugeiconsIcon :icon="PlusSignIcon" width="1em" height="1em" />
      </RouterLink>
    </template>

    <p class="webhook-intro padding">
      Currently supported event: <code>approval.requested</code>
    </p>
    <p v-if="error" class="form-error padding">{{ error }}</p>

    <Table
      v-if="webhooks.length > 0"
      :data="listRows"
      :headers="listHeaders"
      :show-pagination="webhooks.length > 10"
    >
      <template #cell-url="{ value }">
        <code class="webhook-url">{{ value }}</code>
      </template>
      <template #cell-status="{ value }">
        <span :class="value === 'Enabled' ? 'webhook-status-on' : 'webhook-status-off'">{{ value }}</span>
      </template>
      <template #cell-actions="{ row }">
        <button type="button" class="button" @click="editingId = row.id">Edit</button>
      </template>
    </Table>
    <p v-else class="padding">No webhooks configured yet.</p>
  </Section>

  <Section
    v-if="editingWebhook && edits[editingWebhook.id]"
    :title="`Edit webhook #${editingWebhook.id}`"
    :padding="true"
  >
    <form class="form-stack" @submit.prevent="update(editingWebhook.id)">
      <label>
        URL
        <input v-model="edits[editingWebhook.id].url" type="url" required />
      </label>
      <label>
        Secret
        <input v-model="edits[editingWebhook.id].secret" type="text" placeholder="Leave blank to keep current secret" />
      </label>
      <label>
        Event
        <select v-model="edits[editingWebhook.id].event" required>
          <option v-for="event in events" :key="event" :value="event">{{ event }}</option>
        </select>
      </label>
      <label>
        <input type="checkbox" v-model="edits[editingWebhook.id].enabled" />
        Enabled
      </label>
      <div class="quote-edit-actions">
        <button type="submit" class="button">Save</button>
        <button type="button" class="button" @click="editingId = null">Cancel</button>
      </div>
    </form>
  </Section>

  <DangerZone v-if="editingWebhook" :title="`Delete webhook #${editingWebhook.id}`">
    <p>Permanently remove this webhook endpoint.</p>
    <button type="button" class="button bad" @click="destroy(editingWebhook.id)">Delete</button>
  </DangerZone>
</template>
