<script setup>
import { ref } from 'vue'
import { RouterLink } from 'vue-router'
import Section from 'picocrank/vue/components/Section.vue'
import FormLayout from 'picocrank/vue/components/FormLayout.vue'
import FormField from 'picocrank/vue/components/FormField.vue'
import { client } from '../composables/client'

const currentPassword = ref('')
const newPassword = ref('')
const confirmPassword = ref('')
const saving = ref(false)
const error = ref('')
const success = ref('')

async function save() {
  error.value = ''
  success.value = ''
  if (newPassword.value.length < 8) {
    error.value = 'New password must be at least 8 characters.'
    return
  }
  if (newPassword.value !== confirmPassword.value) {
    error.value = 'New password and confirmation do not match.'
    return
  }
  saving.value = true
  try {
    await client.changePassword({
      currentPassword: currentPassword.value,
      newPassword: newPassword.value,
    })
    success.value = 'Password changed successfully.'
    currentPassword.value = ''
    newPassword.value = ''
    confirmPassword.value = ''
  } catch (e) {
    error.value = 'Could not change password. Check your current password and try again.'
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <Section title="Change Password" :padding="true">
    <template #toolbar>
      <RouterLink :to="{ name: 'userControlPanel' }" class="button">Back</RouterLink>
    </template>

    <p v-if="error" class="form-error">{{ error }}</p>
    <p v-if="success" class="form-success">{{ success }}</p>

    <FormLayout @submit.prevent="save">
      <FormField label="Current password" for="change-password-current" :disabled="saving">
        <input
          id="change-password-current"
          v-model="currentPassword"
          type="password"
          autocomplete="current-password"
          required
          :disabled="saving"
        >
      </FormField>

      <FormField
        label="New password"
        for="change-password-new"
        :disabled="saving"
        description="At least 8 characters."
      >
        <input
          id="change-password-new"
          v-model="newPassword"
          type="password"
          autocomplete="new-password"
          required
          minlength="8"
          :disabled="saving"
        >
      </FormField>

      <FormField label="Confirm new password" for="change-password-confirm" :disabled="saving">
        <input
          id="change-password-confirm"
          v-model="confirmPassword"
          type="password"
          autocomplete="new-password"
          required
          minlength="8"
          :disabled="saving"
        >
      </FormField>

      <template #actions>
        <button type="submit" class="good" :disabled="saving">
          {{ saving ? 'Saving…' : 'Change password' }}
        </button>
      </template>
    </FormLayout>
  </Section>
</template>
