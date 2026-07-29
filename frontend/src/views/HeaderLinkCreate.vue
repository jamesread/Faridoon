<script setup>
import { ref } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import Section from 'picocrank/vue/components/Section.vue'
import { client } from '../composables/client'
import { loadInit } from '../composables/useInit'

const router = useRouter()
const title = ref('')
const url = ref('')
const sortOrder = ref(0)
const enabled = ref(true)
const openInNewTab = ref(true)
const error = ref('')

async function submit() {
  error.value = ''
  try {
    await client.createHeaderLink({
      title: title.value,
      url: url.value,
      sortOrder: Number(sortOrder.value) || 0,
      enabled: enabled.value,
      openInNewTab: openInNewTab.value,
    })
    await loadInit()
    router.push('/admin/header-links')
  } catch (e) {
    error.value = e.message || String(e)
  }
}
</script>

<template>
  <Section title="Add header link" :padding="true">
    <template #toolbar>
      <RouterLink to="/admin/header-links" class="button">Back</RouterLink>
    </template>
    <form class="form-stack" @submit.prevent="submit">
      <label>
        Title
        <input v-model="title" type="text" required maxlength="64" placeholder="Documentation" />
      </label>
      <label>
        URL
        <input v-model="url" type="text" required maxlength="2048" placeholder="https://example.com or /quotes" />
      </label>
      <label>
        Sort order
        <input v-model.number="sortOrder" type="number" />
      </label>
      <label>
        <input type="checkbox" v-model="enabled" />
        Enabled
      </label>
      <label>
        <input type="checkbox" v-model="openInNewTab" />
        Open in new tab
      </label>
      <p v-if="error" class="form-error">{{ error }}</p>
      <div class="quote-edit-actions">
        <button type="submit" class="button">Create</button>
        <RouterLink to="/admin/header-links" class="button">Cancel</RouterLink>
      </div>
    </form>
  </Section>
</template>
