<script setup>
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import QuoteCard from '../components/QuoteCard.vue'
import { client } from '../composables/client'
import { initState } from '../composables/useInit'

const props = defineProps({ id: { type: [String, Number], required: true } })
const router = useRouter()
const quote = ref(null)
const error = ref('')

onMounted(async () => {
  try {
    const res = await client.getQuote({ id: Number(props.id) })
    quote.value = res.quote
  } catch (e) {
    error.value = e.message || String(e)
  }
})

async function vote(delta) {
  try {
    const res = await client.voteQuote({ id: Number(props.id), delta })
    if (quote.value) quote.value.voteCount = res.voteCount
  } catch (e) {
    if (String(e).includes('Unauthenticated')) router.push('/login')
  }
}
</script>

<template>
  <p v-if="error" class="form-error">{{ error }}</p>
  <QuoteCard
    v-else-if="quote"
    :quote="quote"
    :voting-enabled="initState.features.votingEnabled"
    :can-edit="!!initState.user?.canApproveQuotes"
    @vote="vote"
  />
</template>
