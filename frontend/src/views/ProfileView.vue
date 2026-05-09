<script setup>
import { computed, reactive, ref } from 'vue'
import { useRoute } from 'vue-router'
import {
  mdiAccount,
  mdiMail,
  mdiAsterisk,
  mdiFormTextboxPassword,
  mdiCamera,
  mdiUpload,
} from '@mdi/js'
import SectionMain from '@/components/SectionMain.vue'
import CardBox from '@/components/CardBox.vue'
import BaseDivider from '@/components/BaseDivider.vue'
import FormField from '@/components/FormField.vue'
import FormControl from '@/components/FormControl.vue'
import PasswordField from '@/components/PasswordField.vue'
import BaseButton from '@/components/BaseButton.vue'
import BaseButtons from '@/components/BaseButtons.vue'
import BaseIcon from '@/components/BaseIcon.vue'
import LayoutAuthenticated from '@/layouts/LayoutAuthenticated.vue'
import SectionTitleLineWithButton from '@/components/SectionTitleLineWithButton.vue'
import { useAuthStore } from '@/stores/auth.js'
import { useNotificationStore } from '@/stores/notification.js'
import { api } from '@/services/api.js'
import { resolveBackendAssetUrl } from '@/utils/assetUrl.js'

const route = useRoute()
const authStore = useAuthStore()
const notificationStore = useNotificationStore()

const title = computed(() => route.meta?.title || 'Profil')
const roleLabel = computed(() => authStore.roleLabel)

const userAvatar = computed(() => {
  const user = authStore.user
  const photo = user?.photo_url || user?.photo || user?.avatar
  if (photo) return resolveBackendAssetUrl(photo)
  return `https://api.dicebear.com/7.x/initials/svg?seed=${user?.username || 'Admin'}&backgroundColor=0033ff`
})

const profileForm = reactive({
  name: authStore.user?.name || '',
  email: authStore.user?.email || '',
  username: authStore.user?.username || '',
})

const passwordForm = reactive({
  password_current: '',
  password: '',
  password_confirmation: '',
})

const photoFile = ref(null)
const photoPreview = ref(null)
const isUploading = ref(false)
const isSavingProfile = ref(false)
const isChangingPassword = ref(false)

const handleFileUpload = (e) => {
  const file = e.target.files[0]
  if (file) {
    photoFile.value = file
    photoPreview.value = URL.createObjectURL(file)
  }
}

const uploadPhoto = async () => {
  if (!photoFile.value) return
  isUploading.value = true
  
  try {
    const formData = new FormData()
    formData.append('file', photoFile.value)
    
    await api.post('/api/v1/me/photo', formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    })
    
    // Refresh user data to update avatar everywhere
    await authStore.loadMe()
    
    notificationStore.pushSuccess('Foto profil berhasil diperbarui!')
    photoFile.value = null
    photoPreview.value = null
  } catch (error) {
    console.error('Upload failed:', error)
    notificationStore.pushError(error?.response?.data?.error?.message || 'Gagal mengunggah foto profil')
  } finally {
    isUploading.value = false
  }
}

const submitProfile = async () => {
  const name = String(profileForm.name || '').trim()
  const email = String(profileForm.email || '').trim()
  if (!name) {
    notificationStore.pushWarning('Nama wajib diisi')
    return
  }

  isSavingProfile.value = true
  try {
    await api.put('/api/v1/me', { name, email })
    await authStore.loadMe()
    notificationStore.pushSuccess('Profil berhasil diperbarui')
  } catch (error) {
    notificationStore.pushError(error?.response?.data?.error?.message || 'Gagal menyimpan profil')
  } finally {
    isSavingProfile.value = false
  }
}

const submitPass = async () => {
  const current = String(passwordForm.password_current || '')
  const next = String(passwordForm.password || '')
  const confirm = String(passwordForm.password_confirmation || '')

  if (!current || !next || !confirm) {
    notificationStore.pushWarning('Semua field password wajib diisi')
    return
  }
  if (next.length < 8) {
    notificationStore.pushWarning('Password baru minimal 8 karakter')
    return
  }
  if (next !== confirm) {
    notificationStore.pushWarning('Konfirmasi password tidak sama')
    return
  }

  isChangingPassword.value = true
  try {
    await api.post('/api/v1/me/password', {
      current_password: current,
      new_password: next,
      password_confirmation: confirm,
    })
    passwordForm.password_current = ''
    passwordForm.password = ''
    passwordForm.password_confirmation = ''
    notificationStore.pushSuccess('Password berhasil diubah')
  } catch (error) {
    notificationStore.pushError(error?.response?.data?.error?.message || 'Gagal mengubah password')
  } finally {
    isChangingPassword.value = false
  }
}
</script>

<template>
  <LayoutAuthenticated>
    <SectionMain>
      <SectionTitleLineWithButton :icon="mdiAccount" :title="title" main />

      <CardBox class="mb-6" color="blue">
        <div class="flex flex-col md:flex-row items-center gap-8 py-4">
          <!-- Photo Section -->
          <div class="relative group">
            <div class="h-32 w-32 overflow-hidden rounded-full border-4 border-blue-50 dark:border-slate-800 bg-slate-100 dark:bg-slate-800 shadow-lg ring-4 ring-white dark:ring-slate-900">
              <img :src="photoPreview || userAvatar" class="h-full w-full object-cover transition-opacity group-hover:opacity-75" alt="Profile" />
            </div>
            <label class="absolute bottom-1 right-1 flex h-10 w-10 cursor-pointer items-center justify-center rounded-full bg-blue-600 text-white shadow-lg transition-all hover:scale-110 hover:bg-blue-700 ring-4 ring-white dark:ring-slate-900">
              <BaseIcon :path="mdiCamera" size="20" />
              <input type="file" class="hidden" accept="image/*" @change="handleFileUpload" />
            </label>
            <div v-if="photoFile" class="absolute -bottom-12 left-1/2 -translate-x-1/2 whitespace-nowrap">
              <BaseButton 
                color="info" 
                :label="isUploading ? 'Mengunggah...' : 'Simpan Foto'" 
                :icon="mdiUpload" 
                small 
                :disabled="isUploading"
                @click="uploadPhoto" 
              />
            </div>
          </div>
          
          <!-- Info Section -->
          <div class="flex-1 text-center md:text-left">
            <h2 class="text-3xl font-bold text-gray-800 dark:text-slate-100">{{ profileForm.name || 'Pengguna' }}</h2>
            <div class="flex flex-wrap items-center justify-center md:justify-start gap-3 mt-2">
              <span class="px-3 py-1 bg-blue-100 dark:bg-blue-900/40 text-blue-700 dark:text-blue-400 rounded-full text-xs font-bold uppercase tracking-wider">{{ roleLabel }}</span>
              <span class="text-slate-400">|</span>
              <span class="text-slate-500 font-medium tracking-tight">@{{ profileForm.username || '-' }}</span>
            </div>
            <p class="text-sm text-slate-400 mt-4 italic">Format foto: JPG, PNG. Maksimal: 2MB.</p>
          </div>

          <div class="hidden lg:block rounded-2xl bg-slate-50 dark:bg-slate-800/40 p-6 border border-emerald-400/60 dark:border-emerald-800/80 max-w-xs transition-colors hover:bg-blue-50/50 dark:hover:bg-blue-900/20">
            <h4 class="font-bold text-slate-700 dark:text-slate-200 mb-1 leading-tight">Keamanan Sesi</h4>
            <p class="text-xs text-slate-500 dark:text-slate-400 leading-relaxed">Data profil Anda dikelola melalui sesi JWT terenkripsi. Pastikan data akun tetap rahasia.</p>
          </div>
        </div>
      </CardBox>

      <div class="grid grid-cols-1 gap-6 lg:grid-cols-2">
        <CardBox is-form @submit.prevent="submitProfile" color="purple">
          <FormField label="Username">
            <FormControl
              v-model="profileForm.username"
              :icon="mdiAccount"
              name="username"
              disabled
            />
          </FormField>

          <FormField label="Nama">
            <FormControl
              v-model="profileForm.name"
              :icon="mdiAccount"
              name="name"
              autocomplete="name"
            />
          </FormField>

          <FormField label="E-mail">
            <FormControl
              v-model="profileForm.email"
              :icon="mdiMail"
              type="email"
              name="email"
              autocomplete="email"
            />
          </FormField>

          <template #footer>
            <BaseButtons>
              <BaseButton color="info" type="submit" :label="isSavingProfile ? 'Menyimpan...' : 'Simpan Profil'" :disabled="isSavingProfile" />
            </BaseButtons>
          </template>
        </CardBox>

        <CardBox is-form @submit.prevent="submitPass" color="blue">
          <FormField label="Password Saat Ini">
            <PasswordField
              v-model="passwordForm.password_current"
              :icon="mdiAsterisk"
              name="password_current"
              autocomplete="current-password"
            />
          </FormField>

          <BaseDivider />

          <FormField label="Password Baru">
            <PasswordField
              v-model="passwordForm.password"
              :icon="mdiFormTextboxPassword"
              name="password"
              autocomplete="new-password"
            />
          </FormField>

          <FormField label="Konfirmasi Password">
            <PasswordField
              v-model="passwordForm.password_confirmation"
              :icon="mdiFormTextboxPassword"
              name="password_confirmation"
              autocomplete="new-password"
            />
          </FormField>

          <template #footer>
            <BaseButtons>
              <BaseButton type="submit" color="info" :label="isChangingPassword ? 'Memproses...' : 'Ubah Password'" :disabled="isChangingPassword" />
            </BaseButtons>
          </template>
        </CardBox>
      </div>
    </SectionMain>
  </LayoutAuthenticated>
</template>
