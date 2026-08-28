<script setup>
import { computed, onMounted, ref } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import Section from 'picocrank/vue/components/Section.vue'
import DangerZone from '../components/DangerZone.vue'
import { client } from '../composables/client'
import { initState } from '../composables/useInit'

const props = defineProps({ id: { type: [String, Number], required: true } })
const router = useRouter()
const user = ref(null)
const groups = ref([])
const groupId = ref(0)
const error = ref('')

const isSelf = computed(() => !!user.value && user.value.id === initState.user?.id)

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
  router.push({ name: 'iamUser', params: { id: String(props.id) } })
}

async function destroy() {
  if (isSelf.value) return
  if (!confirm('Delete this user?')) return
  await client.deleteUser({ id: Number(props.id) })
  router.push({ name: 'iamUsers' })
}
</script>

<template>
  <Section :title="user ? `Edit ${user.username}` : 'Edit user'" :padding="true">
    <template #toolbar>
      <RouterLink :to="{ name: 'iamUser', params: { id: String(id) } }" class="button">Back</RouterLink>
    </template>
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
        <RouterLink :to="{ name: 'iamUser', params: { id: String(id) } }" class="button">Cancel</RouterLink>
      </div>
    </form>
  </Section>
  <DangerZone v-if="user" :title="`Delete ${user.username}`">
    <template v-if="isSelf">
      <p>You cannot delete your own account while signed in.</p>
    </template>
    <template v-else>
      <p>Permanently remove this user.</p>
      <button type="button" class="button bad" @click="destroy">Delete</button>
    </template>
  </DangerZone>
</template>
