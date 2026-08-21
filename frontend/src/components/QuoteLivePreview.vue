<script setup>
import { ref, watch } from 'vue'
import QuoteCard from './QuoteCard.vue'
import { client } from '../composables/client.js'
import { initState } from '../composables/useInit.js'

const props = defineProps({
  content: { type: String, default: '' },
  markdownEnabled: { type: Boolean, default: false },
})

const previewQuote = ref(null)
const previewError = ref('')
let debounceTimer = null
let requestSeq = 0

watch(
  () => [props.content, props.markdownEnabled],
  () => {
    clearTimeout(debounceTimer)
    previewError.value = ''

    if (!props.content.trim()) {
      previewQuote.value = null
      return
    }

    debounceTimer = setTimeout(async () => {
      const seq = ++requestSeq
      try {
        const res = await client.formatQuote({
          content: props.content,
          markdownEnabled: props.markdownEnabled,
        })
        if (seq !== requestSeq) {
          return
        }
        previewQuote.value = {
          id: 0,
          submittedByUsername: initState.user?.username || '',
          formatStyle: res.formatStyle,
          signatureAuthor: res.signatureAuthor,
          signatureAuthorHtml: res.signatureAuthorHtml,
          lines: res.lines,
          voteCount: 0,
        }
      } catch (e) {
        if (seq !== requestSeq) {
          return
        }
        previewQuote.value = null
        previewError.value = e.message || String(e)
      }
    }, 300)
  },
  { immediate: true },
)
</script>

<template>
  <p v-if="!content.trim()" class="quote-preview-empty">Start typing to see how your quote will look.</p>
  <p v-else-if="previewError" class="form-error">{{ previewError }}</p>
  <QuoteCard
    v-else-if="previewQuote"
    :quote="previewQuote"
    draft
    :voting-enabled="initState.features.votingEnabled"
  />
</template>
