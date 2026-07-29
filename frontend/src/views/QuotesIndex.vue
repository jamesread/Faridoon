<script setup>
import { ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import QuoteCard from '../components/QuoteCard.vue'
import { client } from '../composables/client'
import { initState } from '../composables/useInit'

const route = useRoute()
const router = useRouter()
const quotes = ref([])
const page = ref(1)
const totalPages = ref(1)
const error = ref('')

async function load() {
  error.value = ''
  try {
    const res = await client.listQuotes({
      order: route.query.order || 'latest',
      page: Number(route.query.page || 1),
    })
    quotes.value = res.quotes || []
    page.value = res.page || 1
    totalPages.value = res.totalPages || 1
  } catch (e) {
    error.value = e.message || String(e)
  }
}

async function vote(id, delta) {
  try {
    const res = await client.voteQuote({ id, delta })
    const q = quotes.value.find((x) => x.id === id)
    if (q) q.voteCount = res.voteCount
  } catch (e) {
    if (String(e).includes('Unauthenticated')) router.push('/login')
  }
}

watch(() => route.fullPath, load, { immediate: true })
</script>

<template>
  <p v-if="error" class="form-error">{{ error }}</p>
  <p v-else-if="quotes.length === 0">
    There are no quotes in the database... yet. Click "Add" in the navigation to be the first!
  </p>
  <template v-else>
    <QuoteCard
      v-for="q in quotes"
      :key="q.id"
      :quote="q"
      :voting-enabled="initState.features.votingEnabled"
      :can-edit="!!initState.user?.canApproveQuotes"
      @vote="(d) => vote(q.id, d)"
    />
  </template>
  <div v-if="totalPages > 1" class="quotes-pagination">
    <div class="quotes-pagination-controls">
      <button
        type="button"
        class="button"
        :disabled="page <= 1"
        title="First page"
        @click="router.push({ query: { ...route.query, page: 1 } })"
      >
        <svg xmlns="http://www.w3.org/2000/svg" width="1em" height="1em" viewBox="0 0 24 24" aria-hidden="true">
          <path fill="currentColor" d="M18.41 16.59L13.82 12l4.59-4.59L17 6l-6 6l6 6zM6 6h2v12H6z"/>
        </svg>
      </button>
      <button
        type="button"
        class="button"
        :disabled="page <= 1"
        title="Previous page"
        @click="router.push({ query: { ...route.query, page: page - 1 } })"
      >
        <svg xmlns="http://www.w3.org/2000/svg" width="1em" height="1em" viewBox="0 0 24 24" aria-hidden="true">
          <path fill="currentColor" d="M15.41 7.41L14 6l-6 6l6 6l1.41-1.41L10.83 12z"/>
        </svg>
      </button>
      <button type="button" class="button active" disabled>{{ page }}</button>
      <button
        type="button"
        class="button"
        :disabled="page >= totalPages"
        title="Next page"
        @click="router.push({ query: { ...route.query, page: page + 1 } })"
      >
        <svg xmlns="http://www.w3.org/2000/svg" width="1em" height="1em" viewBox="0 0 24 24" aria-hidden="true">
          <path fill="currentColor" d="M8.59 16.59L10 18l6-6l-6-6L8.59 7.41L13.17 12z"/>
        </svg>
      </button>
      <button
        type="button"
        class="button"
        :disabled="page >= totalPages"
        title="Last page"
        @click="router.push({ query: { ...route.query, page: totalPages } })"
      >
        <svg xmlns="http://www.w3.org/2000/svg" width="1em" height="1em" viewBox="0 0 24 24" aria-hidden="true">
          <path fill="currentColor" d="M5.59 7.41L10.18 12l-4.59 4.59L7 18l6-6l-6-6zM16 6h2v12h-2z"/>
        </svg>
      </button>
    </div>
  </div>
</template>

<style scoped>
.quotes-pagination {
  display: flex;
  justify-content: center;
  margin-top: 1rem;
}

.quotes-pagination-controls {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.quotes-pagination-controls .button {
  border: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  min-width: 2.5rem;
  height: 2.5rem;
}

.quotes-pagination-controls .button:disabled {
  background: transparent;
  cursor: not-allowed;
}
</style>
