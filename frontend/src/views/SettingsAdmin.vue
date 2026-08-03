<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import Section from 'picocrank/vue/components/Section.vue'
import { client } from '../composables/client'
import { loadInit } from '../composables/useInit'

const cvars = ref([])
const edits = reactive({})
const error = ref('')
const success = ref('')
const saving = ref(false)
const dirty = ref(false)

function labelFor(cvar) {
  return cvar.title || cvar.key.replace(/_/g, ' ')
}

function markDirty() {
  dirty.value = true
}

const categories = computed(() => {
  const groups = []
  const indexByName = {}
  for (const c of cvars.value) {
    const name = c.category || 'Other'
    if (indexByName[name] === undefined) {
      indexByName[name] = groups.length
      groups.push({ name, cvars: [] })
    }
    groups[indexByName[name]].cvars.push(c)
  }
  return groups
})

function syncEdits() {
  for (const key of Object.keys(edits)) delete edits[key]
  for (const c of cvars.value) {
    edits[c.key] = {
      valueString: c.valueString || '',
      valueInt: c.valueInt || 0,
      boolValue: !!c.valueInt,
    }
  }
  dirty.value = false
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

async function saveAll() {
  saving.value = true
  success.value = ''
  error.value = ''
  try {
    for (const cvar of cvars.value) {
      const { valueInt, valueString } = valuesFor(cvar)
      await client.updateCvar({ key: cvar.key, valueInt, valueString })
    }
    success.value = 'Settings saved.'
    await load()
    await loadInit()
  } catch (e) {
    error.value = e.message || String(e)
  } finally {
    saving.value = false
  }
}

onMounted(load)
</script>

<template>
  <form class="settings-form" @submit.prevent="saveAll">
    <Section title="Settings" subtitle="Configuration variables" :padding="true">
      <p>Site-wide options stored in the database. Edits apply after you save.</p>
      <p v-if="error" class="form-error">{{ error }}</p>
      <p v-if="success" class="flash-success">{{ success }}</p>
      <p v-if="cvars.length === 0 && !error" class="subtle">No configuration variables found.</p>
      <button v-if="cvars.length > 0" type="submit" class="button" :disabled="!dirty || saving">
        {{ saving ? 'Saving…' : 'Save' }}
      </button>
    </Section>

    <Section
      v-for="group in categories"
      :key="group.name"
      :title="group.name"
      :padding="true"
    >
      <div v-for="cvar in group.cvars" :key="cvar.key" class="settings-cvar">
        <label v-if="cvar.mainType === 'string'">
          {{ labelFor(cvar) }}
          <span v-if="cvar.description" class="settings-cvar-description">{{ cvar.description }}</span>
          <input
            v-model="edits[cvar.key].valueString"
            type="text"
            required
            maxlength="255"
            @change="markDirty"
          />
        </label>
        <label v-else-if="cvar.mainType === 'int'">
          {{ labelFor(cvar) }}
          <span v-if="cvar.description" class="settings-cvar-description">{{ cvar.description }}</span>
          <input
            v-model.number="edits[cvar.key].valueInt"
            type="number"
            required
            :min="cvar.key === 'quotes_per_page' ? 1 : undefined"
            :max="cvar.key === 'quotes_per_page' ? 127 : undefined"
            @change="markDirty"
          />
        </label>
        <div v-else-if="cvar.mainType === 'bool'" class="settings-cvar-bool">
          <label>
            <input v-model="edits[cvar.key].boolValue" type="checkbox" @change="markDirty" />
            {{ labelFor(cvar) }}
          </label>
          <p v-if="cvar.description" class="settings-cvar-description">{{ cvar.description }}</p>
        </div>
        <p v-else class="form-error">Unsupported type: {{ cvar.mainType }}</p>
      </div>
    </Section>
  </form>
</template>
