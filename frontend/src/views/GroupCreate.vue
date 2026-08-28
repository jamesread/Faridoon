<script setup>
import { ref } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import Section from 'picocrank/vue/components/Section.vue'
import { client } from '../composables/client'

const router = useRouter()
const title = ref('')
const error = ref('')

async function submit() {
  error.value = ''
  try {
    const res = await client.createGroup({ title: title.value })
    const id = res.group?.id
    router.push(id ? { name: 'iamGroup', params: { id: String(id) } } : { name: 'iamUsers' })
  } catch (e) {
    error.value = e.message || String(e)
  }
}
</script>

<template>
  <Section title="Create group" :padding="true">
    <template #toolbar>
      <RouterLink :to="{ name: 'iamUsers' }" class="button">Back</RouterLink>
    </template>
    <form class="form-stack" @submit.prevent="submit">
      <label>
        Title
        <input v-model="title" type="text" required maxlength="32" placeholder="Moderators" />
      </label>
      <p v-if="error" class="form-error">{{ error }}</p>
      <div class="quote-edit-actions">
        <button type="submit" class="button">Create</button>
        <RouterLink :to="{ name: 'iamUsers' }" class="button">Cancel</RouterLink>
      </div>
    </form>
  </Section>
</template>
