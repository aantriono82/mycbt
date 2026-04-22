<script setup>
import { onMounted, ref, watch } from 'vue'
import { mdiPrinterOutline, mdiRefresh } from '@mdi/js'
import LayoutAuthenticated from '@/layouts/LayoutAuthenticated.vue'
import SectionMain from '@/components/SectionMain.vue'
import SectionTitleLineWithButton from '@/components/SectionTitleLineWithButton.vue'
import CardBox from '@/components/CardBox.vue'
import BaseButton from '@/components/BaseButton.vue'
import FormField from '@/components/FormField.vue'
import FormControl from '@/components/FormControl.vue'
import { api } from '@/services/api.js'
import { useAuthStore } from '@/stores/auth.js'

const authStore = useAuthStore()

const exams = ref([])
const selectedExamId = ref('')
const isLoading = ref(false)
const errorMessage = ref('')
const participantQuery = ref('')
const attendanceFilter = ref('all')
const schoolIdentity = ref({
  school_name: '',
  logo_url: '',
  principal_name: '',
})
const selectedTemplate = ref('default')
const printTemplateStorageKey = 'mycbt.print.templatePreset'
const templatePresets = [
  {
    value: 'default',
    label: 'Default CBT',
    cardTitle: 'KARTU UJIAN CBT',
    cardSubtitle: 'Peserta resmi ujian berbasis komputer',
    signatoryTitle: 'Panitia Ujian',
  },
  {
    value: 'uts',
    label: 'Template UTS',
    cardTitle: 'KARTU UJIAN TENGAH SEMESTER',
    cardSubtitle: 'Wajib dibawa saat pelaksanaan UTS',
    signatoryTitle: 'Koordinator UTS',
  },
  {
    value: 'uas',
    label: 'Template UAS',
    cardTitle: 'KARTU UJIAN AKHIR SEMESTER',
    cardSubtitle: 'Wajib dibawa saat pelaksanaan UAS',
    signatoryTitle: 'Koordinator UAS',
  },
  {
    value: 'tryout',
    label: 'Template Try Out',
    cardTitle: 'KARTU TRY OUT',
    cardSubtitle: 'Simulasi ujian dan evaluasi kesiapan',
    signatoryTitle: 'Koordinator Try Out',
  },
]
const signatoryName = ref('')
const signatoryTitle = ref('Panitia Ujian')

const escapeHtml = (value) =>
  String(value ?? '')
    .replaceAll('&', '&amp;')
    .replaceAll('<', '&lt;')
    .replaceAll('>', '&gt;')
    .replaceAll('"', '&quot;')
    .replaceAll("'", '&#39;')

const openPrintWindow = (title, contentHtml) => {
  const generatedAt = new Date().toLocaleString('id-ID')
  const html = `
    <!doctype html>
    <html lang="id">
      <head>
        <meta charset="UTF-8" />
        <title>${escapeHtml(title)}</title>
        <style>
          * { box-sizing: border-box; }
          body { margin: 24px; color: #111827; font-family: Arial, sans-serif; }
          h1 { margin: 0 0 4px; font-size: 20px; }
          .meta { margin-bottom: 14px; color: #4b5563; font-size: 12px; }
          table { width: 100%; border-collapse: collapse; font-size: 12px; }
          th, td { border: 1px solid #d1d5db; padding: 6px; text-align: left; vertical-align: top; }
          thead { background: #f3f4f6; }
          @page { size: A4 landscape; margin: 14mm; }
        </style>
      </head>
      <body>
        <h1>${escapeHtml(title)}</h1>
        <div class="meta">Dicetak: ${escapeHtml(generatedAt)}</div>
        ${contentHtml}
      </body>
    </html>
  `

  const w = window.open('', '_blank')
  if (!w) return
  w.document.open()
  w.document.write(html)
  w.document.close()
  w.focus()
  w.print()
}

const selectedExamTitle = () => exams.value.find((item) => item.id === selectedExamId.value)?.title || '-'
const selectedExam = () => exams.value.find((item) => item.id === selectedExamId.value) || null
const selectedTemplateConfig = () =>
  templatePresets.find((preset) => preset.value === selectedTemplate.value) || templatePresets[0]

const applyTemplatePreset = () => {
  const preset = selectedTemplateConfig()
  signatoryTitle.value = preset.signatoryTitle
  if (!signatoryName.value.trim()) {
    signatoryName.value = schoolIdentity.value.principal_name || ''
  }
}

const restoreTemplatePreset = () => {
  if (typeof window === 'undefined') return
  const saved = window.localStorage.getItem(printTemplateStorageKey)
  if (!saved) return
  if (templatePresets.some((preset) => preset.value === saved)) {
    selectedTemplate.value = saved
  }
}

const loadExams = async () => {
  if (!authStore.isAuthenticated) return
  isLoading.value = true
  errorMessage.value = ''
  try {
    const [examRes, identityRes] = await Promise.all([
      api.get('/api/v1/exams', { params: { limit: 200, offset: 0 } }),
      api.get('/api/v1/settings/school-identity'),
    ])
    exams.value = examRes?.data?.data || []
    schoolIdentity.value = {
      ...schoolIdentity.value,
      ...(identityRes?.data?.data || {}),
    }
    if (!signatoryName.value) {
      signatoryName.value = schoolIdentity.value.principal_name || ''
    }
    if (!signatoryTitle.value) {
      signatoryTitle.value = selectedTemplateConfig().signatoryTitle
    }
    if (!selectedExamId.value && exams.value.length) selectedExamId.value = exams.value[0].id
  } catch (error) {
    exams.value = []
    errorMessage.value = error?.response?.data?.error?.message || 'Gagal memuat daftar ujian'
  } finally {
    isLoading.value = false
  }
}

const printAttendance = async () => {
  if (!selectedExamId.value) return
  try {
    const { data } = await api.get(`/api/v1/exams/${selectedExamId.value}/attendance`, {
      params: { q: '', limit: 1000, offset: 0 },
    })
    const rows = data?.data || []
    const title = `Daftar Hadir Ujian - ${selectedExamTitle()}`
    const tableRows = rows
      .map(
        (row, index) => `
        <tr>
          <td>${index + 1}</td>
          <td>${escapeHtml(row.name)}</td>
          <td>${escapeHtml(row.username)}</td>
          <td>${escapeHtml(row.nis)}</td>
          <td>${escapeHtml(row.level_name || '-')}</td>
          <td>${escapeHtml(row.group_name || '-')}</td>
          <td>${row.attended ? 'Hadir' : 'Belum'}</td>
          <td>${escapeHtml(row.attended_at || '-')}</td>
          <td style="height: 28px;"></td>
        </tr>
      `,
      )
      .join('')

    openPrintWindow(
      title,
      `
      <div class="meta">Ujian: <strong>${escapeHtml(selectedExamTitle())}</strong></div>
      <table>
        <thead>
          <tr>
            <th style="width:40px;">No</th>
            <th>Nama</th>
            <th>Username</th>
            <th>NIS</th>
            <th>Level</th>
            <th>Group</th>
            <th>Status</th>
            <th>Waktu Hadir</th>
            <th style="width:140px;">Tanda Tangan</th>
          </tr>
        </thead>
        <tbody>${tableRows || '<tr><td colspan="9">Belum ada data.</td></tr>'}</tbody>
      </table>
    `,
    )
  } catch (error) {
    errorMessage.value = error?.response?.data?.error?.message || 'Gagal memuat data absensi'
  }
}

const printResults = async () => {
  if (!selectedExamId.value) return
  try {
    const { data } = await api.get(`/api/v1/exams/${selectedExamId.value}/results`, {
      params: { q: '', limit: 1000, offset: 0 },
    })
    const rows = data?.data || []
    const title = `Laporan Nilai Ujian - ${selectedExamTitle()}`
    const tableRows = rows
      .map(
        (row, index) => `
        <tr>
          <td>${index + 1}</td>
          <td>${escapeHtml(row.student_name)}</td>
          <td>${escapeHtml(row.student_username)}</td>
          <td>${escapeHtml(row.student_nis)}</td>
          <td>${escapeHtml(row.status)}</td>
          <td>${escapeHtml(row.correct_count)}/${escapeHtml(row.auto_scorable_questions)}</td>
          <td>${escapeHtml(row.score)}</td>
        </tr>
      `,
      )
      .join('')
    openPrintWindow(
      title,
      `
      <div class="meta">Ujian: <strong>${escapeHtml(selectedExamTitle())}</strong></div>
      <table>
        <thead>
          <tr>
            <th style="width:40px;">No</th>
            <th>Nama</th>
            <th>Username</th>
            <th>NIS</th>
            <th>Status</th>
            <th>Benar</th>
            <th>Nilai</th>
          </tr>
        </thead>
        <tbody>${tableRows || '<tr><td colspan="7">Belum ada data.</td></tr>'}</tbody>
      </table>
    `,
    )
  } catch (error) {
    errorMessage.value = error?.response?.data?.error?.message || 'Gagal memuat data hasil ujian'
  }
}

const printExamCards = async () => {
  if (!selectedExamId.value) return
  try {
    const [attendanceRes, tokenRes] = await Promise.all([
      api.get(`/api/v1/exams/${selectedExamId.value}/attendance`, {
        params: { q: '', limit: 1000, offset: 0 },
      }),
      api.get(`/api/v1/exams/${selectedExamId.value}/tokens`, {
        params: { limit: 100, offset: 0 },
      }),
    ])
    let participants = attendanceRes?.data?.data || []
    const q = participantQuery.value.trim().toLowerCase()
    if (q) {
      participants = participants.filter((participant) => {
        const haystack = `${participant.name || ''} ${participant.username || ''} ${participant.nis || ''}`.toLowerCase()
        return haystack.includes(q)
      })
    }
    if (attendanceFilter.value === 'attended') {
      participants = participants.filter((participant) => participant.attended)
    }
    if (attendanceFilter.value === 'not_attended') {
      participants = participants.filter((participant) => !participant.attended)
    }
    const activeTokens = (tokenRes?.data?.data || []).filter((item) => item.is_active)
    const tokenText = activeTokens.length ? activeTokens.map((item) => item.token).join(', ') : '-'
    const exam = selectedExam()
    const examTime = exam ? `${exam.starts_at || '-'} s/d ${exam.ends_at || '-'}` : '-'
    const schoolName = schoolIdentity.value.school_name || 'Sekolah'
    const logoHtml = schoolIdentity.value.logo_url
      ? `<img src="${escapeHtml(schoolIdentity.value.logo_url)}" alt="Logo sekolah" class="school-logo" />`
      : ''
    const signatureName = signatoryName.value.trim() || schoolIdentity.value.principal_name || '-'
    const signatureTitle = signatoryTitle.value.trim() || 'Panitia Ujian'
    const template = selectedTemplateConfig()

    const cardsHtml = participants
      .map(
        (participant) => `
        <div class="card">
          <div class="card-header">
            ${logoHtml}
            <div>
              <div class="card-title">${escapeHtml(template.cardTitle)}</div>
              <div class="school-name">${escapeHtml(schoolName)}</div>
              <div class="card-subtitle">${escapeHtml(template.cardSubtitle)}</div>
            </div>
          </div>
          <table>
            <tbody>
              <tr><td>Ujian</td><td>${escapeHtml(selectedExamTitle())}</td></tr>
              <tr><td>Jadwal</td><td>${escapeHtml(examTime)}</td></tr>
              <tr><td>Nama</td><td>${escapeHtml(participant.name || '-')}</td></tr>
              <tr><td>Username</td><td>${escapeHtml(participant.username || '-')}</td></tr>
              <tr><td>NIS</td><td>${escapeHtml(participant.nis || '-')}</td></tr>
              <tr><td>Level / Group</td><td>${escapeHtml(`${participant.level_name || '-'} / ${participant.group_name || '-'}`)}</td></tr>
              <tr><td>Token Aktif</td><td><strong>${escapeHtml(tokenText)}</strong></td></tr>
            </tbody>
          </table>
          <div class="signature-grid">
            <div class="signature">
              <div>Peserta,</div>
              <div class="signature-line">(____________________)</div>
            </div>
            <div class="signature">
              <div>${escapeHtml(signatureTitle)},</div>
              <div class="signature-line">${escapeHtml(signatureName)}</div>
            </div>
          </div>
        </div>
      `,
      )
      .join('')

    openPrintWindow(
      `${template.cardTitle} - ${selectedExamTitle()}`,
      `
      <style>
        @page { size: A4 portrait; margin: 10mm; }
        .cards-grid {
          display: grid;
          grid-template-columns: 1fr 1fr;
          gap: 10px;
        }
        .card {
          border: 1px solid #94a3b8;
          border-radius: 8px;
          padding: 10px;
          break-inside: avoid;
          page-break-inside: avoid;
        }
        .card-header {
          display: flex;
          align-items: center;
          gap: 8px;
          margin-bottom: 8px;
        }
        .school-logo {
          width: 40px;
          height: 40px;
          object-fit: contain;
        }
        .card-title {
          font-size: 13px;
          font-weight: 700;
          margin-bottom: 2px;
        }
        .school-name {
          font-size: 11px;
          color: #475569;
        }
        .card-subtitle {
          font-size: 10px;
          color: #64748b;
          margin-top: 2px;
        }
        .card td {
          border: 0;
          padding: 2px 0;
          font-size: 12px;
          vertical-align: top;
        }
        .card td:first-child {
          width: 100px;
          color: #475569;
        }
        .signature-grid {
          display: grid;
          grid-template-columns: 1fr 1fr;
          gap: 12px;
          margin-top: 12px;
        }
        .signature {
          margin-top: 12px;
          text-align: right;
          font-size: 11px;
          color: #334155;
        }
        .signature-line {
          margin-top: 18px;
          font-size: 12px;
        }
      </style>
      <div class="meta">
        Ujian: <strong>${escapeHtml(selectedExamTitle())}</strong>
        · Total kartu: <strong>${participants.length}</strong>
        ${q ? `· Filter peserta: <strong>${escapeHtml(participantQuery.value.trim())}</strong>` : ''}
      </div>
      <div class="cards-grid">${cardsHtml || '<div>Tidak ada peserta untuk dicetak.</div>'}</div>
    `,
    )
  } catch (error) {
    errorMessage.value = error?.response?.data?.error?.message || 'Gagal memuat data kartu ujian'
  }
}

onMounted(loadExams)
onMounted(() => {
  restoreTemplatePreset()
  applyTemplatePreset()
})

watch(selectedTemplate, (value) => {
  if (typeof window === 'undefined') return
  window.localStorage.setItem(printTemplateStorageKey, value)
})
</script>

<template>
  <LayoutAuthenticated>
    <SectionMain>
      <SectionTitleLineWithButton :icon="mdiPrinterOutline" title="Cetak" main>
        <BaseButton :icon="mdiRefresh" color="info" label="Refresh" @click="loadExams" />
      </SectionTitleLineWithButton>

      <div v-if="!authStore.isAuthenticated" class="mb-6 rounded-xl bg-amber-50 dark:bg-amber-900/20 px-4 py-3 text-sm text-amber-700 dark:text-amber-400 border border-amber-100 dark:border-amber-900/40">
        Login terlebih dulu agar modul cetak dapat memuat data backend.
      </div>
      <div v-else-if="errorMessage" class="mb-6 rounded-xl bg-red-50 dark:bg-red-900/20 px-4 py-3 text-sm text-red-700 dark:text-red-400 border border-red-100 dark:border-red-900/40">
        {{ errorMessage }}
      </div>

      <CardBox>
        <div class="mb-6 border-b dark:border-slate-800 pb-4">
          <h3 class="font-bold text-slate-700 dark:text-slate-100 mb-4">Pengaturan Cetak</h3>
          <FormField label="Pilih Ujian">
            <FormControl
              v-model="selectedExamId"
              :options="exams.map((item) => ({ value: item.id, label: item.title }))"
            />
          </FormField>
        </div>
        <div class="grid gap-4 pb-4 md:grid-cols-2">
          <FormField label="Preset Template Kartu">
            <div class="flex gap-2">
              <FormControl
                v-model="selectedTemplate"
                :options="templatePresets.map((item) => ({ value: item.value, label: item.label }))"
              />
              <BaseButton color="whiteDark" outline label="Terapkan" @click="applyTemplatePreset" />
            </div>
          </FormField>
          <FormField label="Filter Peserta (nama/username/NIS)">
            <FormControl v-model="participantQuery" placeholder="Contoh: Budi / 10231" />
          </FormField>
          <FormField label="Status Kehadiran">
            <FormControl
              v-model="attendanceFilter"
              :options="[
                { value: 'all', label: 'Semua peserta target' },
                { value: 'attended', label: 'Hanya yang sudah hadir' },
                { value: 'not_attended', label: 'Hanya yang belum hadir' },
              ]"
            />
          </FormField>
          <FormField label="Nama Penandatangan">
            <FormControl v-model="signatoryName" placeholder="Nama panitia / kepala sekolah" />
          </FormField>
          <FormField label="Jabatan Penandatangan">
            <FormControl v-model="signatoryTitle" placeholder="Panitia Ujian" />
          </FormField>
        </div>
        <div class="flex flex-wrap gap-3">
          <BaseButton
            :icon="mdiPrinterOutline"
            color="info"
            label="Cetak Daftar Hadir"
            :disabled="isLoading || !selectedExamId"
            @click="printAttendance"
          />
          <BaseButton
            :icon="mdiPrinterOutline"
            color="purple"
            label="Cetak Laporan Nilai"
            :disabled="isLoading || !selectedExamId"
            @click="printResults"
          />
          <BaseButton
            :icon="mdiPrinterOutline"
            color="success"
            label="Cetak Kartu Ujian"
            :disabled="isLoading || !selectedExamId"
            @click="printExamCards"
          />
          <div v-if="isLoading" class="self-center text-sm text-slate-500 dark:text-slate-400 italic">Memuat daftar ujian...</div>
        </div>
      </CardBox>
    </SectionMain>
  </LayoutAuthenticated>
</template>

