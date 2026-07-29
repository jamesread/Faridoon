<script setup>
import { RouterLink } from 'vue-router'

defineProps({
  quote: { type: Object, required: true },
  votingEnabled: { type: Boolean, default: true },
  canEdit: { type: Boolean, default: false },
})

const emit = defineEmits(['vote'])
</script>

<template>
  <section class="quote quote-with-votes">
    <div v-if="votingEnabled" class="vote-controls">
      <button type="button" class="button vote-btn" @click="emit('vote', 1)">+</button>
      <div class="vote-count">{{ quote.voteCount }}</div>
      <button type="button" class="button vote-btn" @click="emit('vote', -1)">−</button>
    </div>
    <div class="quote-body">
      <div class="quote-meta">
        <RouterLink :to="`/quotes/${quote.id}`">#{{ quote.id }}</RouterLink>
        <span>{{ quote.created }}</span>
        <RouterLink v-if="canEdit" :to="`/quotes/${quote.id}/edit`">Edit</RouterLink>
      </div>
      <div v-for="(line, idx) in quote.lines" :key="idx" class="quote-line">
        <span
          v-if="line.username"
          class="quote-username"
          :class="`username-color-${line.usernameColor || 1}`"
        >{{ line.username }}:</span>
        <span>{{ line.content }}</span>
      </div>
    </div>
  </section>
</template>
