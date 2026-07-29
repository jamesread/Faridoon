<script setup>
import { onMounted, reactive, ref } from 'vue'
import Section from 'picocrank/vue/components/Section.vue'
import { client } from '../composables/client'
import { loadInit } from '../composables/useInit'

const cvars = ref([])
const edits = reactive({})
const error = ref('')
const success = ref('')
const saving = ref(false)

function labelFor(key) {
  return key.replace(/_/g, ' ')
}

function syncEdits() {
  for (const key of Object.keys(edits)) delete edits[key]
  for (const c of cvars.value) {
    edits[c.key] = {
      valueString: c.valueString || '',
      valueInt: c.valueInt || 0,
      boolValue: !!c.valueInt,
    }
  }
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
  <Section title="Settings" subtitle="Configuration variables" :padding="true">
    <p v-if="error" class="form-error">{{ error }}</p>
    <p v-if="success" class="flash-success">{{ success }}</p>
    <p v-if="cvars.length === 0 && !error" class="subtle">No configuration variables found.</p>

    <form v-else class="form-stack" @submit.prevent="saveAll">
      <div v-for="cvar in cvars" :key="cvar.key" class="settings-cvar">
        <label v-if="cvar.mainType === 'string'">
          {{ labelFor(cvar.key) }}
          <input v-model="edits[cvar.key].valueString" type="text" required maxlength="255" />
        </label>
        <label v-else-if="cvar.mainType === 'int'">
          {{ labelFor(cvar.key) }}
          <input v-model.number="edits[cvar.key].valueInt" type="number" required />
        </label>
        <label v-else-if="cvar.mainType === 'bool'">
          <input v-model="edits[cvar.key].boolValue" type="checkbox" />
          {{ labelFor(cvar.key) }}
        </label>
        <p v-else class="form-error">Unsupported type: {{ cvar.mainType }}</p>
      </div>
      <button type="submit" class="button" :disabled="saving">
        {{ saving ? 'Saving…' : 'Save' }}
      </button>
    </form>
  </Section>
</template>
