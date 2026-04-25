<script setup>
import { computed, onMounted, ref, watch } from 'vue'
import {
  mdiAccountGroupOutline,
  mdiBookEducationOutline,
  mdiCalendarCheckOutline,
  mdiFilePdfBox,
  mdiRefresh,
} from '@mdi/js'
import LayoutAuthenticated from '@/layouts/LayoutAuthenticated.vue'
import SectionMain from '@/components/SectionMain.vue'
import SectionTitleLineWithButton from '@/components/SectionTitleLineWithButton.vue'
import CardBox from '@/components/CardBox.vue'
import CardBoxWidget from '@/components/CardBoxWidget.vue'
import BaseButton from '@/components/BaseButton.vue'
import FormField from '@/components/FormField.vue'
import FormControl from '@/components/FormControl.vue'
import { api } from '@/services/api.js'
import { useAuthStore } from '@/stores/auth.js'

const formatDateTime = (value) => {
  if (!value) return '-'
  const d = new Date(value)
  if (Number.isNaN(d.getTime())) return value
  const formatted = d.toLocaleString('id-ID', {
    dateStyle: 'medium',
    timeStyle: 'short',
    hour12: false
  }).replace(/\./g, ':')

  const offset = -d.getTimezoneOffset() / 60
  let tz = ''
  if (offset === 7) tz = 'WIB'
  else if (offset === 8) tz = 'WITA'
  else if (offset === 9) tz = 'WIT'
  else tz = offset >= 0 ? `GMT+${offset}` : `GMT${offset}`

  return `${formatted} ${tz}`
}

const authStore = useAuthStore()

const exams = ref([])
const selectedExamId = ref('')
const q = ref('')

const attendanceRows = ref([])
const meta = ref(null)

const isLoadingExams = ref(false)
const isLoadingAttendance = ref(false)
const errorMessage = ref('')

const canLoad = computed(() => authStore.isAuthenticated)

const stats = computed(() => {
  const targeted = Number(meta.value?.targeted_students || 0)
  const attended = Number(meta.value?.attended_students || 0)
  const absent = Math.max(0, targeted - attended)
  const rate = Number(meta.value?.attendance_rate_percent || 0)
  return { targeted, attended, absent, rate }
})

const escapeHtml = (value) =>
  String(value ?? '')
    .replaceAll('&', '&amp;')
    .replaceAll('<', '&lt;')
    .replaceAll('>', '&gt;')
    .replaceAll('"', '&quot;')
    .replaceAll("'", '&#39;')

const exportPdf = () => {
  if (!meta.value?.exam || !attendanceRows.value.length) return

  const generatedAt = new Date().toLocaleString('id-ID')
  const examTitle = escapeHtml(meta.value.exam.title || '-')
  const searchQuery = escapeHtml(q.value || '-')

  const rowsHtml = attendanceRows.value
    .map((row, index) => {
      const attendanceText = row.attended ? 'Hadir' : 'Belum Hadir'
      return `
        <tr>
          <td>${index + 1}</td>
          <td>${escapeHtml(row.name)}<br><small>${escapeHtml(row.username)} · ${escapeHtml(row.nis)}</small></td>
          <td>${escapeHtml(row.level_name || '-')}<br><small>${escapeHtml(row.group_name || '-')}</small></td>
          <td>${attendanceText}</td>
          <td>${escapeHtml(row.attended_at || '-')}</td>
          <td>${escapeHtml(row.attendance_note || '-')}</td>
          <td>${escapeHtml(row.session_status || 'not_joined')}</td>
        </tr>
      `
    })
    .join('')

  const printableHtml = `
    <!doctype html>
    <html lang="id">
      <head>
        <meta charset="UTF-8" />
        <title>Laporan Absensi</title>
        <style>
          * { box-sizing: border-box; }
          body { margin: 24px; color: #111827; font-family: Arial, sans-serif; }
          h1 { margin: 0 0 6px; font-size: 20px; }
          .meta { margin-bottom: 14px; font-size: 12px; line-height: 1.45; color: #374151; }
          .stats { margin-bottom: 14px; border: 1px solid #d1d5db; border-radius: 8px; padding: 10px; }
          .stats span { margin-right: 14px; font-size: 12px; }
          table { width: 100%; border-collapse: collapse; font-size: 11px; }
          th, td { border: 1px solid #d1d5db; padding: 6px; vertical-align: top; text-align: left; }
          thead { background: #f3f4f6; }
          .footer { margin-top: 10px; font-size: 11px; color: #6b7280; }
          @page { size: A4 landscape; margin: 14mm; }
        </style>
      </head>
      <body>
        <h1>Laporan Absensi Peserta</h1>
        <div class="meta">
          Ujian: <strong>${examTitle}</strong><br>
          Filter Pencarian: ${searchQuery}<br>
          Dicetak: ${escapeHtml(generatedAt)}
        </div>
        <div class="stats">
          <span>Target: <strong>${stats.value.targeted}</strong></span>
          <span>Hadir: <strong>${stats.value.attended}</strong></span>
          <span>Belum Hadir: <strong>${stats.value.absent}</strong></span>
          <span>Persentase: <strong>${stats.value.rate}%</strong></span>
        </div>
        <table>
          <thead>
            <tr>
              <th style="width: 36px;">No</th>
              <th style="width: 180px;">Siswa</th>
              <th style="width: 140px;">Kelas</th>
              <th style="width: 90px;">Absensi</th>
              <th style="width: 130px;">Waktu Hadir</th>
              <th>Catatan</th>
              <th style="width: 100px;">Status Sesi</th>
            </tr>
          </thead>
          <tbody>${rowsHtml}</tbody>
        </table>
        <div class="footer">MYCBT - Rekap Absensi Peserta</div>
      </body>
    </html>
  `

  const printWindow = window.open('', '_blank')
  if (!printWindow) return
  printWindow.document.open()
  printWindow.document.write(printableHtml)
  printWindow.document.close()
  printWindow.focus()
  printWindow.print()
}

const loadExams = async () => {
  if (!canLoad.value) return
  isLoadingExams.value = true
  errorMessage.value = ''
  try {
    const { data } = await api.get('/api/v1/exams', { params: { limit: 200, offset: 0 } })
    exams.value = data?.data || []
    if (!selectedExamId.value && exams.value.length) {
      selectedExamId.value = exams.value[0].id
    }
  } catch (error) {
    exams.value = []
    errorMessage.value = error?.response?.data?.error?.message || 'Gagal memuat daftar ujian'
  } finally {
    isLoadingExams.value = false
  }
}

const loadAttendance = async () => {
  if (!canLoad.value || !selectedExamId.value) {
    attendanceRows.value = []
    meta.value = null
    return
  }
  isLoadingAttendance.value = true
  errorMessage.value = ''
  try {
    const { data } = await api.get(`/api/v1/exams/${selectedExamId.value}/attendance`, {
      params: { q: q.value, limit: 300, offset: 0 },
    })
    attendanceRows.value = data?.data || []
    meta.value = data?.meta || null
  } catch (error) {
    attendanceRows.value = []
    meta.value = null
    errorMessage.value = error?.response?.data?.error?.message || 'Gagal memuat rekap absensi'
  } finally {
    isLoadingAttendance.value = false
  }
}

watch(selectedExamId, loadAttendance)

onMounted(async () => {
  await loadExams()
  await loadAttendance()
})
</script>

<template>
  <LayoutAuthenticated>
    <SectionMain>
      <SectionTitleLineWithButton :icon="mdiBookEducationOutline" title="Absensi Peserta" main>
        <BaseButton :icon="mdiRefresh" color="info" label="Refresh" @click="loadExams(); loadAttendance()" />
        <BaseButton
          :icon="mdiCalendarCheckOutline"
          color="purple"
          label="Generate QR"
          :disabled="!selectedExamId"
          :to="`${authStore.role === 'admin' ? '/admin' : '/teacher'}/ujian/absensi/qr/${selectedExamId}`"
        />
        <BaseButton
          :icon="mdiFilePdfBox"
          color="success"
          label="Export PDF"
          :disabled="!attendanceRows.length || isLoadingAttendance"
          @click="exportPdf"
        />
      </SectionTitleLineWithButton>

      <div v-if="!authStore.isAuthenticated" class="mb-6 rounded-xl bg-amber-50 dark:bg-amber-900/20 px-4 py-3 text-sm text-amber-700 dark:text-amber-400 border border-amber-100 dark:border-amber-900/40">
        Login terlebih dulu agar data absensi bisa dimuat.
      </div>
      <div v-else-if="errorMessage" class="mb-6 rounded-xl bg-red-50 dark:bg-red-900/20 px-4 py-3 text-sm text-red-700 dark:text-red-400 border border-red-100 dark:border-red-900/40">
        {{ errorMessage }}
      </div>

      <CardBox class="mb-6">
        <div class="grid gap-4 md:grid-cols-[360px_1fr] md:items-end">
          <FormField label="Pilih Ujian">
            <FormControl
              v-model="selectedExamId"
              :options="exams.map((item) => ({ value: item.id, label: item.title }))"
            />
          </FormField>
          <FormField label="Cari Siswa (nama/username/nis)">
            <FormControl v-model="q" placeholder="Ketik lalu klik Terapkan" />
          </FormField>
        </div>
        <div class="mt-3 flex items-center gap-3">
          <BaseButton color="whiteDark" outline label="Terapkan" :disabled="isLoadingAttendance" @click="loadAttendance" />
          <div class="text-sm text-slate-500 dark:text-slate-400 italic" v-if="isLoadingExams || isLoadingAttendance">Memuat data...</div>
          <div class="text-sm text-slate-500 dark:text-slate-400" v-else-if="meta?.exam">
            Laporan Ujian: <span class="font-bold dark:text-slate-100">{{ meta.exam.title }}</span>
          </div>
        </div>
      </CardBox>

      <div class="mb-6 grid gap-6 md:grid-cols-4">
        <CardBoxWidget
          :icon="mdiAccountGroupOutline"
          color="text-sky-500"
          label="Target Peserta"
          :number="stats.targeted"
        />
        <CardBoxWidget
          :icon="mdiCalendarCheckOutline"
          color="text-emerald-500"
          label="Hadir"
          :number="stats.attended"
        />
        <CardBoxWidget
          :icon="mdiBookEducationOutline"
          color="text-amber-500"
          label="Belum Hadir"
          :number="stats.absent"
        />
        <CardBoxWidget
          :icon="mdiRefresh"
          color="text-indigo-500"
          label="Persentase"
          :number="`${stats.rate}%`"
        />
      </div>

      <CardBox>
        <div class="overflow-x-auto">
          <table class="w-full text-left text-sm">
            <thead class="border-b dark:border-slate-800 bg-slate-50 dark:bg-slate-800/50 text-slate-600 dark:text-slate-300 uppercase text-xs tracking-wider font-bold">
              <tr>
                <th class="px-3 py-3">Siswa</th>
                <th class="px-3 py-3">Kelas</th>
                <th class="px-3 py-3 text-center">Absensi</th>
                <th class="px-3 py-3">Waktu Hadir</th>
                <th class="px-3 py-3">Catatan</th>
                <th class="px-3 py-3 text-center">Status Sesi</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="row in attendanceRows" :key="row.student_id" class="border-b dark:border-slate-800 last:border-b-0 hover:bg-slate-50/50 dark:hover:bg-slate-800/30 transition-colors">
                <td class="px-3 py-3 font-medium">
                  <div class="dark:text-slate-100">{{ row.name }}</div>
                  <div class="text-[10px] text-slate-500 dark:text-slate-400 font-mono italic">{{ row.username }} · {{ row.nis }}</div>
                </td>
                <td class="px-3 py-3">
                  <div class="text-xs dark:text-slate-300">{{ row.level_name || '-' }}</div>
                  <div class="text-[10px] text-slate-500 dark:text-slate-400 italic">{{ row.group_name || '-' }}</div>
                </td>
                <td class="px-3 py-3 text-center">
                  <span
                    class="rounded-full px-2 py-0.5 text-[10px] font-bold uppercase tracking-tight"
                    :class="row.attended ? 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400' : 'bg-slate-100 text-slate-600 dark:bg-slate-800 dark:text-slate-400'"
                  >
                    {{ row.attended ? 'hadir' : 'belum hadir' }}
                  </span>
                </td>
                <td class="px-3 py-3 text-xs text-slate-500 dark:text-slate-400 font-mono">{{ formatDateTime(row.attended_at) }}</td>
                <td class="px-3 py-3 text-xs text-slate-500 dark:text-slate-400 italic">{{ row.attendance_note || '-' }}</td>
                <td class="px-3 py-3 text-center">
                  <span class="text-[10px] font-bold uppercase px-2 py-0.5 rounded border dark:border-slate-800">
                    {{ row.session_status || 'not_joined' }}
                  </span>
                </td>
              </tr>
              <tr v-if="!attendanceRows.length && !isLoadingAttendance">
                <td colspan="6" class="px-3 py-10 text-center text-slate-400 dark:text-slate-500 italic">
                  Belum ada data absensi untuk ujian ini.
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </CardBox>
    </SectionMain>
  </LayoutAuthenticated>
</template>
