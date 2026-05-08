<script setup>
import { computed, onMounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { mdiPrinterOutline, mdiRefresh, mdiUpload, mdiImageMultiple } from '@mdi/js'
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
const router = useRouter()

const exams = ref([])
const selectedExamId = ref('')
const isLoading = ref(false)
const errorMessage = ref('')
const participantQuery = ref('')

const participantsWithoutPhoto = computed(() => {
  return cachedParticipants.value.filter((p) => !p.photo_url).length
})
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
const selectedTemplate = ref('official')
const printTemplateStorageKey = 'atigacbt.print.templatePreset'
const templatePresets = [
  {
    value: 'official',
    label: 'Kartu Dinas (ASAT)',
    cardTitle: 'ASESMEN SUMATIF AKHIR SEMESTER',
    cardSubtitle: 'Template kartu ujian resmi sekolah',
    signatoryTitle: 'Kepala Sekolah',
  },
  {
    value: 'modern',
    label: 'Kartu Modern (ID Card)',
    cardTitle: 'KARTU PESERTA UJIAN',
    cardSubtitle: 'Kartu identitas peserta ujian sekolah',
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
  schoolStampUrl: '',
  principalSignatureUrl: '',
})
const logoFile1 = ref(null)
const logoFile2 = ref(null)
const stampFile = ref(null)
const signatureFile = ref(null)

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

watch(stampFile, (file) => {
  if (!file) return
  const reader = new FileReader()
  reader.onload = (e) => {
    printCardConfig.value.schoolStampUrl = e.target.result
  }
  reader.readAsDataURL(file)
})

watch(signatureFile, (file) => {
  if (!file) return
  const reader = new FileReader()
  reader.onload = (e) => {
    printCardConfig.value.principalSignatureUrl = e.target.result
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
          <td>${escapeHtml(row.participant_no || row.nis || '-')}</td>
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
            <th>No Peserta</th>
            <th>NIS</th>
            <th>Level</th>
            <th>Group</th>
            <th>Status</th>
            <th>Waktu Hadir</th>
            <th style="width:140px;">Tanda Tangan</th>
          </tr>
        </thead>
        <tbody>${tableRows || '<tr><td colspan="10">Belum ada data.</td></tr>'}</tbody>
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
          <td>${escapeHtml(row.participant_no || row.student_nis || '-')}</td>
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
            <th>No Peserta</th>
            <th>NIS</th>
            <th>Status</th>
            <th>Benar</th>
            <th>Nilai</th>
          </tr>
        </thead>
        <tbody>${tableRows || '<tr><td colspan="8">Belum ada data.</td></tr>'}</tbody>
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
            @page { size: A4 portrait; margin: 8mm; }
            body { margin: 0 !important; font-family: "Arial", sans-serif; }
            body > h1 { font-size: 14px; margin: 2mm 0; text-align: center; }
            body > .meta { font-size: 10px; margin: 0 0 4mm; text-align: center; opacity: 0.7; }
            .official-cards {
              display: grid;
              grid-template-columns: repeat(2, 1fr);
              gap: 2mm;
              padding: 2mm;
            }
            .official-card {
              border: 1pt solid #000;
              break-inside: avoid;
              page-break-inside: avoid;
              background: #fff;
              width: 100%;
              min-height: 68mm;
              max-height: 70mm;
              position: relative;
              overflow: hidden;
              display: flex;
              flex-direction: column;
            }
            .official-header-row {
              display: grid;
              grid-template-columns: 18mm 1fr 18mm;
              border-bottom: 1pt solid #000;
              min-height: 18mm;
            }
            .official-logo-wrap {
              display: flex;
              align-items: center;
              justify-content: center;
              padding: 1mm;
            }
            .official-logo-wrap.left { border-right: 0.5pt solid #000; }
            .official-logo-wrap.right { border-left: 0.5pt solid #000; }
            .official-logo { width: 14mm; height: 14mm; object-fit: contain; }
            .official-logo-fallback { font-size: 7px; font-weight: bold; color: #666; }
            
            .official-header-text { text-align: center; padding: 1mm; display: flex; flex-direction: column; justify-content: center; }
            .official-header-text .line { line-height: 1.1; }
            .official-header-text .line.small-bold { font-size: 6.5pt; font-weight: 700; text-transform: uppercase; }
            .official-header-text .line.school { font-size: 8.5pt; font-weight: 700; text-transform: uppercase; }
            .official-header-text .line.exam { font-size: 7.5pt; font-weight: 700; }
            .official-header-text .line.year { font-size: 7pt; font-weight: 700; }

            .official-identity {
              padding: 1.5mm 2mm;
              border-bottom: 0.5pt solid #000;
              display: flex;
              flex-direction: column;
              gap: 0.3mm;
              flex-grow: 1;
            }
            .identity-row {
              display: grid;
              grid-template-columns: 14mm 3mm 1fr;
              font-size: 7.5pt;
              line-height: 1.2;
            }
            .identity-row .label { color: #111; }
            .identity-row .value.bold { font-weight: 700; }

            .official-bottom {
              display: grid;
              grid-template-columns: 1fr 20mm 1fr;
              min-height: 25mm;
            }
            .no-peserta-col {
              display: flex;
              flex-direction: column;
              align-items: center;
              justify-content: flex-start;
              padding-top: 2mm;
              border-right: 0.5pt solid #000;
            }
            .no-peserta-col .title { font-size: 7.5pt; font-weight: 700; text-decoration: underline; margin-bottom: 1mm; }
            .no-peserta-col .number { font-size: 11pt; font-weight: 900; letter-spacing: -0.5px; }

            .photo-col {
              display: flex;
              align-items: center;
              justify-content: center;
              border-right: 0.5pt solid #000;
            }
            .participant-photo-box {
              width: 15mm;
              height: 20mm;
              border: 0.5pt solid #000;
              display: flex;
              align-items: center;
              justify-content: center;
              overflow: hidden;
            }
            .participant-photo { width: 100%; height: 100%; object-fit: cover; }
            .participant-photo-fallback { font-size: 6pt; color: #666; text-align: center; }

            .signature-col {
              display: flex;
              flex-direction: column;
              align-items: center;
              justify-content: flex-start;
              padding: 1.5mm;
              font-size: 6.5pt;
            }
            .signature-box { text-align: center; width: 100%; }
            .signature-box .ttd-space { height: 8mm; }
            .signature-box .name { font-weight: 700; text-decoration: underline; }
            .signature-box .nip { margin-top: 0.5px; }
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

    if (template.value === 'modern') {
      const modernCardsHtml = participants
        .map((participant) => {
          const participantNo = participant.exam_no || participant.participant_no || participant.nis || '-'
          const classLabel = `${participant.level_name || ''} ${participant.group_name || ''}`.trim() || '-'
          const semester = printCardConfig.value.semester?.trim() || '-'
          const schoolLine = printCardConfig.value.schoolLine?.trim() || (schoolIdentity.value.school_name || 'SEKOLAH')
          const examLine = printCardConfig.value.examLine?.trim() || 'KARTU PESERTA UJIAN'
          const academicYear = printCardConfig.value.academicYear?.trim() || '-'
          const placeDate = printCardConfig.value.placeDateLine?.trim() || '-'
          const principalName = printCardConfig.value.principalName?.trim() || signatureName
          const principalNip = printCardConfig.value.principalNip?.trim() || '-'
          const logoURL = resolveAssetURL(printCardConfig.value.schoolLogoUrl || schoolIdentity.value.logo_url)
          const participantPhotoURL = resolveAssetURL(participant.photo_url)

          const logoHtmlMod = logoURL
            ? `<img src="${escapeHtml(logoURL)}" alt="Logo" class="mod-logo" />`
            : `<div class="mod-logo-fallback">LOGO</div>`
          const photoHtmlMod = participantPhotoURL
            ? `<img src="${escapeHtml(participantPhotoURL)}" alt="Foto" class="mod-photo-img" />`
            : `<div class="mod-photo-empty">FOTO<br/>PESERTA</div>`

          return `
            <div class="mod-card">
              <!-- Colored Header Band -->
              <div class="mod-header">
                <div class="mod-header-logo">${logoHtmlMod}</div>
                <div class="mod-header-text">
                  <div class="mod-school">${escapeHtml(schoolLine)}</div>
                  <div class="mod-exam-title">${escapeHtml(examLine)}</div>
                  <div class="mod-year">Tahun Pelajaran ${escapeHtml(academicYear)} &bull; Semester ${escapeHtml(semester)}</div>
                </div>
              </div>

              <!-- Body: data kiri + foto kanan -->
              <div class="mod-body">
                <div class="mod-data">
                  <div class="mod-no-peserta">
                    <span class="mod-no-label">No. Peserta</span>
                    <span class="mod-no-value">${escapeHtml(participantNo)}</span>
                  </div>
                  <div class="mod-row"><span class="mod-key">Nama</span><span class="mod-sep">:</span><span class="mod-val bold">${escapeHtml(participant.name || '-')}</span></div>
                  <div class="mod-row"><span class="mod-key">NIS</span><span class="mod-sep">:</span><span class="mod-val">${escapeHtml(participant.nis || '-')}</span></div>
                  <div class="mod-row"><span class="mod-key">Kelas</span><span class="mod-sep">:</span><span class="mod-val">${escapeHtml(classLabel)}</span></div>
                  <div class="mod-row"><span class="mod-key">Ujian</span><span class="mod-sep">:</span><span class="mod-val">${escapeHtml(selectedExamTitle())}</span></div>
                </div>
                <div class="mod-photo-wrap">
                  <div class="mod-photo-box">${photoHtmlMod}</div>
                  <div class="mod-photo-label">Pas Foto</div>
                </div>
              </div>

              <!-- Footer: tanda tangan -->
              <div class="mod-footer">
                <div class="mod-sig">
                  <div>${escapeHtml(placeDate)}</div>
                  <div class="mod-sig-title">Kepala Sekolah,</div>
                  <div class="mod-sig-space">
                    <!-- Stempel overlap ke nama -->
                    <div class="mod-stamp">${resolveAssetURL(printCardConfig.value.schoolStampUrl) ? `<img src="${escapeHtml(resolveAssetURL(printCardConfig.value.schoolStampUrl))}" alt="Stempel" class="mod-stamp-img" />` : '<span>STEMPEL</span>'}</div>
                    <!-- Tanda tangan kepala sekolah -->
                    ${resolveAssetURL(printCardConfig.value.principalSignatureUrl) ? `<img src="${escapeHtml(resolveAssetURL(printCardConfig.value.principalSignatureUrl))}" alt="TTD" class="mod-ttd-img" />` : ''}
                  </div>
                  <div class="mod-sig-name">${escapeHtml(principalName)}</div>
                  <div class="mod-sig-nip">NIP. ${escapeHtml(principalNip)}</div>
                </div>
              </div>

              <!-- Bottom accent stripe -->
              <div class="mod-bottom-stripe"></div>
            </div>
          `
        })
        .join('')

      openPrintWindow(
        `Kartu Peserta Ujian - ${selectedExamTitle()}`,
        `
          <style>
            @page { size: A4 portrait; margin: 8mm; }
            body { margin: 0 !important; font-family: Arial, sans-serif; background: #f0f0f0; }
            body > .meta { font-size: 9px; margin: 0 0 3mm; text-align: center; color: #555; }
            .mod-grid {
              display: grid;
              grid-template-columns: repeat(2, 1fr);
              gap: 3mm;
            }
            .mod-card {
              background: #fff;
              border-radius: 2mm;
              overflow: hidden;
              break-inside: avoid;
              page-break-inside: avoid;
              box-shadow: 0 0.5mm 1.5mm rgba(0,0,0,0.12);
              display: flex;
              flex-direction: column;
              min-height: 68mm;
              max-height: 72mm;
              border: 0.5pt solid #c8d3e0;
              position: relative;
            }
            /* === Header === */
            .mod-header {
              background: linear-gradient(135deg, #1a3a5c 0%, #1e5799 100%);
              display: flex;
              align-items: center;
              gap: 2mm;
              padding: 2mm 2.5mm;
              min-height: 16mm;
            }
            .mod-header-logo {
              flex-shrink: 0;
              width: 13mm;
              height: 13mm;
              display: flex;
              align-items: center;
              justify-content: center;
            }
            .mod-logo { width: 13mm; height: 13mm; object-fit: contain; }
            .mod-logo-fallback {
              width: 13mm;
              height: 13mm;
              background: rgba(255,255,255,0.15);
              border: 0.5pt solid rgba(255,255,255,0.4);
              display: flex;
              align-items: center;
              justify-content: center;
              font-size: 6pt;
              color: #fff;
              font-weight: bold;
            }
            .mod-header-text { flex: 1; }
            .mod-school {
              font-size: 8pt;
              font-weight: 900;
              color: #fff;
              text-transform: uppercase;
              letter-spacing: 0.3px;
              line-height: 1.15;
            }
            .mod-exam-title {
              font-size: 6.5pt;
              font-weight: 700;
              color: #ffe082;
              text-transform: uppercase;
              margin-top: 0.5mm;
              line-height: 1.1;
            }
            .mod-year {
              font-size: 5.5pt;
              color: rgba(255,255,255,0.75);
              margin-top: 0.5mm;
            }
            /* === Body === */
            .mod-body {
              display: flex;
              flex: 1;
              padding: 2mm 2.5mm;
              gap: 2mm;
              align-items: flex-start;
            }
            .mod-data { flex: 1; }
            .mod-no-peserta {
              display: flex;
              flex-direction: column;
              margin-bottom: 1.5mm;
              padding-bottom: 1.5mm;
              border-bottom: 0.5pt solid #d1d5db;
            }
            .mod-no-label {
              font-size: 5.5pt;
              color: #6b7280;
              font-weight: 600;
              text-transform: uppercase;
              letter-spacing: 0.5px;
            }
            .mod-no-value {
              font-size: 14pt;
              font-weight: 900;
              color: #1a3a5c;
              line-height: 1;
              letter-spacing: -0.5px;
            }
            .mod-row {
              display: grid;
              grid-template-columns: 10mm 3mm 1fr;
              font-size: 7pt;
              line-height: 1.3;
              margin-bottom: 0.3mm;
              color: #1f2937;
            }
            .mod-key { color: #6b7280; }
            .mod-val.bold { font-weight: 700; }
            .mod-photo-wrap {
              flex-shrink: 0;
              display: flex;
              flex-direction: column;
              align-items: center;
              gap: 0.5mm;
            }
            .mod-photo-box {
              width: 18mm;
              height: 23mm;
              border: 1pt solid #1a3a5c;
              overflow: hidden;
              display: flex;
              align-items: center;
              justify-content: center;
              background: #f8fafc;
            }
            .mod-photo-img { width: 100%; height: 100%; object-fit: cover; }
            .mod-photo-empty {
              font-size: 6pt;
              color: #9ca3af;
              text-align: center;
              line-height: 1.4;
            }
            .mod-photo-label {
              font-size: 5.5pt;
              color: #9ca3af;
              text-align: center;
            }
            /* === Footer === */
            .mod-footer {
              display: flex;
              justify-content: flex-end;
              align-items: flex-end;
              padding: 1mm 2.5mm 1.5mm;
              border-top: 0.5pt solid #d1d5db;
            }
            .mod-sig {
              text-align: center;
              font-size: 6pt;
              color: #374151;
              position: relative;
            }
            .mod-sig-title { font-size: 6.5pt; font-weight: 600; margin-top: 0.5mm; }
            .mod-sig-space {
              height: 10mm;
              position: relative;
            }
            /* Stempel overlap ke kiri, menindih nama kepala sekolah */
            .mod-stamp {
              position: absolute;
              left: -10mm;
              bottom: -7mm;
              width: 20mm;
              height: 20mm;
              border: 0.75pt dashed #9ca3af;
              border-radius: 50%;
              display: flex;
              align-items: center;
              justify-content: center;
              color: #9ca3af;
              font-size: 5.5pt;
              overflow: hidden;
              opacity: 0.85;
              z-index: 2;
            }
            .mod-stamp-img {
              width: 100%;
              height: 100%;
              object-fit: contain;
              border-radius: 50%;
            }
            .mod-sig-name {
              font-weight: 700;
              text-decoration: underline;
              font-size: 7pt;
              position: relative;
              z-index: 1;
            }
            .mod-sig-nip { font-size: 5.5pt; color: #6b7280; margin-top: 0.3mm; }
            /* Tanda tangan kepala sekolah */
            .mod-ttd-img {
              position: absolute;
              left: 50%;
              transform: translateX(-50%);
              bottom: 0;
              height: 9mm;
              max-width: 28mm;
              object-fit: contain;
              z-index: 3;
              opacity: 0.88;
            }
            /* === Bottom accent === */
            .mod-bottom-stripe {
              height: 2mm;
              background: linear-gradient(90deg, #1a3a5c 0%, #ffe082 50%, #1a3a5c 100%);
            }
          </style>
          <div class="meta">
            Ujian: <strong>${escapeHtml(selectedExamTitle())}</strong>
            &bull; Total kartu: <strong>${participants.length}</strong>
            ${levelFilter.value !== 'all' ? `&bull; Level: <strong>${escapeHtml(levelFilter.value)}</strong>` : ''}
            ${groupFilter.value !== 'all' ? `&bull; Kelas: <strong>${escapeHtml(groupFilter.value)}</strong>` : ''}
            ${sessionStatusFilter.value !== 'all' ? `&bull; Status sesi: <strong>${escapeHtml(sessionStatusFilter.value)}</strong>` : ''}
          </div>
          <div class="mod-grid">${modernCardsHtml || '<div>Tidak ada peserta untuk dicetak.</div>'}</div>
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

      <!-- Banner: ada peserta tanpa foto -->
      <div
        v-if="selectedExamId && !isLoadingParticipants && participantsWithoutPhoto > 0"
        class="mb-4 flex items-center gap-3 rounded-xl border border-violet-200 dark:border-violet-800 bg-violet-50 dark:bg-violet-900/20 px-4 py-3"
      >
        <div class="flex h-9 w-9 flex-shrink-0 items-center justify-center rounded-lg bg-violet-100 dark:bg-violet-900/40">
          <BaseIcon :path="mdiImageMultiple" size="20" class="text-violet-600 dark:text-violet-400" />
        </div>
        <div class="flex-1 min-w-0">
          <p class="text-sm font-semibold text-violet-800 dark:text-violet-200">
            {{ participantsWithoutPhoto }} peserta belum punya foto
          </p>
          <p class="text-xs text-violet-600 dark:text-violet-400">
            Foto akan muncul kosong pada kartu ujian. Upload sekarang sebelum mencetak.
          </p>
        </div>
        <button
          class="flex-shrink-0 rounded-lg bg-violet-600 px-3 py-1.5 text-xs font-bold text-white hover:bg-violet-700 transition-colors"
          @click="router.push('/admin/students')"
        >
          Import Foto &rarr;
        </button>
      </div>

      <CardBox color="blue">
        <div class="p-3 md:p-5 mb-6 rounded-2xl border border-slate-100 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-900/10">
          <h3 class="font-black text-slate-800 dark:text-slate-100 mb-5 uppercase tracking-tighter text-sm flex items-center gap-2">
            <div class="w-1.5 h-4 bg-info rounded-full"></div>
            Pengaturan & Filter Cetak
          </h3>
          <div class="grid gap-5 md:grid-cols-2">
            <FormField label="Pilih Ujian">
              <FormControl
                v-model="selectedExamId"
                :options="exams.map((item) => ({ value: item.id, label: item.title }))"
              />
            </FormField>
            <FormField label="Preset Template Kartu">
              <div class="flex flex-col sm:flex-row gap-2 sm:gap-3 sm:items-center">
                <div class="flex-1">
                  <FormControl
                    v-model="selectedTemplate"
                    :options="templatePresets.map((item) => ({ value: item.value, label: item.label }))"
                  />
                </div>
                <div class="bg-purple-100/50 dark:bg-purple-900/20 p-1.5 rounded-2xl border border-purple-200 dark:border-purple-800 flex items-center shadow-sm">
                  <BaseButton 
                    color="purple" 
                    label="Terapkan" 
                    @click="applyTemplatePreset" 
                    class="!bg-purple-600 !border-purple-700 hover:!bg-purple-700 text-white font-bold shadow-md w-full sm:min-w-[100px]"
                  />
                </div>
              </div>
            </FormField>
            <FormField label="Cetak Kartu Siswa Tertentu (opsional)">
              <FormControl v-model="participantQuery" placeholder="Ketik nama, username, atau NIS siswa..." />
              <p class="mt-1.5 text-xs text-slate-400 dark:text-slate-500">
                💡 <strong>Kosongkan</strong> untuk mencetak kartu seluruh kelas/sesi. Isi hanya jika ingin mencetak ulang kartu <strong>1 siswa tertentu</strong> (mis. kartu hilang/rusak).
              </p>
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
          </div>
        </div>
        <div class="grid gap-4 md:grid-cols-2 mb-6">
          <FormField label="Nama Penandatangan">
            <FormControl v-model="signatoryName" placeholder="Nama panitia / kepala sekolah" />
          </FormField>
          <FormField label="Jabatan Penandatangan">
            <FormControl v-model="signatoryTitle" placeholder="Panitia Ujian" />
          </FormField>
        </div>
        <div
          v-if="selectedTemplate === 'official' || selectedTemplate === 'modern'"
          class="mt-2 mb-4 rounded-xl border border-slate-200 dark:border-slate-700 p-4 bg-slate-50/70 dark:bg-slate-900/20"
        >
          <h4 class="font-bold text-slate-700 dark:text-slate-100 mb-3">
            {{ selectedTemplate === 'modern' ? 'Form Kartu Modern' : 'Form Kartu Ujian Resmi' }}
          </h4>
          <div class="grid gap-4 md:grid-cols-2">
            <FormField v-if="selectedTemplate === 'official'" label="Header Baris 1">
              <FormControl v-model="printCardConfig.govLine1" placeholder="PEMERINTAH KABUPATEN ..." />
            </FormField>
            <FormField v-if="selectedTemplate === 'official'" label="Header Baris 2">
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
            <FormField v-if="selectedTemplate === 'official'" label="Logo 2 (Kanan)">
              <div class="flex flex-col gap-2">
                <FormControl v-model="printCardConfig.schoolLogo2Url" placeholder="URL Logo 2 atau upload di bawah" />
                <FormFilePicker v-model="logoFile2" label="Ambil dari Komputer" :icon="mdiUpload" accept=".png,.jpg,.jpeg,.webp" />
              </div>
            </FormField>
            <!-- Stempel sekolah — hanya untuk Kartu Modern -->
            <FormField v-if="selectedTemplate === 'modern'" label="Stempel Sekolah">
              <div class="flex flex-col gap-2">
                <FormControl v-model="printCardConfig.schoolStampUrl" placeholder="URL Stempel atau upload di bawah" />
                <FormFilePicker v-model="stampFile" label="Upload Stempel dari Komputer" :icon="mdiUpload" accept=".png,.jpg,.jpeg,.webp" />
                <!-- Preview stempel -->
                <div v-if="printCardConfig.schoolStampUrl" class="flex items-center gap-3 mt-1">
                  <div class="w-14 h-14 rounded-full border-2 border-dashed border-slate-300 overflow-hidden flex items-center justify-center bg-slate-50 dark:bg-slate-800">
                    <img :src="printCardConfig.schoolStampUrl" alt="Preview stempel" class="w-full h-full object-contain" />
                  </div>
                  <div class="text-xs text-slate-500 dark:text-slate-400">
                    Preview stempel bulat
                    <a href="#" class="block text-red-500 hover:underline mt-0.5" @click.prevent="printCardConfig.schoolStampUrl = ''; stampFile = null">× Hapus</a>
                  </div>
                </div>
              </div>
            </FormField>
            <!-- Tanda tangan kepala sekolah — hanya untuk Kartu Modern -->
            <FormField v-if="selectedTemplate === 'modern'" label="Tanda Tangan Kepala Sekolah">
              <div class="flex flex-col gap-2">
                <FormControl v-model="printCardConfig.principalSignatureUrl" placeholder="URL gambar TTD atau upload di bawah" />
                <FormFilePicker v-model="signatureFile" label="Upload Tanda Tangan dari Komputer" :icon="mdiUpload" accept=".png,.jpg,.jpeg,.webp" />
                <p class="text-xs text-slate-400 dark:text-slate-500">💡 Gunakan gambar TTD dengan latar <strong>transparan (PNG)</strong> agar terlihat alami di atas kartu.</p>
                <!-- Preview TTD -->
                <div v-if="printCardConfig.principalSignatureUrl" class="flex items-center gap-3 mt-1">
                  <div class="h-12 px-3 border border-dashed border-slate-300 rounded-lg overflow-hidden flex items-center justify-center bg-white dark:bg-slate-800">
                    <img :src="printCardConfig.principalSignatureUrl" alt="Preview TTD" class="h-full object-contain max-w-[120px]" />
                  </div>
                  <div class="text-xs text-slate-500 dark:text-slate-400">
                    Preview tanda tangan
                    <a href="#" class="block text-red-500 hover:underline mt-0.5" @click.prevent="printCardConfig.principalSignatureUrl = ''; signatureFile = null">× Hapus</a>
                  </div>
                </div>
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
        <div class="flex flex-col sm:flex-row flex-wrap gap-3">
          <BaseButton
            :icon="mdiPrinterOutline"
            color="info"
            label="Cetak Daftar Hadir"
            :disabled="isLoading || !selectedExamId"
            class="w-full sm:w-auto"
            @click="printAttendance"
          />
          <BaseButton
            :icon="mdiPrinterOutline"
            color="purple"
            label="Cetak Laporan Nilai"
            :disabled="isLoading || !selectedExamId"
            class="w-full sm:w-auto"
            @click="printResults"
          />
          <BaseButton
            :icon="mdiPrinterOutline"
            color="success"
            label="Cetak Kartu Ujian"
            :disabled="isLoading || !selectedExamId"
            class="w-full sm:w-auto"
            @click="printExamCards"
          />
          <div v-if="isLoading" class="self-center text-sm text-slate-500 dark:text-slate-400 italic">Memuat daftar ujian...</div>
          <div v-else-if="isLoadingParticipants" class="self-center text-sm text-slate-500 dark:text-slate-400 italic">Memuat daftar peserta...</div>
        </div>
      </CardBox>
    </SectionMain>
  </LayoutAuthenticated>
</template>
