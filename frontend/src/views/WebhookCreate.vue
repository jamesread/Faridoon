<script setup>
import { ref } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import Section from 'picocrank/vue/components/Section.vue'
import { client } from '../composables/client'
import { initState } from '../composables/useInit'

const router = useRouter()
const url = ref('')
const secret = ref('')
const event = ref(initState.webhookEvents?.[0] || 'approval.requested')
const enabled = ref(true)
const error = ref('')

async function submit() {
  error.value = ''
  try {
    await client.createWebhook({
      url: url.value,
      secret: secret.value,
      event: event.value,
      enabled: enabled.value,
    })
    router.push('/admin/webhooks')
  } catch (e) {
    error.value = e.message || String(e)
  }
}
</script>

<template>
  <Section title="Add webhook" :padding="true">
    <template #toolbar>
      <RouterLink to="/admin/webhooks" class="button">Back</RouterLink>
    </template>
    <form class="form-stack" @submit.prevent="submit">
      <label>
        URL
        <input v-model="url" type="url" required placeholder="https://example.com/hooks/faridoon" />
      </label>
      <label>
        Secret
        <input v-model="secret" type="text" required placeholder="Shared signing secret" />
      </label>
      <label>
        Event
        <select v-model="event" required>
          <option v-for="e in initState.webhookEvents" :key="e" :value="e">{{ e }}</option>
        </select>
      </label>
      <label>
        <input type="checkbox" v-model="enabled" />
        Enabled
      </label>
      <p v-if="error" class="form-error">{{ error }}</p>
      <div class="quote-edit-actions">
        <button type="submit" class="button">Create</button>
        <RouterLink to="/admin/webhooks" class="button">Cancel</RouterLink>
      </div>
    </form>
  </Section>
</template>
