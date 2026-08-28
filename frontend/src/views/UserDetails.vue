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
const error = ref('')
const password = ref('')
const passwordConfirmation = ref('')
const passwordError = ref('')
const passwordSuccess = ref('')

const isSelf = computed(() => !!user.value && user.value.id === initState.user?.id)

onMounted(async () => {
  try {
    const res = await client.getUser({ id: Number(props.id) })
    user.value = res.user
  } catch (e) {
    error.value = e.message || String(e)
  }
})

async function resetPassword() {
  passwordError.value = ''
  passwordSuccess.value = ''
  try {
    await client.resetUserPassword({
      id: Number(props.id),
      password: password.value,
      passwordConfirmation: passwordConfirmation.value,
    })
    password.value = ''
    passwordConfirmation.value = ''
    passwordSuccess.value = 'Password updated.'
  } catch (e) {
    passwordError.value = e.message || String(e)
  }
}

async function destroy() {
  if (isSelf.value) return
  if (!confirm('Delete this user?')) return
  await client.deleteUser({ id: Number(props.id) })
  router.push({ name: 'iamUsers' })
}
</script>

<template>
  <Section :title="user ? user.username : 'User'" :padding="true">
    <template #toolbar>
      <RouterLink :to="{ name: 'iamUsers' }" class="button">Back</RouterLink>
      <RouterLink v-if="user" :to="{ name: 'iamUserEdit', params: { id: String(user.id) } }" class="button">Edit</RouterLink>
    </template>

    <p v-if="error" class="form-error">{{ error }}</p>
    <dl v-else-if="user" class="detail-list">
      <dt>ID</dt>
      <dd>{{ user.id }}</dd>

      <dt>Username</dt>
      <dd>{{ user.username }}</dd>

      <dt>Group</dt>
      <dd>
        <RouterLink :to="{ name: 'iamGroup', params: { id: String(user.groupId) } }">{{ user.groupTitle || `Group #${user.groupId}` }}</RouterLink>
      </dd>

      <dt>Admin</dt>
      <dd>{{ user.isAdmin ? 'Yes' : 'No' }}</dd>

      <dt>Can approve quotes</dt>
      <dd>{{ user.canApproveQuotes ? 'Yes' : 'No' }}</dd>

      <dt>Can bypass approval</dt>
      <dd>{{ user.canBypassApproval ? 'Yes' : 'No' }}</dd>

      <dt>Privileges</dt>
      <dd>
        <ul v-if="user.privileges?.length" class="detail-list-inline">
          <li v-for="priv in user.privileges" :key="priv"><code>{{ priv }}</code></li>
        </ul>
        <span v-else class="subtle">None</span>
      </dd>
    </dl>
  </Section>

  <Section v-if="user" title="Reset password" :padding="true">
    <form class="form-stack" @submit.prevent="resetPassword">
      <label>
        New password
        <input v-model="password" type="password" required minlength="4" autocomplete="new-password" />
      </label>
      <label>
        Confirm password
        <input v-model="passwordConfirmation" type="password" required minlength="4" autocomplete="new-password" />
      </label>
      <p v-if="passwordError" class="form-error">{{ passwordError }}</p>
      <p v-if="passwordSuccess" class="flash-success">{{ passwordSuccess }}</p>
      <button type="submit" class="button">Reset password</button>
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
