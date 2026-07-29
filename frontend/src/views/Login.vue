<script setup>
import { ref } from 'vue'
import { useRouter, RouterLink } from 'vue-router'
import Login from 'picocrank/vue/components/Login.vue'
import Section from 'picocrank/vue/components/Section.vue'
import { client } from '../composables/client'
import { initState, loadInit, setUser } from '../composables/useInit'

const router = useRouter()
const loginRef = ref(null)
const error = ref('')

async function onLocalLogin({ username, password }) {
  error.value = ''
  try {
    const res = await client.login({ username, password })
    setUser(res.user)
    await loadInit()
    const redirect = typeof router.currentRoute.value.query.redirect === 'string'
      ? router.currentRoute.value.query.redirect
      : '/account'
    router.push(redirect || '/account')
  } catch (e) {
    const message = e.message || String(e)
    error.value = message
    loginRef.value?.setLocalLoginError?.(message)
  }
}
</script>

<template>
  <Section title="Login" :padding="true">
    <p v-if="error" class="form-error">{{ error }}</p>
    <Login
      ref="loginRef"
      :show-default-tabs="false"
      :custom-tabs="[{ id: 'local', label: 'Username & Password' }]"
      @local-login="onLocalLogin"
    />
    <p v-if="initState.features.registrationEnabled">
      <RouterLink to="/register">Create an account</RouterLink>
    </p>
  </Section>
</template>
