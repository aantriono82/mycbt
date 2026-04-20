<script setup>
import { onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth.js'
import SectionFullScreen from '@/components/SectionFullScreen.vue'
import CardBox from '@/components/CardBox.vue'
import LayoutGuest from '@/layouts/LayoutGuest.vue'
import BaseButton from '@/components/BaseButton.vue'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()

const type = ref(route.params.type) // success, error, info
const message = ref(route.query.message || '')
const isLoading = ref(true)

onMounted(async () => {
  if (type.value === 'success') {
    const token = route.query.token
    if (token) {
      const ok = await authStore.loginWithToken(token)
      if (ok) {
        router.push('/dashboard')
        return
      }
    }
    type.value = 'error'
    message.value = 'Gagal memproses token login Google.'
  }
  isLoading.value = false
})
</script>

<template>
  <LayoutGuest>
    <SectionFullScreen v-slot="{ cardClass }" bg="purplePink">
      <CardBox :class="cardClass">
        <div v-if="isLoading" class="text-center py-10">
          <div class="animate-spin inline-block w-8 h-8 border-4 border-current border-t-transparent text-purple-600 rounded-full dark:text-purple-500" role="status" aria-label="loading">
            <span class="sr-only">Loading...</span>
          </div>
          <p class="mt-4 text-slate-500 dark:text-slate-400">Memproses login Google...</p>
        </div>
        
        <div v-else class="text-center py-10">
          <div v-if="type === 'error'" class="mb-4">
            <div class="mx-auto flex items-center justify-center h-12 w-12 rounded-full bg-red-100 dark:bg-red-900/30">
              <svg class="h-6 w-6 text-red-600 dark:text-red-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l18 18" />
              </svg>
            </div>
            <h3 class="mt-4 text-lg font-bold text-slate-900 dark:text-slate-100">Gagal Login</h3>
            <p class="mt-2 text-sm text-slate-500 dark:text-slate-400">{{ message }}</p>
          </div>

          <div v-else-if="type === 'info'" class="mb-4">
            <div class="mx-auto flex items-center justify-center h-12 w-12 rounded-full bg-blue-100 dark:bg-blue-900/30">
              <svg class="h-6 w-6 text-blue-600 dark:text-blue-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
            </div>
            <h3 class="mt-4 text-lg font-bold text-slate-900 dark:text-slate-100">Informasi</h3>
            <p class="mt-2 text-sm text-slate-500 dark:text-slate-400">{{ message }}</p>
          </div>

          <div class="mt-6">
            <BaseButton color="info" label="Kembali ke Login" to="/login" />
          </div>
        </div>
      </CardBox>
    </SectionFullScreen>
  </LayoutGuest>
</template>
