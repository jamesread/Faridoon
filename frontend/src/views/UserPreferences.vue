<script setup>
import { computed, onMounted, ref } from 'vue'
import { RouterLink } from 'vue-router'
import Section from 'picocrank/vue/components/Section.vue'
import FormLayout from 'picocrank/vue/components/FormLayout.vue'
import FormField from 'picocrank/vue/components/FormField.vue'
import RadioGroup from 'picocrank/vue/components/RadioGroup.vue'
import { useTheme } from 'picocrank/vue/composables/useTheme.js'
import { useCustomTheme } from 'picocrank/vue/composables/useCustomTheme.js'
import { client } from '../composables/client'
import { initState } from '../composables/useInit'
import { themeControlFromFeatures } from '../composables/applyAppTheming.js'
import {
  applyUserLanguage,
  languageLabel,
} from '../composables/userPreferences'

const { theme } = useTheme()
const { availableThemes, themeLabels, themePreference, setTheme } = useCustomTheme()

const language = ref('')
const availableLanguages = ref([])
const saved = ref({ language: '' })
const saving = ref(false)
const error = ref('')
const success = ref('')

const initFeatures = computed(() => initState.features)
const themeControl = computed(() => themeControlFromFeatures(initFeatures.value))
const userThemeControl = computed(() => themeControl.value === 'user')

const dirty = computed(() => language.value !== saved.value.language)

const colorSchemeOptions = [
  { label: 'Auto (system)', value: 'auto' },
  { label: 'Light', value: 'light' },
  { label: 'Dark', value: 'dark' },
]

function syncSaved(res) {
  language.value = res.language || ''
  availableLanguages.value = res.availableLanguages?.length ? [...res.availableLanguages] : []
  saved.value = { language: language.value }
}

function onThemeNameChange(event) {
  setTheme(event.target.value)
}

async function load() {
  error.value = ''
  success.value = ''
  const res = await client.getUserPreferences({})
  syncSaved(res)
}

async function save() {
  saving.value = true
  error.value = ''
  success.value = ''
  try {
    await client.saveUserPreferences({
      language: language.value,
      sidebarEnabled: false,
    })
    saved.value = { language: language.value }
    applyUserLanguage(null, 'en', language.value)
    success.value = 'Preferences saved.'
  } catch (e) {
    error.value = e.message || String(e)
  } finally {
    saving.value = false
  }
}

onMounted(load)
</script>

<template>
  <Section title="User Preferences" :padding="true">
    <template #toolbar>
      <RouterLink :to="{ name: 'userControlPanel' }" class="button">Back</RouterLink>
    </template>

    <p v-if="error" class="form-error">{{ error }}</p>
    <p v-if="success" class="form-success">{{ success }}</p>

    <FormLayout @submit.prevent="save">
      <FormField
        label="Language"
        for="user-preferences-language"
        :disabled="saving"
        description="Empty (browser default) follows the browser on the next full page load."
      >
        <select id="user-preferences-language" v-model="language" :disabled="saving">
          <option value="">Browser default</option>
          <option v-for="code in availableLanguages" :key="code" :value="code">
            {{ languageLabel(code) }}
          </option>
        </select>
      </FormField>

      <FormField
        label="Color scheme preference"
        fake
        description="Choose light, dark, or match your operating system. Saved in this browser."
      >
        <RadioGroup
          v-model="theme"
          name="user-preferences-color-scheme"
          variant="list"
          :options="colorSchemeOptions"
          aria-label="Color scheme preference"
        />
      </FormField>

      <FormField
        v-if="userThemeControl"
        label="Theme name"
        for="user-preferences-theme-name"
        description="Override the administrator default theme. Saved in this browser."
      >
        <select
          id="user-preferences-theme-name"
          :value="themePreference"
          @change="onThemeNameChange"
        >
          <option value="">Default (Femtocrank only)</option>
          <option v-for="name in availableThemes" :key="name" :value="name">
            {{ themeLabels[name] || name }}
          </option>
        </select>
      </FormField>

      <template #actions>
        <button type="submit" class="good" :disabled="saving || !dirty">
          {{ saving ? 'Saving…' : 'Save preferences' }}
        </button>
      </template>
    </FormLayout>
  </Section>
</template>
