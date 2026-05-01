<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { mdiBullhornOutline, mdiRefresh, mdiPlus, mdiDelete, mdiPencil, mdiContentSave, mdiEmailOutline, mdiWhatsapp, mdiSend } from '@mdi/js'
import LayoutAuthenticated from '@/layouts/LayoutAuthenticated.vue'
import SectionMain from '@/components/SectionMain.vue'
import SectionTitleLineWithButton from '@/components/SectionTitleLineWithButton.vue'
import CardBox from '@/components/CardBox.vue'
import BaseButton from '@/components/BaseButton.vue'
import BaseButtons from '@/components/BaseButtons.vue'
import FormField from '@/components/FormField.vue'
import FormControl from '@/components/FormControl.vue'
import QuillEditor from '@/components/QuillEditor.vue'
import { api } from '@/services/api.js'
import { useAuthStore } from '@/stores/auth.js'

const authStore = useAuthStore()

const announcements = ref([])
const meta = ref({ total: 0 })
const isLoading = ref(false)
const isSaving = ref(false)
const errorMessage = ref('')
const successMessage = ref('')
const query = ref('')
const statusFilter = ref('all')
const editingId = ref('')
const isBlasting = ref({}) // map of id -> boolean
const blastChannels = ref(['email', 'whatsapp'])

const levels = ref([])
const groups = ref([])
const students = ref([])

const form = reactive({
  title: '',
  body: '',
  category: 'pengumuman',
  is_active: true,
  published_date: '',
  published_time: '',
  expires_date: '',
  expires_time: '',
  target_type: 'all',
  target_id: '',
})

const isEditing = computed(() => !!editingId.value)

const statusOptions = [
  { value: 'all', label: 'Semua' },
  { value: 'active', label: 'Aktif' },
  { value: 'inactive', label: 'Nonaktif' },
]

const categoryOptions = [
  { value: 'pengumuman', label: 'Pengumuman' },
  { value: 'informasi', label: 'Informasi' },
  { value: 'general', label: 'General' },
]

const targetTypeOptions = [
  { value: 'all', label: 'Semua siswa (broadcast)' },
  { value: 'level', label: 'Target level' },
  { value: 'group', label: 'Target group' },
  { value: 'student', label: 'Target siswa' },
]

const targetOptions = computed(() => {
  if (form.target_type === 'level') {
    return [{ value: '', label: 'Pilih level' }, ...levels.value.map((item) => ({ value: item.id, label: item.name }))]
  }
  if (form.target_type === 'group') {
    return [{ value: '', label: 'Pilih group' }, ...groups.value.map((item) => ({ value: item.id, label: item.name }))]
  }
  if (form.target_type === 'student') {
    return [
      { value: '', label: 'Pilih siswa' },
      ...students.value.map((item) => ({ value: item.id, label: `${item.name} (${item.nis})` })),
    ]
  }
  return [{ value: '', label: 'Tidak ada target khusus' }]
})

const pad2 = (value) => String(value).padStart(2, '0')

const toDatetimeInput = (value) => {
  if (!value) return ''
  const parsed = new Date(value)
  if (Number.isNaN(parsed.getTime())) return ''
  return `${parsed.getFullYear()}-${pad2(parsed.getMonth() + 1)}-${pad2(parsed.getDate())} ${pad2(parsed.getHours())}:${pad2(parsed.getMinutes())}`
}

const toRFC3339 = (value) => {
  if (!value) return ''
  const normalized = String(value).trim().replace('T', ' ')
  const match = normalized.match(/^(\d{4})-(\d{2})-(\d{2})\s+(\d{2}):(\d{2})$/)
  if (match) {
    const [, y, m, d, hh, mm] = match
    const parsed = new Date(Number(y), Number(m) - 1, Number(d), Number(hh), Number(mm), 0)
    if (!Number.isNaN(parsed.getTime())) return parsed.toISOString()
  }
  const parsed = new Date(value)
  if (Number.isNaN(parsed.getTime())) return ''
  return parsed.toISOString()
}

const splitDateTimeParts = (value) => {
  const text = toDatetimeInput(value)
  if (!text) return { date: '', time: '' }
  const [date = '', time = ''] = text.split(' ')
  return { date, time }
}

const combineDateTimeParts = (date, time) => {
  const d = String(date || '').trim()
  const t = String(time || '').trim()
  if (!d) return ''
  if (!t) return `${d} 00:00`
  if (!/^\d{2}:\d{2}$/.test(t)) return ''
  return `${d} ${t}`
}

const formatDateTime = (value) => {
  if (!value) return '-'
  const parsed = new Date(value)
  if (Number.isNaN(parsed.getTime())) return value
  const formatted = parsed.toLocaleString('id-ID', {
    dateStyle: 'medium',
    timeStyle: 'short',
    hour12: false,
  }).replace(/\./g, ':')
  const offset = -parsed.getTimezoneOffset() / 60
  let tz = ''
  if (offset === 7) tz = 'WIB'
  else if (offset === 8) tz = 'WITA'
  else if (offset === 9) tz = 'WIT'
  else tz = offset >= 0 ? `GMT+${offset}` : `GMT${offset}`
  return `${formatted} ${tz}`
}

const resetForm = () => {
  editingId.value = ''
  form.title = ''
  form.body = ''
  form.category = 'pengumuman'
  form.is_active = true
  form.published_date = ''
  form.published_time = ''
  form.expires_date = ''
  form.expires_time = ''
  form.target_type = 'all'
  form.target_id = ''
}

const loadLookups = async () => {
  try {
    const [levelRes, groupRes, studentRes] = await Promise.all([
      api.get('/api/v1/lookups/levels'),
      api.get('/api/v1/lookups/groups'),
      api.get('/api/v1/lookups/students', { params: { limit: 200, offset: 0 } }),
    ])
    levels.value = levelRes?.data?.data || []
    groups.value = groupRes?.data?.data || []
    students.value = studentRes?.data?.data || []
  } catch {
    levels.value = []
    groups.value = []
    students.value = []
  }
}

const loadAnnouncements = async () => {
  if (!authStore.isAuthenticated) return
  isLoading.value = true
  errorMessage.value = ''
  try {
    const params = {
      q: query.value,
      limit: 50,
      offset: 0,
    }
    if (statusFilter.value !== 'all') {
      params.is_active = String(statusFilter.value === 'active')
    }
    const { data } = await api.get('/api/v1/announcements', { params })
    announcements.value = data?.data || []
    meta.value = data?.meta || { total: announcements.value.length }
  } catch (error) {
    errorMessage.value = error?.response?.data?.error?.message || 'Gagal memuat pengumuman'
  } finally {
    isLoading.value = false
  }
}

const buildPayload = () => {
  const payload = {
    title: form.title,
    body: form.body,
    category: form.category,
    is_active: form.is_active,
    published_at: toRFC3339(combineDateTimeParts(form.published_date, form.published_time)),
    expires_at: toRFC3339(combineDateTimeParts(form.expires_date, form.expires_time)),
    target_level_id: '',
    target_group_id: '',
    target_student_id: '',
  }
  if (form.target_type === 'level') payload.target_level_id = form.target_id
  if (form.target_type === 'group') payload.target_group_id = form.target_id
  if (form.target_type === 'student') payload.target_student_id = form.target_id
  return payload
}

const saveAnnouncement = async () => {
  successMessage.value = ''
  errorMessage.value = ''
  isSaving.value = true
  try {
    const payload = buildPayload()
    if (isEditing.value) {
      await api.patch(`/api/v1/announcements/${editingId.value}`, payload)
      successMessage.value = 'Pengumuman berhasil diperbarui'
    } else {
      await api.post('/api/v1/announcements', payload)
      successMessage.value = 'Pengumuman berhasil ditambahkan'
    }
    resetForm()
    await loadAnnouncements()
  } catch (error) {
    errorMessage.value = error?.response?.data?.error?.message || 'Gagal menyimpan pengumuman'
  } finally {
    isSaving.value = false
  }
}

const startEdit = (item) => {
  editingId.value = item.id
  form.title = item.title || ''
  form.body = item.body || ''
  form.category = item.category || 'pengumuman'
  form.is_active = !!item.is_active
  {
    const published = splitDateTimeParts(item.published_at)
    form.published_date = published.date
    form.published_time = published.time
  }
  {
    const expires = splitDateTimeParts(item.expires_at)
    form.expires_date = expires.date
    form.expires_time = expires.time
  }
  if (item.target_level_id) {
    form.target_type = 'level'
    form.target_id = item.target_level_id
  } else if (item.target_group_id) {
    form.target_type = 'group'
    form.target_id = item.target_group_id
  } else if (item.target_student_id) {
    form.target_type = 'student'
    form.target_id = item.target_student_id
  } else {
    form.target_type = 'all'
    form.target_id = ''
  }
}

const deleteAnnouncement = async (id) => {
  successMessage.value = ''
  errorMessage.value = ''
  try {
    await api.delete(`/api/v1/announcements/${id}`)
    successMessage.value = 'Pengumuman berhasil dihapus'
    if (editingId.value === id) {
      resetForm()
    }
    await loadAnnouncements()
  } catch (error) {
    errorMessage.value = error?.response?.data?.error?.message || 'Gagal menghapus pengumuman'
  }
}

const blastAnnouncement = async (id) => {
  if (!confirm('Kirim pengumuman ini ke target (Email/WA) sesuai pengaturan?')) return
  isBlasting.value[id] = true
  successMessage.value = ''
  errorMessage.value = ''
  try {
    const { data } = await api.post(`/api/v1/announcements/${id}/blast`, {
      channels: blastChannels.value,
    })
    successMessage.value = `Blast selesai. Terkirim: ${data.data.sent_count}, Gagal: ${data.data.failed_count}`
  } catch (error) {
    errorMessage.value = error?.response?.data?.error?.message || 'Gagal mengirim blast'
  } finally {
    isBlasting.value[id] = false
  }
}

const targetLabel = (item) => {
  if (item.target_student_id) return 'Siswa khusus'
  if (item.target_group_id) return 'Group'
  if (item.target_level_id) return 'Level'
  return 'Broadcast'
}

const stripHtml = (value) => String(value || '').replace(/<[^>]*>/g, ' ').replace(/\s+/g, ' ').trim()

onMounted(async () => {
  await loadLookups()
  await loadAnnouncements()
})
</script>

<template>
  <LayoutAuthenticated>
    <SectionMain>
      <SectionTitleLineWithButton :icon="mdiBullhornOutline" title="Kelola Pengumuman" main>
        <BaseButton :icon="mdiRefresh" color="info" label="Muat Ulang" @click="loadAnnouncements" />
      </SectionTitleLineWithButton>

      <div class="mb-6 grid min-w-0 gap-6 xl:grid-cols-5">
        <CardBox class="min-w-0 overflow-hidden xl:col-span-2">
          <h3 class="mb-4 text-xl font-black uppercase tracking-tight dark:text-slate-100">
            {{ isEditing ? 'Edit Pengumuman' : 'Tambah Pengumuman' }}
          </h3>
          <div class="flex flex-col min-w-0 gap-4">
            <FormField label="Judul">
              <FormControl v-model="form.title" placeholder="Contoh: Jadwal Ujian Diperbarui" />
            </FormField>
            <FormField label="Isi Pengumuman">
              <QuillEditor v-model="form.body" :height="170" placeholder="Isi pengumuman..." />
            </FormField>
            <FormField label="Kategori">
              <FormControl v-model="form.category" :options="categoryOptions" />
            </FormField>
            <FormField label="Status">
              <FormControl
                v-model="form.is_active"
                :options="[
                  { value: true, label: 'Aktif' },
                  { value: false, label: 'Nonaktif' },
                ]"
              />
            </FormField>
            <FormField label="Publikasi (opsional)">
              <div class="grid grid-cols-1 gap-2 md:grid-cols-2">
                <FormControl v-model="form.published_date" type="date" />
                <FormControl v-model="form.published_time" type="text" placeholder="HH:mm (24 jam)" />
              </div>
            </FormField>
            <FormField label="Kedaluwarsa (opsional)">
              <div class="grid grid-cols-1 gap-2 md:grid-cols-2">
                <FormControl v-model="form.expires_date" type="date" />
                <FormControl v-model="form.expires_time" type="text" placeholder="HH:mm (24 jam)" />
              </div>
            </FormField>
            <FormField label="Target Pengumuman">
              <FormControl
                v-model="form.target_type"
                :options="targetTypeOptions"
                @update:model-value="form.target_id = ''"
              />
            </FormField>
            <FormField v-if="form.target_type !== 'all'" label="Pilih Target">
              <FormControl v-model="form.target_id" :options="targetOptions" />
            </FormField>
            <BaseButtons no-wrap class-addon="mr-2 last:mr-0 mb-0" mb="mb-0">
              <BaseButton
                :icon="isEditing ? mdiContentSave : mdiPlus"
                color="info"
                :label="isEditing ? 'Simpan Perubahan' : 'Tambah Pengumuman'"
                :disabled="isSaving"
                @click="saveAnnouncement"
              />
              <BaseButton
                :icon="mdiPencil"
                color="purple"
                :label="isEditing ? 'Batal Edit' : 'Reset Form'"
                @click="resetForm"
              />
            </BaseButtons>
          </div>
        </CardBox>

        <CardBox class="min-w-0 overflow-hidden xl:col-span-3">
          <div class="mb-4 flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
            <div>
              <h3 class="text-xl font-black uppercase tracking-tight dark:text-slate-100">Daftar Pengumuman</h3>
              <p class="text-xs font-bold uppercase tracking-widest text-slate-500 dark:text-slate-400">Total data: {{ meta.total || announcements.length }}</p>
            </div>
            <div class="grid w-full gap-3 lg:max-w-xl lg:grid-cols-2">
              <FormField label="Cari">
                <FormControl v-model="query" placeholder="Judul, isi, kategori" />
              </FormField>
              <FormField label="Status">
                <FormControl v-model="statusFilter" :options="statusOptions" />
              </FormField>
            </div>
          </div>

          <div v-if="!authStore.isAuthenticated" class="rounded-lg bg-amber-50 dark:bg-amber-900/20 px-4 py-3 text-sm text-amber-700 dark:text-amber-400 border border-amber-100 dark:border-amber-900/40">
            Login terlebih dulu agar data dapat dimuat dari backend.
          </div>
          <div v-else-if="errorMessage" class="rounded-lg bg-red-50 dark:bg-red-900/20 px-4 py-3 text-sm text-red-700 dark:text-red-400 border border-red-100 dark:border-red-900/40">
            {{ errorMessage }}
          </div>
          <div v-if="successMessage" class="mb-4 rounded-lg bg-emerald-50 dark:bg-emerald-900/20 px-4 py-3 text-sm text-emerald-700 dark:text-emerald-400 border border-emerald-100 dark:border-emerald-900/40">
            {{ successMessage }}
          </div>

          <div class="mb-4 flex flex-wrap items-center gap-4 bg-slate-50 dark:bg-slate-800/40 p-3 rounded-xl border border-slate-100 dark:border-slate-800">
            <span class="text-xs font-bold uppercase tracking-widest text-slate-500">Channel Blast:</span>
            <label class="flex items-center gap-2 text-xs font-medium dark:text-slate-300">
              <input type="checkbox" v-model="blastChannels" value="email" /> Email
            </label>
            <label class="flex items-center gap-2 text-xs font-medium dark:text-slate-300">
              <input type="checkbox" v-model="blastChannels" value="whatsapp" /> WhatsApp
            </label>
            <BaseButton color="info" label="Terapkan Filter" @click="loadAnnouncements" small />
          </div>

          <div v-if="isLoading" class="text-sm text-slate-500 dark:text-slate-400 italic">Memuat data...</div>
          <div v-else class="overflow-x-auto">
            <table class="w-full text-left text-sm">
              <thead class="border-b dark:border-slate-800 bg-slate-50 dark:bg-slate-800/50 text-slate-600 dark:text-slate-300 uppercase text-xs tracking-wider">
                <tr>
                  <th class="px-3 py-3">Judul</th>
                  <th class="px-3 py-3 text-center">Kategori</th>
                  <th class="px-3 py-3 text-center">Target</th>
                  <th class="px-3 py-3 text-center">Status</th>
                  <th class="px-3 py-3">Publikasi</th>
                  <th class="px-3 py-3">Aksi</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="item in announcements" :key="item.id" class="border-b dark:border-slate-800 last:border-b-0 hover:bg-slate-50 dark:hover:bg-slate-800/20 transition-colors">
                  <td class="px-3 py-3">
                    <div class="font-medium dark:text-slate-200">{{ item.title }}</div>
                    <div class="mt-1 text-xs text-slate-500 dark:text-slate-400 line-clamp-2">{{ stripHtml(item.body) }}</div>
                  </td>
                  <td class="px-3 py-3 text-center">
                    <span class="px-2 py-0.5 bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-400 rounded text-[10px] uppercase font-bold">{{ item.category || '-' }}</span>
                  </td>
                  <td class="px-3 py-3 text-center text-xs dark:text-slate-300">{{ targetLabel(item) }}</td>
                  <td class="px-3 py-3 text-center">
                    <span
                      class="rounded-full px-2 py-1 text-[10px] font-bold uppercase tracking-tight"
                      :class="item.is_active ? 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400' : 'bg-slate-200 text-slate-700 dark:bg-slate-800 dark:text-slate-400'"
                    >
                      {{ item.is_active ? 'Aktif' : 'Nonaktif' }}
                    </span>
                  </td>
                  <td class="px-3 py-3">
                    <div>{{ formatDateTime(item.published_at) }}</div>
                    <div class="text-xs text-slate-500">s/d {{ formatDateTime(item.expires_at) }}</div>
                  </td>
                  <td class="px-3 py-3">
                    <BaseButtons>
                      <BaseButton :icon="mdiSend" color="emerald" small label="Blast" :disabled="isBlasting[item.id]" @click="blastAnnouncement(item.id)" />
                      <BaseButton :icon="mdiPencil" color="info" small label="Edit" @click="startEdit(item)" />
                      <BaseButton :icon="mdiDelete" color="danger" small label="Hapus" @click="deleteAnnouncement(item.id)" />
                    </BaseButtons>
                  </td>
                </tr>
                <tr v-if="!announcements.length">
                  <td colspan="6" class="px-3 py-10 text-center text-slate-400 dark:text-slate-500 italic">
                    Belum ada data pengumuman.
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </CardBox>
      </div>
    </SectionMain>
  </LayoutAuthenticated>
</template>
