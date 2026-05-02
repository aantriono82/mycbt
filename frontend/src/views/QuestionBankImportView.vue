<script setup>
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  mdiFileDocumentOutline,
  mdiRefresh,
  mdiDownload,
  mdiContentCopy,
  mdiEye,
  mdiContentSave,
  mdiPencil,
  mdiPlus,
} from '@mdi/js'
import LayoutAuthenticated from '@/layouts/LayoutAuthenticated.vue'
import SectionMain from '@/components/SectionMain.vue'
import SectionTitleLineWithButton from '@/components/SectionTitleLineWithButton.vue'
import CardBox from '@/components/CardBox.vue'
import BaseButton from '@/components/BaseButton.vue'
import BaseButtons from '@/components/BaseButtons.vue'
import FormField from '@/components/FormField.vue'
import FormControl from '@/components/FormControl.vue'
import FormFilePicker from '@/components/FormFilePicker.vue'
import { api } from '@/services/api.js'

const route = useRoute()
const router = useRouter()

const questionSets = ref([])
const subjects = ref([])
const levels = ref([])
const selectedSetId = ref(String(route.query.set_id || ''))
const docxFile = ref(null)
const previewQuestions = ref([])
const previewWarnings = ref([])
const isLoading = ref(false)
const isPreviewing = ref(false)
const isImporting = ref(false)
const isCreatingSet = ref(false)
const errorMessage = ref('')
const successMessage = ref('')
const createErrors = reactive({
  subject_id: '',
  title: '',
})

const createForm = reactive({
  subject_id: '',
  title: '',
  jenjang: '',
  level_id: '',
})

const rolePrefix = computed(() => (route.path.startsWith('/admin') ? '/admin' : '/teacher'))
const templateDocxUrl = computed(() => {
  const base = import.meta.env.BASE_URL || '/'
  const normalized = base.endsWith('/') ? base : `${base}/`
  return `${normalized}templates/template-soal-docx.docx`
})

const selectedSet = computed(() => questionSets.value.find((item) => item.id === selectedSetId.value) || null)
const canCreateSet = computed(() => !!createForm.subject_id && !!String(createForm.title || '').trim() && !isCreatingSet.value)

const previewByType = computed(() => {
  const counts = {}
  for (const item of previewQuestions.value) {
    const key = item?.type || 'unknown'
    counts[key] = (counts[key] || 0) + 1
  }
  return Object.entries(counts)
})

const normalizedOptionLabel = (label) => String(label || '').trim().toUpperCase()

const buildFixSuggestion = (type, message) => {
  if (message === 'Tipe soal belum terdeteksi.') {
    return 'Tambahkan tag tipe pada baris soal, misalnya [mc_single], [matching], atau [essay].'
  }
  if (message === 'Isi soal / stem kosong.') {
    return 'Isi teks soal setelah nomor, contoh: 1. Berapakah hasil 2 + 2?'
  }
  if (message.startsWith('Tipe soal tidak didukung:')) {
    return 'Gunakan salah satu tipe yang didukung: mc_single, mc_multiple, matching, short_answer, essay, true_false.'
  }
  if (message === 'Pilihan ganda minimal harus punya 2 opsi.') {
    return 'Tambahkan minimal dua opsi, misalnya A. ... dan B. ...'
  }
  if (message === 'Semua opsi harus punya label dan isi.') {
    return 'Pastikan format opsi lengkap, contoh: A. Jakarta'
  }
  if (message.startsWith('Label opsi duplikat:')) {
    return 'Gunakan label unik berurutan seperti A, B, C, D.'
  }
  if (message === 'mc_single harus memiliki tepat 1 jawaban benar.') {
    return 'Tambahkan satu kunci jawaban saja, misalnya Answer: A'
  }
  if (message === 'mc_multiple harus memiliki minimal 1 jawaban benar.') {
    return 'Tambahkan kunci dengan satu atau lebih opsi benar, misalnya Answer: A,C'
  }
  if (message === 'Soal menjodohkan minimal harus punya 2 pasangan.') {
    return 'Tambahkan minimal dua baris pasangan, contoh: 1 => satu dan 2 => dua'
  }
  if (message === 'Setiap pasangan matching harus punya sisi kiri dan kanan.') {
    return 'Lengkapi format pasangan menjadi kiri => kanan pada setiap baris.'
  }
  if (message === 'Isian singkat minimal harus punya 1 jawaban yang diterima.') {
    return 'Tambahkan kunci isian singkat, misalnya Answer: 3/4 | 0.75'
  }
  if (message === 'Jawaban isian singkat tidak boleh kosong.') {
    return 'Hapus jawaban kosong atau isi semua alternatif jawaban yang diterima.'
  }
  if (message === 'Semua pernyataan true/false harus memiliki isi.') {
    return 'Isi semua pernyataan true/false atau hapus baris yang kosong.'
  }
  if (message === 'Soal true_false harus memiliki kunci benar/salah.') {
    return 'Tambahkan Answer: benar atau Answer: salah'
  }
  if (message === 'Nilai maksimum essay tidak boleh negatif.') {
    return 'Gunakan nilai maksimum 0 atau lebih besar, misalnya 10 atau 100.'
  }

  if (type === 'mc_single' || type === 'mc_multiple') {
    return 'Periksa format opsi dan kunci jawaban pada soal pilihan ganda.'
  }
  if (type === 'matching') {
    return 'Periksa kembali format pasangan dengan pola kiri => kanan.'
  }
  if (type === 'short_answer') {
    return 'Periksa kembali format Answer: untuk isian singkat.'
  }
  if (type === 'true_false') {
    return 'Periksa kembali kunci benar/salah atau isi pernyataan.'
  }
  if (type === 'essay') {
    return 'Periksa kembali konfigurasi nilai maksimum atau rubric essay.'
  }

  return 'Periksa kembali format soal pada template impor.'
}

const validatePreviewQuestion = (item) => {
  const errors = []
  const type = String(item?.type || '').trim()
  const stem = String(item?.stem || '').trim()

  if (!type) {
    errors.push('Tipe soal belum terdeteksi.')
    return errors
  }

  if (!stem) {
    errors.push('Isi soal / stem kosong.')
  }

  if (!['mc_single', 'mc_multiple', 'matching', 'short_answer', 'essay', 'true_false'].includes(type)) {
    errors.push(`Tipe soal tidak didukung: ${type}`)
    return errors
  }

  if (type === 'mc_single' || type === 'mc_multiple') {
    const options = Array.isArray(item?.options) ? item.options : []
    const seen = new Set()
    let correctCount = 0

    if (options.length < 2) {
      errors.push('Pilihan ganda minimal harus punya 2 opsi.')
    }

    for (const opt of options) {
      const label = normalizedOptionLabel(opt?.label)
      const content = String(opt?.content || '').trim()

      if (!label || !content) {
        errors.push('Semua opsi harus punya label dan isi.')
        continue
      }
      if (seen.has(label)) {
        errors.push(`Label opsi duplikat: ${label}`)
      }
      seen.add(label)
      if (opt?.is_correct) correctCount++
    }

    if (type === 'mc_single' && correctCount !== 1) {
      errors.push('mc_single harus memiliki tepat 1 jawaban benar.')
    }
    if (type === 'mc_multiple' && correctCount < 1) {
      errors.push('mc_multiple harus memiliki minimal 1 jawaban benar.')
    }
  }

  if (type === 'matching') {
    const pairs = Array.isArray(item?.pairs) ? item.pairs : []
    if (pairs.length < 2) {
      errors.push('Soal menjodohkan minimal harus punya 2 pasangan.')
    }
    for (const pair of pairs) {
      const left = String(pair?.left_content || '').trim()
      const right = String(pair?.right_content || '').trim()
      if (!left || !right) {
        errors.push('Setiap pasangan matching harus punya sisi kiri dan kanan.')
        break
      }
    }
  }

  if (type === 'short_answer') {
    const answers = Array.isArray(item?.answers) ? item.answers : []
    if (answers.length < 1) {
      errors.push('Isian singkat minimal harus punya 1 jawaban yang diterima.')
    }
    for (const answer of answers) {
      const answerText = String(answer?.answer_text || '').trim()
      if (!answerText) {
        errors.push('Jawaban isian singkat tidak boleh kosong.')
        break
      }
    }
  }

  if (type === 'true_false') {
    const statements = Array.isArray(item?.statements) ? item.statements : []
    if (statements.length > 0) {
      for (const statement of statements) {
        const content = String(statement?.content || '').trim()
        if (!content) {
          errors.push('Semua pernyataan true/false harus memiliki isi.')
          break
        }
      }
    } else if (typeof item?.true_false?.correct !== 'boolean') {
      errors.push('Soal true_false harus memiliki kunci benar/salah.')
    }
  }

  if (type === 'essay') {
    const maxScore = item?.essay?.max_score
    if (maxScore != null && Number(maxScore) < 0) {
      errors.push('Nilai maksimum essay tidak boleh negatif.')
    }
  }

  return [...new Set(errors)]
}

const previewValidation = computed(() =>
  previewQuestions.value.map((item) => ({
    orderNo: item?.order_no,
    type: item?.type,
    errors: validatePreviewQuestion(item),
  })),
)

const previewValidationDetailed = computed(() =>
  previewValidation.value.map((item) => ({
    orderNo: item.orderNo,
    type: item.type,
    issues: item.errors.map((message) => ({
      message,
      suggestion: buildFixSuggestion(item.type, message),
    })),
  })),
)

const previewValidationMap = computed(() => {
  const map = new Map()
  for (const item of previewValidationDetailed.value) {
    map.set(item.orderNo, item.issues)
  }
  return map
})

const previewBlockingIssues = computed(() =>
  previewValidationDetailed.value.filter((item) => item.issues.length > 0),
)

const hasPreviewBlockingIssues = computed(() => previewBlockingIssues.value.length > 0)
const isPreviewReady = computed(() => previewQuestions.value.length > 0)
const previewRepairHints = computed(() => {
  const seen = new Set()
  const out = []
  for (const item of previewBlockingIssues.value) {
    for (const issue of item.issues) {
      if (seen.has(issue.suggestion)) continue
      seen.add(issue.suggestion)
      out.push(issue.suggestion)
    }
  }
  return out
})

const validationReportText = computed(() => {
  const lines = []
  const now = new Date().toLocaleString('id-ID')

  lines.push('LAPORAN VALIDASI IMPORT SOAL')
  lines.push(`Waktu: ${now}`)
  lines.push(`File: ${docxFile.value?.name || '-'}`)
  lines.push(`Bank soal tujuan: ${selectedSet.value?.title || '-'}`)
  lines.push(`Jumlah soal terdeteksi: ${previewQuestions.value.length}`)
  lines.push(`Status validasi: ${hasPreviewBlockingIssues.value ? 'PERLU PERBAIKAN' : 'VALID'}`)
  lines.push('')

  if (previewByType.value.length) {
    lines.push('Ringkasan tipe soal:')
    for (const [type, count] of previewByType.value) {
      lines.push(`- ${type}: ${count}`)
    }
    lines.push('')
  }

  if (previewWarnings.value.length) {
    lines.push('Peringatan parser:')
    for (const warning of previewWarnings.value) {
      lines.push(`- ${warning}`)
    }
    lines.push('')
  }

  if (previewBlockingIssues.value.length) {
    lines.push('Error validasi per soal:')
    for (const issue of previewBlockingIssues.value) {
      lines.push(`- Soal ${issue.orderNo} (${issue.type || 'unknown'})`)
      for (const item of issue.issues) {
        lines.push(`  * ${item.message}`)
        lines.push(`    Saran: ${item.suggestion}`)
      }
    }
  } else if (previewQuestions.value.length) {
    lines.push('Tidak ada error validasi blocking.')
  } else {
    lines.push('Belum ada hasil preview.')
  }

  return lines.join('\n')
})

const loadLookups = async () => {
  isLoading.value = true
  errorMessage.value = ''
  try {
    const [setsRes, subjectsRes, levelsRes] = await Promise.allSettled([
      api.get('/api/v1/question-sets', { params: { limit: 100, offset: 0 } }),
      api.get('/api/v1/lookups/subjects'),
      api.get('/api/v1/lookups/levels'),
    ])

    questionSets.value = setsRes.status === 'fulfilled' ? setsRes.value?.data?.data || [] : []
    subjects.value = subjectsRes.status === 'fulfilled' ? subjectsRes.value?.data?.data || [] : []
    levels.value = levelsRes.status === 'fulfilled' ? levelsRes.value?.data?.data || [] : []

    if (setsRes.status !== 'fulfilled' || subjectsRes.status !== 'fulfilled') {
      throw new Error('Gagal memuat data bank soal')
    }
  } catch (error) {
    errorMessage.value = error?.response?.data?.error?.message || 'Gagal memuat data bank soal'
  } finally {
    isLoading.value = false
  }
}

const resetCreateErrors = () => {
  createErrors.subject_id = ''
  createErrors.title = ''
}

const resetCreateForm = () => {
  createForm.subject_id = ''
  createForm.title = ''
  createForm.jenjang = ''
  createForm.level_id = ''
  resetCreateErrors()
}

const goToEditor = () => {
  if (!selectedSetId.value) return
  router.push(`${rolePrefix.value}/bank-soal/new?id=${selectedSetId.value}`)
}

const resetPreview = () => {
  previewQuestions.value = []
  previewWarnings.value = []
}

const createQuestionSet = async () => {
  errorMessage.value = ''
  successMessage.value = ''
  resetCreateErrors()

  if (!createForm.subject_id) {
    createErrors.subject_id = 'Mata pelajaran wajib dipilih'
  }
  if (!createForm.title.trim()) {
    createErrors.title = 'Judul bank soal wajib diisi'
  }
  if (createErrors.subject_id || createErrors.title) return

  isCreatingSet.value = true
  try {
    const { data } = await api.post('/api/v1/question-sets', {
      subject_id: createForm.subject_id,
      title: createForm.title.trim(),
      jenjang: createForm.jenjang.trim(),
      level_id: createForm.level_id,
    })
    selectedSetId.value = data?.data?.id || ''
    await loadLookups()
    await router.replace({ query: { ...route.query, set_id: selectedSetId.value } })
    successMessage.value = 'Bank soal baru berhasil dibuat dan siap dipakai untuk impor.'
    resetCreateForm()
  } catch (error) {
    errorMessage.value = error?.response?.data?.error?.message || 'Gagal membuat bank soal baru'
  } finally {
    isCreatingSet.value = false
  }
}

const validateImportForm = () => {
  errorMessage.value = ''
  successMessage.value = ''

  if (!selectedSetId.value) {
    errorMessage.value = 'Pilih bank soal tujuan terlebih dahulu'
    return false
  }

  if (!docxFile.value) {
    errorMessage.value = 'Pilih file .docx terlebih dahulu'
    return false
  }

  return true
}

const previewDocx = async () => {
  if (!validateImportForm()) return

  isPreviewing.value = true
  resetPreview()

  const formData = new FormData()
  formData.append('file', docxFile.value)

  try {
    const { data } = await api.post(`/api/v1/question-sets/${selectedSetId.value}/import-docx/preview`, formData)
    previewQuestions.value = data?.data?.questions || []
    previewWarnings.value = data?.data?.warnings || []
    successMessage.value = `Preview berhasil. Terdeteksi ${previewQuestions.value.length} soal.`
  } catch (error) {
    errorMessage.value = error?.response?.data?.error?.message || 'Gagal mem-preview file DOCX'
  } finally {
    isPreviewing.value = false
  }
}

const importDocx = async () => {
  if (!validateImportForm()) return

  isImporting.value = true

  const formData = new FormData()
  formData.append('file', docxFile.value)

  try {
    const { data } = await api.post(`/api/v1/question-sets/${selectedSetId.value}/import-docx`, formData)
    const count = data?.meta?.count || data?.data?.questions?.length || 0
    previewWarnings.value = data?.data?.warnings || []
    successMessage.value = `Import selesai. ${count} soal ditambahkan ke bank soal.`
    docxFile.value = null
    await loadLookups()
  } catch (error) {
    errorMessage.value = error?.response?.data?.error?.message || 'Gagal mengimpor soal dari DOCX'
  } finally {
    isImporting.value = false
  }
}

const copyValidationReport = async () => {
  if (!isPreviewReady.value) return
  try {
    await navigator.clipboard.writeText(validationReportText.value)
    successMessage.value = 'Laporan validasi disalin ke clipboard.'
    errorMessage.value = ''
  } catch {
    errorMessage.value = 'Gagal menyalin laporan validasi ke clipboard'
  }
}

const downloadValidationReport = () => {
  if (!isPreviewReady.value) return

  const setName = String(selectedSet.value?.title || 'import-soal')
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-+|-+$/g, '') || 'import-soal'

  const blob = new Blob([validationReportText.value], { type: 'text/plain;charset=utf-8' })
  const href = window.URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = href
  link.download = `laporan-validasi-${setName}-${Date.now()}.txt`
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  window.setTimeout(() => window.URL.revokeObjectURL(href), 1000)
  successMessage.value = 'Laporan validasi berhasil diunduh.'
  errorMessage.value = ''
}

watch(
  () => route.query.set_id,
  (value) => {
    selectedSetId.value = String(value || '')
    resetPreview()
  },
)

watch(selectedSetId, () => {
  resetPreview()
  successMessage.value = ''
  errorMessage.value = ''

  const nextQuery = { ...route.query }
  if (selectedSetId.value) nextQuery.set_id = selectedSetId.value
  else delete nextQuery.set_id

  if ((route.query.set_id || '') !== (selectedSetId.value || '')) {
    router.replace({ query: nextQuery })
  }
})

watch(docxFile, () => {
  resetPreview()
  successMessage.value = ''
  errorMessage.value = ''
})

onMounted(async () => {
  await loadLookups()
})
</script>

<template>
  <LayoutAuthenticated>
    <SectionMain>
      <SectionTitleLineWithButton :icon="mdiFileDocumentOutline" title="Impor Soal" main>
        <div class="flex items-center gap-2">
          <BaseButton :icon="mdiRefresh" color="info" label="Refresh" @click="loadLookups" />
          <BaseButton :icon="mdiPencil" color="purple" label="Buka Editor" :disabled="!selectedSetId" @click="goToEditor" />
        </div>
      </SectionTitleLineWithButton>

      <div class="grid gap-6 xl:grid-cols-5">
        <CardBox class="xl:col-span-2 shadow-md">
          <div class="grid gap-5">
            <div>
              <h3 class="text-lg font-bold text-slate-900 dark:text-slate-100">Tujuan Import</h3>
              <p class="mt-1 text-sm text-slate-500">Pilih bank soal yang akan menerima hasil import dari file DOCX.</p>
            </div>

            <div v-if="errorMessage" class="rounded-2xl border border-red-100 bg-red-50 px-4 py-3 text-sm text-red-700">
              {{ errorMessage }}
            </div>
            <div v-if="successMessage" class="rounded-2xl border border-emerald-100 bg-emerald-50 px-4 py-3 text-sm text-emerald-700">
              {{ successMessage }}
            </div>

            <FormField label="Bank Soal Tujuan">
              <FormControl
                v-model="selectedSetId"
                :options="[
                  { value: '', label: isLoading ? 'Memuat bank soal...' : 'Pilih bank soal' },
                  ...questionSets.map((item) => ({
                    value: item.id,
                    label: `${item.title} (${subjects.find((s) => s.id === item.subject_id)?.name || item.subject_id})`,
                  })),
                ]"
                :disabled="isLoading"
              />
            </FormField>

            <div class="rounded-2xl border border-slate-200 bg-slate-50 px-4 py-4 dark:border-slate-800 dark:bg-slate-900/40">
              <div class="flex items-center justify-between gap-3">
                <div>
                  <div class="text-xs font-black uppercase tracking-widest text-slate-400">Buat Bank Soal Baru</div>
                  <p class="mt-1 text-sm text-slate-500">
                    Jika belum ada set tujuan yang cocok, buat langsung di sini lalu lanjut preview/import tanpa pindah halaman.
                  </p>
                </div>
                <BaseButton
                  :icon="mdiPlus"
                  color="success"
                  :label="isCreatingSet ? 'Membuat...' : 'Buat Set'"
                  :disabled="!canCreateSet"
                  @click="createQuestionSet"
                />
              </div>

              <div class="mt-4 grid gap-4">
                <FormField label="Mata Pelajaran" :error="createErrors.subject_id">
                  <FormControl
                    v-model="createForm.subject_id"
                    :options="[
                      { value: '', label: 'Pilih mata pelajaran' },
                      ...subjects.map((item) => ({
                        value: item.id,
                        label: `${item.code || '-'} - ${item.name}`,
                      })),
                    ]"
                    :disabled="isLoading"
                  />
                </FormField>

                <FormField label="Judul Bank Soal" :error="createErrors.title">
                  <FormControl v-model="createForm.title" placeholder="Contoh: MTK Kelas X - Paket DOCX 01" :disabled="isLoading || isCreatingSet" />
                </FormField>

                <div class="grid gap-4 xl:grid-cols-2">
                  <FormField label="Jenjang">
                    <FormControl
                      v-model="createForm.jenjang"
                      :options="[
                        { value: '', label: 'Pilih jenjang' },
                        { value: 'SD', label: 'SD' },
                        { value: 'SMP', label: 'SMP' },
                        { value: 'SMA', label: 'SMA' },
                        { value: 'SMK', label: 'SMK' },
                      ]"
                      :disabled="isLoading || isCreatingSet"
                    />
                  </FormField>

                  <FormField label="Level / Kelas">
                    <FormControl
                      v-model="createForm.level_id"
                      :options="[
                        { value: '', label: isLoading ? 'Memuat level...' : 'Opsional' },
                        ...levels.map((item) => ({
                          value: item.id,
                          label: item.name,
                        })),
                      ]"
                      :disabled="isLoading || isCreatingSet"
                    />
                  </FormField>
                </div>
              </div>
            </div>

            <div v-if="selectedSet" class="rounded-2xl border border-slate-200 bg-slate-50 px-4 py-4 dark:border-slate-800 dark:bg-slate-900/40">
              <div class="text-xs font-black uppercase tracking-widest text-slate-400">Set Dipilih</div>
              <div class="mt-2 text-lg font-bold text-slate-900 dark:text-white">{{ selectedSet.title }}</div>
              <div class="mt-1 text-sm text-slate-500">
                {{ subjects.find((s) => s.id === selectedSet.subject_id)?.name || selectedSet.subject_id }}
              </div>
            </div>

            <div class="rounded-2xl border border-purple-200 bg-purple-50 p-4 text-purple-900 dark:border-purple-500/20 dark:bg-purple-500/10 dark:text-purple-100">
              <div class="flex items-start justify-between gap-4">
                <div>
                  <div class="text-xs font-black uppercase tracking-widest text-purple-500">Template DOCX</div>
                  <p class="mt-2 text-sm leading-relaxed text-purple-800 dark:text-purple-100/80">
                    Template sudah menyiapkan format semua tipe soal, termasuk LaTeX dengan delimiter <span class="font-mono">$...$</span> dan <span class="font-mono">$$...$$</span>.
                  </p>
                </div>
                <BaseButton
                  :icon="mdiDownload"
                  color="purple"
                  label="Template"
                  :href="templateDocxUrl"
                  download="template-soal-docx.docx"
                />
              </div>
            </div>

            <div class="rounded-2xl border border-slate-200 bg-white px-4 py-4 dark:border-slate-800 dark:bg-slate-900/40">
              <div class="text-xs font-black uppercase tracking-widest text-slate-400">File Import</div>
              <div class="mt-3 flex flex-wrap items-center gap-3">
                <FormFilePicker
                  v-model="docxFile"
                  label="Pilih File DOCX"
                  accept=".docx,application/vnd.openxmlformats-officedocument.wordprocessingml.document"
                />
                <span class="text-xs text-slate-400">Format: `.docx`</span>
              </div>
              <BaseButtons class="mt-4">
                <BaseButton :icon="mdiEye" color="info" :label="isPreviewing ? 'Mem-preview...' : 'Preview'" :disabled="isPreviewing || isImporting || !docxFile || !selectedSetId" @click="previewDocx" />
                <BaseButton :icon="mdiContentSave" color="purple" :label="isImporting ? 'Mengimpor...' : 'Impor Sekarang'" :disabled="isImporting || isPreviewing || !docxFile || !selectedSetId || !isPreviewReady || hasPreviewBlockingIssues" @click="importDocx" />
              </BaseButtons>
              <p class="mt-3 text-xs text-slate-400">
                Import aktif setelah preview berhasil dan semua soal lolos validasi.
              </p>
            </div>
          </div>
        </CardBox>

        <CardBox class="xl:col-span-3 shadow-md">
          <div class="grid gap-5">
            <div class="flex items-center justify-between gap-4">
              <div>
                <h3 class="text-lg font-bold text-slate-900 dark:text-slate-100">Preview Hasil Parse</h3>
                <p class="mt-1 text-sm text-slate-500">Periksa hasil pembacaan parser sebelum file diimpor ke bank soal.</p>
              </div>
              <div class="rounded-2xl border border-slate-200 bg-slate-50 px-4 py-3 text-center dark:border-slate-800 dark:bg-slate-900/40">
                <div class="text-[10px] font-black uppercase tracking-widest text-slate-400">Jumlah Soal</div>
                <div class="mt-1 text-2xl font-black text-slate-900 dark:text-white">{{ previewQuestions.length }}</div>
              </div>
            </div>

            <div v-if="previewByType.length" class="flex flex-wrap gap-2">
              <div
                v-for="[type, count] in previewByType"
                :key="type"
                class="rounded-full border border-blue-200 bg-blue-50 px-3 py-1 text-xs font-bold text-blue-700 dark:border-blue-500/20 dark:bg-blue-500/10 dark:text-blue-300"
              >
                {{ type }}: {{ count }}
              </div>
            </div>

            <div
              v-if="isPreviewReady"
              class="rounded-2xl border px-4 py-4 text-sm"
              :class="hasPreviewBlockingIssues ? 'border-red-200 bg-red-50 text-red-800' : 'border-emerald-200 bg-emerald-50 text-emerald-800'"
            >
              <div class="text-xs font-black uppercase tracking-widest" :class="hasPreviewBlockingIssues ? 'text-red-500' : 'text-emerald-500'">
                Status Validasi
              </div>
              <p class="mt-2">
                <span v-if="hasPreviewBlockingIssues">
                  Ditemukan {{ previewBlockingIssues.length }} soal yang belum valid. Perbaiki file sumber lalu preview ulang sebelum import.
                </span>
                <span v-else>
                  Semua soal lolos validasi dasar dan siap diimpor.
                </span>
              </p>
              <div class="mt-4 flex flex-wrap gap-2">
                <BaseButton :icon="mdiContentCopy" color="whiteDark" outline small label="Copy Error" :disabled="!isPreviewReady" @click="copyValidationReport" />
                <BaseButton :icon="mdiDownload" color="whiteDark" outline small label="Download Laporan" :disabled="!isPreviewReady" @click="downloadValidationReport" />
              </div>
            </div>

            <div v-if="previewRepairHints.length" class="rounded-2xl border border-blue-200 bg-blue-50 px-4 py-4 text-sm text-blue-800">
              <div class="text-xs font-black uppercase tracking-widest text-blue-500">Saran Cepat</div>
              <ul class="mt-3 list-disc space-y-2 pl-5">
                <li v-for="(hint, index) in previewRepairHints" :key="`hint-${index}`">{{ hint }}</li>
              </ul>
            </div>

            <div v-if="previewWarnings.length" class="rounded-2xl border border-amber-200 bg-amber-50 px-4 py-4 text-sm text-amber-800">
              <div class="text-xs font-black uppercase tracking-widest text-amber-500">Peringatan Parser</div>
              <ul class="mt-3 list-disc space-y-1 pl-5">
                <li v-for="(warning, index) in previewWarnings" :key="`${warning}-${index}`">{{ warning }}</li>
              </ul>
            </div>

            <div v-if="!previewQuestions.length" class="rounded-2xl border-2 border-dashed border-slate-200 px-8 py-16 text-center text-sm text-slate-400 dark:border-slate-800">
              Preview belum tersedia. Pilih bank soal, unggah file DOCX, lalu klik <span class="font-semibold text-slate-600 dark:text-slate-300">Preview</span>.
            </div>

            <div v-else class="max-h-[44rem] space-y-4 overflow-y-auto pr-1">
              <div
                v-for="item in previewQuestions"
                :key="item.order_no"
                class="rounded-2xl border border-slate-200 bg-slate-50 px-4 py-4 dark:border-slate-800 dark:bg-slate-900/40"
              >
                <div class="flex items-center justify-between gap-4">
                  <div class="flex items-center gap-3">
                    <div class="text-sm font-black text-slate-900 dark:text-white">Soal {{ item.order_no }}</div>
                    <div
                      class="rounded-full px-3 py-1 text-[11px] font-black uppercase tracking-widest"
                      :class="(previewValidationMap.get(item.order_no) || []).length ? 'bg-red-100 text-red-700 dark:bg-red-500/10 dark:text-red-300' : 'bg-emerald-100 text-emerald-700 dark:bg-emerald-500/10 dark:text-emerald-300'"
                    >
                      {{ (previewValidationMap.get(item.order_no) || []).length ? 'Perlu Perbaikan' : 'Valid' }}
                    </div>
                  </div>
                  <div class="rounded-full bg-purple-100 px-3 py-1 text-[11px] font-black uppercase tracking-widest text-purple-700 dark:bg-purple-500/10 dark:text-purple-300">
                    {{ item.type || 'unknown' }}
                  </div>
                </div>
                <div class="mt-3 whitespace-pre-line text-sm leading-relaxed text-slate-700 dark:text-slate-200">{{ item.stem }}</div>

                <div
                  v-if="(previewValidationMap.get(item.order_no) || []).length"
                  class="mt-4 rounded-xl border border-red-200 bg-red-50 px-3 py-3 text-sm text-red-800 dark:border-red-500/20 dark:bg-red-500/10 dark:text-red-300"
                >
                  <div class="text-[11px] font-black uppercase tracking-widest text-red-500">Error Validasi</div>
                  <ul class="mt-2 list-disc space-y-2 pl-5">
                    <li
                      v-for="(issue, index) in previewValidationMap.get(item.order_no) || []"
                      :key="`${item.order_no}-error-${index}`"
                    >
                      <div>{{ issue.message }}</div>
                      <div class="mt-1 text-xs text-red-700/80 dark:text-red-200/80">
                        Saran: {{ issue.suggestion }}
                      </div>
                    </li>
                  </ul>
                </div>

                <div v-if="item.options?.length" class="mt-4 grid gap-2">
                  <div
                    v-for="opt in item.options"
                    :key="`${item.order_no}-${opt.label}`"
                    class="rounded-xl border px-3 py-2 text-sm"
                    :class="opt.is_correct ? 'border-emerald-200 bg-emerald-50 text-emerald-700 dark:border-emerald-500/20 dark:bg-emerald-500/10 dark:text-emerald-300' : 'border-slate-200 bg-white text-slate-600 dark:border-slate-800 dark:bg-slate-950/30 dark:text-slate-300'"
                  >
                    <span class="font-bold">{{ opt.label }}.</span> {{ opt.content }}
                  </div>
                </div>

                <div v-if="item.pairs?.length" class="mt-4 grid gap-2">
                  <div
                    v-for="(pair, index) in item.pairs"
                    :key="`${item.order_no}-pair-${index}`"
                    class="grid gap-2 rounded-xl border border-slate-200 bg-white px-3 py-2 text-sm text-slate-700 dark:border-slate-800 dark:bg-slate-950/30 dark:text-slate-200 xl:grid-cols-[1fr_auto_1fr]"
                  >
                    <span>{{ pair.left_content }}</span>
                    <span class="font-black text-slate-400">=&gt;</span>
                    <span>{{ pair.right_content }}</span>
                  </div>
                </div>

                <div v-if="item.answers?.length" class="mt-4 rounded-xl border border-blue-200 bg-blue-50 px-3 py-3 text-sm text-blue-700 dark:border-blue-500/20 dark:bg-blue-500/10 dark:text-blue-300">
                  Jawaban diterima: {{ item.answers.map((ans) => ans.answer_text).join(' | ') }}
                </div>

                <div v-if="item.true_false" class="mt-4 rounded-xl border border-violet-200 bg-violet-50 px-3 py-3 text-sm text-violet-700 dark:border-violet-500/20 dark:bg-violet-500/10 dark:text-violet-300">
                  Kunci: {{ item.true_false.correct ? 'Benar' : 'Salah' }}
                </div>
              </div>
            </div>
          </div>
        </CardBox>
      </div>
    </SectionMain>
  </LayoutAuthenticated>
</template>
