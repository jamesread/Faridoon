<script setup>
import Section from 'picocrank/vue/components/Section.vue'

defineProps({
  summary: { type: String, default: 'Show destructive actions' },
  title: { type: String, default: 'Danger zone' },
})
</script>

<template>
  <Section :title="title" :padding="true">
    <details class="danger-zone-details">
      <summary>{{ summary }}</summary>
      <div class="danger-zone-content bad">
        <slot />
      </div>
    </details>
  </Section>
</template>
