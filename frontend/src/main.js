import { createApp } from 'vue'
import { createPinia } from 'pinia'

import App from './App.vue'
import router from './router'
import { useAuthStore } from '@/stores/auth.js'
import { createPersistedStatePlugin } from '@/stores/plugins/persistedstate.js'

import './css/fonts.css'
import './css/main.css'

// Init Pinia
const pinia = createPinia()
pinia.use(createPersistedStatePlugin())

// Create Vue app
createApp(App).use(router).use(pinia).mount('#app')

if (import.meta.env.PROD) {
  const registerServiceWorker = () => {
    import('virtual:pwa-register').then(({ registerSW }) => {
      registerSW({ immediate: true })
    })
  }

  if ('requestIdleCallback' in window) {
    window.requestIdleCallback(registerServiceWorker, { timeout: 3000 })
  } else {
    window.setTimeout(registerServiceWorker, 1500)
  }
} else if ('serviceWorker' in navigator) {
  navigator.serviceWorker.getRegistrations().then((registrations) => {
    registrations.forEach((registration) => registration.unregister())
  })
}

const authStore = useAuthStore(pinia)

authStore.loadMe()

// Dark mode
// Uncomment, if you'd like to restore persisted darkMode setting, or use `prefers-color-scheme: dark`.
// import { useDarkModeStore } from '@/stores/darkMode'

// const darkModeStore = useDarkModeStore(pinia)
// darkModeStore.init()

// Default title tag
const defaultDocumentTitle = 'AtigaCBT'

// Set document title from route meta
router.afterEach((to) => {
  document.title = to.meta?.title
    ? `${to.meta.title} — ${defaultDocumentTitle}`
    : defaultDocumentTitle
})
