<script setup>
import { nextTick, onMounted, ref } from 'vue'
import Section from 'picocrank/vue/components/Section.vue'
import ReadOnlyTextArea from 'picocrank/vue/components/ReadOnlyTextArea.vue'
import { client } from '../composables/client'

const dump = ref(null)
const yaml = ref('')
const error = ref('')

function fillDump(stats) {
  const el = dump.value
  if (!el) {
    return
  }
  el.clear()
  el.appendSection('Application')
  el.appendYamlProperty('version', stats.version || 'development')
  el.appendYamlProperty('site_title', stats.siteTitle || '')
  el.appendYamlProperty('go_version', stats.goVersion || '')
  el.appendSection('Database')
  el.appendYamlProperty('migration', stats.databaseMigration || '')
  el.appendSection('Counts')
  el.appendYamlProperty('users', stats.userCount ?? 0)
  el.appendYamlProperty('pending_approvals', stats.pendingApprovals ?? 0)
  el.appendYamlProperty('approved_quotes', stats.approvedQuotes ?? 0)
  yaml.value = el.getContentAsString()
}

onMounted(async () => {
  try {
    const res = await client.getDiagnostics({})
    await nextTick()
    fillDump(res)
  } catch (e) {
    error.value = e.message || String(e)
  }
})
</script>

<template>
  <Section title="Diagnostics" subtitle="Runtime and database stats" :padding="true">
    <template #toolbar>
      <router-link :to="{ name: 'controlPanel' }" class="button inline-icon neutral">
        <span>System Control Panel</span>
      </router-link>
    </template>
    <p v-if="error" class="form-error">{{ error }}</p>
    <ReadOnlyTextArea
      v-else
      ref="dump"
      v-model="yaml"
      label="Diagnostics dump"
      :rows="16"
      markdown-lang="yaml"
      :markdown-ticks="true"
    />
  </Section>
</template>
