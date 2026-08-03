<script setup>
import { onMounted, ref } from 'vue'
import { RouterView } from 'vue-router'
import AppLayout from './components/AppLayout.vue'
import { loadInit } from './composables/useInit'
import { useInstallPrompt } from './composables/useInstallPrompt'

const { isInstallable, isInstalled, promptInstall } = useInstallPrompt()
const showInstallBanner = ref(true)

function handleInstall() {
  promptInstall().then((accepted) => {
    if (accepted) {
      showInstallBanner.value = false
    }
  })
}

function dismissInstall() {
  showInstallBanner.value = false
  localStorage.setItem('faridoon_install_dismissed', 'true')
}

onMounted(() => {
  loadInit()
  if (localStorage.getItem('faridoon_install_dismissed') === 'true') {
    showInstallBanner.value = false
  }
})
</script>

<template>
  <div :class="{ 'has-install-banner': isInstallable && !isInstalled && showInstallBanner }">
    <div
      v-if="isInstallable && !isInstalled && showInstallBanner"
      class="install-banner"
    >
      <div class="install-banner-content">
        <div class="install-banner-text">
          <strong>Install Faridoon</strong>
          <span>Add to your home screen for quick access</span>
        </div>
        <div class="install-banner-actions">
          <button
            type="button"
            class="install-dismiss-btn"
            aria-label="Dismiss"
            @click="dismissInstall"
          >
            ×
          </button>
          <button type="button" class="install-btn" @click="handleInstall">
            Install
          </button>
        </div>
      </div>
    </div>

    <AppLayout>
      <RouterView />
    </AppLayout>
  </div>
</template>

<style scoped>
.has-install-banner {
  padding-bottom: calc(2rem + 100px);
}

.install-banner {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  z-index: 1000;
  background: #3f3f3f;
  color: #fff;
  padding: 1rem;
  box-shadow: 0 -2px 8px rgba(0, 0, 0, 0.15);
  animation: slideUp 0.3s ease-out;
}

@keyframes slideUp {
  from {
    transform: translateY(100%);
  }
  to {
    transform: translateY(0);
  }
}

.install-banner-content {
  max-width: 600px;
  margin: 0 auto;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
}

.install-banner-text {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  flex: 1;
}

.install-banner-text strong {
  font-size: 1rem;
  font-weight: 600;
}

.install-banner-text span {
  font-size: 0.875rem;
  opacity: 0.9;
}

.install-banner-actions {
  display: flex;
  gap: 0.5rem;
  align-items: center;
}

.install-btn {
  padding: 0.5rem 1rem;
  background: #488448;
  color: #fff;
  border: none;
  border-radius: 4px;
  font-weight: 600;
  cursor: pointer;
  white-space: nowrap;
}

.install-btn:hover {
  background: #3a6f3a;
}

.install-dismiss-btn {
  width: 32px;
  height: 32px;
  border: none;
  background: rgba(255, 255, 255, 0.15);
  color: white;
  border-radius: 4px;
  font-size: 1.25rem;
  line-height: 1;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
}

.install-dismiss-btn:hover {
  background: rgba(255, 255, 255, 0.25);
}

@media (max-width: 640px) {
  .install-banner-content {
    flex-direction: column;
    align-items: stretch;
  }

  .install-banner-actions {
    width: 100%;
    justify-content: space-between;
  }

  .install-btn {
    flex: 1;
  }
}
</style>
