<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import Section from 'picocrank/vue/components/Section.vue'
import FormLayout from 'picocrank/vue/components/FormLayout.vue'
import FormField from 'picocrank/vue/components/FormField.vue'
import RadioGroup from 'picocrank/vue/components/RadioGroup.vue'
import NotificationBlock from 'picocrank/vue/components/NotificationBlock.vue'
import { useCustomTheme } from 'picocrank/vue/composables/useCustomTheme.js'
import { client } from '../composables/client'
import { initState, loadInit } from '../composables/useInit'
import { applyAppTheming } from '../composables/applyAppTheming.js'

const cvars = ref([])
const edits = reactive({})
const dirtySections = reactive({})
const error = ref('')
const success = ref('')
const savingSection = ref('')

const { availableThemes, themeLabels } = useCustomTheme()

const booleanOptions = [
  { label: 'On', value: true },
  { label: 'Off', value: false },
]

const themeControlOptions = [
  { label: 'System preference', value: 'system' },
  { label: 'User preference', value: 'user' },
]

const themeKeys = new Set([
  'theme_color_scheme_switcher_enabled',
  'theme_name',
  'theme_control',
])

function labelFor(cvar) {
  return cvar.title || cvar.key.replace(/_/g, ' ')
}

function fieldId(cvar) {
  return `cvar-${cvar.key}`
}

function markDirty(sectionName) {
  dirtySections[sectionName] = true
}

function previewThemeSection() {
  markDirty('Theme')
  applyAppTheming({
    themeName: edits.theme_name?.valueString || '',
    themeControl: edits.theme_control?.valueString || 'user',
  })
}

function clearDirty() {
  for (const key of Object.keys(dirtySections)) {
    delete dirtySections[key]
  }
}

const categories = computed(() => {
  const groups = []
  const indexByName = {}
  for (const c of cvars.value) {
    if (themeKeys.has(c.key)) {
      continue
    }
    const name = c.category || 'Other'
    if (indexByName[name] === undefined) {
      indexByName[name] = groups.length
      groups.push({ name, cvars: [] })
    }
    groups[indexByName[name]].cvars.push(c)
  }
  return groups
})

const themeCvars = computed(() => cvars.value.filter((c) => themeKeys.has(c.key)))

function syncEdits() {
  for (const key of Object.keys(edits)) delete edits[key]
  for (const c of cvars.value) {
    edits[c.key] = {
      valueString: c.valueString || '',
      valueInt: c.valueInt || 0,
      boolValue: !!c.valueInt,
    }
  }
  clearDirty()
}

function valuesFor(cvar) {
  const edit = edits[cvar.key]
  if (cvar.mainType === 'bool') {
    return { valueInt: edit.boolValue ? 1 : 0, valueString: '' }
  }
  if (cvar.mainType === 'int') {
    return { valueInt: Number(edit.valueInt) || 0, valueString: '' }
  }
  return { valueInt: 0, valueString: edit.valueString || '' }
}

async function load() {
  try {
    const res = await client.listCvars({})
    cvars.value = res.cvars || []
    syncEdits()
    error.value = ''
  } catch (e) {
    error.value = e.message || String(e)
  }
}

async function saveSection(group) {
  savingSection.value = group.name
  success.value = ''
  error.value = ''
  try {
    for (const cvar of group.cvars) {
      const { valueInt, valueString } = valuesFor(cvar)
      await client.updateCvar({ key: cvar.key, valueInt, valueString })
    }
    success.value = `${group.name} settings saved.`
    await load()
    await loadInit()
    if (group.name === 'Theme') {
      applyAppTheming(initState.features)
    }
  } catch (e) {
    error.value = e.message || String(e)
  } finally {
    savingSection.value = ''
  }
}

const themeSection = computed(() => ({
  name: 'Theme',
  cvars: themeCvars.value,
}))

onMounted(load)
</script>

<template>
  <Section title="Settings" subtitle="Configuration variables" :padding="true">
    <template #toolbar>
      <router-link :to="{ name: 'controlPanel' }" class="button inline-icon neutral">
        <span>System Control Panel</span>
      </router-link>
    </template>
    <p>Site-wide options stored in the database. Edits apply after you save.</p>
    <p v-if="error" class="form-error">{{ error }}</p>
    <NotificationBlock
      v-if="success"
      type="success"
      :message="success"
    />
    <p v-if="cvars.length === 0 && !error" class="subtle">No configuration variables found.</p>
  </Section>

  <Section
    v-if="themeSection.cvars.length"
    title="Theme"
    subtitle="Appearance and theme policy"
    :padding="true"
  >
    <FormLayout @submit.prevent="saveSection(themeSection)">
      <FormField
        label="Color scheme switcher"
        fake
        description="Show the auto/light/dark control in the header."
      >
        <RadioGroup
          v-model="edits.theme_color_scheme_switcher_enabled.boolValue"
          name="settings-theme-color-scheme-switcher"
          variant="boolean"
          :options="booleanOptions"
          aria-label="Color scheme switcher"
          @change="previewThemeSection"
        />
      </FormField>

      <FormField
        label="Theme name"
        for="settings-theme-name"
        description="Drop-in CSS theme for the app."
      >
        <select
          id="settings-theme-name"
          v-model="edits.theme_name.valueString"
          @change="previewThemeSection"
        >
          <option value="">Default (Femtocrank only)</option>
          <option v-for="name in availableThemes" :key="name" :value="name">
            {{ themeLabels[name] || name }}
          </option>
        </select>
      </FormField>

      <FormField
        label="Theme control"
        fake
        description="System preference forces the theme name for all users. User preference uses this theme as the default; users may override on User Preferences."
      >
        <RadioGroup
          v-model="edits.theme_control.valueString"
          name="settings-theme-control"
          variant="list"
          :options="themeControlOptions"
          aria-label="Theme control"
          @change="previewThemeSection"
        />
      </FormField>

      <template #actions>
        <button
          type="submit"
          class="button good"
          :disabled="!dirtySections.Theme || savingSection === 'Theme'"
        >
          {{ savingSection === 'Theme' ? 'Saving…' : 'Save' }}
        </button>
      </template>
    </FormLayout>
  </Section>

  <Section
    v-for="group in categories"
    :key="group.name"
    :title="group.name"
    :padding="true"
  >
    <FormLayout @submit.prevent="saveSection(group)">
      <template v-for="cvar in group.cvars" :key="cvar.key">
        <FormField
          v-if="cvar.mainType === 'string'"
          :label="labelFor(cvar)"
          :for="fieldId(cvar)"
        >
          <div>
            <input
              :id="fieldId(cvar)"
              v-model="edits[cvar.key].valueString"
              type="text"
              required
              maxlength="255"
              @change="markDirty(group.name)"
            />
            <p v-if="cvar.description" class="subtle">{{ cvar.description }}</p>
          </div>
        </FormField>

        <FormField
          v-else-if="cvar.mainType === 'int'"
          :label="labelFor(cvar)"
          :for="fieldId(cvar)"
        >
          <div>
            <input
              :id="fieldId(cvar)"
              v-model.number="edits[cvar.key].valueInt"
              type="number"
              required
              :min="cvar.key === 'quotes_per_page' ? 1 : undefined"
              :max="cvar.key === 'quotes_per_page' ? 127 : undefined"
              @change="markDirty(group.name)"
            />
            <p v-if="cvar.description" class="subtle">{{ cvar.description }}</p>
          </div>
        </FormField>

        <FormField
          v-else-if="cvar.mainType === 'bool'"
          :label="labelFor(cvar)"
          fake
        >
          <div>
            <RadioGroup
              v-model="edits[cvar.key].boolValue"
              :name="fieldId(cvar)"
              variant="boolean"
              :options="booleanOptions"
              :aria-label="labelFor(cvar)"
              @change="markDirty(group.name)"
            />
            <p v-if="cvar.description" class="subtle">{{ cvar.description }}</p>
          </div>
        </FormField>

        <p v-else class="form-error">Unsupported type: {{ cvar.mainType }}</p>
      </template>

      <template #actions>
        <button
          type="submit"
          class="button"
          :disabled="!dirtySections[group.name] || savingSection === group.name"
        >
          {{ savingSection === group.name ? 'Saving…' : 'Save' }}
        </button>
      </template>
    </FormLayout>
  </Section>
</template>
