<script setup>
import { onMounted, ref } from 'vue'
import {
  mdiDatabaseExportOutline,
  mdiAccountSchool,
  mdiChartBoxOutline,
  mdiDownload,
  mdiRefresh,
  mdiInformationOutline,
  mdiLinkVariant,
} from '@mdi/js'
import LayoutAuthenticated from '@/layouts/LayoutAuthenticated.vue'
import SectionMain from '@/components/SectionMain.vue'
import SectionTitleLineWithButton from '@/components/SectionTitleLineWithButton.vue'
import CardBox from '@/components/CardBox.vue'
import BaseButton from '@/components/BaseButton.vue'
import BaseIcon from '@/components/BaseIcon.vue'
import FormField from '@/components/FormField.vue'
import FormControl from '@/components/FormControl.vue'
import { api } from '@/services/api.js'

// ─── State ────────────────────────────────────────────────────────────────────
const isLoadingSummary = ref(false)
const isLoadingExams   = ref(false)
const summary = ref({ total_students: 0, total_exams: 0, total_sessions: 0, generated_at: '' })
const exams   = ref([])
const selectedExamId  = ref('')
const selectedFormat  = ref('csv')
const isExporting     = ref(false)
const errorMessage    = ref('')

const formatOptions = [
  { value: 'csv',  label: 'CSV (.csv)  — Excel / Google Sheets compatible' },
  { value: 'json', label: 'JSON (.json) — API / LMS Platform compatible' },
]

// ─── Data fetch ───────────────────────────────────────────────────────────────
const loadSummary = async () => {
  isLoadingSummary.value = true
  try {
    const { data } = await api.get('/api/v1/lms/summary')
    summary.value = data?.data || summary.value
  } catch (e) {
    errorMessage.value = 'Gagal memuat ringkasan.'
  } finally {
    isLoadingSummary.value = false
  }
}

const loadExams = async () => {
  isLoadingExams.value = true
  try {
    const { data } = await api.get('/api/v1/lms/exams')
    exams.value = data?.data || []
  } catch (e) {
    errorMessage.value = 'Gagal memuat daftar ujian.'
  } finally {
    isLoadingExams.value = false
  }
}

const load = async () => {
  errorMessage.value = ''
  await Promise.all([loadSummary(), loadExams()])
}

// ─── Export helpers ───────────────────────────────────────────────────────────
const triggerDownload = (url, desc) => {
  const link = document.createElement('a')
  link.href = url
  link.setAttribute('download', '')
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
}

const exportStudents = async () => {
  isExporting.value = true
  errorMessage.value = ''
  try {
    const response = await api.get('/api/v1/lms/export/students', {
      params: { format: selectedFormat.value },
      responseType: 'blob',
    })
    const ext  = selectedFormat.value
    const url  = window.URL.createObjectURL(new Blob([response.data]))
    const link = document.createElement('a')
    link.href  = url
    link.setAttribute('download', `lms-students-${new Date().toISOString().slice(0,10)}.${ext}`)
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
  } catch (e) {
    errorMessage.value = 'Gagal mengekspor data siswa.'
  } finally {
    isExporting.value = false
  }
}

const exportResults = async () => {
  isExporting.value = true
  errorMessage.value = ''
  try {
    const params = { format: selectedFormat.value }
    if (selectedExamId.value) params.exam_id = selectedExamId.value
    const response = await api.get('/api/v1/lms/export/results', {
      params,
      responseType: 'blob',
    })
    const ext  = selectedFormat.value
    const label = selectedExamId.value ? selectedExamId.value.slice(0,8) : 'all'
    const url  = window.URL.createObjectURL(new Blob([response.data]))
    const link = document.createElement('a')
    link.href  = url
    link.setAttribute('download', `lms-results-${label}-${new Date().toISOString().slice(0,10)}.${ext}`)
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
  } catch (e) {
    errorMessage.value = 'Gagal mengekspor hasil ujian.'
  } finally {
    isExporting.value = false
  }
}

const examOptions = () => [
  { value: '', label: '— Semua Ujian —' },
  ...exams.value.map(e => ({
    value: e.id,
    label: `${e.title} (${e.start_at})`,
  })),
]

onMounted(load)
</script>

<template>
  <LayoutAuthenticated>
    <SectionMain>
      <SectionTitleLineWithButton :icon="mdiDatabaseExportOutline" title="Integrasi LMS & Data Portability" main>
        <BaseButton :icon="mdiRefresh" color="info" label="Refresh" @click="load" :disabled="isLoadingSummary" />
      </SectionTitleLineWithButton>

      <!-- Alert strip -->
      <div v-if="errorMessage" class="mb-5 rounded-xl bg-red-50 dark:bg-red-900/20 px-4 py-3 text-sm text-red-700 dark:text-red-400 border border-red-200 dark:border-red-900/40">
        {{ errorMessage }}
      </div>

      <!-- Info Banner -->
      <div class="mb-6 rounded-2xl border border-blue-100 dark:border-blue-900/40 bg-gradient-to-br from-blue-50 to-indigo-50 dark:from-blue-950/40 dark:to-indigo-950/30 p-5 flex gap-4 items-start">
        <div class="flex-shrink-0 flex h-10 w-10 items-center justify-center rounded-xl bg-blue-100 dark:bg-blue-900/50 text-blue-600 dark:text-blue-400">
          <BaseIcon :path="mdiInformationOutline" size="22" />
        </div>
        <div>
          <h3 class="font-bold text-blue-800 dark:text-blue-200 mb-1">Tentang Fitur Ini</h3>
          <p class="text-sm text-blue-700 dark:text-blue-300 leading-relaxed">
            Ekspor data AtigaCBT ke format <strong>CSV</strong> atau <strong>JSON</strong> standar industri untuk diimpor ke
            Google Classroom, Moodle, SisDiknas, atau sistem LMS manapun.
            Semua ekspor menggunakan format kolom yang kompatibel dengan standar <strong>IMS Global LTI / xAPI</strong>.
          </p>
        </div>
      </div>

      <!-- Stats Cards -->
      <div class="grid grid-cols-3 gap-4 mb-6">
        <div class="rounded-2xl border border-slate-100 dark:border-slate-800 bg-white dark:bg-slate-900/60 p-5 flex items-center gap-4 shadow-sm">
          <div class="flex h-12 w-12 items-center justify-center rounded-2xl bg-emerald-100 dark:bg-emerald-900/30 text-emerald-600 dark:text-emerald-400">
            <BaseIcon :path="mdiAccountSchool" size="24" />
          </div>
          <div>
            <p class="text-2xl font-black dark:text-white">
              <span v-if="isLoadingSummary" class="animate-pulse bg-slate-200 dark:bg-slate-700 rounded w-10 h-6 inline-block" />
              <span v-else>{{ summary.total_students.toLocaleString('id-ID') }}</span>
            </p>
            <p class="text-xs text-slate-500 dark:text-slate-400 uppercase tracking-widest font-semibold mt-0.5">Total Siswa</p>
          </div>
        </div>
        <div class="rounded-2xl border border-slate-100 dark:border-slate-800 bg-white dark:bg-slate-900/60 p-5 flex items-center gap-4 shadow-sm">
          <div class="flex h-12 w-12 items-center justify-center rounded-2xl bg-violet-100 dark:bg-violet-900/30 text-violet-600 dark:text-violet-400">
            <BaseIcon :path="mdiChartBoxOutline" size="24" />
          </div>
          <div>
            <p class="text-2xl font-black dark:text-white">
              <span v-if="isLoadingSummary" class="animate-pulse bg-slate-200 dark:bg-slate-700 rounded w-10 h-6 inline-block" />
              <span v-else>{{ summary.total_exams.toLocaleString('id-ID') }}</span>
            </p>
            <p class="text-xs text-slate-500 dark:text-slate-400 uppercase tracking-widest font-semibold mt-0.5">Total Ujian</p>
          </div>
        </div>
        <div class="rounded-2xl border border-slate-100 dark:border-slate-800 bg-white dark:bg-slate-900/60 p-5 flex items-center gap-4 shadow-sm">
          <div class="flex h-12 w-12 items-center justify-center rounded-2xl bg-amber-100 dark:bg-amber-900/30 text-amber-600 dark:text-amber-400">
            <BaseIcon :path="mdiLinkVariant" size="24" />
          </div>
          <div>
            <p class="text-2xl font-black dark:text-white">
              <span v-if="isLoadingSummary" class="animate-pulse bg-slate-200 dark:bg-slate-700 rounded w-10 h-6 inline-block" />
              <span v-else>{{ summary.total_sessions.toLocaleString('id-ID') }}</span>
            </p>
            <p class="text-xs text-slate-500 dark:text-slate-400 uppercase tracking-widest font-semibold mt-0.5">Sesi Ujian</p>
          </div>
        </div>
      </div>

      <!-- Main export cards -->
      <div class="grid gap-6 xl:grid-cols-2">

        <!-- Export Format picker (shared) -->
        <CardBox class="xl:col-span-2" color="blue">
          <h3 class="text-base font-bold dark:text-slate-100 mb-4 flex items-center gap-2">
            <span class="flex h-7 w-7 items-center justify-center rounded-lg bg-slate-100 dark:bg-slate-800">
              <BaseIcon :path="mdiDatabaseExportOutline" size="16" class="text-slate-500 dark:text-slate-400" />
            </span>
            Format Ekspor
          </h3>
          <div class="grid sm:grid-cols-2 gap-3">
            <label
              v-for="opt in formatOptions"
              :key="opt.value"
              :for="`fmt-${opt.value}`"
              class="flex items-center gap-3 rounded-xl border-2 cursor-pointer px-4 py-3 transition-all duration-150"
              :class="selectedFormat === opt.value
                ? 'border-indigo-500 bg-indigo-50 dark:bg-indigo-950/40 dark:border-indigo-400'
                : 'border-slate-200 dark:border-slate-700 hover:border-indigo-300 dark:hover:border-indigo-600'"
            >
              <input
                :id="`fmt-${opt.value}`"
                type="radio"
                v-model="selectedFormat"
                :value="opt.value"
                class="accent-indigo-600"
              />
              <span class="text-sm font-medium dark:text-slate-200">{{ opt.label }}</span>
            </label>
          </div>
        </CardBox>

        <!-- Export: Students Roster -->
        <CardBox color="indigo">
          <div class="flex items-center gap-4 mb-5">
            <div class="flex h-14 w-14 items-center justify-center rounded-2xl bg-gradient-to-br from-emerald-400 to-green-600 shadow-lg shadow-emerald-200 dark:shadow-emerald-900/30">
              <BaseIcon :path="mdiAccountSchool" size="26" class="text-white" />
            </div>
            <div>
              <h3 class="text-base font-bold dark:text-slate-100">Roster Siswa</h3>
              <p class="text-xs text-slate-500 dark:text-slate-400">Ekspor seluruh data siswa terdaftar</p>
            </div>
          </div>

          <div class="mb-5 rounded-xl bg-slate-50 dark:bg-slate-800/40 border border-slate-100 dark:border-slate-800 p-4 space-y-1.5">
            <p class="text-xs text-slate-600 dark:text-slate-300 font-medium uppercase tracking-wide mb-2">Kolom yang diekspor:</p>
            <div class="flex flex-wrap gap-2">
              <span
                v-for="col in ['id','name','username','nis','jenjang','level','group','program','email']"
                :key="col"
                class="px-2 py-0.5 bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-700 rounded text-[11px] font-mono text-slate-600 dark:text-slate-300"
              >{{ col }}</span>
            </div>
          </div>

          <BaseButton
            :icon="mdiDownload"
            color="success"
            label="Download Roster Siswa"
            :disabled="isExporting"
            @click="exportStudents"
            class="w-full"
          />
          <p v-if="isExporting" class="mt-2 text-center text-xs text-emerald-600 animate-pulse font-bold uppercase tracking-widest">
            Mengekspor data...
          </p>
        </CardBox>

        <!-- Export: Exam Results -->
        <CardBox color="blue">
          <div class="flex items-center gap-4 mb-5">
            <div class="flex h-14 w-14 items-center justify-center rounded-2xl bg-gradient-to-br from-violet-500 to-indigo-600 shadow-lg shadow-violet-200 dark:shadow-violet-900/30">
              <BaseIcon :path="mdiChartBoxOutline" size="26" class="text-white" />
            </div>
            <div>
              <h3 class="text-base font-bold dark:text-slate-100">Hasil Ujian</h3>
              <p class="text-xs text-slate-500 dark:text-slate-400">Ekspor skor, status, dan progres peserta</p>
            </div>
          </div>

          <FormField label="Filter Ujian (opsional)">
            <FormControl
              v-model="selectedExamId"
              :options="examOptions()"
              :disabled="isLoadingExams"
            />
          </FormField>

          <div class="mb-5 rounded-xl bg-slate-50 dark:bg-slate-800/40 border border-slate-100 dark:border-slate-800 p-4 space-y-1.5">
            <p class="text-xs text-slate-600 dark:text-slate-300 font-medium uppercase tracking-wide mb-2">Kolom yang diekspor:</p>
            <div class="flex flex-wrap gap-2">
              <span
                v-for="col in ['exam_id','exam_title','exam_date','subject','student_id','student_name','username','nis','status','score','max_score','correct_count','total_items','started_at','finished_at']"
                :key="col"
                class="px-2 py-0.5 bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-700 rounded text-[11px] font-mono text-slate-600 dark:text-slate-300"
              >{{ col }}</span>
            </div>
          </div>

          <BaseButton
            :icon="mdiDownload"
            color="info"
            :label="selectedExamId ? 'Download Hasil Ujian Ini' : 'Download Semua Hasil'"
            :disabled="isExporting"
            @click="exportResults"
            class="w-full"
          />
          <p v-if="isExporting" class="mt-2 text-center text-xs text-blue-600 animate-pulse font-bold uppercase tracking-widest">
            Mengekspor data...
          </p>
        </CardBox>

        <!-- LTI / Platform Guide -->
        <CardBox class="xl:col-span-2" color="emerald">
          <h3 class="text-base font-bold dark:text-slate-100 mb-4 flex items-center gap-2">
            <span class="flex h-7 w-7 items-center justify-center rounded-lg bg-blue-100 dark:bg-blue-900/40">
              <BaseIcon :path="mdiLinkVariant" size="16" class="text-blue-600 dark:text-blue-400" />
            </span>
            Panduan Integrasi Platform LMS
          </h3>
          <div class="grid sm:grid-cols-3 gap-4">
            <div
              v-for="platform in [
                { name: 'Google Classroom', steps: ['Export Roster Siswa → CSV', 'Import di People → Import CSV', 'Export Hasil → Gunakan di Gradebook'], color: 'from-blue-50 to-sky-50 dark:from-blue-950/30 dark:to-sky-950/20', border: 'border-blue-100 dark:border-blue-900/40', tagColor: 'bg-blue-100 text-blue-700 dark:bg-blue-900/40 dark:text-blue-300' },
                { name: 'Moodle', steps: ['Export Hasil → CSV', 'Grades → Import Grades → CSV', 'Map kolom: username + score'], color: 'from-orange-50 to-amber-50 dark:from-orange-950/30 dark:to-amber-950/20', border: 'border-orange-100 dark:border-orange-900/40', tagColor: 'bg-orange-100 text-orange-700 dark:bg-orange-900/40 dark:text-orange-300' },
                { name: 'Dapodik / SisDiknas', steps: ['Export Roster → CSV/JSON', 'Gunakan kolom NIS & nama', 'Hasil ujian untuk rapor digital'], color: 'from-emerald-50 to-green-50 dark:from-emerald-950/30 dark:to-green-950/20', border: 'border-emerald-100 dark:border-emerald-900/40', tagColor: 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/40 dark:text-emerald-300' },
              ]"
              :key="platform.name"
              class="rounded-xl border p-4 bg-gradient-to-br"
              :class="[platform.color, platform.border]"
            >
              <div class="mb-3">
                <span class="text-xs font-bold px-2.5 py-1 rounded-full" :class="platform.tagColor">{{ platform.name }}</span>
              </div>
              <ol class="space-y-1.5">
                <li
                  v-for="(step, i) in platform.steps"
                  :key="i"
                  class="flex gap-2 text-xs text-slate-600 dark:text-slate-300"
                >
                  <span class="flex-shrink-0 w-4 h-4 rounded-full bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 flex items-center justify-center text-[10px] font-bold text-slate-500">{{ i + 1 }}</span>
                  {{ step }}
                </li>
              </ol>
            </div>
          </div>
        </CardBox>
      </div>

      <p class="mt-4 text-center text-xs text-slate-400 dark:text-slate-500">
        Data dihasilkan: {{ summary.generated_at ? new Date(summary.generated_at).toLocaleString('id-ID', { dateStyle: 'medium', timeStyle: 'short', hour12: false }).replace(/\./g, ':') + ' WIB' : '—' }}
      </p>
    </SectionMain>
  </LayoutAuthenticated>
</template>
