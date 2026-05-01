<script setup>
import { onMounted, ref } from 'vue'
import { mdiCogOutline, mdiRefresh, mdiContentSave, mdiEmailOutline } from '@mdi/js'
import LayoutAuthenticated from '@/layouts/LayoutAuthenticated.vue'
import SectionMain from '@/components/SectionMain.vue'
import SectionTitleLineWithButton from '@/components/SectionTitleLineWithButton.vue'
import CardBox from '@/components/CardBox.vue'
import BaseButton from '@/components/BaseButton.vue'
import FormField from '@/components/FormField.vue'
import FormControl from '@/components/FormControl.vue'
import FormFilePicker from '@/components/FormFilePicker.vue'
import { api } from '@/services/api.js'
import { useAuthStore } from '@/stores/auth.js'

const authStore = useAuthStore()

const isLoading = ref(false)
const isSavingIdentity = ref(false)
const isSavingSystem = ref(false)
const isUploadingLogo = ref(false)
const errorMessage = ref('')
const successMessage = ref('')
const logoFile = ref(null)
const isSavingSMTP = ref(false)

const schoolIdentity = ref({
  school_name: '',
  address: '',
  phone: '',
  email: '',
  website: '',
  principal_name: '',
  logo_url: '',
})

const systemSettings = ref({
  timezone: 'Asia/Jakarta',
  token_required: true,
  allow_reset_login: true,
  max_active_sessions: 1,
  attendance_require_ip: false,
})

const smtpConfig = ref({
  host: '',
  port: 587,
  user: '',
  password: '',
  from: '',
  use_tls: true,
})

const loadSettings = async () => {
  if (!authStore.isAuthenticated) return
  isLoading.value = true
  errorMessage.value = ''
  successMessage.value = ''
  try {
    const [identityRes, systemRes, smtpRes] = await Promise.all([
      api.get('/api/v1/settings/school-identity'),
      api.get('/api/v1/settings/system'),
      api.get('/api/v1/settings/smtp'),
    ])
    schoolIdentity.value = { ...schoolIdentity.value, ...(identityRes?.data?.data || {}) }
    systemSettings.value = { ...systemSettings.value, ...(systemRes?.data?.data || {}) }
    const smtpData = smtpRes?.data?.data || {}
    smtpConfig.value = {
      ...smtpConfig.value,
      ...smtpData,
      port: Number(smtpData.port || smtpConfig.value.port || 587),
      use_tls: typeof smtpData.use_tls === 'boolean' ? smtpData.use_tls : smtpConfig.value.use_tls,
    }
  } catch (error) {
    errorMessage.value = error?.response?.data?.error?.message || 'Gagal memuat settings'
  } finally {
    isLoading.value = false
  }
}

const saveSMTP = async () => {
  if (!authStore.isAuthenticated) return
  isSavingSMTP.value = true
  errorMessage.value = ''
  successMessage.value = ''
  try {
    const payload = {
      host: String(smtpConfig.value.host || '').trim(),
      port: Number(smtpConfig.value.port),
      user: String(smtpConfig.value.user || '').trim(),
      password: String(smtpConfig.value.password || ''),
      from: String(smtpConfig.value.from || '').trim(),
      use_tls: !!smtpConfig.value.use_tls,
    }
    await api.put('/api/v1/settings/smtp', payload)
    successMessage.value = 'Pengaturan SMTP berhasil disimpan.'
  } catch (error) {
    errorMessage.value = error?.response?.data?.error?.message || 'Gagal menyimpan SMTP'
  } finally {
    isSavingSMTP.value = false
  }
}

const saveSchoolIdentity = async () => {
  if (!authStore.isAuthenticated) return
  isSavingIdentity.value = true
  errorMessage.value = ''
  successMessage.value = ''
  try {
    const payload = {
      school_name: schoolIdentity.value.school_name?.trim(),
      address: schoolIdentity.value.address?.trim(),
      phone: schoolIdentity.value.phone?.trim(),
      email: schoolIdentity.value.email?.trim(),
      website: schoolIdentity.value.website?.trim(),
      principal_name: schoolIdentity.value.principal_name?.trim(),
      logo_url: schoolIdentity.value.logo_url?.trim(),
    }
    const { data } = await api.put('/api/v1/settings/school-identity', payload)
    schoolIdentity.value = { ...schoolIdentity.value, ...(data?.data || {}) }
    successMessage.value = 'Identitas sekolah berhasil disimpan.'
  } catch (error) {
    errorMessage.value = error?.response?.data?.error?.message || 'Gagal menyimpan identitas sekolah'
  } finally {
    isSavingIdentity.value = false
  }
}

const uploadSchoolLogo = async () => {
  if (!authStore.isAuthenticated || !logoFile.value) return
  isUploadingLogo.value = true
  errorMessage.value = ''
  successMessage.value = ''
  try {
    const formData = new FormData()
    formData.append('file', logoFile.value)
    const { data } = await api.post('/api/v1/settings/school-identity/logo', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    })
    schoolIdentity.value.logo_url = data?.data?.logo_url || schoolIdentity.value.logo_url
    logoFile.value = null
    successMessage.value = 'Logo sekolah berhasil diunggah.'
  } catch (error) {
    errorMessage.value = error?.response?.data?.error?.message || 'Gagal upload logo sekolah'
  } finally {
    isUploadingLogo.value = false
  }
}

const saveSystemSettings = async () => {
  if (!authStore.isAuthenticated) return
  isSavingSystem.value = true
  errorMessage.value = ''
  successMessage.value = ''
  try {
    const payload = {
      timezone: systemSettings.value.timezone?.trim() || 'Asia/Jakarta',
      token_required: !!systemSettings.value.token_required,
      allow_reset_login: !!systemSettings.value.allow_reset_login,
      max_active_sessions: Math.max(1, Number(systemSettings.value.max_active_sessions || 1)),
      attendance_require_ip: !!systemSettings.value.attendance_require_ip,
    }
    const { data } = await api.put('/api/v1/settings/system', payload)
    systemSettings.value = { ...systemSettings.value, ...(data?.data || {}) }
    successMessage.value = 'Pengaturan sistem berhasil disimpan.'
  } catch (error) {
    errorMessage.value = error?.response?.data?.error?.message || 'Gagal menyimpan pengaturan sistem'
  } finally {
    isSavingSystem.value = false
  }
}

onMounted(loadSettings)
</script>

<template>
  <LayoutAuthenticated>
    <SectionMain>
      <SectionTitleLineWithButton :icon="mdiCogOutline" title="Config / Settings" main>
        <BaseButton :icon="mdiRefresh" color="info" label="Refresh" @click="loadSettings" />
      </SectionTitleLineWithButton>

      <div v-if="!authStore.isAuthenticated" class="mb-6 rounded-xl bg-amber-50 dark:bg-amber-900/20 px-4 py-3 text-sm text-amber-700 dark:text-amber-400 border border-amber-100 dark:border-amber-900/40">
        Login terlebih dulu agar settings bisa dimuat.
      </div>
      <div v-else-if="errorMessage" class="mb-6 rounded-xl bg-red-50 dark:bg-red-900/20 px-4 py-3 text-sm text-red-700 dark:text-red-400 border border-red-100 dark:border-red-900/40">
        {{ errorMessage }}
      </div>
      <div v-else-if="successMessage" class="mb-6 rounded-xl bg-emerald-50 dark:bg-emerald-900/20 px-4 py-3 text-sm text-emerald-700 dark:text-emerald-400 border border-emerald-100 dark:border-emerald-900/40">
        {{ successMessage }}
      </div>

      <div class="grid gap-6 xl:grid-cols-12">
        <CardBox class="xl:col-span-6">
          <h3 class="mb-4 text-lg font-semibold dark:text-slate-100">Identitas Sekolah</h3>
          <FormField label="Nama Sekolah">
            <FormControl v-model="schoolIdentity.school_name" placeholder="Nama sekolah" />
          </FormField>
          <FormField label="Alamat">
            <FormControl v-model="schoolIdentity.address" type="textarea" placeholder="Alamat lengkap" />
          </FormField>
          <FormField label="Telepon">
            <FormControl v-model="schoolIdentity.phone" placeholder="021-xxxx" />
          </FormField>
          <FormField label="Email">
            <FormControl v-model="schoolIdentity.email" placeholder="admin@sekolah.sch.id" />
          </FormField>
          <FormField label="Website">
            <FormControl v-model="schoolIdentity.website" placeholder="https://sekolah.sch.id" />
          </FormField>
          <FormField label="Nama Kepala Sekolah">
            <FormControl v-model="schoolIdentity.principal_name" placeholder="Nama kepala sekolah" />
          </FormField>
          <FormField label="URL Logo">
            <FormControl v-model="schoolIdentity.logo_url" placeholder="https://.../logo.png" />
          </FormField>
          <FormField label="Upload Logo (png/jpg/webp)">
            <div class="flex flex-wrap items-center gap-3">
              <FormFilePicker v-model="logoFile" label="Pilih File Logo" accept=".png,.jpg,.jpeg,.webp" />
              <BaseButton
                :icon="mdiContentSave"
                color="whiteDark"
                outline
                label="Upload Logo"
                :disabled="isLoading || isUploadingLogo || !logoFile"
                @click="uploadSchoolLogo"
              />
              <div v-if="isUploadingLogo" class="text-sm text-slate-500 dark:text-slate-400 italic">Mengunggah logo...</div>
            </div>
            <div v-if="schoolIdentity.logo_url" class="mt-3">
              <img
                :src="schoolIdentity.logo_url"
                alt="Logo sekolah"
                class="h-16 w-16 rounded border border-slate-200 dark:border-slate-800 object-contain p-1 bg-white dark:bg-transparent"
              />
            </div>
          </FormField>
          <div class="flex items-center gap-3">
            <BaseButton
              :icon="mdiContentSave"
              color="info"
              label="Simpan Identitas"
              :disabled="isLoading || isSavingIdentity"
              @click="saveSchoolIdentity"
            />
            <div v-if="isLoading || isSavingIdentity" class="text-sm text-slate-500 dark:text-slate-400 italic">Menyimpan...</div>
          </div>
        </CardBox>

        <CardBox class="xl:col-span-6">
          <h3 class="mb-4 text-lg font-semibold dark:text-slate-100">Pengaturan Sistem</h3>
          <FormField label="Timezone">
            <FormControl
              v-model="systemSettings.timezone"
              :options="[
                { value: 'Asia/Jakarta', label: 'Asia/Jakarta (WIB)' },
                { value: 'Asia/Makassar', label: 'Asia/Makassar (WITA)' },
                { value: 'Asia/Jayapura', label: 'Asia/Jayapura (WIT)' },
              ]"
            />
          </FormField>

          <FormField label="Maksimal Sesi Aktif per Siswa">
            <FormControl v-model="systemSettings.max_active_sessions" type="number" />
          </FormField>

          <div class="mb-4 space-y-2 rounded-xl border border-slate-200 dark:border-slate-800 p-4 bg-slate-50/50 dark:bg-slate-800/20">
            <label class="flex items-center gap-3 text-sm dark:text-slate-300">
              <input v-model="systemSettings.token_required" type="checkbox" />
              Token ujian wajib saat join
            </label>
            <label class="flex items-center gap-3 text-sm dark:text-slate-300">
              <input v-model="systemSettings.allow_reset_login" type="checkbox" />
              Izinkan reset login dari menu monitor
            </label>
            <label class="flex items-center gap-3 text-sm dark:text-slate-300">
              <input v-model="systemSettings.attendance_require_ip" type="checkbox" />
              Catat kebijakan absensi berbasis IP
            </label>
          </div>

          <div class="flex items-center gap-3">
            <BaseButton
              :icon="mdiContentSave"
              color="info"
              label="Simpan Sistem"
              :disabled="isLoading || isSavingSystem"
              @click="saveSystemSettings"
            />
            <div v-if="isLoading || isSavingSystem" class="text-sm text-slate-500 dark:text-slate-400 italic">Menyimpan...</div>
          </div>
        </CardBox>

        <CardBox class="xl:col-span-12">
          <div class="mb-4 flex items-center gap-2">
            <BaseIcon :path="mdiEmailOutline" class="text-blue-500" />
            <h3 class="text-lg font-semibold dark:text-slate-100">Konfigurasi Email (SMTP)</h3>
          </div>
          <div class="grid grid-cols-1 gap-x-4 md:grid-cols-2">
            <FormField label="SMTP Host">
              <FormControl v-model="smtpConfig.host" placeholder="smtp.gmail.com" />
            </FormField>
            <FormField label="SMTP Port">
              <FormControl v-model="smtpConfig.port" type="number" placeholder="587" />
            </FormField>
            <FormField label="Username / Email">
              <FormControl v-model="smtpConfig.user" placeholder="user@gmail.com" />
            </FormField>
            <FormField label="Password / App Password">
              <FormControl v-model="smtpConfig.password" type="password" placeholder="••••••••" />
            </FormField>
            <FormField label="From Email">
              <FormControl v-model="smtpConfig.from" placeholder="noreply@gmail.com" />
            </FormField>
            <FormField label="Keamanan Koneksi">
              <label class="flex items-center gap-3 text-sm dark:text-slate-300">
                <input v-model="smtpConfig.use_tls" type="checkbox" />
                Gunakan TLS/STARTTLS
              </label>
            </FormField>
          </div>
          <div class="flex items-center gap-3 mt-4">
            <BaseButton
              :icon="mdiContentSave"
              color="info"
              label="Simpan SMTP"
              :disabled="isLoading || isSavingSMTP"
              @click="saveSMTP"
            />
            <div v-if="isSavingSMTP" class="text-sm text-slate-500 dark:text-slate-400 italic">Menyimpan...</div>
          </div>
        </CardBox>

      </div>
    </SectionMain>
  </LayoutAuthenticated>
</template>
