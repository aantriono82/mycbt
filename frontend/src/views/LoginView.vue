<script setup>
import { computed, reactive, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { mdiAccount, mdiLock, mdiEye, mdiEyeOff, mdiEmailOutline, mdiLockOutline } from '@mdi/js'
import BaseIcon from '@/components/BaseIcon.vue'
import SectionFullScreen from '@/components/SectionFullScreen.vue'
import CardBox from '@/components/CardBox.vue'
import FormCheckRadio from '@/components/FormCheckRadio.vue'
import FormField from '@/components/FormField.vue'
import FormControl from '@/components/FormControl.vue'
import BaseButton from '@/components/BaseButton.vue'
import BaseButtons from '@/components/BaseButtons.vue'
import LayoutGuest from '@/layouts/LayoutGuest.vue'
import NotificationBarInCard from '@/components/NotificationBarInCard.vue'
import { useAuthStore } from '@/stores/auth.js'
import { useSchoolIdentityStore } from '@/stores/schoolIdentity.js'
import { api } from '@/services/api.js'

const schoolStore = useSchoolIdentityStore()
const authStore = useAuthStore()

const form = reactive({
  login: '',
  pass: '',
  remember: true,
  showPassword: false
})

const showRegisterOptions = ref(false)

const router = useRouter()
const route = useRoute()
const appVersion = __APP_VERSION__

const submit = async () => {
  const ok = await authStore.login({
    username: form.login,
    password: form.pass,
  })
  if (!ok) return
  const redirect = typeof route.query.redirect === 'string' ? route.query.redirect : '/dashboard'
  router.push(redirect)
}

const loginGoogle = (role) => {
  const safeRole = role === 'teacher' ? 'teacher' : 'student'
  const endpoint = '/api/v1/auth/google/redirect'
  const configuredBase = String(api.defaults.baseURL || '').trim()

  if (configuredBase) {
    try {
      const url = new URL(configuredBase, window.location.origin)
      const basePath = url.pathname.replace(/\/+$/, '')
      const prefix = basePath.endsWith('/api') ? basePath.slice(0, -4) : basePath
      const cleanPrefix = `${prefix}${endpoint}`.replace(/\/{2,}/g, '/')
      const finalUrl = url.origin + cleanPrefix + `?role=${encodeURIComponent(safeRole)}`
      console.log('Redirecting to Google OAuth:', finalUrl)
      window.location.replace(finalUrl)
      return
    } catch (e) {
      console.error('Error constructing Google Login URL:', e)
      // fallback below
    }
  }

  const fallbackUrl = window.location.origin + endpoint + `?role=${encodeURIComponent(safeRole)}`
  console.log('Fallback Redirect to Google OAuth:', fallbackUrl)
  window.location.replace(fallbackUrl)
}

const startRegistration = (role) => {
  router.push({
    path: '/auth/google/register',
    query: { role }
  })
}

onMounted(async () => {
  schoolStore.loadSchoolIdentity()
  const queryToken = route.query.token
  if (queryToken && typeof queryToken === 'string') {
    const ok = await authStore.loginWithToken(queryToken)
    if (ok) {
      const redirect = typeof route.query.redirect === 'string' ? route.query.redirect : '/dashboard'
      router.push(redirect)
    }
  }
})
</script>

<template>
  <LayoutGuest>
    <SectionFullScreen v-slot="{ cardClass }" class="bg-slate-50 dark:bg-slate-950">
      <div class="absolute inset-0 bg-gradient-to-tr from-blue-50 via-transparent to-indigo-50 opacity-60 dark:from-blue-900/10 dark:to-transparent"></div>
      <div class="relative z-10 mx-auto w-full max-w-[440px] animate-fade-in px-4 sm:px-0">
        <form 
          class="relative overflow-hidden rounded-[2rem] border border-white/40 bg-white/95 p-6 shadow-[0_24px_50px_-20px_rgba(0,0,0,0.28)] backdrop-blur-3xl sm:rounded-[2.5rem] sm:p-10 dark:border-white/5 dark:bg-slate-900/90"
          @submit.prevent="submit"
        >
          <!-- Logo Section -->
          <div class="mb-5 flex flex-col items-center sm:mb-8">
            <div class="mb-2 sm:mb-4">
              <div class="flex h-24 w-24 items-center justify-center rounded-full border border-blue-200 bg-blue-50 shadow-sm sm:h-28 sm:w-28 dark:border-slate-700 dark:bg-slate-900">
                <img src="/logo_atiga.png" alt="Atiga CBT Logo" class="h-16 w-16 sm:h-20 sm:w-20 object-contain rounded-full" />
              </div>
            </div>
            <h1 class="mb-1 text-center text-2xl font-extrabold text-slate-800 sm:text-3xl dark:text-white">Login Atiga CBT</h1>
            <p class="max-w-[260px] text-center text-sm leading-snug text-slate-500 sm:max-w-[280px] sm:text-base dark:text-slate-400">
              Masuk ke sistem ujian berbasis komputer.
            </p>
          </div>

          <!-- Notification -->
          <div v-if="authStore.errorMessage" class="mb-4 rounded-xl bg-red-50 dark:bg-red-900/20 px-4 py-2 text-xs sm:text-sm text-red-700 dark:text-red-400 border border-red-100 dark:border-red-900/40 text-center font-medium">
            {{ authStore.errorMessage }}
          </div>

          <!-- Login Form -->
          <div class="space-y-3 sm:space-y-4">
            <div>
              <label class="block text-sm font-bold text-slate-700 dark:text-slate-300 mb-1.5 px-1">Email atau Username</label>
              <div class="relative group">
                <span class="absolute inset-y-0 left-0 pl-4 flex items-center text-slate-400 group-focus-within:text-blue-600 transition-colors">
                  <base-icon :path="mdiEmailOutline" size="18" />
                </span>
                <input 
                  v-model="form.login"
                  type="text" 
                  placeholder="Masukkan email atau username"
                  class="w-full pl-11 pr-4 py-2.5 sm:py-3 bg-slate-50 dark:bg-slate-800 border border-slate-100 dark:border-slate-700 rounded-2xl focus:ring-4 focus:ring-blue-100 dark:focus:ring-blue-900/20 focus:border-blue-500 focus:bg-white dark:focus:bg-slate-900 outline-none transition-all placeholder:text-slate-300 text-sm sm:text-base"
                  autocomplete="username"
                  required
                />
              </div>
            </div>

            <div>
              <label class="block text-sm font-bold text-slate-700 dark:text-slate-300 mb-1.5 px-1">Password</label>
              <div class="relative group">
                <span class="absolute inset-y-0 left-0 pl-4 flex items-center text-slate-400 group-focus-within:text-blue-600 transition-colors">
                  <base-icon :path="mdiLockOutline" size="18" />
                </span>
                  <input 
                    v-model="form.pass"
                    :type="form.showPassword ? 'text' : 'password'" 
                    placeholder="Masukkan password"
                    class="w-full pl-11 pr-11 py-2.5 sm:py-3 bg-slate-50 dark:bg-slate-800 border border-slate-100 dark:border-slate-700 rounded-2xl focus:ring-4 focus:ring-blue-100 dark:focus:ring-blue-900/20 focus:border-blue-500 focus:bg-white dark:focus:bg-slate-900 outline-none transition-all placeholder:text-slate-300 text-sm sm:text-base"
                    autocomplete="current-password"
                    required
                  />
                <button 
                  type="button"
                  class="absolute inset-y-0 right-0 pr-4 flex items-center text-slate-400 hover:text-blue-600 transition-colors"
                  @click="form.showPassword = !form.showPassword"
                >
                  <base-icon :path="form.showPassword ? mdiEyeOff : mdiEye" size="20" />
                </button>
              </div>
            </div>

            <div class="flex items-center justify-between px-1">
              <label class="flex items-center group cursor-pointer">
                <input 
                  v-model="form.remember"
                  type="checkbox" 
                  class="w-4 h-4 rounded border-slate-200 text-blue-600 focus:ring-blue-500 transition-all cursor-pointer"
                />
                <span class="ml-2 text-sm text-slate-600 dark:text-slate-400 font-medium group-hover:text-slate-800">Ingat saya</span>
              </label>
              <router-link to="/forgot-password" class="text-sm font-semibold text-blue-600 hover:text-blue-700 transition-colors">Lupa Password?</router-link>
            </div>

            <button 
              type="submit" 
              class="w-full rounded-2xl bg-blue-600 py-3 sm:py-3.5 text-sm sm:text-base font-extrabold tracking-wide text-white shadow-lg shadow-blue-200 transition-all hover:-translate-y-0.5 hover:bg-blue-700 active:translate-y-0 disabled:bg-blue-400 dark:shadow-none"
              :disabled="authStore.isLoading"
            >
              {{ authStore.isLoading ? 'MEMPROSES...' : 'Masuk' }}
            </button>
          </div>

          <!-- Divider -->
          <div class="relative my-4 sm:my-6">
            <div class="absolute inset-0 flex items-center" aria-hidden="true">
              <div class="w-full border-t border-slate-100 dark:border-slate-800"></div>
            </div>
            <div class="relative flex justify-center text-[10px] sm:text-xs uppercase">
              <span class="bg-white dark:bg-slate-900 px-4 text-slate-400 font-bold tracking-widest">Atau</span>
            </div>
          </div>

          <!-- Google Section -->
          <div class="space-y-2.5">
              <button
                type="button"
                id="btn-login-google"
                class="w-full flex items-center justify-center py-2.5 sm:py-3 px-4 bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-2xl hover:bg-slate-50 dark:hover:bg-slate-700 transition-all font-bold text-slate-700 dark:text-slate-200 text-sm sm:text-base"
                @click="loginGoogle('student')"
              >
                <img src="https://www.gstatic.com/firebasejs/ui/2.0.0/images/auth/google.svg" class="w-4 h-4 mr-3" />
                Masuk dengan Google
              </button>
          </div>

          <!-- Footer -->
          <div class="mt-4 sm:mt-6 space-y-2 pb-[max(0px,env(safe-area-inset-bottom))] text-center">
            <p class="text-xs sm:text-sm text-slate-500 dark:text-slate-400">
              Belum punya akun? <br class="sm:hidden" />
              <button 
                type="button"
                @click="showRegisterOptions = !showRegisterOptions"
                class="text-blue-600 font-bold hover:underline"
              >
                Daftar Sekarang
              </button>
            </p>
            <p class="text-[11px] sm:text-xs text-slate-400 dark:text-slate-500">
              Versi aplikasi: v{{ appVersion }}
            </p>
            
            <div v-if="showRegisterOptions" class="grid grid-cols-2 gap-2 mt-2 animate-fade-in-up">
              <button 
                type="button"
                @click="startRegistration('student')"
                class="py-2 text-xs bg-blue-50 text-blue-700 font-bold rounded-lg border border-blue-100 hover:bg-blue-100 transition-colors uppercase tracking-wider"
              >
                Daftar Siswa
              </button>
              <button 
                type="button"
                @click="startRegistration('teacher')"
                class="py-2 text-xs bg-indigo-50 text-indigo-700 font-bold rounded-lg border border-indigo-100 hover:bg-indigo-100 transition-colors uppercase tracking-wider"
              >
                Daftar Guru
              </button>
            </div>
          </div>
        </form>
        
        <div class="mt-8 text-center">
            <p class="text-sm text-slate-400 dark:text-slate-500 font-medium tracking-wide">
              Butuh bantuan? <a href="mailto:aantriono82@gmail.com" class="text-blue-600 dark:text-blue-400 font-bold hover:underline transition-all">Hubungi Admin</a>
            </p>
        </div>
      </div>
    </SectionFullScreen>
  </LayoutGuest>
</template>

<style scoped>
/* Optional: Soft glass effect for card */
form {
  border: 1px solid rgba(255, 255, 255, 0.1);
}
</style>
