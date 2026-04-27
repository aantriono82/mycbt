<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { mdiAccountSchool, mdiRefresh, mdiPlus, mdiDelete, mdiPencil, mdiContentSave, mdiEye, mdiContentCopy, mdiFileExcel, mdiDownload, mdiUpload, mdiAccountSwitch } from '@mdi/js'
import LayoutAuthenticated from '@/layouts/LayoutAuthenticated.vue'
import SectionMain from '@/components/SectionMain.vue'
import SectionTitleLineWithButton from '@/components/SectionTitleLineWithButton.vue'
import CardBox from '@/components/CardBox.vue'
import BaseButton from '@/components/BaseButton.vue'
import BaseButtons from '@/components/BaseButtons.vue'
import BaseIcon from '@/components/BaseIcon.vue'
import FormField from '@/components/FormField.vue'
import FormControl from '@/components/FormControl.vue'
import { api } from '@/services/api.js'
import { useAuthStore } from '@/stores/auth.js'

const authStore = useAuthStore()

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

const programs = ref([])
const levels = ref([])
const groups = ref([])

const form = reactive({
  username: '',
  password: '',
  name: '',
  email: '',
  nis: '',
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
  alert('ID disalin ke clipboard')
}

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
    const formData = new FormData()
    formData.append('file', photoFile.value)

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
  if (photo) {
    if (photo.startsWith('/uploads')) {
      const baseUrl = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'
      return `${baseUrl}${photo}`
    }
    return photo
  }
  return `https://api.dicebear.com/7.x/initials/svg?seed=${student.username || 'Student'}&backgroundColor=0033ff`
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
        <div class="flex items-center gap-2">
          <BaseButton :icon="mdiDownload" color="purple" label="Template" @click="downloadTemplate" small />
          <label class="inline-flex cursor-pointer items-center justify-center rounded-lg border border-emerald-600 bg-emerald-600 px-3 py-1.5 text-xs font-bold text-white hover:bg-emerald-700 transition-colors">
            <BaseIcon :path="mdiUpload" size="16" class="mr-1" />
            Impor Excel
            <input type="file" class="hidden" accept=".xlsx" @change="uploadImport" />
          </label>
          <BaseButton :icon="mdiRefresh" color="info" label="Muat Ulang" @click="loadStudents" small />
        </div>
      </SectionTitleLineWithButton>

      <div class="mb-6 grid gap-6 xl:grid-cols-5">
        <CardBox class="xl:col-span-2">
          <h3 class="mb-4 text-lg font-semibold dark:text-slate-100">
            {{ isEditing ? 'Edit Siswa' : 'Tambah Siswa' }}
          </h3>

          <div v-if="isEditing" class="mb-6 flex flex-col items-center gap-4 border-b dark:border-slate-800 pb-6">
            <div class="relative group">
              <div class="h-24 w-24 overflow-hidden rounded-full border-4 border-blue-50 dark:border-slate-800 bg-slate-100 dark:bg-slate-800 shadow-md">
                <img :src="photoPreview || getStudentAvatar(form)" class="h-full w-full object-cover" alt="Student Photo" />
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
              <FormControl v-model="form.password" type="password" placeholder="Minimal 8 karakter" />
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

        <CardBox class="xl:col-span-3">
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
                  <th class="px-3 py-3">NIS</th>
                  <th class="px-3 py-3">Jenjang</th>
                  <th class="px-3 py-3">Email</th>
                  <th class="px-3 py-3">ID Internal</th>
                  <th class="px-3 py-3 text-center">Status</th>
                  <th class="px-3 py-3">Aksi</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="student in students" :key="student.id" class="border-b dark:border-slate-800 last:border-b-0 hover:bg-slate-50 dark:hover:bg-slate-800/20 transition-colors">
                  <td class="px-3 py-3">
                    <div class="h-10 w-10 overflow-hidden rounded-full border border-slate-200 dark:border-slate-700 bg-slate-100">
                      <img :src="getStudentAvatar(student)" class="h-full w-full object-cover" alt="" />
                    </div>
                  </td>
                  <td class="px-3 py-3 font-medium dark:text-slate-200">{{ student.name }}</td>
                  <td class="px-3 py-3 text-slate-500 dark:text-slate-400">{{ student.username }}</td>
                  <td class="px-3 py-3 text-slate-500 dark:text-slate-400">{{ student.nis }}</td>
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
                       <span class="truncate w-16">{{ student.id }}</span>
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
                <tr v-if="!students.length && !isLoading">
                  <td colspan="7" class="px-3 py-8 text-center text-slate-400 dark:text-slate-500 italic">Belum ada data siswa.</td>
                </tr>
              </tbody>
            </table>
          </div>
        </CardBox>
      </div>
    </SectionMain>
  </LayoutAuthenticated>
</template>
