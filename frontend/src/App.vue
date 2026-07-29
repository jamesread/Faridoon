<script setup>
import { onMounted } from 'vue'
import { RouterView } from 'vue-router'
import AppLayout from './components/AppLayout.vue'
import { loadInit } from './composables/useInit'

onMounted(() => {
  loadInit()
})
</script>

<template>
  <AppLayout>
    <RouterView />
  </AppLayout>
</template>
