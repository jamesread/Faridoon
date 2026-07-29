<script setup>
import { ref } from 'vue'
import { useRouter, RouterLink } from 'vue-router'
import Section from 'picocrank/vue/components/Section.vue'
import { client } from '../composables/client'
import { loadInit, setUser } from '../composables/useInit'

const router = useRouter()
const username = ref('')
const password = ref('')
const passwordConfirmation = ref('')
const error = ref('')

async function submit() {
  error.value = ''
  try {
    const res = await client.register({
      username: username.value,
      password: password.value,
      passwordConfirmation: passwordConfirmation.value,
    })
    setUser(res.user)
    await loadInit()
    router.push('/account')
  } catch (e) {
    error.value = e.message || String(e)
  }
}
</script>

<template>
  <Section title="Register" :padding="true">
    <form class="form-stack" @submit.prevent="submit">
      <label>
        Username
        <input v-model="username" required />
      </label>
      <label>
        Password
        <input v-model="password" type="password" required minlength="4" />
      </label>
      <label>
        Confirm password
        <input v-model="passwordConfirmation" type="password" required />
      </label>
      <p v-if="error" class="form-error">{{ error }}</p>
      <button type="submit" class="button">Register</button>
    </form>
    <p><RouterLink to="/login">Already have an account?</RouterLink></p>
  </Section>
</template>
