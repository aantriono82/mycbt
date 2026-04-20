<script setup>
import { onMounted, ref } from 'vue'
import { mdiCogOutline, mdiRefresh, mdiContentSave, mdiDatabaseExport, mdiDatabaseImport, mdiAlert, mdiEmailOutline, mdiWhatsapp, mdiInformationOutline } from '@mdi/js'
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
const restoreFile = ref(null)

const isBackingUp = ref(false)
const isRestoring = ref(false)
const isSavingSMTP = ref(false)
const isSavingWA = ref(false)

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
  username: '',
  password: '',
  from_name: '',
  from_email: '',
  encryption: 'tls',
})

const whatsappConfig = ref({
  api_provider: 'wagw', // default is 'wagw'
  api_url: '',
  api_token: '',
  sender_number: '',
})

const loadSettings = async () => {
  if (!authStore.isAuthenticated) return
  isLoading.value = true
  errorMessage.value = ''
  successMessage.value = ''
  try {
    const [identityRes, systemRes, smtpRes, waRes] = await Promise.all([
      api.get('/api/v1/settings/school-identity'),
      api.get('/api/v1/settings/system'),
      api.get('/api/v1/settings/smtp'),
      api.get('/api/v1/settings/whatsapp'),
    ])
    schoolIdentity.value = { ...schoolIdentity.value, ...(identityRes?.data?.data || {}) }
    systemSettings.value = { ...systemSettings.value, ...(systemRes?.data?.data || {}) }
    smtpConfig.value = { ...smtpConfig.value, ...(smtpRes?.data?.data || {}) }
    whatsappConfig.value = { ...whatsappConfig.value, ...(waRes?.data?.data || {}) }
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
      ...smtpConfig.value,
      port: Number(smtpConfig.value.port),
    }
    await api.put('/api/v1/settings/smtp', payload)
    successMessage.value = 'Pengaturan SMTP berhasil disimpan.'
  } catch (error) {
    errorMessage.value = error?.response?.data?.error?.message || 'Gagal menyimpan SMTP'
  } finally {
    isSavingSMTP.value = false
  }
}

const saveWhatsApp = async () => {
  if (!authStore.isAuthenticated) return
  isSavingWA.value = true
  errorMessage.value = ''
  successMessage.value = ''
  try {
    await api.put('/api/v1/settings/whatsapp', whatsappConfig.value)
    successMessage.value = 'Pengaturan WhatsApp berhasil disimpan.'
  } catch (error) {
    errorMessage.value = error?.response?.data?.error?.message || 'Gagal menyimpan WhatsApp'
  } finally {
    isSavingWA.value = false
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

const downloadBackup = async () => {
  if (!authStore.isAuthenticated) return
  isBackingUp.value = true
  errorMessage.value = ''
  successMessage.value = ''
  try {
    const response = await api.get('/api/v1/maintenance/backup', {
      responseType: 'blob',
    })
    
    // Create a download link
    const url = window.URL.createObjectURL(new Blob([response.data]))
    const link = document.createElement('a')
    link.href = url
    
    const contentDisposition = response.headers['content-disposition']
    let filename = `mycbt_backup_${new Date().toISOString().slice(0, 10)}.sql`
    if (contentDisposition) {
      const filenameMatch = contentDisposition.match(/filename=(.+)/)
      if (filenameMatch.length > 1) {
        filename = filenameMatch[1]
      }
    }
    
    link.setAttribute('download', filename)
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    
    successMessage.value = 'Database backup berhasil diunduh.'
  } catch (error) {
    errorMessage.value = 'Gagal melakukan backup database.'
    console.error(error)
  } finally {
    isBackingUp.value = false
  }
}

const restoreDatabase = async () => {
  if (!authStore.isAuthenticated || !restoreFile.value) return
  
  if (!confirm('PERINGATAN: Restore akan menimpa database saat ini. Pastikan Anda sudah memiliki backup terbaru. Lanjutkan?')) {
    return
  }

  isRestoring.value = true
  errorMessage.value = ''
  successMessage.value = ''
  try {
    const formData = new FormData()
    formData.append('file', restoreFile.value)
    await api.post('/api/v1/maintenance/restore', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    })
    successMessage.value = 'Database berhasil di-restore. Silakan muat ulang halaman.'
    restoreFile.value = null
  } catch (error) {
    errorMessage.value = error?.response?.data?.error?.message || 'Gagal restore database'
    if (error?.response?.data?.error?.detail) {
      console.error('Restore Error Detail:', error.response.data.error.detail)
    }
  } finally {
    isRestoring.value = false
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

      <div class="grid gap-6 xl:grid-cols-2">
        <CardBox>
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

        <CardBox>
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

        <CardBox>
          <div class="flex items-center gap-2 mb-4">
            <BaseIcon :path="mdiEmailOutline" class="text-blue-500" />
            <h3 class="text-lg font-semibold dark:text-slate-100 uppercase tracking-tight">Konfigurasi Email (SMTP)</h3>
          </div>
          <div class="grid grid-cols-1 md:grid-cols-2 gap-x-4">
            <FormField label="SMTP Host">
              <FormControl v-model="smtpConfig.host" placeholder="smtp.gmail.com" />
            </FormField>
            <FormField label="SMTP Port">
              <FormControl v-model="smtpConfig.port" type="number" placeholder="587" />
            </FormField>
            <FormField label="Username / Email">
              <FormControl v-model="smtpConfig.username" placeholder="user@gmail.com" />
            </FormField>
            <FormField label="Password / App Password">
              <FormControl v-model="smtpConfig.password" type="password" placeholder="••••••••" />
            </FormField>
            <FormField label="From Name">
              <FormControl v-model="smtpConfig.from_name" placeholder="AtigaCBT Notifikasi" />
            </FormField>
            <FormField label="From Email">
              <FormControl v-model="smtpConfig.from_email" placeholder="noreply@gmail.com" />
            </FormField>
            <FormField label="Encryption">
              <FormControl
                v-model="smtpConfig.encryption"
                :options="[
                  { value: 'tls', label: 'TLS (StartTLS/587)' },
                  { value: 'ssl', label: 'SSL (Implicit/465)' },
                  { value: 'none', label: 'None (25/8025)' },
                ]"
              />
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

        <CardBox>
          <div class="flex items-center gap-2 mb-4">
            <BaseIcon :path="mdiWhatsapp" class="text-emerald-500" />
            <h3 class="text-lg font-semibold dark:text-slate-100 uppercase tracking-tight">Konfigurasi WhatsApp Gateway</h3>
          </div>
          
          <div class="mb-4 p-4 rounded-xl bg-slate-50 dark:bg-slate-800/40 border border-slate-100 dark:border-slate-800 flex gap-3">
             <BaseIcon :path="mdiInformationOutline" class="text-blue-500 shrink-0" size="20" />
             <p class="text-xs text-slate-600 dark:text-slate-400 leading-relaxed">
               Gunakan provider WhatsApp Gateway lokal (WAGW/WAMD). <br/>
               Format Request: POST ke <code>API URL</code> dengan payload JSON: <code>{"target": "number", "message": "msg"}</code> dan header Auth: <code>Authorization: api_token</code>.
             </p>
          </div>

          <FormField label="Provider Template">
            <FormControl
              v-model="whatsappConfig.api_provider"
              :options="[
                { value: 'wagw', label: 'Custom HTTP API (Standard)' },
                { value: 'wa-local', label: 'Local Server / Local Provider' },
              ]"
            />
          </FormField>

          <FormField label="API URL Endpoint">
            <FormControl v-model="whatsappConfig.api_url" placeholder="https://api.wagw.com/send-message" />
          </FormField>

          <FormField label="API Token / Key">
            <FormControl v-model="whatsappConfig.api_token" type="password" placeholder="Bearer your-token-here" />
          </FormField>

          <FormField label="Sender Number (Optional)">
            <FormControl v-model="whatsappConfig.sender_number" placeholder="62812xxxx" />
          </FormField>

          <div class="flex items-center gap-3 mt-6">
            <BaseButton
              :icon="mdiContentSave"
              color="info"
              label="Simpan WhatsApp"
              :disabled="isLoading || isSavingWA"
              @click="saveWhatsApp"
            />
            <div v-if="isSavingWA" class="text-sm text-slate-500 dark:text-slate-400 italic">Menyimpan...</div>
          </div>
        </CardBox>

        <CardBox class="xl:col-span-2">
          <div class="flex items-center gap-2 mb-4">
             <h3 class="text-lg font-semibold dark:text-slate-100 uppercase tracking-tight">Maintenance (Backup & Restore)</h3>
             <span class="px-2 py-0.5 bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400 rounded text-[10px] uppercase font-bold tracking-widest">Admin Only</span>
          </div>

          <div class="grid gap-6 md:grid-cols-2">
            <div class="rounded-2xl border border-slate-100 p-6 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-900/50">
               <div class="flex items-center gap-4 mb-4">
                  <div class="flex h-12 w-12 items-center justify-center rounded-2xl bg-blue-100 text-blue-600 dark:bg-blue-900/30 dark:text-blue-400">
                     <BaseIcon :path="mdiDatabaseExport" size="24" />
                  </div>
                  <div>
                    <h4 class="font-bold dark:text-slate-100">Ekspor Database</h4>
                    <p class="text-xs text-slate-500">Unduh seluruh data dalam format SQL.</p>
                  </div>
               </div>
               <BaseButton
                 :icon="mdiDatabaseExport"
                 color="info"
                 label="Download Backup (.sql)"
                 :disabled="isBackingUp || isRestoring"
                 @click="downloadBackup"
                 class="w-full"
               />
               <p v-if="isBackingUp" class="mt-2 text-center text-xs text-blue-600 animate-pulse font-bold uppercase tracking-widest">Mengekspor data...</p>
            </div>

            <div class="rounded-2xl border border-slate-100 p-6 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-900/50">
               <div class="flex items-center gap-4 mb-4">
                  <div class="flex h-12 w-12 items-center justify-center rounded-2xl bg-amber-100 text-amber-600 dark:bg-amber-900/30 dark:text-amber-400">
                     <BaseIcon :path="mdiDatabaseImport" size="24" />
                  </div>
                  <div>
                    <h4 class="font-bold dark:text-slate-100">Impor Database</h4>
                    <p class="text-xs text-slate-500 uppercase tracking-tight font-black text-amber-600">⚠ BERBAHAYA: Timpa Data</p>
                  </div>
               </div>
               <div class="flex flex-col gap-3">
                  <FormFilePicker v-model="restoreFile" label="Pilih File SQL" accept=".sql" class="w-full" />
                  <BaseButton
                    :icon="mdiAlert"
                    color="danger"
                    label="Restore Database Sekarang"
                    :disabled="!restoreFile || isRestoring || isBackingUp"
                    @click="restoreDatabase"
                    class="w-full"
                  />
               </div>
               <p v-if="isRestoring" class="mt-2 text-center text-xs text-amber-600 animate-pulse font-bold uppercase tracking-widest">Memulihkan data (Jangan tutup tab ini)...</p>
            </div>
          </div>
        </CardBox>
      </div>
    </SectionMain>
  </LayoutAuthenticated>
</template>
