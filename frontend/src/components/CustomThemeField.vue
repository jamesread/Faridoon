<script setup>
import { watch } from 'vue'
import ThemeSwitcher from 'picocrank/vue/components/ThemeSwitcher.vue'
import { useCustomTheme } from 'picocrank/vue/composables/useCustomTheme.js'

const props = defineProps({
  modelValue: {
    type: String,
    default: '',
  },
})

const emit = defineEmits(['update:modelValue', 'change'])

const { setTheme, themePreference } = useCustomTheme()

watch(
  () => props.modelValue,
  (value) => {
    const next = value || ''
    if (themePreference.value !== next) {
      setTheme(next)
    }
  },
  { immediate: true },
)

function onThemeChange(value) {
  const next = value || ''
  emit('update:modelValue', next)
  emit('change', next)
}
</script>

<template>
  <ThemeSwitcher
    :include-supplemental-themes="true"
    :show-empty-hint="true"
    @change="onThemeChange"
  />
</template>
