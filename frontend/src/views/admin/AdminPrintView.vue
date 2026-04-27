<script setup>
import { computed, onMounted, ref, watch } from 'vue'
import { mdiPrinterOutline, mdiRefresh, mdiUpload } from '@mdi/js'
import FormFilePicker from '@/components/FormFilePicker.vue'
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
const levelFilter = ref('all')
const groupFilter = ref('all')
const sessionFilter = ref('all')
const sessionStatusFilter = ref('all')
const cachedParticipants = ref([])
const isLoadingParticipants = ref(false)
const schoolIdentity = ref({
  school_name: '',
  logo_url: '',
  principal_name: '',
})
const masterLevels = ref([])
const masterGroups = ref([])
const masterSessions = ref([])
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
  {
    value: 'official',
    label: 'Kartu Dinas (ASAT)',
    cardTitle: 'ASESMEN SUMATIF AKHIR SEMESTER',
    cardSubtitle: 'Template kartu ujian resmi sekolah',
    signatoryTitle: 'Kepala Sekolah',
  },
]
const signatoryName = ref('')
const signatoryTitle = ref('Panitia Ujian')
const printCardConfig = ref({
  govLine1: 'PEMERINTAH KABUPATEN',
  govLine2: 'DINAS PENDIDIKAN DAN KEBUDAYAAN',
  schoolLine: '',
  schoolLogoUrl: '',
  examLine: 'ASESMEN SUMATIF AKHIR SEMESTER',
  academicYear: '',
  semester: 'Ganjil',
  placeDateLine: '',
  principalName: '',
  principalNip: '',
  schoolLogo2Url: '',
})
const logoFile1 = ref(null)
const logoFile2 = ref(null)

watch(logoFile1, (file) => {
  if (!file) return
  const reader = new FileReader()
  reader.onload = (e) => {
    printCardConfig.value.schoolLogoUrl = e.target.result
  }
  reader.readAsDataURL(file)
})

watch(logoFile2, (file) => {
  if (!file) return
  const reader = new FileReader()
  reader.onload = (e) => {
    printCardConfig.value.schoolLogo2Url = e.target.result
  }
  reader.readAsDataURL(file)
})

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

const pickLevelName = (item) => String(item?.level_name || item?.levelName || '').trim()
const pickGroupName = (item) => String(item?.group_name || item?.groupName || '').trim()
const pickSessionName = (item) => String(item?.session_name || item?.sessionName || item?.master_session_name || item?.masterSessionName || '').trim()
const pickSessionStatus = (item) => String(item?.session_status || item?.sessionStatus || '').trim()

const levelOptions = computed(() => {
  const uniq = Array.from(new Set(masterLevels.value.map((l) => l.name).filter(Boolean))).sort()
  if (uniq.length === 0) {
    // fallback to participants if lookups failed or empty
    const fromParts = Array.from(new Set(cachedParticipants.value.map((p) => pickLevelName(p)).filter(Boolean))).sort()
    return [{ value: 'all', label: 'Semua level' }, ...fromParts.map((n) => ({ value: n, label: n }))]
  }
  return [{ value: 'all', label: 'Semua level' }, ...uniq.map((name) => ({ value: name, label: name }))]
})

const groupOptions = computed(() => {
  const uniq = Array.from(new Set(masterGroups.value.map((g) => g.name).filter(Boolean))).sort()
  if (uniq.length === 0) {
    // fallback
    const fromParts = Array.from(new Set(cachedParticipants.value.map((p) => pickGroupName(p)).filter(Boolean))).sort()
    return [{ value: 'all', label: 'Semua kelas/group' }, ...fromParts.map((n) => ({ value: n, label: n }))]
  }
  return [{ value: 'all', label: 'Semua kelas/group' }, ...uniq.map((name) => ({ value: name, label: name }))]
})

const sessionOptions = computed(() => {
  const uniq = Array.from(new Set(masterSessions.value.map((s) => s.name).filter(Boolean))).sort()
  if (uniq.length === 0) {
    const fromParts = Array.from(new Set(cachedParticipants.value.map((p) => pickSessionName(p)).filter(Boolean))).sort()
    return [{ value: 'all', label: 'Semua sesi' }, ...fromParts.map((n) => ({ value: n, label: n }))]
  }
  return [{ value: 'all', label: 'Semua sesi' }, ...uniq.map((name) => ({ value: name, label: name }))]
})

const sessionStatusOptions = computed(() => {
  const uniq = Array.from(
    new Set((cachedParticipants.value || []).map((item) => pickSessionStatus(item)).filter(Boolean)),
  ).sort((a, b) => a.localeCompare(b))
  return [{ value: 'all', label: 'Semua status sesi' }, ...uniq.map((status) => ({ value: status, label: status }))]
})

const formatLongDateID = (value) => {
  if (!value) return ''
  const dt = new Date(value)
  if (Number.isNaN(dt.getTime())) return ''
  return dt.toLocaleDateString('id-ID', { day: '2-digit', month: 'long', year: 'numeric' })
}

const resolveAssetURL = (value) => {
  const raw = String(value || '').trim()
  if (!raw) return ''
  if (/^https?:\/\//i.test(raw) || raw.startsWith('data:')) return raw
  if (raw.startsWith('/')) {
    if (typeof window !== 'undefined' && window.location?.origin) return `${window.location.origin}${raw}`
    return raw
  }
  if (typeof window !== 'undefined' && window.location?.origin) return `${window.location.origin}/${raw}`
  return raw
}

const applyTemplatePreset = () => {
  const preset = selectedTemplateConfig()
  signatoryTitle.value = preset.signatoryTitle
  if (!signatoryName.value.trim()) {
    signatoryName.value = schoolIdentity.value.principal_name || ''
  }
  if (preset.value === 'official') {
    signatoryTitle.value = 'Kepala Sekolah'
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

const loadLookups = async () => {
  try {
    const [levelRes, groupRes, sessionRes] = await Promise.all([
      api.get('/api/v1/admin/levels'),
      api.get('/api/v1/admin/groups'),
      api.get('/api/v1/admin/sessions'),
    ])
    masterLevels.value = levelRes?.data?.data || []
    masterGroups.value = groupRes?.data?.data || []
    masterSessions.value = sessionRes?.data?.data || []
  } catch (e) {
    console.error('Failed to load lookups', e)
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
    if (!printCardConfig.value.schoolLine) {
      printCardConfig.value.schoolLine = schoolIdentity.value.school_name || 'UPTD SATUAN PENDIDIKAN'
    }
    if (!printCardConfig.value.schoolLogoUrl) {
      printCardConfig.value.schoolLogoUrl = schoolIdentity.value.logo_url || ''
    }
    if (!printCardConfig.value.principalName) {
      printCardConfig.value.principalName = schoolIdentity.value.principal_name || ''
    }
    if (!printCardConfig.value.academicYear) {
      const examYear = new Date(selectedExam()?.starts_at || Date.now()).getFullYear()
      printCardConfig.value.academicYear = `${examYear}/${examYear + 1}`
    }
    if (!printCardConfig.value.placeDateLine) {
      const examDate = formatLongDateID(selectedExam()?.starts_at || Date.now())
      printCardConfig.value.placeDateLine = `Bandarjaya, ${examDate}`
    }
    if (!signatoryName.value) {
      signatoryName.value = schoolIdentity.value.principal_name || ''
    }
    if (!signatoryTitle.value) {
      signatoryTitle.value = selectedTemplateConfig().signatoryTitle
    }
    if (!selectedExamId.value && exams.value.length) selectedExamId.value = exams.value[0].id
    if (selectedExamId.value) {
      await Promise.all([loadParticipantsCache(), loadLookups()])
    } else {
      await loadLookups()
    }
  } catch (error) {
    exams.value = []
    errorMessage.value = error?.response?.data?.error?.message || 'Gagal memuat daftar ujian'
  } finally {
    isLoading.value = false
  }
}

const loadParticipantsCache = async () => {
  if (!authStore.isAuthenticated || !selectedExamId.value) {
    cachedParticipants.value = []
    return
  }
  isLoadingParticipants.value = true
  try {
    const { data } = await api.get(`/api/v1/exams/${selectedExamId.value}/attendance`, {
      params: { q: '', limit: 1000, offset: 0 },
    })
    cachedParticipants.value = data?.data || []
  } catch {
    cachedParticipants.value = []
  } finally {
    isLoadingParticipants.value = false
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
    if (levelFilter.value !== 'all') {
      participants = participants.filter((participant) => pickLevelName(participant) === levelFilter.value)
    }
    if (groupFilter.value !== 'all') {
      participants = participants.filter((participant) => pickGroupName(participant) === groupFilter.value)
    }
    if (sessionFilter.value !== 'all') {
      participants = participants.filter((participant) => pickSessionName(participant) === sessionFilter.value)
    }
    if (sessionStatusFilter.value !== 'all') {
      participants = participants.filter((participant) => pickSessionStatus(participant) === sessionStatusFilter.value)
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

    if (template.value === 'official') {
      const officialCardsHtml = participants
        .map((participant) => {
          const participantNo = participant.exam_no || participant.participant_no || participant.nis || '-'
          const classLabel = `${participant.level_name || ''} ${participant.group_name || ''}`.trim() || '-'
          const semester = printCardConfig.value.semester?.trim() || '-'
          const govLine1 = printCardConfig.value.govLine1?.trim() || 'PEMERINTAH KABUPATEN'
          const govLine2 = printCardConfig.value.govLine2?.trim() || 'DINAS PENDIDIKAN DAN KEBUDAYAAN'
          const schoolLine = printCardConfig.value.schoolLine?.trim() || (schoolIdentity.value.school_name || 'SEKOLAH')
          const examLine = printCardConfig.value.examLine?.trim() || 'ASESMEN SUMATIF AKHIR SEMESTER'
          const academicYear = printCardConfig.value.academicYear?.trim() || '-'
          const placeDate = printCardConfig.value.placeDateLine?.trim() || '-'
          const principalName = printCardConfig.value.principalName?.trim() || signatureName
          const principalNip = printCardConfig.value.principalNip?.trim() || '-'
          const logoURL = resolveAssetURL(printCardConfig.value.schoolLogoUrl || schoolIdentity.value.logo_url)
          const logo2URL = resolveAssetURL(printCardConfig.value.schoolLogo2Url)
          const participantPhotoURL = resolveAssetURL(participant.photo_url)

          const logoHtml = logoURL
            ? `<img src="${escapeHtml(logoURL)}" alt="Logo sekolah" class="official-logo" />`
            : '<div class="official-logo-fallback">LOGO</div>'
          const logo2Html = logo2URL
            ? `<img src="${escapeHtml(logo2URL)}" alt="Logo 2" class="official-logo" />`
            : '<div class="official-logo-fallback">LOGO 2</div>'
          const participantPhotoHtml = participantPhotoURL
            ? `<img src="${escapeHtml(participantPhotoURL)}" alt="Foto peserta" class="participant-photo" />`
            : '<div class="participant-photo-fallback">FOTO<br/>PESERTA</div>'

          return `
            <div class="official-card">
              <div class="official-header-row">
                <div class="official-logo-wrap left">${logoHtml}</div>
                <div class="official-header-text">
                  <div class="line small-bold">${escapeHtml(govLine1)}</div>
                  <div class="line small-bold">${escapeHtml(govLine2)}</div>
                  <div class="line school">${escapeHtml(schoolLine)}</div>
                  <div class="line exam">${escapeHtml(examLine)}</div>
                  <div class="line year">TAHUN PELAJARAN ${escapeHtml(academicYear)}</div>
                </div>
                <div class="official-logo-wrap right">${logo2Html}</div>
              </div>

              <div class="official-identity">
                <div class="identity-row">
                  <div class="label">Nama</div>
                  <div class="colon">:</div>
                  <div class="value bold">${escapeHtml(participant.name || '-')}</div>
                </div>
                <div class="identity-row">
                  <div class="label">Kelas</div>
                  <div class="colon">:</div>
                  <div class="value">${escapeHtml(classLabel)}</div>
                </div>
                <div class="identity-row">
                  <div class="label">Semester</div>
                  <div class="colon">:</div>
                  <div class="value">${escapeHtml(semester)}</div>
                </div>
              </div>

              <div class="official-bottom">
                <div class="no-peserta-col">
                  <div class="title">Nomor Peserta</div>
                  <div class="number">${escapeHtml(participantNo)}</div>
                </div>
                <div class="photo-col">
                  <div class="participant-photo-box">
                    ${participantPhotoHtml}
                  </div>
                </div>
                <div class="signature-col">
                  <div class="signature-box">
                    <div>${escapeHtml(placeDate)}</div>
                    <div>Kepala Sekolah,</div>
                    <div class="ttd-space"></div>
                    <div class="name">${escapeHtml(principalName)}</div>
                    <div class="nip">NIP ${escapeHtml(principalNip)}</div>
                  </div>
                </div>
              </div>
            </div>
          `
        })
        .join('')

      openPrintWindow(
        `Kartu Ujian - ${selectedExamTitle()}`,
        `
          <style>
            @page { size: A4 portrait; margin: 6mm; }
            body { margin: 6mm !important; }
            body > h1 { font-size: 14px; margin: 0 0 2mm; text-align: center; }
            body > .meta { font-size: 10px; margin: 0 0 4mm; text-align: center; }
            .official-cards {
              display: grid;
              grid-template-columns: 1fr;
              gap: 4mm;
            }
            .official-card {
              border: 1.5pt solid #000;
              break-inside: avoid;
              page-break-inside: avoid;
              background: #fff;
              width: 100%;
              max-width: 180mm;
              margin: 0 auto;
              position: relative;
            }
            .official-header-row {
              display: grid;
              grid-template-columns: 25mm 1fr 25mm;
              border-bottom: 1.5pt solid #000;
              min-height: 25mm;
            }
            .official-logo-wrap {
              display: flex;
              align-items: center;
              justify-content: center;
              padding: 2mm;
            }
            .official-logo-wrap.left { border-right: 1pt solid #000; }
            .official-logo-wrap.right { border-left: 1pt solid #000; }
            .official-logo { width: 18mm; height: 18mm; object-fit: contain; }
            .official-logo-fallback { font-size: 9px; font-weight: bold; color: #666; }
            
            .official-header-text { text-align: center; padding: 2mm; display: flex; flex-direction: column; justify-content: center; }
            .official-header-text .line { line-height: 1.2; font-family: "Arial", sans-serif; }
            .official-header-text .line.small-bold { font-size: 9pt; font-weight: 700; }
            .official-header-text .line.school { font-size: 12pt; font-weight: 700; }
            .official-header-text .line.exam { font-size: 10pt; font-weight: 700; }
            .official-header-text .line.year { font-size: 9pt; font-weight: 700; }

            .official-identity {
              padding: 2mm 3mm;
              border-bottom: 1pt solid #000;
              display: flex;
              flex-direction: column;
              gap: 0.5mm;
            }
            .identity-row {
              display: grid;
              grid-template-columns: 20mm 4mm 1fr;
              font-size: 9pt;
              line-height: 1.3;
            }
            .identity-row .label { color: #111; }
            .identity-row .colon { text-align: left; }
            .identity-row .value { color: #111; }
            .identity-row .value.bold { font-weight: 700; }

            .official-bottom {
              display: grid;
              grid-template-columns: 1fr 1fr 1fr;
              min-height: 30mm;
            }
            .no-peserta-col {
              display: flex;
              flex-direction: column;
              align-items: center;
              justify-content: flex-start;
              padding-top: 4mm;
              border-right: 1pt solid #000;
            }
            .no-peserta-col .title { font-size: 10pt; font-weight: 700; text-decoration: underline; margin-bottom: 2mm; }
            .no-peserta-col .number { font-size: 14pt; font-weight: 700; }

            .photo-col {
              display: flex;
              align-items: center;
              justify-content: center;
              border-right: 1pt solid #000;
            }
            .participant-photo-box {
              width: 20mm;
              height: 25mm;
              border: 0.5pt solid #000;
              display: flex;
              align-items: center;
              justify-content: center;
              overflow: hidden;
            }
            .participant-photo { width: 100%; height: 100%; object-fit: cover; }
            .participant-photo-fallback { 
              font-size: 7pt; color: #666; text-align: center; 
            }

            .signature-col {
              display: flex;
              flex-direction: column;
              align-items: flex-end;
              justify-content: flex-start;
              padding: 2mm 4mm;
              font-size: 8pt;
            }
            .signature-box { text-align: right; width: 100%; }
            .signature-box .ttd-space { height: 12mm; }
            .signature-box .name { font-weight: 700; text-decoration: none; border-bottom: 0px; position: relative; }
            .signature-box .name::after { content: ""; position: absolute; left: 0; right: 0; bottom: -1px; border-bottom: 1pt solid #000; }
            .signature-box .nip { margin-top: 1px; }
          </style>
          <div class="meta">
            Ujian: <strong>${escapeHtml(selectedExamTitle())}</strong>
            · Total kartu: <strong>${participants.length}</strong>
            ${levelFilter.value !== 'all' ? `· Level: <strong>${escapeHtml(levelFilter.value)}</strong>` : ''}
            ${groupFilter.value !== 'all' ? `· Kelas: <strong>${escapeHtml(groupFilter.value)}</strong>` : ''}
            ${sessionStatusFilter.value !== 'all' ? `· Status sesi: <strong>${escapeHtml(sessionStatusFilter.value)}</strong>` : ''}
          </div>
          <div class="official-cards">${officialCardsHtml || '<div>Tidak ada peserta untuk dicetak.</div>'}</div>
        `,
      )
      return
    }

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
              <div class="signature-line">${escapeHtml(participant.name || '-')}</div>
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
        @page { size: A4 portrait; margin: 6mm; }
        body { margin: 6mm !important; }
        body > h1 { font-size: 14px; margin: 0 0 2mm; }
        body > .meta { font-size: 10px; margin: 0 0 2mm; }
        .meta { font-size: 10px; margin-bottom: 2mm; }
        .cards-grid {
          display: grid;
          grid-template-columns: repeat(2, minmax(0, 1fr));
          gap: 2.2mm;
        }
        .card {
          border: 1px solid #94a3b8;
          border-radius: 8px;
          padding: 2.2mm;
          break-inside: avoid;
          page-break-inside: avoid;
          height: 66mm;
          overflow: hidden;
        }
        .card-header {
          display: flex;
          align-items: center;
          gap: 2mm;
          margin-bottom: 1.5mm;
        }
        .school-logo {
          width: 10mm;
          height: 10mm;
          object-fit: contain;
        }
        .card-title {
          font-size: 10px;
          font-weight: 700;
          margin-bottom: 0;
        }
        .school-name {
          font-size: 9px;
          color: #475569;
        }
        .card-subtitle {
          font-size: 8px;
          color: #64748b;
          margin-top: 0.5mm;
        }
        .card td {
          border: 0;
          padding: 0.35mm 0;
          font-size: 8px;
          vertical-align: top;
          line-height: 1.15;
        }
        .card td:first-child {
          width: 22mm;
          color: #475569;
        }
        .signature-grid {
          display: grid;
          grid-template-columns: 1fr 1fr;
          gap: 1.6mm;
          margin-top: 2mm;
        }
        .signature {
          margin-top: 1.5mm;
          text-align: right;
          font-size: 8px;
          color: #334155;
        }
        .signature-line {
          margin-top: 4.5mm;
          font-size: 8px;
        }
      </style>
      <div class="meta">
        Ujian: <strong>${escapeHtml(selectedExamTitle())}</strong>
        · Total kartu: <strong>${participants.length}</strong>
        ${q ? `· Filter peserta: <strong>${escapeHtml(participantQuery.value.trim())}</strong>` : ''}
        ${levelFilter.value !== 'all' ? `· Level: <strong>${escapeHtml(levelFilter.value)}</strong>` : ''}
        ${groupFilter.value !== 'all' ? `· Kelas: <strong>${escapeHtml(groupFilter.value)}</strong>` : ''}
        ${sessionFilter.value !== 'all' ? `· Sesi: <strong>${escapeHtml(sessionFilter.value)}</strong>` : ''}
        ${sessionStatusFilter.value !== 'all' ? `· Status sesi: <strong>${escapeHtml(sessionStatusFilter.value)}</strong>` : ''}
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
onMounted(loadParticipantsCache)

watch(selectedTemplate, (value) => {
  if (typeof window === 'undefined') return
  window.localStorage.setItem(printTemplateStorageKey, value)
})
watch(selectedExamId, () => {
  levelFilter.value = 'all'
  groupFilter.value = 'all'
  sessionFilter.value = 'all'
  sessionStatusFilter.value = 'all'
  loadParticipantsCache()
}, { immediate: true })
watch(levelFilter, () => {
  if (groupFilter.value === 'all') return
  const exists = groupOptions.value.some((item) => item.value === groupFilter.value)
  if (!exists) groupFilter.value = 'all'
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
          <FormField label="Filter Level">
            <FormControl v-model="levelFilter" :options="levelOptions" />
          </FormField>
          <FormField label="Filter Kelas / Group">
            <FormControl v-model="groupFilter" :options="groupOptions" />
          </FormField>
          <FormField label="Filter Sesi">
            <FormControl v-model="sessionFilter" :options="sessionOptions" />
          </FormField>
          <FormField label="Status Sesi Ujian">
            <FormControl v-model="sessionStatusFilter" :options="sessionStatusOptions" />
          </FormField>
          <FormField label="Nama Penandatangan">
            <FormControl v-model="signatoryName" placeholder="Nama panitia / kepala sekolah" />
          </FormField>
          <FormField label="Jabatan Penandatangan">
            <FormControl v-model="signatoryTitle" placeholder="Panitia Ujian" />
          </FormField>
        </div>
        <div
          v-if="selectedTemplate === 'official'"
          class="mt-2 mb-4 rounded-xl border border-slate-200 dark:border-slate-700 p-4 bg-slate-50/70 dark:bg-slate-900/20"
        >
          <h4 class="font-bold text-slate-700 dark:text-slate-100 mb-3">Form Kartu Ujian Resmi</h4>
          <div class="grid gap-4 md:grid-cols-2">
            <FormField label="Header Baris 1">
              <FormControl v-model="printCardConfig.govLine1" placeholder="PEMERINTAH KABUPATEN ..." />
            </FormField>
            <FormField label="Header Baris 2">
              <FormControl v-model="printCardConfig.govLine2" placeholder="DINAS PENDIDIKAN ..." />
            </FormField>
            <FormField label="Nama Sekolah / UPTD">
              <FormControl v-model="printCardConfig.schoolLine" placeholder="UPTD SATUAN PENDIDIKAN ..." />
            </FormField>
            <FormField label="Logo Sekolah (Utama)">
              <div class="flex flex-col gap-2">
                <FormControl v-model="printCardConfig.schoolLogoUrl" placeholder="URL Logo atau upload di bawah" />
                <FormFilePicker v-model="logoFile1" label="Ambil dari Komputer" :icon="mdiUpload" accept=".png,.jpg,.jpeg,.webp" />
              </div>
            </FormField>
            <FormField label="Logo 2 (Kanan)">
              <div class="flex flex-col gap-2">
                <FormControl v-model="printCardConfig.schoolLogo2Url" placeholder="URL Logo 2 atau upload di bawah" />
                <FormFilePicker v-model="logoFile2" label="Ambil dari Komputer" :icon="mdiUpload" accept=".png,.jpg,.jpeg,.webp" />
              </div>
            </FormField>
            <FormField label="Nama Ujian">
              <FormControl v-model="printCardConfig.examLine" placeholder="ASESMEN SUMATIF AKHIR SEMESTER" />
            </FormField>
            <FormField label="Tahun Pelajaran">
              <FormControl v-model="printCardConfig.academicYear" placeholder="2025/2026" />
            </FormField>
            <FormField label="Semester">
              <FormControl v-model="printCardConfig.semester" placeholder="Ganjil / Genap" />
            </FormField>
            <FormField label="Tempat & Tanggal">
              <FormControl v-model="printCardConfig.placeDateLine" placeholder="Bangunrejo, 01 Desember 2025" />
            </FormField>
            <FormField label="Nama Kepala Sekolah">
              <FormControl v-model="printCardConfig.principalName" placeholder="Nama kepala sekolah" />
            </FormField>
            <FormField label="NIP Kepala Sekolah">
              <FormControl v-model="printCardConfig.principalNip" placeholder="1976..." />
            </FormField>
          </div>
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
          <div v-else-if="isLoadingParticipants" class="self-center text-sm text-slate-500 dark:text-slate-400 italic">Memuat daftar peserta...</div>
        </div>
      </CardBox>
    </SectionMain>
  </LayoutAuthenticated>
</template>
