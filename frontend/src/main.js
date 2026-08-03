import { createApp } from 'vue'
import { registerSW } from 'virtual:pwa-register'
import 'femtocrank/style.css'
import './style.css'
import App from './App.vue'
import router from './router'

registerSW({
  onNeedRefresh() {
    console.log('New content available, please refresh.')
  },
  onOfflineReady() {
    console.log('App ready to work offline')
  },
})

createApp(App).use(router).mount('#app')
