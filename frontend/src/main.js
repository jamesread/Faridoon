import { createApp } from 'vue'
import { registerSW } from 'virtual:pwa-register'
import './style.css'
import { initSiteThemes } from './composables/useSiteTheme.js'
import App from './App.vue'
import router from './router'

void initSiteThemes()

registerSW({
  onNeedRefresh() {
    console.log('New content available, please refresh.')
  },
  onOfflineReady() {
    console.log('App ready to work offline')
  },
})

createApp(App).use(router).mount('#app')
