<script setup>
import { computed, ref } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import Section from 'picocrank/vue/components/Section.vue'
import CheckGroup from 'picocrank/vue/components/CheckGroup.vue'
import RadioGroup from 'picocrank/vue/components/RadioGroup.vue'
import FormField from 'picocrank/vue/components/FormField.vue'
import { client } from '../composables/client'
import { initState } from '../composables/useInit'

const router = useRouter()
const url = ref('')
const secret = ref('')
const selectedEvents = ref(
  initState.webhookEvents?.length ? [...initState.webhookEvents] : ['approval.requested'],
)
const enabled = ref(true)
const error = ref('')

const booleanOptions = [
  { label: 'On', value: true },
  { label: 'Off', value: false },
]

const eventOptions = computed(() =>
  (initState.webhookEvents?.length ? initState.webhookEvents : ['approval.requested']).map((e) => ({
    label: e,
    value: e,
  })),
)

async function submit() {
  error.value = ''
  try {
    await client.createWebhook({
      url: url.value,
      secret: secret.value,
      events: selectedEvents.value,
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
      <FormField label="Events" fake>
        <div>
          <CheckGroup
            v-model="selectedEvents"
            :options="eventOptions"
            name="webhook-events-create"
            aria-label="Webhook events"
          />
          <p class="subtle">Select one or more events that should POST to this URL.</p>
        </div>
      </FormField>
      <FormField label="Enabled" fake>
        <RadioGroup
          v-model="enabled"
          name="webhook-enabled-create"
          variant="boolean"
          :options="booleanOptions"
          aria-label="Webhook enabled"
        />
      </FormField>
      <p v-if="error" class="form-error">{{ error }}</p>
      <div class="quote-edit-actions">
        <button type="submit" class="button">Create</button>
        <RouterLink to="/admin/webhooks" class="button">Cancel</RouterLink>
      </div>
    </form>
  </Section>
</template>
