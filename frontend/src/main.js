import { createApp } from 'vue'
import { createPinia } from 'pinia'

import App from './App.vue'
import router from './router'
import { useAuthStore } from '@/stores/auth.js'
import { createPersistedStatePlugin } from '@/stores/plugins/persistedstate.js'

import '@fontsource/plus-jakarta-sans/200.css'
import '@fontsource/plus-jakarta-sans/400.css'
import '@fontsource/plus-jakarta-sans/500.css'
import '@fontsource/plus-jakarta-sans/600.css'
import '@fontsource/plus-jakarta-sans/700.css'
import '@fontsource/plus-jakarta-sans/800.css'
import '@fontsource/amiri/400.css'
import '@fontsource/amiri/700.css'
import './css/main.css'

// Init Pinia
const pinia = createPinia()
pinia.use(createPersistedStatePlugin())

// Create Vue app
createApp(App).use(router).use(pinia).mount('#app')

import { registerSW } from 'virtual:pwa-register'

// Register PWA Service Worker
registerSW({ immediate: true })

const authStore = useAuthStore(pinia)

authStore.loadMe()

// Dark mode
// Uncomment, if you'd like to restore persisted darkMode setting, or use `prefers-color-scheme: dark`.
// import { useDarkModeStore } from '@/stores/darkMode'

// const darkModeStore = useDarkModeStore(pinia)
// darkModeStore.init()

// Default title tag
const defaultDocumentTitle = 'mycbt'

// Set document title from route meta
router.afterEach((to) => {
  document.title = to.meta?.title
    ? `${to.meta.title} — ${defaultDocumentTitle}`
    : defaultDocumentTitle
})
