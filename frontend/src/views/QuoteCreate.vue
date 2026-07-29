<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import Section from 'picocrank/vue/components/Section.vue'
import { client } from '../composables/client'
import { initState, loadInit } from '../composables/useInit'

const router = useRouter()
const content = ref('')
const syntax = ref('')
const error = ref('')
const pending = ref(false)

async function submit() {
  error.value = ''
  try {
    const res = await client.createQuote({
      content: content.value,
      syntaxHighlighting: syntax.value,
    })
    await loadInit()
    if (res.pendingApproval) {
      pending.value = true
    } else {
      router.push(`/quotes/${res.quote.id}`)
    }
  } catch (e) {
    error.value = e.message || String(e)
  }
}
</script>

<template>
  <Section title="Add Quote" :padding="true">
    <p v-if="pending" class="flash-success">Quote submitted and waiting for approval.</p>
    <form v-else class="form-stack" @submit.prevent="submit">
      <label>
        Quote
        <textarea v-model="content" rows="10" required />
      </label>
      <label v-if="initState.features.syntaxHighlightingEnabled">
        Syntax highlighting
        <input v-model="syntax" type="text" />
      </label>
      <p v-if="error" class="form-error">{{ error }}</p>
      <button type="submit" class="button">Add</button>
    </form>
  </Section>
</template>
