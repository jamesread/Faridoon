import { ref } from 'vue'

export function useInstallPrompt() {
  const deferredPrompt = ref(null)
  const isInstallable = ref(false)
  const isInstalled = ref(false)

  if (typeof window !== 'undefined') {
    if (window.matchMedia('(display-mode: standalone)').matches) {
      isInstalled.value = true
    }

    window.addEventListener('beforeinstallprompt', (e) => {
      e.preventDefault()
      deferredPrompt.value = e
      isInstallable.value = true
    })

    window.addEventListener('appinstalled', () => {
      isInstalled.value = true
      isInstallable.value = false
      deferredPrompt.value = null
    })
  }

  async function promptInstall() {
    if (!deferredPrompt.value) {
      return false
    }

    deferredPrompt.value.prompt()

    const { outcome } = await deferredPrompt.value.userChoice

    deferredPrompt.value = null
    isInstallable.value = false

    return outcome === 'accepted'
  }

  return {
    isInstallable,
    isInstalled,
    promptInstall,
  }
}
