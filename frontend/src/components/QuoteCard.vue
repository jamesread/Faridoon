<script setup>
import { computed } from 'vue'
import { RouterLink } from 'vue-router'

const props = defineProps({
  quote: { type: Object, required: true },
  votingEnabled: { type: Boolean, default: false },
  canEdit: { type: Boolean, default: false },
  draft: { type: Boolean, default: false },
})

const emit = defineEmits(['vote'])

const hasSignature = computed(
  () => props.quote.formatStyle === 'signatureQuote' && props.quote.signatureAuthor,
)
</script>

<template>
  <section class="quote" :class="{ 'quote-with-votes': votingEnabled }">
    <div v-if="votingEnabled" class="vote-controls">
      <button type="button" class="button vote-btn" @click="emit('vote', 1)">+</button>
      <div class="vote-count">{{ quote.voteCount }}</div>
      <button type="button" class="button vote-btn" @click="emit('vote', -1)">−</button>
    </div>
    <div class="quote-body">
      <div v-if="draft" class="quote-meta">
        <span class="subtle quote-preview-label">Preview</span>
        <span v-if="quote.submittedByUsername" class="subtle">by {{ quote.submittedByUsername }}</span>
      </div>
      <div v-else class="quote-meta">
        <RouterLink :to="`/quotes/${quote.id}`">#{{ quote.id }}</RouterLink>
        <span>{{ quote.created }}</span>
        <span v-if="quote.submittedByUsername" class="subtle">by {{ quote.submittedByUsername }}</span>
        <RouterLink v-if="canEdit" :to="`/quotes/${quote.id}/edit`">Edit</RouterLink>
      </div>
      <div v-for="(line, idx) in quote.lines" :key="idx" class="quote-line">
        <span
          v-if="line.username"
          class="quote-username"
          :class="`username-color-${line.usernameColor || 1}`"
        >{{ line.username }}:</span>
        <div
          v-if="line.contentHtml"
          class="quote-line-content"
          v-html="line.contentHtml"
        />
        <span v-else class="quote-line-content">{{ line.content }}</span>
      </div>
      <p
        v-if="hasSignature"
        class="quote-signature subtle"
      >
        <span v-if="quote.signatureAuthorHtml" v-html="quote.signatureAuthorHtml" />
        <em v-else>{{ quote.signatureAuthor }}</em>
      </p>
      <div v-if="$slots.default" class="quote-actions">
        <slot />
      </div>
    </div>
  </section>
</template>
