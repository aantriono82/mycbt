<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { mdiAccountTie, mdiRefresh, mdiPlus, mdiDelete, mdiPencil, mdiContentSave, mdiEye, mdiContentCopy, mdiDownload, mdiUpload, mdiAccountSwitch } from '@mdi/js'
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

const teachers = ref([])
const subjects = ref([])
const meta = ref({ total: 0 })
const isLoading = ref(false)
const isSaving = ref(false)
const errorMessage = ref('')
const successMessage = ref('')
const query = ref('')
const editingId = ref('')
const allGroups = ref([])
const allLevels = ref([])

const form = reactive({
  username: '',
  password: '',
  name: '',
  email: '',
  nip: '',
  jenjang: '',
  mapel_codes: '',
  group_names: '',
  level_names: '',
  phone: '',
  is_active: true,
})

const canLoad = computed(() => authStore.isAuthenticated)
const isEditing = computed(() => !!editingId.value)

const resetForm = () => {
  editingId.value = ''
  form.username = ''
  form.password = ''
  form.name = ''
  form.email = ''
  form.nip = ''
  form.jenjang = ''
  form.mapel_codes = ''
  form.group_names = ''
  form.level_names = ''
  form.phone = ''
  form.is_active = true
}

const loadSubjects = async () => {
  if (!canLoad.value) return
  try {
    const { data } = await api.get('/api/v1/admin/subjects')
    subjects.value = data?.data || []
  } catch {
    subjects.value = []
  }
}

const loadGroups = async () => {
  if (!canLoad.value) return
  try {
    const { data } = await api.get('/api/v1/admin/groups')
    allGroups.value = data?.data || []
  } catch {
    allGroups.value = []
  }
}

const loadLevels = async () => {
  if (!canLoad.value) return
  try {
    const { data } = await api.get('/api/v1/admin/levels')
    allLevels.value = data?.data || []
  } catch {
    allLevels.value = []
  }
}

const loadTeachers = async () => {
  if (!canLoad.value) return
  isLoading.value = true
  errorMessage.value = ''
  try {
    const { data } = await api.get('/api/v1/admin/teachers', {
      params: {
        q: query.value,
        limit: 50,
        offset: 0,
      },
    })
    teachers.value = data?.data || []
    meta.value = data?.meta || { total: teachers.value.length }
  } catch (error) {
    errorMessage.value = error?.response?.data?.error?.message || 'Gagal memuat data guru'
  } finally {
    isLoading.value = false
  }
}

const startEditTeacher = async (teacher) => {
  editingId.value = teacher.id
  form.username = teacher.username || ''
  form.password = ''
  form.name = teacher.name || ''
  form.email = teacher.email || ''
  form.nip = teacher.nip || ''
  form.jenjang = teacher.jenjang || ''
  form.phone = teacher.phone || ''
  form.is_active = !!teacher.is_active
  form.mapel_codes = ''
  form.group_names = ''
  form.level_names = ''

  try {
    const { data } = await api.get(`/api/v1/admin/teachers/${teacher.id}/subjects`)
    form.mapel_codes = (data?.data || []).map((item) => item.code).filter(Boolean).join(',')
  } catch {
    form.mapel_codes = ''
  }

  try {
    const { data } = await api.get(`/api/v1/admin/teachers/${teacher.id}/groups`)
    form.group_names = (data?.data || []).map((item) => item.name).filter(Boolean).join(',')
  } catch {
    form.group_names = ''
  }

  try {
    const { data } = await api.get(`/api/v1/admin/teachers/${teacher.id}/levels`)
    form.level_names = (data?.data || []).map((item) => item.name).filter(Boolean).join(',')
  } catch {
    form.level_names = ''
  }
}

const saveTeacher = async () => {
  isSaving.value = true
  successMessage.value = ''
  errorMessage.value = ''
  try {
    if (isEditing.value) {
      await api.patch(`/api/v1/admin/teachers/${editingId.value}`, {
        username: form.username,
        password: form.password || '',
        name: form.name,
        email: form.email,
        nip: form.nip,
        jenjang: form.jenjang,
        phone: form.phone,
        is_active: form.is_active,
      })
      await api.put(`/api/v1/admin/teachers/${editingId.value}/subjects`, {
        subject_codes: form.mapel_codes
          .split(',')
          .map((item) => item.trim())
          .filter(Boolean),
      })
      await api.put(`/api/v1/admin/teachers/${editingId.value}/groups`, {
        group_names: form.group_names
          .split(',')
          .map((item) => item.trim())
          .filter(Boolean),
      })
      await api.put(`/api/v1/admin/teachers/${editingId.value}/levels`, {
        level_names: form.level_names
          .split(',')
          .map((item) => item.trim())
          .filter(Boolean),
      })
      successMessage.value = 'Data guru berhasil diperbarui'
    } else {
      await api.post('/api/v1/admin/teachers', {
        ...form,
        phone: form.phone,
      })
      successMessage.value = 'Data guru berhasil ditambahkan'
    }
    resetForm()
    await loadTeachers()
  } catch (error) {
    errorMessage.value = error?.response?.data?.error?.message || 'Gagal menyimpan guru'
  } finally {
    isSaving.value = false
  }
}

const deleteTeacher = async (id) => {
  successMessage.value = ''
  errorMessage.value = ''
  try {
    await api.delete(`/api/v1/admin/teachers/${id}`)
    successMessage.value = 'Data guru berhasil dihapus'
    if (editingId.value === id) {
      resetForm()
    }
    await loadTeachers()
  } catch (error) {
    errorMessage.value = error?.response?.data?.error?.message || 'Gagal menghapus guru'
  }
}

const copyTeacher = async (teacher) => {
  resetForm()
  form.name = `${teacher.name} (Copy)`
  form.email = teacher.email
  form.nip = teacher.nip
  form.jenjang = teacher.jenjang
  form.phone = teacher.phone
  
  // Load associations to the form
  try {
    const [subRes, grpRes, lvlRes] = await Promise.all([
      api.get(`/api/v1/admin/teachers/${teacher.id}/subjects`),
      api.get(`/api/v1/admin/teachers/${teacher.id}/groups`),
      api.get(`/api/v1/admin/teachers/${teacher.id}/levels`)
    ])
    form.mapel_codes = (subRes.data?.data || []).map(i => i.code).join(',')
    form.group_names = (grpRes.data?.data || []).map(i => i.name).join(',')
    form.level_names = (lvlRes.data?.data || []).map(i => i.name).join(',')
  } catch {
    // fine if failed
  }
  
  successMessage.value = 'Data guru disalin ke form. Silakan isi username & password baru.'
}

const switchRoleTeacher = async (teacher) => {
  const ok = window.confirm(`Ubah role "${teacher.name}" dari GURU menjadi SISWA?\nPastikan data pendukung (NIS, dll) diisi setelahnya.`)
  if (!ok) return
  successMessage.value = ''
  errorMessage.value = ''
  try {
    await api.post(`/api/v1/admin/users/${teacher.user_id}/switch-role`)
    successMessage.value = `Role "${teacher.name}" berhasil diubah menjadi Siswa.`
    await loadTeachers()
  } catch (error) {
    errorMessage.value = error?.response?.data?.error?.message || 'Gagal mengubah role'
  }
}

const downloadTemplate = async () => {
  try {
    const response = await api.get('/api/v1/admin/teachers/template', { responseType: 'blob' })
    const url = window.URL.createObjectURL(new Blob([response.data]))
    const link = document.createElement('a')
    link.href = url
    link.setAttribute('download', 'template_guru.xlsx')
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
    const { data } = await api.post('/api/v1/admin/teachers/import', formData, {
      headers: { 'Content-Type': 'multipart/form-data' }
    })
    const res = data?.data
    successMessage.value = `Berhasil impor: ${res.inserted} data. Gagal: ${res.errors.length} data.`
    await loadTeachers()
  } catch (e) {
    errorMessage.value = e?.response?.data?.error?.message || 'Gagal mengimpor file.'
  } finally {
    isLoading.value = false
    event.target.value = ''
  }
}


onMounted(async () => {
  await loadSubjects()
  await loadGroups()
  await loadLevels()
  await loadTeachers()
})
</script>

<template>
  <LayoutAuthenticated>
    <SectionMain>
      <SectionTitleLineWithButton :icon="mdiAccountTie" title="Guru" main>
        <div class="flex items-center gap-2">
          <BaseButton :icon="mdiDownload" color="purple" label="Template" @click="downloadTemplate" small />
          <label class="inline-flex cursor-pointer items-center justify-center rounded-lg border border-emerald-600 bg-emerald-600 px-3 py-1.5 text-xs font-bold text-white hover:bg-emerald-700 transition-colors">
            <BaseIcon :path="mdiUpload" size="16" class="mr-1" />
            Impor Excel
            <input type="file" class="hidden" accept=".xlsx" @change="uploadImport" />
          </label>
          <BaseButton :icon="mdiRefresh" color="info" label="Muat Ulang" @click="loadTeachers" small />
        </div>
      </SectionTitleLineWithButton>

      <div class="mb-6 grid gap-6 xl:grid-cols-5">
        <CardBox class="xl:col-span-2">
          <h3 class="mb-4 text-lg font-semibold dark:text-slate-100">
            {{ isEditing ? 'Edit Guru' : 'Tambah Guru' }}
          </h3>
          <div class="grid gap-4">
            <FormField label="Username">
              <FormControl v-model="form.username" placeholder="guru.baru" />
            </FormField>
            <FormField :label="isEditing ? 'Password Baru (Opsional)' : 'Password'">
              <FormControl v-model="form.password" type="password" placeholder="Minimal 8 karakter" />
            </FormField>
            <FormField label="Nama">
              <FormControl v-model="form.name" placeholder="Nama lengkap guru" />
            </FormField>
            <FormField label="Email">
              <FormControl v-model="form.email" placeholder="guru@sekolah.sch.id" />
            </FormField>
            <FormField label="NIP">
              <FormControl v-model="form.nip" placeholder="Nomor induk pegawai" />
            </FormField>
            <FormField label="No. WhatsApp">
              <FormControl v-model="form.phone" placeholder="62812xxxx" />
            </FormField>
            <FormField label="Jenjang">
              <FormControl
                v-model="form.jenjang"
                :options="[
                  { value: '', label: 'Kosong/Umum' },
                  { value: 'SMA', label: 'SMA' },
                  { value: 'SMP', label: 'SMP' },
                  { value: 'SD', label: 'SD' },
                ]"
              />
            </FormField>
            <FormField label="Mapel Diampu (Kode)">
              <FormControl v-model="form.mapel_codes" placeholder="MTK,BIO,FIS" />
            </FormField>
            <FormField label="Level Diampu (Nama)">
              <FormControl v-model="form.level_names" placeholder="Kelas 10,Kelas 11" />
            </FormField>
            <FormField label="Kelas/Group Diampu (Nama)">
              <FormControl v-model="form.group_names" placeholder="X-IPA-1,X-IPA-2" />
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

            <div v-if="subjects.length" class="rounded-xl bg-slate-50 dark:bg-slate-800/60 px-4 py-3 text-[10px] text-slate-600 dark:text-slate-400 border border-slate-100 dark:border-slate-800">
              <span class="font-bold text-slate-700 dark:text-slate-300">Mapel tersedia:</span>
              {{ subjects.map((item) => `${item.code || '-'}`).join(' • ') }}
            </div>

            <div v-if="allLevels.length" class="rounded-xl bg-slate-50 dark:bg-slate-800/60 px-4 py-3 text-[10px] text-slate-600 dark:text-slate-400 border border-slate-100 dark:border-slate-800">
              <span class="font-bold text-slate-700 dark:text-slate-300">Level tersedia:</span>
              {{ allLevels.map((item) => item.name).join(' • ') }}
            </div>

            <div v-if="allGroups.length" class="rounded-xl bg-slate-50 dark:bg-slate-800/60 px-4 py-3 text-[10px] text-slate-600 dark:text-slate-400 border border-slate-100 dark:border-slate-800">
              <span class="font-bold text-slate-700 dark:text-slate-300">Kelas/Group tersedia:</span>
              {{ allGroups.map((item) => item.name).join(' • ') }}
            </div>

            <BaseButtons>
              <BaseButton
                :icon="isEditing ? mdiContentSave : mdiPlus"
                color="info"
                :label="isEditing ? 'Simpan Perubahan' : 'Tambah Guru'"
                :disabled="isSaving"
                @click="saveTeacher"
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
              <h3 class="text-lg font-semibold dark:text-slate-100">Daftar Guru</h3>
              <p class="text-sm text-slate-500 dark:text-slate-400">Total data: {{ meta.total || teachers.length }}</p>
            </div>
            <div class="w-full lg:max-w-sm">
              <FormField label="Cari">
                <FormControl v-model="query" placeholder="Nama, username, email, nip" />
              </FormField>
            </div>
          </div>

          <div v-if="!authStore.isAuthenticated" class="rounded-lg bg-amber-50 dark:bg-amber-900/20 px-4 py-3 text-sm text-amber-700 dark:text-amber-400 border border-amber-100 dark:border-amber-900/40">
            Login terlebih dulu agar data guru dapat dimuat dari backend.
          </div>
          <div v-else-if="errorMessage" class="rounded-lg bg-red-50 dark:bg-red-900/20 px-4 py-3 text-sm text-red-700 dark:text-red-400 border border-red-100 dark:border-red-900/40">
            {{ errorMessage }}
          </div>
          <div v-if="successMessage" class="mb-4 rounded-lg bg-emerald-50 dark:bg-emerald-900/20 px-4 py-3 text-sm text-emerald-700 dark:text-emerald-400 border border-emerald-100 dark:border-emerald-900/40">
            {{ successMessage }}
          </div>

          <div class="mb-4">
            <BaseButton color="whiteDark" outline label="Terapkan Pencarian" @click="loadTeachers" />
          </div>

          <div v-if="isLoading" class="text-sm text-slate-500 dark:text-slate-400">Memuat data guru...</div>
          <div v-else class="overflow-x-auto">
            <table class="w-full text-left text-sm">
              <thead class="border-b dark:border-slate-800 bg-slate-50 dark:bg-slate-800/50 text-slate-600 dark:text-slate-300 uppercase text-xs tracking-wider">
                <tr>
                  <th class="px-3 py-3">Nama</th>
                  <th class="px-3 py-3">Username</th>
                  <th class="px-3 py-3">Email</th>
                  <th class="px-3 py-3">NIP</th>
                  <th class="px-3 py-3">Level/Grup Diampu</th>
                  <th class="px-3 py-3 text-center">Status</th>
                  <th class="px-3 py-3">Aksi</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="teacher in teachers" :key="teacher.id" class="border-b dark:border-slate-800 last:border-b-0 hover:bg-slate-50 dark:hover:bg-slate-800/30 transition-colors">
                  <td class="px-3 py-3 font-medium dark:text-slate-200">
                    <div>{{ teacher.name }}</div>
                    <div v-if="teacher.jenjang" class="inline-block rounded bg-amber-100 dark:bg-amber-900/30 px-1 text-[8px] font-bold text-amber-700 dark:text-amber-400 mr-2 uppercase">{{ teacher.jenjang }}</div>
                    <div v-if="teacher.mapel_summary" class="inline-block text-[10px] text-slate-500 dark:text-slate-400 font-normal">Mapel: {{ teacher.mapel_summary }}</div>
                  </td>
                  <td class="px-3 py-3 text-slate-500 dark:text-slate-400">{{ teacher.username }}</td>
                  <td class="px-3 py-3 text-slate-500 dark:text-slate-400">{{ teacher.email || '-' }}</td>
                  <td class="px-3 py-3 text-slate-500 dark:text-slate-400">{{ teacher.nip || '-' }}</td>
                  <td class="px-3 py-3">
                    <div v-if="teacher.level_summary || teacher.group_summary" class="flex flex-col gap-1">
                      <div v-if="teacher.level_summary" class="flex flex-wrap gap-1">
                        <span v-for="lvl in teacher.level_summary.split(', ')" :key="lvl" class="rounded bg-sky-50 dark:bg-sky-900/30 px-1.5 py-0.5 text-[9px] font-bold text-sky-600 dark:text-sky-400 border border-sky-100 dark:border-sky-800">
                          {{ lvl }}
                        </span>
                      </div>
                      <div v-if="teacher.group_summary" class="flex flex-wrap gap-1">
                        <span v-for="grp in teacher.group_summary.split(', ')" :key="grp" class="rounded bg-indigo-50 dark:bg-indigo-900/30 px-1.5 py-0.5 text-[9px] text-indigo-600 dark:text-indigo-400 border border-indigo-100 dark:border-indigo-800">
                          {{ grp }}
                        </span>
                      </div>
                    </div>
                    <span v-else class="text-slate-400 italic text-xs">-</span>
                  </td>
                  <td class="px-3 py-3 text-center">
                    <span
                      class="rounded-full px-2 py-1 text-[10px] font-bold uppercase tracking-tighter"
                      :class="teacher.is_active ? 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400' : 'bg-slate-200 text-slate-700 dark:bg-slate-800 dark:text-slate-400'"
                    >
                      {{ teacher.is_active ? 'Aktif' : 'Nonaktif' }}
                    </span>
                  </td>
                  <td class="px-3 py-3">
                    <div class="flex items-center justify-start lg:justify-end gap-3">
                      <BaseIcon
                        :path="mdiEye"
                        size="18"
                        class="text-emerald-500 cursor-pointer hover:scale-125 transition-transform"
                        title="Lihat"
                        @click="startEditTeacher(teacher)"
                      />
                      <BaseIcon
                        :path="mdiPencil"
                        size="18"
                        class="text-blue-500 cursor-pointer hover:scale-125 transition-transform"
                        title="Edit"
                        @click="startEditTeacher(teacher)"
                      />
                      <BaseIcon
                        :path="mdiContentCopy"
                        size="18"
                        class="text-purple-500 cursor-pointer hover:scale-125 transition-transform"
                        title="Duplikat"
                        @click="copyTeacher(teacher)"
                      />
                      <BaseIcon
                        :path="mdiAccountSwitch"
                        size="18"
                        class="text-orange-500 cursor-pointer hover:scale-125 transition-transform"
                        title="Ubah Role ke Siswa"
                        @click="switchRoleTeacher(teacher)"
                      />
                      <BaseIcon
                        :path="mdiDelete"
                        size="18"
                        class="text-red-500 cursor-pointer hover:scale-125 transition-transform"
                        title="Hapus"
                        @click="deleteTeacher(teacher.id)"
                      />
                    </div>
                  </td>
                </tr>
                <tr v-if="!teachers.length && !isLoading">
                  <td colspan="6" class="px-3 py-8 text-center text-slate-400 dark:text-slate-500 italic">Belum ada data guru.</td>
                </tr>
              </tbody>
            </table>
          </div>
        </CardBox>
      </div>
    </SectionMain>
  </LayoutAuthenticated>
</template>
