<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { mdiClipboardCheckOutline } from '@mdi/js'
import LayoutAuthenticated from '@/layouts/LayoutAuthenticated.vue'
import SectionMain from '@/components/SectionMain.vue'
import SectionTitleLineWithButton from '@/components/SectionTitleLineWithButton.vue'
import CardBox from '@/components/CardBox.vue'
import BaseButton from '@/components/BaseButton.vue'
import FormField from '@/components/FormField.vue'
import FormControl from '@/components/FormControl.vue'
import { api } from '@/services/api.js'

const exams = ref([])
const history = ref([])

const isLoadingExams = ref(false)
const isLoadingHistory = ref(false)
const isSubmitting = ref(false)

const errorMessage = ref('')
const submitMessage = ref('')

const form = reactive({
  exam_id: '',
  note: '',
})

const formatDateTime = (value) => {
  if (!value) return '-'
  const parsed = new Date(value)
  if (Number.isNaN(parsed.getTime())) return value
  return parsed.toLocaleString('id-ID', {
    dateStyle: 'medium',
    timeStyle: 'short',
  })
}

const examOptions = computed(() =>
  exams.value.map((exam) => ({
    value: exam.id,
    label: `${exam.title} (${formatDateTime(exam.starts_at)})`,
  })),
)

const loadExams = async () => {
  isLoadingExams.value = true
  errorMessage.value = ''
  try {
    const { data } = await api.get('/api/v1/student/exams', {
      params: { limit: 100, offset: 0 },
    })
    exams.value = data?.data || []
    if (!form.exam_id && exams.value.length) {
      form.exam_id = exams.value[0].id
    }
  } catch (error) {
    exams.value = []
    errorMessage.value = error?.response?.data?.error?.message || 'Gagal memuat daftar ujian'
  } finally {
    isLoadingExams.value = false
  }
}

const loadHistory = async () => {
  isLoadingHistory.value = true
  try {
    const { data } = await api.get('/api/v1/student/attendance/history', {
      params: { limit: 50, offset: 0 },
    })
    history.value = data?.data || []
  } catch (error) {
    history.value = []
    errorMessage.value = error?.response?.data?.error?.message || 'Gagal memuat riwayat absensi'
  } finally {
    isLoadingHistory.value = false
  }
}

const submitAttendance = async () => {
  submitMessage.value = ''
  errorMessage.value = ''
  if (!form.exam_id) {
    errorMessage.value = 'Pilih ujian terlebih dulu.'
    return
  }

  isSubmitting.value = true
  try {
    await api.post('/api/v1/student/attendance', {
      exam_id: form.exam_id,
      note: form.note,
    })
    submitMessage.value = 'Absensi berhasil disimpan.'
    await loadHistory()
  } catch (error) {
    errorMessage.value = error?.response?.data?.error?.message || 'Gagal menyimpan absensi'
  } finally {
    isSubmitting.value = false
  }
}

const refreshAll = async () => {
  await Promise.all([loadExams(), loadHistory()])
}

onMounted(refreshAll)
</script>

<template>
  <LayoutAuthenticated>
    <SectionMain>
      <SectionTitleLineWithButton :icon="mdiClipboardCheckOutline" title="Absensi" main>
        <BaseButton color="info" label="Refresh" @click="refreshAll" />
      </SectionTitleLineWithButton>

      <CardBox class="mb-6">
        <div class="grid gap-4 lg:max-w-3xl">
          <FormField label="Pilih Ujian">
            <FormControl
              v-model="form.exam_id"
              :options="examOptions"
              :disabled="isLoadingExams || isSubmitting || !examOptions.length"
            />
          </FormField>

          <FormField label="Catatan Kehadiran (Opsional)">
            <FormControl
              v-model="form.note"
              type="textarea"
              placeholder="Contoh: Perangkat siap, jaringan stabil."
              :disabled="isSubmitting"
            />
          </FormField>

          <div class="flex gap-3">
            <BaseButton
              color="info"
              :label="isSubmitting ? 'Menyimpan...' : 'Kirim Absensi'"
              :disabled="isSubmitting || isLoadingExams || !form.exam_id"
              @click="submitAttendance"
            />
          </div>
        </div>

        <div v-if="isLoadingExams" class="mt-4 text-sm text-slate-500 dark:text-slate-400 italic">Memuat daftar ujian...</div>
        <div v-if="submitMessage" class="mt-4 rounded-lg bg-emerald-50 dark:bg-emerald-900/20 px-4 py-3 text-sm text-emerald-700 dark:text-emerald-400 border border-emerald-100 dark:border-emerald-900/40">
          {{ submitMessage }}
        </div>
        <div v-if="errorMessage" class="mt-4 rounded-lg bg-red-50 dark:bg-red-900/20 px-4 py-3 text-sm text-red-700 dark:text-red-400 border border-red-100 dark:border-red-900/40">
          {{ errorMessage }}
        </div>
      </CardBox>

      <CardBox>
        <h3 class="mb-4 text-lg font-bold dark:text-slate-100 uppercase tracking-tight">Riwayat Absensi</h3>
        <div v-if="isLoadingHistory" class="text-sm text-slate-500 dark:text-slate-400 italic">Memuat riwayat absensi...</div>
        <div v-else-if="!history.length" class="text-sm text-slate-500 dark:text-slate-400 italic">Belum ada riwayat absensi.</div>

        <div v-else class="overflow-x-auto">
          <table class="w-full text-left text-sm">
            <thead class="border-b dark:border-slate-800 bg-slate-50 dark:bg-slate-800/50 text-slate-600 dark:text-slate-300 uppercase text-xs tracking-wider font-bold">
              <tr>
                <th class="px-3 py-3">Ujian</th>
                <th class="px-3 py-3">Mapel</th>
                <th class="px-3 py-3">Jadwal</th>
                <th class="px-3 py-3">Catatan</th>
                <th class="px-3 py-3">Waktu Absensi</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="item in history" :key="item.id" class="border-b dark:border-slate-800 last:border-b-0 hover:bg-slate-50/50 dark:hover:bg-slate-800/30 transition-colors">
                <td class="px-3 py-3 font-medium dark:text-slate-100">{{ item.exam_title }}</td>
                <td class="px-3 py-3 text-slate-600 dark:text-slate-300 text-xs">{{ item.subject }}</td>
                <td class="px-3 py-3 text-xs text-slate-500 dark:text-slate-400 italic">
                  {{ formatDateTime(item.starts_at) }} s/d {{ formatDateTime(item.ends_at) }}
                </td>
                <td class="px-3 py-3 text-xs text-slate-500 dark:text-slate-400 italic">{{ item.note || '-' }}</td>
                <td class="px-3 py-3 text-[11px] font-mono font-bold dark:text-slate-100">{{ formatDateTime(item.attended_at) }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </CardBox>
    </SectionMain>
  </LayoutAuthenticated>
</template>
