<script setup>
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { mdiEmailOutline, mdiSend, mdiArrowLeft } from '@mdi/js'
import BaseIcon from '@/components/BaseIcon.vue'
import SectionFullScreen from '@/components/SectionFullScreen.vue'
import LayoutGuest from '@/layouts/LayoutGuest.vue'
import { api } from '@/services/api.js'

const router = useRouter()

const form = reactive({
  email: ''
})

const isLoading = ref(false)
const errorMessage = ref('')
const successMessage = ref('')

const submit = async () => {
  isLoading.value = true
  errorMessage.value = ''
  successMessage.value = ''
  
  try {
    const { data } = await api.post('/api/v1/auth/forgot-password', {
      email: form.email
    })
    successMessage.value = data.message || 'Tautan reset kata sandi telah dikirim ke email Anda.'
  } catch (err) {
    errorMessage.value = err.response?.data?.error || 'Gagal mengirim permintaan reset kata sandi. Pastikan email terdaftar.'
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
            <h1 class="text-3xl font-extrabold text-blue-600 dark:text-blue-500 mb-3">Lupa Kata Sandi</h1>
            <p class="text-slate-500 dark:text-slate-400 text-center max-w-[320px] leading-relaxed text-base">
              Masukkan email Anda dan kami akan mengirimkan proses untuk mengatur ulang kata sandi Anda.
            </p>
          </div>

          <!-- Notification -->
          <div v-if="errorMessage" class="mb-5 rounded-xl bg-red-50 dark:bg-red-900/20 px-4 py-3 text-sm text-red-700 dark:text-red-400 border border-red-100 dark:border-red-900/40 text-center font-medium">
            {{ errorMessage }}
          </div>
          <div v-if="successMessage" class="mb-5 rounded-xl bg-emerald-50 dark:bg-emerald-900/20 px-4 py-3 text-sm text-emerald-700 dark:text-emerald-400 border border-emerald-100 dark:border-emerald-900/40 text-center font-medium">
            {{ successMessage }}
          </div>

          <!-- Form -->
          <form @submit.prevent="submit" class="space-y-6">
            <div>
              <div class="relative group">
                <span class="absolute inset-y-0 left-0 pl-4 flex items-center text-slate-400 group-focus-within:text-blue-600 transition-colors">
                  <base-icon :path="mdiEmailOutline" size="20" />
                </span>
                <input 
                  v-model="form.email"
                  type="email" 
                  placeholder="Alamat Email"
                  class="w-full pl-12 pr-4 py-3.5 bg-slate-50 dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-2xl focus:ring-4 focus:ring-blue-100 dark:focus:ring-blue-900/20 focus:border-blue-500 focus:bg-white dark:focus:bg-slate-900 outline-none transition-all placeholder:text-slate-400 text-base"
                  required
                />
              </div>
            </div>

            <button 
              type="submit" 
              class="w-full py-4 bg-blue-600 hover:bg-blue-700 disabled:bg-blue-400 text-white font-bold rounded-2xl shadow-lg shadow-blue-200 dark:shadow-none transition-all transform flex items-center justify-center gap-2 text-base"
              :disabled="isLoading"
            >
              <base-icon :path="mdiSend" size="18" />
              {{ isLoading ? 'Mengirim...' : 'Kirim Tautan Reset Kata Sandi' }}
            </button>
          </form>

          <!-- Back to login -->
          <div class="mt-10 text-center">
            <router-link 
              to="/login"
              class="inline-flex items-center gap-2 text-blue-600 dark:text-blue-400 font-bold hover:underline transition-colors text-base"
            >
              <base-icon :path="mdiArrowLeft" size="20" />
              Kembali ke masuk
            </router-link>
          </div>
        </div>
      </div>
    </SectionFullScreen>
  </LayoutGuest>
</template>
