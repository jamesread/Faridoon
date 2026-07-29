<script setup>
import { onMounted, ref } from 'vue'
import Section from 'picocrank/vue/components/Section.vue'
import QuoteCard from '../components/QuoteCard.vue'
import { client } from '../composables/client'
import { loadInit } from '../composables/useInit'

const quotes = ref([])
const error = ref('')

async function load() {
  try {
    const res = await client.listApprovals({})
    quotes.value = res.quotes || []
  } catch (e) {
    error.value = e.message || String(e)
  }
}

async function approve(id) {
  await client.approveQuote({ id })
  await loadInit()
  await load()
}

onMounted(load)
</script>

<template>
  <Section title="Approvals" :padding="true">
    <p v-if="error" class="form-error">{{ error }}</p>
    <p v-else-if="quotes.length === 0">No quotes waiting for approval.</p>
    <div v-for="q in quotes" :key="q.id">
      <QuoteCard :quote="q" :voting-enabled="false" />
      <button type="button" class="button" @click="approve(q.id)">Approve</button>
    </div>
  </Section>
</template>
