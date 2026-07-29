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
    router.push(id ? `/admin/groups/${id}` : '/admin/users')
  } catch (e) {
    error.value = e.message || String(e)
  }
}
</script>

<template>
  <Section title="Create group" :padding="true">
    <template #toolbar>
      <RouterLink to="/admin/users" class="button">Back</RouterLink>
    </template>
    <form class="form-stack" @submit.prevent="submit">
      <label>
        Title
        <input v-model="title" type="text" required maxlength="32" placeholder="Moderators" />
      </label>
      <p v-if="error" class="form-error">{{ error }}</p>
      <div class="quote-edit-actions">
        <button type="submit" class="button">Create</button>
        <RouterLink to="/admin/users" class="button">Cancel</RouterLink>
      </div>
    </form>
  </Section>
</template>
