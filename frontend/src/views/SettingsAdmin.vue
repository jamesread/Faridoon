<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import Section from 'picocrank/vue/components/Section.vue'
import FormLayout from 'picocrank/vue/components/FormLayout.vue'
import FormField from 'picocrank/vue/components/FormField.vue'
import RadioGroup from 'picocrank/vue/components/RadioGroup.vue'
import { client } from '../composables/client'
import { loadInit } from '../composables/useInit'

const cvars = ref([])
const edits = reactive({})
const dirtySections = reactive({})
const error = ref('')
const success = ref('')
const savingSection = ref('')

const booleanOptions = [
  { label: 'On', value: true },
  { label: 'Off', value: false },
]

function labelFor(cvar) {
  return cvar.title || cvar.key.replace(/_/g, ' ')
}

function fieldId(cvar) {
  return `cvar-${cvar.key}`
}

function markDirty(sectionName) {
  dirtySections[sectionName] = true
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
  } catch (e) {
    error.value = e.message || String(e)
  } finally {
    savingSection.value = ''
  }
}

onMounted(load)
</script>

<template>
  <Section title="Settings" subtitle="Configuration variables" :padding="true">
    <p>Site-wide options stored in the database. Edits apply after you save.</p>
    <p v-if="error" class="form-error">{{ error }}</p>
    <p v-if="success" class="flash-success">{{ success }}</p>
    <p v-if="cvars.length === 0 && !error" class="subtle">No configuration variables found.</p>
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
