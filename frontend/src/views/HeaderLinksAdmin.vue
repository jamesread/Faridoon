<script setup>
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { RouterLink } from 'vue-router'
import Section from 'picocrank/vue/components/Section.vue'
import Table from 'picocrank/vue/components/Table.vue'
import DangerZone from '../components/DangerZone.vue'
import { HugeiconsIcon } from '@hugeicons/vue'
import { PlusSignIcon } from '@hugeicons/core-free-icons'
import { client } from '../composables/client'
import { loadInit } from '../composables/useInit'

const links = ref([])
const edits = reactive({})
const editingId = ref(null)
const error = ref('')

const listHeaders = [
  { key: 'id', label: 'ID', sortable: true, width: '4rem' },
  { key: 'title', label: 'Title', sortable: true },
  { key: 'url', label: 'URL', sortable: true },
  { key: 'sortOrder', label: 'Order', sortable: true, width: '5rem' },
  { key: 'status', label: 'Status', sortable: true, width: '6rem' },
  { key: 'actions', label: 'Actions', sortable: false, width: '6rem' },
]

const listRows = computed(() =>
  links.value.map((link) => ({
    id: link.id,
    title: link.title,
    url: link.url,
    sortOrder: link.sortOrder,
    status: link.enabled ? 'Enabled' : 'Disabled',
    actions: '',
  })),
)

function syncEdits() {
  for (const key of Object.keys(edits)) delete edits[key]
  for (const link of links.value) {
    edits[link.id] = {
      title: link.title,
      url: link.url,
      sortOrder: link.sortOrder,
      enabled: !!link.enabled,
      openInNewTab: !!link.openInNewTab,
    }
  }
  if (editingId.value && !edits[editingId.value]) editingId.value = null
}

async function load() {
  try {
    const res = await client.listHeaderLinks({})
    links.value = res.links || []
    syncEdits()
  } catch (e) {
    error.value = e.message || String(e)
  }
}

const editingLink = computed(() => links.value.find((link) => link.id === editingId.value) || null)

async function update(id) {
  const e = edits[id]
  await client.updateHeaderLink({
    id,
    title: e.title,
    url: e.url,
    sortOrder: Number(e.sortOrder) || 0,
    enabled: e.enabled,
    openInNewTab: e.openInNewTab,
  })
  editingId.value = null
  await load()
  await loadInit()
}

async function destroy(id) {
  if (!confirm('Delete this header link?')) return
  await client.deleteHeaderLink({ id })
  editingId.value = null
  await load()
  await loadInit()
}

onMounted(load)
watch(links, syncEdits, { deep: true })
</script>

<template>
  <Section title="Header links" subtitle="Custom links shown in the site header" :padding="false">
    <template #toolbar>
      <RouterLink
        to="/admin/header-links/create"
        class="button"
        title="Add header link"
        aria-label="Add header link"
      >
        <HugeiconsIcon :icon="PlusSignIcon" width="1em" height="1em" />
      </RouterLink>
    </template>

    <p class="padding">
      Enabled links appear in the header after the built-in navigation items. Use a site path like
      <code>/quotes</code> or an absolute <code>https://</code> URL.
    </p>
    <p v-if="error" class="form-error padding">{{ error }}</p>

    <Table
      v-if="links.length > 0"
      :data="listRows"
      :headers="listHeaders"
      :show-pagination="links.length > 10"
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
    <p v-else class="padding">No header links configured yet.</p>
  </Section>

  <Section
    v-if="editingLink && edits[editingLink.id]"
    :title="`Edit header link #${editingLink.id}`"
    :padding="true"
  >
    <form class="form-stack" @submit.prevent="update(editingLink.id)">
      <label>
        Title
        <input v-model="edits[editingLink.id].title" type="text" required maxlength="64" />
      </label>
      <label>
        URL
        <input v-model="edits[editingLink.id].url" type="text" required maxlength="2048" />
      </label>
      <label>
        Sort order
        <input v-model.number="edits[editingLink.id].sortOrder" type="number" />
      </label>
      <label>
        <input type="checkbox" v-model="edits[editingLink.id].enabled" />
        Enabled
      </label>
      <label>
        <input type="checkbox" v-model="edits[editingLink.id].openInNewTab" />
        Open in new tab
      </label>
      <div class="quote-edit-actions">
        <button type="submit" class="button">Save</button>
        <button type="button" class="button" @click="editingId = null">Cancel</button>
      </div>
    </form>
  </Section>

  <DangerZone v-if="editingLink" :title="`Delete header link #${editingLink.id}`">
    <p>Permanently remove this header link.</p>
    <button type="button" class="button bad" @click="destroy(editingLink.id)">Delete</button>
  </DangerZone>
</template>
