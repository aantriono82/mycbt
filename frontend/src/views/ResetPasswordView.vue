<script setup>
import { reactive, ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { mdiLockOutline, mdiEye, mdiEyeOff, mdiCheckCircleOutline, mdiArrowLeft } from '@mdi/js'
import BaseIcon from '@/components/BaseIcon.vue'
import SectionFullScreen from '@/components/SectionFullScreen.vue'
import LayoutGuest from '@/layouts/LayoutGuest.vue'
import { api } from '@/services/api.js'

const route = useRoute()
const router = useRouter()

const form = reactive({
  password: '',
  confirmPassword: '',
  showPassword: false
})

const token = ref('')
const isLoading = ref(false)
const errorMessage = ref('')
const successMessage = ref('')

onMounted(() => {
  token.value = route.query.token || ''
  if (!token.value) {
    errorMessage.value = 'Token tidak ditemukan. Silakan klik tautan dari email kembali.'
  }
})

const submit = async () => {
  if (form.password !== form.confirmPassword) {
    errorMessage.value = 'Konfirmasi kata sandi tidak cocok.'
    return
  }

  isLoading.value = true
  errorMessage.value = ''
  successMessage.value = ''
  
  try {
    const { data } = await api.post('/api/v1/auth/reset-password', {
      token: token.value,
      new_password: form.password
    })
    successMessage.value = data.message || 'Kata sandi Anda telah berhasil diperbarui.'
    
    // Redirect to login after 3 seconds
    setTimeout(() => {
      router.push('/login')
    }, 3000)
  } catch (err) {
    errorMessage.value = err.response?.data?.error || 'Gagal mengatur ulang kata sandi. Tautan mungkin sudah kedaluwarsa.'
  } finally {
    isLoading.value = false
  }
}
</script>

<template>
  <LayoutGuest>
    <SectionFullScreen v-slot="{ cardClass }" bg="white">
      <div class="w-full max-w-md mx-auto">
        <div class="bg-white dark:bg-slate-900 rounded-[2.5rem] shadow-2xl p-6 md:p-10 relative overflow-hidden border border-slate-100 dark:border-slate-800">
          <!-- Logo Section -->
          <div class="flex flex-col items-center mb-8">
            <div class="mb-4">
               <img src="/logo_a3_blue.png" alt="A3 Logo" class="h-16 w-16 object-contain rounded-2xl" />
            </div>
            <h1 class="text-3xl font-extrabold text-blue-600 dark:text-blue-500 mb-3 text-center">Atur Ulang Kata Sandi</h1>
            <p class="text-slate-500 dark:text-slate-400 text-center max-w-[320px] leading-relaxed text-base">
              Silakan masukkan kata sandi baru Anda di bawah ini.
            </p>
          </div>

          <!-- Notification -->
          <div v-if="errorMessage" class="mb-5 rounded-xl bg-red-50 dark:bg-red-900/20 px-4 py-3 text-sm text-red-700 dark:text-red-400 border border-red-100 dark:border-red-900/40 text-center font-medium">
            {{ errorMessage }}
          </div>
          
          <div v-if="successMessage" class="mb-5 flex flex-col items-center gap-3 py-6 px-4 bg-emerald-50 dark:bg-emerald-900/20 rounded-3xl border border-emerald-100 dark:border-emerald-900/40 text-center animate-fade-in">
            <base-icon :path="mdiCheckCircleOutline" size="48" class="text-emerald-500" />
            <h3 class="text-lg font-bold text-emerald-700 dark:text-emerald-400">Berhasil!</h3>
            <p class="text-sm text-emerald-600 dark:text-emerald-500">{{ successMessage }}</p>
            <p class="text-xs text-emerald-500 mt-2 italic">Mengalihkan ke halaman login dalam 3 detik...</p>
          </div>

          <!-- Form -->
          <form v-else @submit.prevent="submit" class="space-y-5">
            <div>
              <label class="block text-sm font-bold text-slate-700 dark:text-slate-300 mb-1.5 px-1">Kata Sandi Baru</label>
              <div class="relative group">
                <span class="absolute inset-y-0 left-0 pl-4 flex items-center text-slate-400 group-focus-within:text-blue-600 transition-colors">
                  <base-icon :path="mdiLockOutline" size="20" />
                </span>
                <input 
                  v-model="form.password"
                  :type="form.showPassword ? 'text' : 'password'" 
                  placeholder="Masukkan kata sandi baru"
                  class="w-full pl-12 pr-12 py-3.5 bg-slate-50 dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-2xl focus:ring-4 focus:ring-blue-100 dark:focus:ring-blue-900/20 focus:border-blue-500 focus:bg-white dark:focus:bg-slate-900 outline-none transition-all placeholder:text-slate-400 text-base"
                  required
                  minlength="6"
                />
                <button 
                  type="button"
                  class="absolute inset-y-0 right-0 pr-4 flex items-center text-slate-300 hover:text-slate-500"
                  @click="form.showPassword = !form.showPassword"
                >
                  <base-icon :path="form.showPassword ? mdiEyeOff : mdiEye" size="20" />
                </button>
              </div>
            </div>

            <div>
              <label class="block text-sm font-bold text-slate-700 dark:text-slate-300 mb-1.5 px-1">Konfirmasi Kata Sandi Baru</label>
              <div class="relative group">
                <span class="absolute inset-y-0 left-0 pl-4 flex items-center text-slate-400 group-focus-within:text-blue-600 transition-colors">
                   <base-icon :path="mdiLockOutline" size="20" />
                </span>
                <input 
                  v-model="form.confirmPassword"
                  :type="form.showPassword ? 'text' : 'password'" 
                  placeholder="Ulangi kata sandi baru"
                  class="w-full pl-12 pr-4 py-3.5 bg-slate-50 dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-2xl focus:ring-4 focus:ring-blue-100 dark:focus:ring-blue-900/20 focus:border-blue-500 focus:bg-white dark:focus:bg-slate-900 outline-none transition-all placeholder:text-slate-400 text-base"
                  required
                />
              </div>
            </div>

            <button 
              type="submit" 
              class="w-full py-4 bg-blue-600 hover:bg-blue-700 disabled:bg-blue-400 text-white font-bold rounded-2xl shadow-lg shadow-blue-200 dark:shadow-none transition-all transform text-base"
              :disabled="isLoading || !token"
            >
              {{ isLoading ? 'Menyimpan...' : 'Simpan Kata Sandi' }}
            </button>
          </form>

          <!-- Back to login -->
          <div v-if="!successMessage" class="mt-8 text-center">
            <router-link 
              to="/login"
              class="inline-flex items-center gap-2 text-slate-500 hover:text-blue-600 transition-colors text-sm font-medium"
            >
              <base-icon :path="mdiArrowLeft" size="18" />
              Batal dan kembali ke login
            </router-link>
          </div>
        </div>
      </div>
    </SectionFullScreen>
  </LayoutGuest>
</template>
