<script setup>
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import Section from 'picocrank/vue/components/Section.vue'
import DangerZone from '../components/DangerZone.vue'
import { client } from '../composables/client'
import { initState } from '../composables/useInit'

const props = defineProps({ id: { type: [String, Number], required: true } })
const router = useRouter()
const content = ref('')
const syntax = ref('')
const error = ref('')

onMounted(async () => {
  const res = await client.getQuote({ id: Number(props.id) })
  content.value = res.quote.content
  syntax.value = res.quote.syntaxHighlighting || ''
})

async function save() {
  error.value = ''
  try {
    await client.updateQuote({
      id: Number(props.id),
      content: content.value,
      syntaxHighlighting: syntax.value,
    })
    router.push(`/quotes/${props.id}`)
  } catch (e) {
    error.value = e.message || String(e)
  }
}

async function destroy() {
  if (!confirm('Delete this quote?')) return
  await client.deleteQuote({ id: Number(props.id) })
  router.push('/quotes')
}
</script>

<template>
  <Section title="Edit Quote" :padding="true">
    <form class="form-stack" @submit.prevent="save">
      <label>
        Quote
        <textarea v-model="content" rows="10" required />
      </label>
      <label v-if="initState.features.syntaxHighlightingEnabled">
        Syntax highlighting
        <input v-model="syntax" type="text" />
      </label>
      <p v-if="error" class="form-error">{{ error }}</p>
      <div class="quote-edit-actions">
        <button type="submit" class="button">Save</button>
        <button type="button" class="button" @click="router.push(`/quotes/${id}`)">Cancel</button>
      </div>
    </form>
  </Section>
  <DangerZone v-if="initState.user?.isAdmin" :title="`Delete quote #${id}`">
    <p>Permanently remove this quote.</p>
    <button type="button" class="button bad" @click="destroy">Delete</button>
  </DangerZone>
</template>
