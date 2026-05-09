<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { mdiAccountSchool, mdiRefresh, mdiPlus, mdiDelete, mdiPencil, mdiContentSave, mdiEye, mdiEyeOff, mdiContentCopy, mdiFileExcel, mdiDownload, mdiUpload, mdiAccountSwitch, mdiImageMultiple, mdiClose, mdiCheckCircle, mdiAlertCircle, mdiHelpCircle, mdiMinus, mdiCamera } from '@mdi/js'
import LayoutAuthenticated from '@/layouts/LayoutAuthenticated.vue'
import SectionMain from '@/components/SectionMain.vue'
import SectionTitleLineWithButton from '@/components/SectionTitleLineWithButton.vue'
import CardBox from '@/components/CardBox.vue'
import BaseEmptyState from '@/components/BaseEmptyState.vue'
import BaseButton from '@/components/BaseButton.vue'
import BaseButtons from '@/components/BaseButtons.vue'
import BaseIcon from '@/components/BaseIcon.vue'
import FormField from '@/components/FormField.vue'
import FormControl from '@/components/FormControl.vue'
import PasswordField from '@/components/PasswordField.vue'
import { api } from '@/services/api.js'
import { useNotificationStore } from '@/stores/notification.js'
import { shortCode2 } from '@/utils/shortCode.js'
import { resolveBackendAssetUrl } from '@/utils/assetUrl.js'
import { compressImage } from '@/utils/image.js'

const authStore = useAuthStore()
const notificationStore = useNotificationStore()

const students = ref([])
const meta = ref({ total: 0 })
const isLoading = ref(false)
const isSaving = ref(false)
const errorMessage = ref('')
const successMessage = ref('')
const query = ref('')
const editingId = ref('')
const editingUserId = ref('')

const photoFile = ref(null)
const photoPreview = ref(null)
const isUploadingPhoto = ref(false)

// Bulk photo upload
const showBulkPhotoModal = ref(false)
const bulkPhotoFile = ref(null)
const bulkPhotoFileName = ref('')
const bulkPhotoUploading = ref(false)
const bulkPhotoResult = ref(null)
const bulkPhotoError = ref('')

const programs = ref([])
const levels = ref([])
const groups = ref([])

const form = reactive({
  username: '',
  password: '',
  name: '',
  email: '',
  nis: '',
  participant_no: '',
  jenjang: '',
  program_id: '',
  level_id: '',
  group_id: '',
  phone: '',
  photo_url: '',
  is_active: true,
})

// Pemetaan jenjang -> rentang kelas
const JENJANG_KELAS = {
  SD:  [1, 2, 3, 4, 5, 6],
  SMP: [7, 8, 9],
  SMA: [10, 11, 12],
}

// Level yang ditampilkan bergantung pada jenjang yang dipilih
// Jika belum ada data level dengan kelas yang sesuai, tampilkan semua level (fallback)
const filteredLevels = computed(() => {
  if (!form.jenjang) return levels.value
  const allowed = JENJANG_KELAS[form.jenjang] || []
  const matched = levels.value.filter((l) => allowed.includes(l.kelas))
  // Fallback: jika tidak ada yang cocok (data belum punya kelas), tampilkan semua
  return matched.length > 0 ? matched : levels.value
})

// Reset program_id & level_id saat jenjang berubah
const onJenjangChange = () => {
  form.level_id = ''
  if (form.jenjang !== 'SMA') {
    form.program_id = ''
  }
}

const canLoad = computed(() => authStore.isAuthenticated)
const isEditing = computed(() => !!editingId.value)

const resetForm = () => {
  editingId.value = ''
  form.username = ''
  form.password = ''
  form.name = ''
  form.email = ''
  form.nis = ''
  form.participant_no = ''
  form.jenjang = ''
  form.program_id = ''
  form.level_id = ''
  form.group_id = ''
  form.phone = ''
  form.photo_url = ''
  form.is_active = true
  photoFile.value = null
  photoPreview.value = null
}

const loadLookups = async () => {
  if (!canLoad.value) return
  try {
    const [programRes, levelRes, groupRes] = await Promise.all([
      api.get('/api/v1/admin/programs'),
      api.get('/api/v1/admin/levels'),
      api.get('/api/v1/admin/groups'),
    ])
    programs.value = programRes?.data?.data || []
    levels.value = levelRes?.data?.data || []
    groups.value = groupRes?.data?.data || []
  } catch {
    programs.value = []
    levels.value = []
    groups.value = []
  }
}

const loadStudents = async () => {
  if (!canLoad.value) return
  isLoading.value = true
  errorMessage.value = ''
  try {
    const { data } = await api.get('/api/v1/admin/students', {
      params: {
        q: query.value,
        limit: 50,
        offset: 0,
      },
    })
    students.value = data?.data || []
    meta.value = data?.meta || { total: students.value.length }
  } catch (error) {
    errorMessage.value = error?.response?.data?.error?.message || 'Gagal memuat data siswa'
  } finally {
    isLoading.value = false
  }
}

const startEditStudent = (student) => {
  editingId.value = student.id
  form.username = student.username || ''
  form.password = ''
  form.name = student.name || ''
  form.email = student.email || ''
  form.nis = student.nis || ''
  form.participant_no = student.participant_no || ''
  form.jenjang = student.jenjang || ''
  form.program_id = student.program_id || ''
  form.level_id = student.level_id || ''
  form.group_id = student.group_id || ''
  form.phone = student.phone || ''
  form.photo_url = student.photo_url || ''
  form.is_active = !!student.is_active
  editingUserId.value = student.user_id || ''
  photoFile.value = null
  photoPreview.value = null
}

const saveStudent = async () => {
  successMessage.value = ''
  errorMessage.value = ''
  isSaving.value = true
  try {
    if (isEditing.value) {
      await api.patch(`/api/v1/admin/students/${editingId.value}`, {
        username: form.username,
        password: form.password || '',
        name: form.name,
        email: form.email,
        nis: form.nis,
        participant_no: form.participant_no,
        jenjang: form.jenjang,
        program_id: form.jenjang === 'SMA' ? form.program_id : '',
        level_id: form.level_id,
        group_id: form.group_id,
        phone: form.phone,
        is_active: form.is_active,
      })
      successMessage.value = 'Data siswa berhasil diperbarui'
    } else {
      await api.post('/api/v1/admin/students', {
        username: form.username,
        password: form.password,
        name: form.name,
        email: form.email,
        nis: form.nis,
        participant_no: form.participant_no,
        jenjang: form.jenjang,
        program_id: form.jenjang === 'SMA' ? form.program_id : '',
        level_id: form.level_id,
        group_id: form.group_id,
        phone: form.phone,
      })
      successMessage.value = 'Data siswa berhasil ditambahkan'
    }
    resetForm()
    await loadStudents()
  } catch (error) {
    errorMessage.value = error?.response?.data?.error?.message || 'Gagal menyimpan siswa'
  } finally {
    isSaving.value = false
  }
}

const deleteStudent = async (id) => {
  successMessage.value = ''
  errorMessage.value = ''
  try {
    await api.delete(`/api/v1/admin/students/${id}`)
    successMessage.value = 'Data siswa berhasil dihapus'
    if (editingId.value === id) {
      resetForm()
    }
    await loadStudents()
  } catch (error) {
    errorMessage.value = error?.response?.data?.error?.message || 'Gagal menghapus siswa'
  }
}

const copyStudent = (student) => {
  resetForm()
  form.name = `${student.name} (Copy)`
  form.email = student.email
  form.nis = student.nis
  form.participant_no = student.participant_no
  form.jenjang = student.jenjang
  form.program_id = student.program_id
  form.level_id = student.level_id
  form.group_id = student.group_id
  form.phone = student.phone
  
  successMessage.value = 'Data siswa disalin ke form. Silakan isi username & password baru.'
}

const switchRoleStudent = async (student) => {
  const ok = window.confirm(`Ubah role "${student.name}" dari SISWA menjadi GURU?\nPastikan data pendukung (NIP, mapel, dll) diisi setelahnya.`)
  if (!ok) return
  successMessage.value = ''
  errorMessage.value = ''
  try {
    await api.post(`/api/v1/admin/users/${student.user_id}/switch-role`)
    successMessage.value = `Role "${student.name}" berhasil diubah menjadi Guru.`
    await loadStudents()
  } catch (error) {
    errorMessage.value = error?.response?.data?.error?.message || 'Gagal mengubah role'
  }
}

const downloadTemplate = async () => {
  try {
    const response = await api.get('/api/v1/admin/students/template', { responseType: 'blob' })
    const url = window.URL.createObjectURL(new Blob([response.data]))
    const link = document.createElement('a')
    link.href = url
    link.setAttribute('download', 'template_siswa.xlsx')
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
  } catch (e) {
    errorMessage.value = 'Gagal mendownload template'
  }
}

const uploadImport = async (event) => {
  const file = event.target.files[0]
  if (!file) return
  
  const formData = new FormData()
  formData.append('file', file)
  
  isLoading.value = true
  errorMessage.value = ''
  successMessage.value = ''
  try {
    const { data } = await api.post('/api/v1/admin/students/import', formData, {
      headers: { 'Content-Type': 'multipart/form-data' }
    })
    const res = data?.data
    successMessage.value = `Berhasil impor: ${res.inserted} data. Gagal: ${res.errors.length} data.`
    await loadStudents()
  } catch (e) {
    errorMessage.value = e?.response?.data?.error?.message || 'Gagal mengimpor file.'
  } finally {
    isLoading.value = false
    event.target.value = ''
  }
}

const copyId = (id) => {
  navigator.clipboard.writeText(id)
  notificationStore.pushInfo('ID disalin ke clipboard')
}

const shortId = (id) => shortCode2(id)

const handlePhotoUpload = (e) => {
  const file = e.target.files[0]
  if (file) {
    photoFile.value = file
    photoPreview.value = URL.createObjectURL(file)
  }
}

const uploadStudentPhoto = async () => {
  if (!photoFile.value || !editingUserId.value) return
  isUploadingPhoto.value = true
  errorMessage.value = ''
  successMessage.value = ''

  try {
    const optimizedFile = await compressImage(photoFile.value, 400, 400, 0.8)
    const formData = new FormData()
    formData.append('file', optimizedFile)

    const { data } = await api.post(`/api/v1/me/photo?target_user_id=${editingUserId.value}`, formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    })

    form.photo_url = data?.data?.photo_url || ''
    successMessage.value = 'Foto siswa berhasil diperbarui'
    photoFile.value = null
    photoPreview.value = null

    // Update list to show new photo if visible
    await loadStudents()
  } catch (error) {
    errorMessage.value = error?.response?.data?.error?.message || 'Gagal mengunggah foto'
  } finally {
    isUploadingPhoto.value = false
  }
}

const getStudentAvatar = (student) => {
  const photo = student.photo_url
  if (photo) return resolveBackendAssetUrl(photo)
  return `https://api.dicebear.com/7.x/initials/svg?seed=${student.username || 'Student'}&backgroundColor=0033ff`
}

const openBulkPhotoModal = () => {
  showBulkPhotoModal.value = true
  bulkPhotoFile.value = null
  bulkPhotoFileName.value = ''
  bulkPhotoResult.value = null
  bulkPhotoError.value = ''
}

const closeBulkPhotoModal = () => {
  showBulkPhotoModal.value = false
  if (bulkPhotoResult.value) {
    loadStudents()
  }
}

const handleBulkPhotoFile = (e) => {
  const file = e.target.files[0]
  if (!file) return
  bulkPhotoFile.value = file
  bulkPhotoFileName.value = file.name
  bulkPhotoResult.value = null
  bulkPhotoError.value = ''
}

const uploadBulkPhotos = async () => {
  if (!bulkPhotoFile.value) return
  bulkPhotoUploading.value = true
  bulkPhotoError.value = ''
  bulkPhotoResult.value = null
  try {
    const formData = new FormData()
    formData.append('file', bulkPhotoFile.value)
    const { data } = await api.post('/api/v1/admin/students/bulk-photos', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    })
    bulkPhotoResult.value = data?.data || null
  } catch (e) {
    bulkPhotoError.value = e?.response?.data?.error?.message || 'Gagal mengunggah file ZIP'
  } finally {
    bulkPhotoUploading.value = false
  }
}

onMounted(async () => {
  await loadLookups()
  await loadStudents()
})
</script>

<template>
  <LayoutAuthenticated>
    <SectionMain>
      <SectionTitleLineWithButton :icon="mdiAccountSchool" title="Siswa" main>
        <div class="flex items-center gap-2 flex-wrap">
          <BaseButton :icon="mdiDownload" color="purple" label="Template" @click="downloadTemplate" small />
          <label class="inline-flex cursor-pointer items-center justify-center rounded-lg border border-emerald-600 bg-emerald-600 px-3 py-1.5 text-xs font-bold text-white hover:bg-emerald-700 transition-colors">
            <BaseIcon :path="mdiUpload" size="16" class="mr-1" />
            Impor Excel
            <input type="file" class="hidden" accept=".xlsx" @change="uploadImport" />
          </label>
          <button
            class="inline-flex cursor-pointer items-center justify-center rounded-lg border border-violet-600 bg-violet-600 px-3 py-1.5 text-xs font-bold text-white hover:bg-violet-700 transition-colors gap-1"
            @click="openBulkPhotoModal"
          >
            <BaseIcon :path="mdiImageMultiple" size="16" />
            Import Foto Massal
          </button>
          <BaseButton :icon="mdiRefresh" color="info" label="Muat Ulang" @click="loadStudents" small />
        </div>
      </SectionTitleLineWithButton>

      <div class="mb-6 grid gap-6 xl:grid-cols-5">
        <CardBox class="xl:col-span-2" color="blue">
          <h3 class="mb-4 text-lg font-semibold dark:text-slate-100">
            {{ isEditing ? 'Edit Siswa' : 'Tambah Siswa' }}
          </h3>

          <div v-if="isEditing" class="mb-6 flex flex-col items-center gap-4 border-b dark:border-slate-800 pb-6">
            <div class="relative group">
              <div class="h-24 w-24 overflow-hidden rounded-full border-4 border-blue-50 dark:border-slate-800 bg-slate-100 dark:bg-slate-800 shadow-md">
                <img :src="photoPreview || getStudentAvatar(form)" class="h-full w-full object-cover" alt="Student Photo" loading="lazy" />
              </div>
              <label class="absolute bottom-0 right-0 flex h-8 w-8 cursor-pointer items-center justify-center rounded-full bg-blue-600 text-white shadow-lg transition-all hover:scale-110 hover:bg-blue-700 ring-2 ring-white dark:ring-slate-900">
                <BaseIcon :path="mdiCamera" size="16" />
                <input type="file" class="hidden" accept="image/*" @change="handlePhotoUpload" />
              </label>
            </div>
            <div v-if="photoFile" class="flex flex-col items-center gap-2">
              <BaseButton
                color="info"
                :label="isUploadingPhoto ? 'Mengunggah...' : 'Simpan Foto Baru'"
                :icon="mdiUpload"
                small
                :disabled="isUploadingPhoto"
                @click="uploadStudentPhoto"
              />
              <p class="text-[10px] text-slate-400">Klik simpan untuk menerapkan foto baru</p>
            </div>
          </div>

          <div class="grid gap-4">
            <FormField label="Username">
              <FormControl v-model="form.username" placeholder="siswa.baru" />
            </FormField>
            <FormField :label="isEditing ? 'Password Baru (Opsional)' : 'Password'">
              <PasswordField v-model="form.password" placeholder="Minimal 8 karakter" autocomplete="new-password" />
            </FormField>
            <FormField label="Nama">
              <FormControl v-model="form.name" placeholder="Nama lengkap siswa" />
            </FormField>
            <FormField label="Email">
              <FormControl v-model="form.email" placeholder="siswa@sekolah.sch.id" />
            </FormField>
            <FormField label="NIS">
              <FormControl v-model="form.nis" placeholder="Nomor induk siswa" />
            </FormField>
            <FormField label="Nomor Peserta Ujian">
              <FormControl v-model="form.participant_no" placeholder="Contoh: ASAT-2026-001" />
            </FormField>
            <FormField label="No. WhatsApp">
              <FormControl v-model="form.phone" placeholder="62812xxxx" />
            </FormField>
            <FormField label="Jenjang">
              <FormControl
                v-model="form.jenjang"
                :options="[
                  { value: '', label: 'Pilih jenjang' },
                  { value: 'SMA', label: 'Tingkat SMA' },
                  { value: 'SMP', label: 'Tingkat SMP' },
                  { value: 'SD', label: 'Tingkat SD' },
                ]"
                @change="onJenjangChange"
              />
            </FormField>
            <FormField v-if="form.jenjang === 'SMA'" label="Program">
              <FormControl
                v-model="form.program_id"
                :options="[{ value: '', label: 'Pilih program' }, ...programs.map((item) => ({ value: item.id, label: item.name }))]"
              />
            </FormField>
            <FormField label="Level">
              <FormControl
                v-model="form.level_id"
                :options="[
                  { value: '', label: 'Pilih level' },
                  ...filteredLevels.map((item) => ({ value: item.id, label: item.name }))
                ]"
              />
            </FormField>
            <FormField label="Group">
              <FormControl
                v-model="form.group_id"
                :options="[{ value: '', label: 'Pilih group' }, ...groups.map((item) => ({ value: item.id, label: item.name }))]"
              />
            </FormField>
            <FormField label="Status Akun">
              <FormControl
                v-model="form.is_active"
                :options="[
                  { value: true, label: 'Aktif' },
                  { value: false, label: 'Nonaktif' },
                ]"
              />
            </FormField>
            <BaseButtons>
              <BaseButton
                :icon="isEditing ? mdiContentSave : mdiPlus"
                color="info"
                :label="isEditing ? 'Simpan Perubahan' : 'Tambah Siswa'"
                :disabled="isSaving"
                @click="saveStudent"
              />
              <BaseButton
                v-if="isEditing"
                :icon="mdiPencil"
                color="whiteDark"
                outline
                label="Batal Edit"
                @click="resetForm"
              />
              <BaseButton v-else color="whiteDark" outline label="Reset" @click="resetForm" />
            </BaseButtons>
          </div>
        </CardBox>

        <CardBox class="xl:col-span-3" color="purple">
          <div class="mb-4 flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
            <div>
              <h3 class="text-lg font-semibold dark:text-slate-100">Daftar Siswa</h3>
              <p class="text-sm text-slate-500 dark:text-slate-400">Total data: {{ meta.total || students.length }}</p>
            </div>
            <div class="w-full lg:max-w-sm">
              <FormField label="Cari">
                <FormControl v-model="query" placeholder="Nama, username, email, nis" />
              </FormField>
            </div>
          </div>

          <div v-if="!authStore.isAuthenticated" class="rounded-lg bg-amber-50 dark:bg-amber-900/20 px-4 py-3 text-sm text-amber-700 dark:text-amber-400 border border-amber-100 dark:border-amber-900/40">
            Login terlebih dulu agar data siswa dapat dimuat dari backend.
          </div>
          <div v-else-if="errorMessage" class="rounded-lg bg-red-50 dark:bg-red-900/20 px-4 py-3 text-sm text-red-700 dark:text-red-400 border border-red-100 dark:border-red-900/40">
            {{ errorMessage }}
          </div>
          <div v-if="successMessage" class="mb-4 rounded-lg bg-emerald-50 dark:bg-emerald-900/20 px-4 py-3 text-sm text-emerald-700 dark:text-emerald-400 border border-emerald-100 dark:border-emerald-900/40">
            {{ successMessage }}
          </div>

          <div class="mb-4">
            <BaseButton color="whiteDark" outline label="Terapkan Pencarian" @click="loadStudents" />
          </div>

          <div v-if="isLoading" class="text-sm text-slate-500 dark:text-slate-400 italic">Memuat data siswa...</div>
          <div v-else class="overflow-x-auto">
            <table class="w-full text-left text-sm">
              <thead class="border-b dark:border-slate-800 bg-slate-50 dark:bg-slate-800/50 text-slate-600 dark:text-slate-300 uppercase text-xs tracking-wider">
                <tr>
                  <th class="px-3 py-3">Foto</th>
                  <th class="px-3 py-3">Nama</th>
                  <th class="px-3 py-3">Username</th>
                  <th class="px-3 py-3">Password</th>
                  <th class="px-3 py-3">NIS</th>
                  <th class="px-3 py-3">No Peserta</th>
                  <th class="px-3 py-3">Jenjang</th>
                  <th class="px-3 py-3">Email</th>
                  <th class="px-3 py-3">SISWA_ID</th>
                  <th class="px-3 py-3 text-center">Status</th>
                  <th class="px-3 py-3">Aksi</th>
                </tr>
              </thead>
              <tbody v-if="students.length || isLoading">
                <tr v-for="student in students" :key="student.id" class="border-b dark:border-slate-800 last:border-b-0 hover:bg-slate-50 dark:hover:bg-slate-800/20 transition-colors">
                  <td class="px-3 py-3">
                    <div class="h-10 w-10 overflow-hidden rounded-full border border-slate-200 dark:border-slate-700 bg-slate-100">
                      <img :src="getStudentAvatar(student)" class="h-full w-full object-cover" alt="" loading="lazy" />
                    </div>
                  </td>
                  <td class="px-3 py-3 font-medium dark:text-slate-200">{{ student.name }}</td>
                  <td class="px-3 py-3 text-slate-500 dark:text-slate-400">{{ student.username }}</td>
                  <td class="px-3 py-3 text-slate-500 dark:text-slate-400 font-mono">{{ student.password_plain || '-' }}</td>
                  <td class="px-3 py-3 text-slate-500 dark:text-slate-400">{{ student.nis }}</td>
                  <td class="px-3 py-3 text-slate-500 dark:text-slate-400 font-mono">{{ student.participant_no || '-' }}</td>
                  <td class="px-3 py-3">
                    <span
                      v-if="student.jenjang"
                      class="rounded-full px-2 py-1 text-[10px] font-bold uppercase tracking-tighter"
                      :class="{
                        'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400': student.jenjang === 'SMA',
                        'bg-violet-100 text-violet-700 dark:bg-violet-900/30 dark:text-violet-400': student.jenjang === 'SMP',
                        'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400': student.jenjang === 'SD',
                      }"
                    >{{ student.jenjang }}</span>
                    <span v-else class="text-slate-400 dark:text-slate-500">-</span>
                  </td>
                  <td class="px-3 py-3 text-slate-500 dark:text-slate-400">{{ student.email || '-' }}</td>
                  <td class="px-3 py-3 font-mono text-[10px] text-slate-400">
                    <div class="flex items-center gap-2">
                       <span class="truncate w-16 font-black">{{ shortId(student.id) }}</span>
                       <BaseIcon :path="mdiContentCopy" size="14" class="cursor-pointer hover:text-blue-500" @click="copyId(student.id)" title="Salin ID" />
                    </div>
                  </td>
                  <td class="px-3 py-3 text-center">
                    <span
                      class="rounded-full px-2 py-1 text-[10px] font-bold uppercase tracking-tighter"
                      :class="student.is_active ? 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400' : 'bg-slate-200 text-slate-700 dark:bg-slate-800 dark:text-slate-400'"
                    >
                      {{ student.is_active ? 'Aktif' : 'Nonaktif' }}
                    </span>
                  </td>
                  <td class="px-3 py-3">
                    <div class="flex items-center justify-start lg:justify-end gap-3">
                      <BaseIcon
                        :path="mdiEye"
                        size="18"
                        class="text-emerald-500 cursor-pointer hover:scale-125 transition-transform"
                        title="Lihat"
                        @click="startEditStudent(student)"
                      />
                      <BaseIcon
                        :path="mdiPencil"
                        size="18"
                        class="text-blue-500 cursor-pointer hover:scale-125 transition-transform"
                        title="Edit"
                        @click="startEditStudent(student)"
                      />
                      <BaseIcon
                        :path="mdiContentCopy"
                        size="18"
                        class="text-purple-500 cursor-pointer hover:scale-125 transition-transform"
                        title="Duplikat"
                        @click="copyStudent(student)"
                      />
                      <BaseIcon
                        :path="mdiAccountSwitch"
                        size="18"
                        class="text-orange-500 cursor-pointer hover:scale-125 transition-transform"
                        title="Ubah Role ke Guru"
                        @click="switchRoleStudent(student)"
                      />
                      <BaseIcon
                        :path="mdiDelete"
                        size="18"
                        class="text-red-500 cursor-pointer hover:scale-125 transition-transform"
                        title="Hapus"
                        @click="deleteStudent(student.id)"
                      />
                    </div>
                  </td>
                </tr>
              </tbody>
            </table>

            <div v-if="!students.length && !isLoading" class="py-12 border-t dark:border-slate-800">
               <BaseEmptyState 
                  title="Siswa Tidak Ditemukan" 
                  description="Tidak ada data siswa yang cocok dengan kriteria pencarian Anda."
                  button-label="Muat Ulang Data"
                  :button-icon="mdiRefresh"
                  @click="query = ''; loadStudents()"
               />
            </div>
          </div>
        </CardBox>
      </div>
    </SectionMain>
  </LayoutAuthenticated>

  <!-- Modal Bulk Upload Foto -->
  <Teleport to="body">
    <Transition name="modal-fade">
      <div
        v-if="showBulkPhotoModal"
        class="fixed inset-0 z-50 flex items-center justify-center p-4"
        @click.self="closeBulkPhotoModal"
      >
        <!-- Backdrop -->
        <div class="absolute inset-0 bg-black/60 backdrop-blur-sm" @click="closeBulkPhotoModal" />

        <!-- Panel -->
        <div class="relative z-10 w-full max-w-2xl rounded-2xl bg-white dark:bg-slate-900 shadow-2xl ring-1 ring-slate-200 dark:ring-slate-700 overflow-hidden">
          <!-- Header -->
          <div class="flex items-center justify-between bg-gradient-to-r from-violet-600 to-purple-700 px-6 py-4">
            <div class="flex items-center gap-3">
              <div class="flex h-9 w-9 items-center justify-center rounded-xl bg-white/20">
                <BaseIcon :path="mdiImageMultiple" size="20" class="text-white" />
              </div>
              <div>
                <h2 class="text-base font-bold text-white">Import Foto Massal</h2>
                <p class="text-xs text-violet-200">Upload ZIP berisi foto siswa sekaligus</p>
              </div>
            </div>
            <button
              class="flex h-8 w-8 items-center justify-center rounded-lg bg-white/10 text-white hover:bg-white/20 transition-colors"
              @click="closeBulkPhotoModal"
            >
              <BaseIcon :path="mdiClose" size="18" />
            </button>
          </div>

          <div class="p-6 space-y-5">
            <!-- Panduan -->
            <div class="rounded-xl border border-violet-100 dark:border-violet-900/40 bg-violet-50 dark:bg-violet-900/20 p-4 text-sm">
              <p class="font-semibold text-violet-800 dark:text-violet-300 mb-2">📋 Cara penggunaan:</p>
              <ol class="list-decimal list-inside space-y-1 text-violet-700 dark:text-violet-400">
                <li>Siapkan foto-foto siswa dalam format <strong>JPG, PNG, atau WebP</strong></li>
                <li>Beri nama file sesuai <strong>NIS</strong> atau <strong>username</strong> siswa<br/>
                  <span class="ml-4 text-xs text-violet-500">Contoh: <code class="bg-violet-100 dark:bg-violet-900 px-1 rounded">1234567890.jpg</code> atau <code class="bg-violet-100 dark:bg-violet-900 px-1 rounded">budi.santoso.png</code></span>
                </li>
                <li>Kumpulkan semua foto ke dalam satu file <strong>ZIP</strong></li>
                <li>Upload file ZIP tersebut di sini</li>
              </ol>
            </div>

            <!-- Upload area -->
            <div>
              <label
                class="flex flex-col items-center justify-center w-full h-32 border-2 border-dashed rounded-xl cursor-pointer transition-colors"
                :class="bulkPhotoFile ? 'border-violet-400 bg-violet-50 dark:bg-violet-900/20' : 'border-slate-300 dark:border-slate-600 bg-slate-50 dark:bg-slate-800/50 hover:border-violet-400 hover:bg-violet-50/50 dark:hover:bg-violet-900/10'"
              >
                <div class="flex flex-col items-center gap-2 text-center px-4">
                  <BaseIcon
                    :path="bulkPhotoFile ? mdiCheckCircle : mdiUpload"
                    size="28"
                    :class="bulkPhotoFile ? 'text-violet-500' : 'text-slate-400'"
                  />
                  <div v-if="bulkPhotoFile">
                    <p class="text-sm font-semibold text-violet-700 dark:text-violet-300">{{ bulkPhotoFileName }}</p>
                    <p class="text-xs text-slate-400">Klik untuk ganti file</p>
                  </div>
                  <div v-else>
                    <p class="text-sm font-medium text-slate-600 dark:text-slate-300">Klik atau seret file ZIP ke sini</p>
                    <p class="text-xs text-slate-400">Maksimum 50 MB</p>
                  </div>
                </div>
                <input type="file" class="hidden" accept=".zip" @change="handleBulkPhotoFile" />
              </label>
            </div>

            <!-- Error -->
            <div
              v-if="bulkPhotoError"
              class="flex items-start gap-2 rounded-lg bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-900/40 px-4 py-3 text-sm text-red-700 dark:text-red-400"
            >
              <BaseIcon :path="mdiAlertCircle" size="18" class="flex-shrink-0 mt-0.5" />
              <span>{{ bulkPhotoError }}</span>
            </div>

            <!-- Hasil -->
            <div v-if="bulkPhotoResult" class="space-y-3">
              <!-- Summary -->
              <div class="grid grid-cols-3 gap-3">
                <div class="rounded-xl bg-emerald-50 dark:bg-emerald-900/20 border border-emerald-200 dark:border-emerald-900/40 p-3 text-center">
                  <div class="text-2xl font-black text-emerald-600 dark:text-emerald-400">{{ bulkPhotoResult.uploaded }}</div>
                  <div class="text-xs text-emerald-700 dark:text-emerald-400 font-medium">Berhasil</div>
                </div>
                <div class="rounded-xl bg-amber-50 dark:bg-amber-900/20 border border-amber-200 dark:border-amber-900/40 p-3 text-center">
                  <div class="text-2xl font-black text-amber-600 dark:text-amber-400">{{ bulkPhotoResult.skipped }}</div>
                  <div class="text-xs text-amber-700 dark:text-amber-400 font-medium">Dilewati</div>
                </div>
                <div class="rounded-xl bg-slate-50 dark:bg-slate-800 border border-slate-200 dark:border-slate-700 p-3 text-center">
                  <div class="text-2xl font-black text-slate-600 dark:text-slate-300">{{ bulkPhotoResult.total_files }}</div>
                  <div class="text-xs text-slate-500 font-medium">Total File</div>
                </div>
              </div>

              <!-- Detail tabel -->
              <div class="max-h-52 overflow-y-auto rounded-xl border border-slate-200 dark:border-slate-700">
                <table class="w-full text-xs">
                  <thead class="sticky top-0 bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-300">
                    <tr>
                      <th class="px-3 py-2 text-left">File</th>
                      <th class="px-3 py-2 text-left">Kunci (NIS/Username)</th>
                      <th class="px-3 py-2 text-center">Status</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr
                      v-for="(item, idx) in bulkPhotoResult.results"
                      :key="idx"
                      class="border-t border-slate-100 dark:border-slate-800"
                    >
                      <td class="px-3 py-2 font-mono text-slate-500 truncate max-w-[160px]">{{ item.filename }}</td>
                      <td class="px-3 py-2 text-slate-600 dark:text-slate-300">{{ item.key || '-' }}</td>
                      <td class="px-3 py-2 text-center">
                        <span
                          class="inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-[10px] font-bold"
                          :class="{
                            'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400': item.status === 'ok',
                            'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400': item.status === 'not_found' || item.status === 'skipped',
                            'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400': item.status === 'error',
                          }"
                          :title="item.error || ''"
                        >
                          <BaseIcon
                            :path="item.status === 'ok' ? mdiCheckCircle : item.status === 'error' ? mdiAlertCircle : mdiMinus"
                            size="10"
                          />
                          {{ item.status === 'ok' ? 'OK' : item.status === 'not_found' ? 'Tidak ditemukan' : item.status === 'skipped' ? 'Dilewati' : 'Error' }}
                        </span>
                        <p v-if="item.error" class="text-[9px] text-red-400 mt-0.5">{{ item.error }}</p>
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>

            <!-- Actions -->
            <div class="flex items-center justify-end gap-3 pt-2">
              <button
                class="rounded-lg border border-slate-200 dark:border-slate-700 px-4 py-2 text-sm font-medium text-slate-600 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-800 transition-colors"
                @click="closeBulkPhotoModal"
              >
                {{ bulkPhotoResult ? 'Tutup' : 'Batal' }}
              </button>
              <button
                v-if="!bulkPhotoResult"
                class="inline-flex items-center gap-2 rounded-lg bg-violet-600 px-5 py-2 text-sm font-bold text-white hover:bg-violet-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                :disabled="!bulkPhotoFile || bulkPhotoUploading"
                @click="uploadBulkPhotos"
              >
                <BaseIcon v-if="!bulkPhotoUploading" :path="mdiUpload" size="16" />
                <svg v-else class="h-4 w-4 animate-spin" fill="none" viewBox="0 0 24 24">
                  <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"/>
                  <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8v8H4z"/>
                </svg>
                {{ bulkPhotoUploading ? 'Memproses...' : 'Upload & Proses' }}
              </button>
              <button
                v-else
                class="inline-flex items-center gap-2 rounded-lg bg-violet-600 px-5 py-2 text-sm font-bold text-white hover:bg-violet-700 transition-colors"
                @click="() => { bulkPhotoResult = null; bulkPhotoFile = null; bulkPhotoFileName = '' }"
              >
                <BaseIcon :path="mdiUpload" size="16" />
                Upload Lagi
              </button>
            </div>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.modal-fade-enter-active,
.modal-fade-leave-active {
  transition: opacity 0.2s ease;
}
.modal-fade-enter-from,
.modal-fade-leave-to {
  opacity: 0;
}
</style>
