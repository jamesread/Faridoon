<script setup>
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import Section from 'picocrank/vue/components/Section.vue'
import Table from 'picocrank/vue/components/Table.vue'
import Pagination from 'picocrank/vue/components/Pagination.vue'
import { client } from '../composables/client'

const route = useRoute()
const router = useRouter()
const logs = ref([])
const page = ref(1)
const pageSize = ref(25)
const total = ref(0)
const error = ref('')

const headers = [
  { key: 'id', label: 'ID', sortable: true, width: '4rem' },
  { key: 'created', label: 'When', sortable: true, width: '10rem' },
  { key: 'actorUsername', label: 'Actor', sortable: true },
  { key: 'action', label: 'Action', sortable: true },
  { key: 'entity', label: 'Entity', sortable: false },
  { key: 'detail', label: 'Detail', sortable: false },
  { key: 'ip', label: 'IP', sortable: false, width: '8rem' },
]

const rows = computed(() =>
  logs.value.map((l) => ({
    ...l,
    entity: [l.entityType, l.entityId || ''].filter(Boolean).join(' #'),
  })),
)

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

onMounted(load)
watch(() => route.query.page, load)
</script>

<template>
  <Section title="Audit logs" subtitle="Administrative activity" :padding="false">
    <p v-if="error" class="form-error padding">{{ error }}</p>
    <template v-else>
      <Table :data="rows" :headers="headers" :show-pagination="false">
        <template #cell-actorUsername="{ value }">
          {{ value || '—' }}
        </template>
        <template #cell-detail="{ value }">
          <code v-if="value">{{ value }}</code>
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
</template>

<style scoped>
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
</style>
