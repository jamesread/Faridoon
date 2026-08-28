<script setup>
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { HugeiconsIcon } from '@hugeicons/vue'
import { ArrowLeftRightIcon } from '@hugeicons/core-free-icons'
import Section from 'picocrank/vue/components/Section.vue'
import Table from 'picocrank/vue/components/Table.vue'
import Pagination from 'picocrank/vue/components/Pagination.vue'
import { useMediaQuery } from 'picocrank/vue/composables/useMediaQuery.js'
import { client } from '../composables/client'
import { formatIpDisplay } from '../utils/formatIp.js'
import { formatDateTimeFull, formatDateTimeShort } from '../utils/formatDateTime.js'

const WIDE_LAYOUT_STORAGE_KEY = 'faridoon-audit-logs-wide'
const MOBILE_LAYOUT_QUERY = '(max-width: 768px)'

const route = useRoute()
const router = useRouter()
const logs = ref([])
const page = ref(1)
const pageSize = ref(25)
const total = ref(0)
const error = ref('')
const selectedLog = ref(null)
const detailDialog = ref(null)
const ipCopied = ref(false)
const wideLayout = ref(false)
const isMobile = useMediaQuery(MOBILE_LAYOUT_QUERY)
const isWide = computed(() => wideLayout.value && !isMobile.value)

const widthToggleLabel = computed(() =>
  (isWide.value ? 'Use default width' : 'Use full width'))

const headers = [
  { key: 'created', label: 'When', sortable: true, width: '7.5rem' },
  { key: 'actorUsername', label: 'Actor', sortable: true, colPriority: 4 },
  { key: 'action', label: 'Action', sortable: true },
  { key: 'entity', label: 'Entity', sortable: false, colPriority: 3 },
  { key: 'detailPreview', label: 'Detail', sortable: false, colPriority: 2 },
  { key: 'ip', label: 'IP', sortable: false, colPriority: 1 },
]

const defaultColumnVisibility = {
  ip: false,
}

const columnVisibility = ref({ ...defaultColumnVisibility })

const rows = computed(() =>
  logs.value.map((l) => {
    const ip = formatIpDisplay(l.ip || '')
    return {
      ...l,
      entity: [l.entityType, l.entityId || ''].filter(Boolean).join(' #') || '—',
      detailPreview: truncateDetail(l.detail),
      ipDisplay: ip.display,
      ipTitle: ip.full || l.ip || '',
    }
  }),
)

const selectedActor = computed(() => {
  const log = selectedLog.value
  if (!log) {
    return '—'
  }
  if (!log.actorUsername) {
    return '—'
  }
  if (log.actorUserId > 0) {
    return `${log.actorUsername} (#${log.actorUserId})`
  }
  return log.actorUsername
})

const selectedEntity = computed(() => {
  const log = selectedLog.value
  if (!log) {
    return '—'
  }
  const parts = [log.entityType, log.entityId || ''].filter(Boolean)
  return parts.length ? parts.join(' #') : '—'
})

const selectedIp = computed(() => formatIpDisplay(selectedLog.value?.ip || ''))

function truncateDetail(value, max = 40) {
  if (!value) {
    return ''
  }
  if (value.length <= max) {
    return value
  }
  return `${value.slice(0, max)}…`
}

function openDetail(row) {
  selectedLog.value = row
  ipCopied.value = false
  detailDialog.value?.showModal()
}

function closeDetail() {
  detailDialog.value?.close('close')
  selectedLog.value = null
}

function onDialogBackdropClick(event) {
  if (event.target === detailDialog.value) {
    closeDetail()
  }
}

async function copyIp() {
  const ip = selectedLog.value?.ip?.trim()
  if (!ip || !navigator.clipboard?.writeText) {
    return
  }
  await navigator.clipboard.writeText(ip)
  ipCopied.value = true
}

async function load() {
  error.value = ''
  try {
    const res = await client.listLogs({
      page: Number(route.query.page || 1),
      pageSize: pageSize.value,
    })
    logs.value = res.logs || []
    page.value = res.page || 1
    total.value = res.total || 0
  } catch (e) {
    error.value = e.message || String(e)
  }
}

function onPageChange(nextPage) {
  if (Number(route.query.page || 1) === nextPage) {
    load()
    return
  }
  router.push({ query: { ...route.query, page: nextPage } })
}

function onPageSizeChange(size) {
  pageSize.value = Number(size)
  if (Number(route.query.page || 1) === 1) {
    load()
    return
  }
  router.push({ query: { ...route.query, page: 1 } })
}

function toggleWideLayout() {
  wideLayout.value = !wideLayout.value
  localStorage.setItem(WIDE_LAYOUT_STORAGE_KEY, wideLayout.value ? '1' : '0')
}

onMounted(() => {
  wideLayout.value = localStorage.getItem(WIDE_LAYOUT_STORAGE_KEY) === '1'
  load()
})
watch(() => route.query.page, load)
</script>

<template>
  <div class="audit-logs-page" :class="{ 'audit-logs-page--wide': isWide }">
    <Section title="Audit logs" subtitle="Administrative activity" :padding="false">
      <template #toolbar>
        <router-link :to="{ name: 'controlPanel' }" class="button inline-icon neutral">
          <span>System Control Panel</span>
        </router-link>
        <button
          v-if="!isMobile"
          type="button"
          class="button neutral audit-logs-width-toggle"
          :title="widthToggleLabel"
          :aria-label="widthToggleLabel"
          :aria-pressed="isWide ? 'true' : 'false'"
          @click="toggleWideLayout"
        >
          <HugeiconsIcon
            :icon="ArrowLeftRightIcon"
            width="1em"
            height="1em"
            aria-hidden="true"
          />
        </button>
      </template>

      <p v-if="error" class="form-error padding">{{ error }}</p>
      <template v-else>
        <Table
          table-id="audit-logs"
          v-model:column-visibility="columnVisibility"
          :default-column-visibility="defaultColumnVisibility"
          :data="rows"
          :headers="headers"
          :horizontal-scroll="true"
          :show-pagination="false"
        >
          <template #cell-created="{ row }">
            <button
              type="button"
              class="audit-log-date-link"
              :title="formatDateTimeFull(row.created)"
              @click="openDetail(row)"
            >
              {{ formatDateTimeShort(row.created) }}
            </button>
          </template>
          <template #cell-actorUsername="{ value }">
            {{ value || '—' }}
          </template>
          <template #cell-detailPreview="{ value }">
            <code v-if="value">{{ value }}</code>
            <span v-else>—</span>
          </template>
          <template #cell-ip="{ row }">
            <code
              v-if="row.ip"
              class="audit-log-ip-table"
              :title="row.ipTitle"
            >{{ row.ipDisplay }}</code>
            <span v-else>—</span>
          </template>
        </Table>
        <div v-if="total > 0" class="logs-pagination padding">
          <Pagination
            :total="total"
            :page="page"
            :page-size="pageSize"
            item-title="logs"
            @page-change="onPageChange"
            @page-size-change="onPageSizeChange"
          />
        </div>
      </template>
    </Section>

    <dialog
      ref="detailDialog"
      class="audit-log-dialog"
      @close="selectedLog = null"
      @click="onDialogBackdropClick"
    >
      <div v-if="selectedLog" class="audit-log-dialog-panel" @click.stop>
      <h3>Audit log #{{ selectedLog.id }}</h3>
      <dl class="detail-list">
        <dt>When</dt>
        <dd>{{ formatDateTimeFull(selectedLog.created) }}</dd>
        <dt>Actor</dt>
        <dd>{{ selectedActor }}</dd>
        <dt>Action</dt>
        <dd><code>{{ selectedLog.action }}</code></dd>
        <dt>Entity</dt>
        <dd>{{ selectedEntity }}</dd>
        <dt>Detail</dt>
        <dd>
          <code v-if="selectedLog.detail">{{ selectedLog.detail }}</code>
          <span v-else>—</span>
        </dd>
        <dt>IP</dt>
        <dd class="audit-log-ip">
          <code
            :title="selectedIp.full || undefined"
          >{{ selectedIp.display }}</code>
          <button
            v-if="selectedLog.ip"
            type="button"
            class="button neutral small audit-log-copy-ip"
            @click="copyIp"
          >
            {{ ipCopied ? 'Copied' : 'Copy' }}
          </button>
        </dd>
      </dl>
      <form method="dialog" class="audit-log-dialog-actions">
        <button type="submit" value="close" class="button neutral">Close</button>
      </form>
    </div>
  </dialog>
  </div>
</template>

<style scoped>
.audit-logs-page {
  --audit-logs-max-width: calc(768px * 1.3);
  box-sizing: border-box;
  width: min(var(--audit-logs-max-width), 100vw);
  max-width: var(--audit-logs-max-width);
  margin-left: calc(50% - min(var(--audit-logs-max-width), 100vw) / 2);
  margin-right: calc(50% - min(var(--audit-logs-max-width), 100vw) / 2);
}

.audit-logs-page--wide {
  box-sizing: border-box;
  width: 100vw;
  max-width: 100vw;
  margin-left: calc(50% - 50vw);
  margin-right: calc(50% - 50vw);
  padding-inline: 1rem;
  overflow-x: clip;
}

@media (max-width: 768px) {
  .audit-logs-page {
    width: auto;
    max-width: none;
    margin-inline: 0;
  }

  .audit-logs-page--wide {
    padding-inline: 0;
    overflow-x: visible;
  }
}

.audit-logs-width-toggle {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 0.35em 0.55em;
  line-height: 1;
}

.logs-pagination :deep(.pagination) {
  justify-content: center;
  flex-direction: column;
  gap: 1rem;
  align-items: center;
}

.logs-pagination :deep(.pagination-info) {
  flex: 0;
  text-align: center;
}

.logs-pagination :deep(.pagination-controls) {
  justify-content: center;
}

.audit-log-dialog-panel h3 {
  margin: 0 0 1rem;
}

.audit-log-dialog-actions {
  margin-top: 1.25rem;
}

.audit-log-copy-ip {
  margin-left: 0.5rem;
  vertical-align: middle;
}
</style>
