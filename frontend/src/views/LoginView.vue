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
import { api } from '@/services/api.js'

const form = reactive({
  login: 'admin',
  pass: 'admin12345',
  remember: true,
  showPassword: false
})

const showRegisterOptions = ref(false)

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()

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
  const baseUrl = api.defaults.baseURL
  window.location.href = `${baseUrl}/api/v1/auth/google/redirect?role=${role}`
}

const startRegistration = (role) => {
  router.push({
    path: '/auth/google/register',
    query: { role }
  })
}

onMounted(async () => {
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
    <SectionFullScreen v-slot="{ cardClass }" bg="white">
      <div class="w-full max-w-md mx-auto">
        <form 
          class="bg-white/90 dark:bg-slate-900/80 backdrop-blur-2xl rounded-[3rem] shadow-[0_32px_64px_-12px_rgba(0,0,0,0.1)] p-6 md:p-10 relative overflow-hidden border border-white/60 dark:border-white/5"
          @submit.prevent="submit"
        >
          <!-- Logo Section -->
          <div class="flex flex-col items-center mb-6">
            <div class="mb-4">
               <img src="/logo_a3_blue.png" alt="A3 Logo" class="h-16 w-16 object-contain rounded-2xl" />
            </div>
            <h1 class="text-2xl font-extrabold text-slate-800 dark:text-white mb-1">Login Atiga CBT</h1>
            <p class="text-slate-500 dark:text-slate-400 text-center max-w-[280px] leading-snug text-base">
              Masuk ke sistem ujian berbasis komputer.
            </p>
          </div>

          <!-- Notification -->
          <div v-if="authStore.errorMessage" class="mb-5 rounded-xl bg-red-50 dark:bg-red-900/20 px-4 py-2 text-sm text-red-700 dark:text-red-400 border border-red-100 dark:border-red-900/40 text-center font-medium">
            {{ authStore.errorMessage }}
          </div>

          <!-- Login Form -->
          <div class="space-y-4">
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
                  class="w-full pl-11 pr-4 py-3 bg-slate-50 dark:bg-slate-800 border border-slate-100 dark:border-slate-700 rounded-2xl focus:ring-4 focus:ring-blue-100 dark:focus:ring-blue-900/20 focus:border-blue-500 focus:bg-white dark:focus:bg-slate-900 outline-none transition-all placeholder:text-slate-300 text-base"
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
                  class="w-full pl-11 pr-11 py-3 bg-slate-50 dark:bg-slate-800 border border-slate-100 dark:border-slate-700 rounded-2xl focus:ring-4 focus:ring-blue-100 dark:focus:ring-blue-900/20 focus:border-blue-500 focus:bg-white dark:focus:bg-slate-900 outline-none transition-all placeholder:text-slate-300 text-base"
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
              class="w-full py-3.5 bg-blue-600 hover:bg-blue-700 disabled:bg-blue-400 text-white font-extrabold rounded-2xl shadow-lg shadow-blue-200 dark:shadow-none hover:-translate-y-0.5 active:translate-y-0 transition-all transform tracking-wide text-base"
              :disabled="authStore.isLoading"
            >
              {{ authStore.isLoading ? 'MEMPROSES...' : 'Masuk' }}
            </button>
          </div>

          <!-- Divider -->
          <div class="relative my-6">
            <div class="absolute inset-0 flex items-center" aria-hidden="true">
              <div class="w-full border-t border-slate-100 dark:border-slate-800"></div>
            </div>
            <div class="relative flex justify-center text-xs uppercase">
              <span class="bg-white dark:bg-slate-900 px-4 text-slate-400 font-bold tracking-widest">Atau</span>
            </div>
          </div>

          <!-- Google Section -->
          <div class="space-y-2.5">
              <button
                type="button"
                id="btn-login-google"
                class="w-full flex items-center justify-center py-3 px-4 bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-2xl hover:bg-slate-50 dark:hover:bg-slate-700 transition-all font-bold text-slate-700 dark:text-slate-200 text-base"
                @click="loginGoogle('student')"
              >
                <img src="https://www.gstatic.com/firebasejs/ui/2.0.0/images/auth/google.svg" class="w-4 h-4 mr-3" />
                Masuk dengan Google
              </button>
          </div>

          <!-- Footer -->
          <div class="mt-6 text-center space-y-2">
            <p class="text-sm text-slate-500 dark:text-slate-400">
              Belum punya akun? <br class="sm:hidden" />
              <button 
                type="button"
                @click="showRegisterOptions = !showRegisterOptions"
                class="text-blue-600 font-bold hover:underline"
              >
                Daftar Sekarang
              </button>
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
        
        <div class="mt-6 text-center">
            <p class="text-sm text-slate-400 dark:text-slate-500 font-medium">
              Butuh bantuan? <a href="mailto:aantriono82@gmail.com" class="text-blue-600 font-bold hover:underline transition-colors">Hubungi Admin</a>
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

