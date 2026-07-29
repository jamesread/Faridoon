<script setup>
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import Section from 'picocrank/vue/components/Section.vue'
import DangerZone from '../components/DangerZone.vue'
import { client } from '../composables/client'

const props = defineProps({ id: { type: [String, Number], required: true } })
const router = useRouter()
const user = ref(null)
const groups = ref([])
const groupId = ref(0)
const error = ref('')

onMounted(async () => {
  try {
    const res = await client.getUser({ id: Number(props.id) })
    user.value = res.user
    groups.value = res.groups || []
    groupId.value = res.user.groupId
  } catch (e) {
    error.value = e.message || String(e)
  }
})

async function save() {
  await client.updateUser({ id: Number(props.id), groupId: Number(groupId.value) })
  router.push('/admin/users')
}

async function destroy() {
  if (!confirm('Delete this user?')) return
  await client.deleteUser({ id: Number(props.id) })
  router.push('/admin/users')
}
</script>

<template>
  <Section :title="user ? `Edit ${user.username}` : 'Edit user'" :padding="true">
    <p v-if="error" class="form-error">{{ error }}</p>
    <form v-else-if="user" class="form-stack" @submit.prevent="save">
      <label>
        Group
        <select v-model="groupId">
          <option v-for="g in groups" :key="g.id" :value="g.id">{{ g.title }}</option>
        </select>
      </label>
      <div class="quote-edit-actions">
        <button type="submit" class="button">Save</button>
        <button type="button" class="button" @click="router.push('/admin/users')">Cancel</button>
      </div>
    </form>
  </Section>
  <DangerZone v-if="user" :title="`Delete ${user.username}`">
    <p>Permanently remove this user.</p>
    <button type="button" class="button bad" @click="destroy">Delete</button>
  </DangerZone>
</template>
