<script setup>
import { onMounted, ref } from 'vue'
import { RouterLink } from 'vue-router'
import Section from 'picocrank/vue/components/Section.vue'
import QuoteCard from '../components/QuoteCard.vue'
import { client } from '../composables/client'
import { loadInit } from '../composables/useInit'

const quotes = ref([])
const error = ref('')
const ready = ref(false)

async function load() {
  try {
    const res = await client.listApprovals({})
    quotes.value = res.quotes || []
    error.value = ''
  } catch (e) {
    error.value = e.message || String(e)
  } finally {
    ready.value = true
  }
}

async function approve(id) {
  await client.approveQuote({ id })
  await loadInit()
  await load()
}

async function reject(id) {
  if (!confirm('Reject and permanently delete this quote?')) return
  await client.rejectQuote({ id })
  await loadInit()
  await load()
}

onMounted(load)
</script>

<template>
  <p v-if="error" class="form-error">{{ error }}</p>
  <Section v-else-if="ready && quotes.length === 0" title="Approvals" :padding="true">
    <p>No quotes waiting for approval.</p>
  </Section>
  <QuoteCard
    v-for="q in quotes"
    :key="q.id"
    :quote="q"
    :voting-enabled="false"
  >
    <div class="quote-edit-actions">
      <button type="button" class="button good" @click="approve(q.id)">Approve</button>
      <button type="button" class="button bad" @click="reject(q.id)">Reject</button>
      <RouterLink :to="`/quotes/${q.id}/edit`" class="button">Edit</RouterLink>
    </div>
  </QuoteCard>
</template>
