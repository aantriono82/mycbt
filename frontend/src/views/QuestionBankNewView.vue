<script setup>
import { onMounted, reactive, ref, computed, watch, nextTick } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import {
  mdiDatabasePlus,
  mdiArrowLeft,
  mdiContentSave,
  mdiPlus,
  mdiRefresh,
  mdiDelete,
  mdiFileDocumentOutline,
  mdiPencil,
  mdiContentCopy,
  mdiBookOpenOutline,
  mdiCheckCircleOutline,
  mdiEye,
} from '@mdi/js'
import LayoutAuthenticated from '@/layouts/LayoutAuthenticated.vue'
import SectionMain from '@/components/SectionMain.vue'
import SectionTitleLineWithButton from '@/components/SectionTitleLineWithButton.vue'
import CardBox from '@/components/CardBox.vue'
import BaseButton from '@/components/BaseButton.vue'
import BaseIcon from '@/components/BaseIcon.vue'
import FormField from '@/components/FormField.vue'
import FormControl from '@/components/FormControl.vue'
import FormFilePicker from '@/components/FormFilePicker.vue'
import { api } from '@/services/api.js'
import BaseButtons from '@/components/BaseButtons.vue'
import CardBoxModal from '@/components/CardBoxModal.vue'
import RichEditor from '@/components/RichEditor.vue'
import QuestionQuickAddCard from '@/components/QuestionQuickAddCard.vue'


const router = useRouter()
const route = useRoute()
const routeRolePrefix = computed(() => (route.path.startsWith('/admin') ? '/admin' : '/teacher'))

// State management
const currentSetId = ref(route.query.id || '')
const isStepEditor = computed(() => !!currentSetId.value)
const subjects = ref([])
const levels = ref([])
const currentSet = ref(null)
const questions = ref([])
const isLoading = ref(false)
const errorMessage = ref('')
const successMessage = ref('')
const isEditingMetadata = ref(false)
const initializedSetId = ref('')
const questionPopulateSeq = ref(0)

const getNextOrderNo = () => {
  const maxOrder = (questions.value || []).reduce((m, q) => {
    const n = Number(q?.order_no) || 0
    return n > m ? n : m
  }, 0)
  return maxOrder + 1
}

// Step 1: Creation Form
const createForm = reactive({
  subject_id: '',
  title: '',
  jenjang: '',
  level_id: '',
})

// Step 2: Question Editor
const editingQuestionId = ref('')
const editorRenderVersion = ref(0)
const questionForm = reactive({
  type: 'mc_single',
  stem: '',
  order_no: 1,
  weight: 1,
  options_text: 'A|Opsi A|true\nB|Opsi B|false',
  answers_text: '',
  pairs_text: '',
  correct: true,
  rubric_text: '',
  max_score: 100,
})

const activeEditorScopeKey = computed(() => {
  const base = editingQuestionId.value || `new-${questionForm.type}-${questionForm.order_no}`
  return `${base}-v${editorRenderVersion.value}`
})

const questionTypeOptions = reactive([
  { value: 'mc_single', label: 'Pilihan Ganda' },
  { value: 'mc_multiple', label: 'PG Kompleks' },
  { value: 'matching', label: 'Menjodohkan' },
  { value: 'short_answer', label: 'Isian Singkat' },
  { value: 'essay', label: 'Uraian' },
  { value: 'true_false', label: 'Benar / Salah' },
])
const allowedQuestionTypes = computed(() => questionTypeOptions.map((item) => item.value))
const normalizeEditorType = (value) => {
  const normalized = String(value || '').trim()
  if (!normalized || !allowedQuestionTypes.value.includes(normalized)) return 'mc_single'
  return normalized
}

const mcOptions = reactive([
  { label: 'A', content: '', is_correct: true },
  { label: 'B', content: '', is_correct: false },
  { label: 'C', content: '', is_correct: false },
  { label: 'D', content: '', is_correct: false },
  { label: 'E', content: '', is_correct: false },
])

const mcCorrectLabel = computed({
  get: () => mcOptions.find(o => o.is_correct)?.label || 'A',
  set: (label) => {
    mcOptions.forEach(o => o.is_correct = o.label === label)
  }
})

const alphaLabels = 'ABCDEFGHIJKLMNOPQRSTUVWXYZ'

const addMcOption = () => {
  const nextLabel = alphaLabels[mcOptions.length] || String(mcOptions.length + 1)
  const wasCorrect = mcOptions.length === 0
  mcOptions.push({ label: nextLabel, content: '', is_correct: wasCorrect })
}

const removeMcOption = (idx) => {
  if (mcOptions.length <= 2) return
  const wasCorrect = mcOptions[idx].is_correct
  mcOptions.splice(idx, 1)
  mcOptions.forEach((o, i) => { o.label = alphaLabels[i] || String(i + 1) })
  // Ensure at least one correct
  if (wasCorrect && mcOptions.length > 0) mcOptions[0].is_correct = true
}

// mc_multiple state (PG Kompleks) — each option has its own is_correct checkbox
const mcMultipleOptions = reactive([
  { label: 'A', content: '', is_correct: false },
  { label: 'B', content: '', is_correct: false },
  { label: 'C', content: '', is_correct: false },
  { label: 'D', content: '', is_correct: false },
])

const addMcMultipleOption = () => {
  const nextLabel = alphaLabels[mcMultipleOptions.length] || String(mcMultipleOptions.length + 1)
  mcMultipleOptions.push({ label: nextLabel, content: '', is_correct: false })
}

const removeMcMultipleOption = (idx) => {
  if (mcMultipleOptions.length <= 2) return
  mcMultipleOptions.splice(idx, 1)
  // Re-assign labels
  mcMultipleOptions.forEach((o, i) => { o.label = alphaLabels[i] || String(i + 1) })
}

const resetMcMultiple = () => {
  mcMultipleOptions.splice(0, mcMultipleOptions.length)
  ;['A', 'B', 'C', 'D'].forEach(l => mcMultipleOptions.push({ label: l, content: '', is_correct: false }))
}

// short_answer state — list of accepted answer texts
const shortAnswers = reactive([
  { text: '' }
])

const addShortAnswer = () => {
  shortAnswers.push({ text: '' })
}

const removeShortAnswer = (idx) => {
  if (shortAnswers.length <= 1) return
  shortAnswers.splice(idx, 1)
}

const resetShortAnswers = () => {
  shortAnswers.splice(0, shortAnswers.length)
  shortAnswers.push({ text: '' })
}


// true_false state — list of statements each with their own correct value
const trueFalseStatements = reactive([
  { content: '', correct: true }
])

const alphaLabels2 = 'ABCDEFGHIJKLMNOPQRSTUVWXYZ'

const addTrueFalseStatement = () => {
  trueFalseStatements.push({ content: '', correct: true })
}

const removeTrueFalseStatement = (idx) => {
  if (trueFalseStatements.length <= 1) return
  trueFalseStatements.splice(idx, 1)
}

const resetTrueFalseStatements = () => {
  trueFalseStatements.splice(0, trueFalseStatements.length)
  trueFalseStatements.push({ content: '', correct: true })
}

// matching state — list of left/right pairs
const matchingPairs = reactive([
  { left_content: '', right_content: '' },
  { left_content: '', right_content: '' },
])

const addMatchingPair = () => {
  matchingPairs.push({ left_content: '', right_content: '' })
}

const removeMatchingPair = (idx) => {
  if (matchingPairs.length <= 1) return
  matchingPairs.splice(idx, 1)
}

const resetMatchingPairs = () => {
  matchingPairs.splice(0, matchingPairs.length)
  matchingPairs.push({ left_content: '', right_content: '' })
  matchingPairs.push({ left_content: '', right_content: '' })
}

// Helper: strip HTML tags to plain text for backend matching
function stripHtml(html) {
  const div = document.createElement('div')
  div.innerHTML = html
  return div.textContent || div.innerText || ''
}

const extractAcceptedAnswerText = (html) => {
  const raw = String(html || '')
  if (!raw.trim()) return ''

  try {
    const doc = new DOMParser().parseFromString(raw, 'text/html')
    doc.querySelectorAll('.math-tex').forEach((node) => {
      const latex = String(node.getAttribute('data-latex') || '').trim()
      const fallback = String(node.textContent || '').trim()
      node.replaceWith(doc.createTextNode(latex || fallback))
    })
    return String(doc.body.textContent || '').replace(/\s+/g, ' ').trim()
  } catch {
    return String(stripHtml(raw) || '').replace(/\s+/g, ' ').trim()
  }
}

const escapeHtml = (value) =>
  String(value ?? '')
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#39;')

const summarizeText = (value, maxLen = 120) => {
  const text = String(stripHtml(value || '') || '').replace(/\s+/g, ' ').trim()
  if (!text) return ''
  if (text.length <= maxLen) return text
  return `${text.slice(0, maxLen - 1)}…`
}

const formatQuestionId = (value) => {
  const id = String(value || '').trim()
  if (!id) return '-'
  if (id.length <= 16) return id
  return `${id.slice(0, 8)}...${id.slice(-4)}`
}

const buildQuestionCardAnswerPreviewHtml = (item) => {
  if (!item || typeof item !== 'object') return ''
  const type = String(item.type || '').trim()
  const lines = []

  if (type === 'mc_single' || type === 'mc_multiple') {
    const options = Array.isArray(item.options) ? item.options : []
    for (const opt of options) {
      const label = String(opt?.label || '').trim()
      const content = summarizeText(opt?.content, 80)
      if (!label && !content) continue
      const keyTag = opt?.is_correct ? ' (kunci)' : ''
      lines.push(`${label ? `${label}. ` : ''}${content}${keyTag}`)
    }
  } else if (type === 'short_answer') {
    const answers = Array.isArray(item.answers) ? item.answers : []
    for (const ans of answers) {
      const text = summarizeText(ans?.answer_text, 90)
      if (!text) continue
      lines.push(text)
    }
  } else if (type === 'matching') {
    const pairs = Array.isArray(item.pairs) ? item.pairs : []
    for (const pair of pairs) {
      const left = summarizeText(pair?.left_content, 36)
      const right = summarizeText(pair?.right_content, 36)
      if (!left && !right) continue
      lines.push(`${left} → ${right}`)
    }
  } else if (type === 'true_false') {
    const statements = Array.isArray(item.statements) ? item.statements : []
    if (statements.length) {
      statements.forEach((st, idx) => {
        const content = summarizeText(st?.content, 70)
        const truth = st?.correct ? 'Benar' : 'Salah'
        if (!content) return
        lines.push(`${idx + 1}. ${content} (${truth})`)
      })
    } else if (item.true_false && typeof item.true_false.correct === 'boolean') {
      lines.push(`Kunci: ${item.true_false.correct ? 'Benar' : 'Salah'}`)
    }
  } else if (type === 'essay') {
    const rubric = summarizeText(item?.essay?.rubric_text, 90)
    const maxScore = item?.essay?.max_score
    if (rubric) lines.push(`Rubrik: ${rubric}`)
    if (maxScore !== null && maxScore !== undefined && String(maxScore) !== '') {
      lines.push(`Skor maksimal: ${maxScore}`)
    }
  }

  const compact = lines.filter(Boolean).slice(0, 4)
  if (!compact.length) return ''
  return `<ul class="space-y-1">${compact.map((line) => `<li class="line-clamp-1">${escapeHtml(line)}</li>`).join('')}</ul>`
}


const docxFile = ref(null)
const isImporting = ref(false)

const loadSubjects = async () => {
  try {
    const { data } = await api.get('/api/v1/lookups/subjects')
    subjects.value = data?.data || []
  } catch {
    subjects.value = []
  }
}

const loadLevels = async () => {
  try {
    const { data } = await api.get('/api/v1/lookups/levels')
    levels.value = data?.data || []
  } catch {
    levels.value = []
  }
}

const loadSetDetail = async () => {
  if (!currentSetId.value) return
  isLoading.value = true
  try {
    const { data } = await api.get(`/api/v1/question-sets/${currentSetId.value}`)
    currentSet.value = data?.data
    if (currentSet.value) {
      createForm.title = currentSet.value.title
      createForm.subject_id = currentSet.value.subject_id
      createForm.jenjang = currentSet.value.jenjang || ''
      createForm.level_id = currentSet.value.level_id || ''
    }
  } catch {
    errorMessage.value = 'Gagal memuat detail bank soal'
  } finally {
    isLoading.value = false
  }
}

const loadQuestions = async () => {
  if (!currentSetId.value) return
  try {
    const { data } = await api.get(`/api/v1/question-sets/${currentSetId.value}/questions`)
    questions.value = data?.data || []
    if (!editingQuestionId.value) {
      questionForm.order_no = questions.value.length + 1
      questionForm.weight = 1
    }
  } catch {
    errorMessage.value = 'Gagal memuat pertanyaan'
  }
}

const initEditorFromFirstQuestion = async () => {
  if (!currentSetId.value) return
  if (initializedSetId.value === currentSetId.value) return
  initializedSetId.value = currentSetId.value

  if (questions.value.length > 0) {
    await populateQuestionForm(questions.value[0])
    return
  }
  resetQuestionForm(false)
}

const createAndContinue = async () => {
  if (!createForm.title || !createForm.subject_id) {
    errorMessage.value = 'Mohon lengkapi semua data'
    return
  }
  isLoading.value = true
  try {
    const { data } = await api.post('/api/v1/question-sets', {
      subject_id: createForm.subject_id,
      title: createForm.title,
      jenjang: createForm.jenjang,
      level_id: createForm.level_id,
    })
    currentSetId.value = data.data.id
    // Update URL without reloading to keep user on "the same page" logically but track ID
    router.replace({ query: { id: data.data.id } })
    await loadSetDetail()
    await loadQuestions()
  } catch (err) {
    errorMessage.value = 'Gagal membuat bank soal'
  } finally {
    isLoading.value = false
  }
}

const updateMetadata = async () => {
  if (!createForm.title) {
    errorMessage.value = 'Judul wajib diisi'
    return
  }
  isLoading.value = true
  try {
    await api.patch(`/api/v1/question-sets/${currentSetId.value}`, {
      title: createForm.title,
      status: currentSet.value.status,
      jenjang: createForm.jenjang,
      level_id: createForm.level_id,
    })
    await loadSetDetail()
    successMessage.value = 'Informasi bank soal diperbarui'
    isEditingMetadata.value = false
  } catch (err) {
    errorMessage.value = 'Gagal memperbarui informasi'
  } finally {
    isLoading.value = false
  }
}

const resetQuestionForm = (keepType = true) => {
  const lastType = questionForm.type
  editingQuestionId.value = ''
  if (!keepType) questionForm.type = 'mc_single'
  else questionForm.type = lastType
  
  questionForm.stem = ''
  questionForm.order_no = getNextOrderNo()
  questionForm.weight = 1
  questionForm.options_text = 'A|Opsi A|true\nB|Opsi B|false'
  questionForm.answers_text = ''
  questionForm.pairs_text = ''
  questionForm.correct = true
  questionForm.rubric_text = ''
  questionForm.max_score = 100
  mcOptions.forEach((o, i) => {
    o.content = ''
    o.is_correct = i === 0
  })
  // If it was more than 5, reset to 5
  if (mcOptions.length > 5) mcOptions.splice(5)
  resetMcMultiple()
  resetShortAnswers()
  resetTrueFalseStatements()
  resetMatchingPairs()
  nextTick(() => {
    editorRenderVersion.value += 1
  })
}

const isQuickAddOpen = ref(false)
const isQuickAddProcessing = ref(false)
const quickAddProgress = reactive({ done: 0, total: 0 })
const quickAddForm = reactive({
  type: 'mc_single',
  count: 10,
})

const openQuickAddModal = () => {
  quickAddForm.type = questionForm.type || 'mc_single'
  quickAddForm.count = 10
  quickAddProgress.done = 0
  quickAddProgress.total = 0
  isQuickAddOpen.value = true
}

const buildTemplateQuestionPayload = (type, orderNo) => {
  const stem = `Soal ${orderNo}`
  const payload = { type, stem, order_no: orderNo, weight: 1 }

  if (type === 'mc_single') {
    payload.options = [
      { label: 'A', content: 'Opsi A', is_correct: true },
      { label: 'B', content: 'Opsi B', is_correct: false },
      { label: 'C', content: 'Opsi C', is_correct: false },
      { label: 'D', content: 'Opsi D', is_correct: false },
    ]
    return payload
  }

  if (type === 'mc_multiple') {
    payload.options = [
      { label: 'A', content: 'Opsi A', is_correct: true },
      { label: 'B', content: 'Opsi B', is_correct: false },
      { label: 'C', content: 'Opsi C', is_correct: false },
      { label: 'D', content: 'Opsi D', is_correct: false },
    ]
    return payload
  }

  if (type === 'matching') {
    payload.pairs = [
      { left_content: 'Kolom A1', right_content: 'Kolom B1', order_no: 1 },
      { left_content: 'Kolom A2', right_content: 'Kolom B2', order_no: 2 },
    ]
    return payload
  }

  if (type === 'short_answer') {
    payload.answers = [{ answer_text: 'Jawaban', order_no: 1 }]
    return payload
  }

  if (type === 'true_false') {
    payload.statements = [
      { content: 'Pernyataan 1', correct: true, order_no: 1 },
      { content: 'Pernyataan 2', correct: false, order_no: 2 },
    ]
    payload.correct = true
    return payload
  }

  if (type === 'essay') {
    payload.rubric_text = ''
    payload.max_score = 100
    return payload
  }

  return payload
}

const confirmQuickAdd = async () => {
  if (!currentSetId.value) return

  const count = Math.max(1, Math.min(100, Number(quickAddForm.count) || 1))
  const type = quickAddForm.type || 'mc_single'

  successMessage.value = ''
  errorMessage.value = ''
  isQuickAddProcessing.value = true
  quickAddProgress.done = 0
  quickAddProgress.total = count

  try {
    const startOrder = getNextOrderNo()

    const batchSize = 4
    for (let i = 0; i < count; i += batchSize) {
      const batch = []
      for (let j = 0; j < batchSize && i + j < count; j++) {
        const orderNo = startOrder + i + j
        const payload = buildTemplateQuestionPayload(type, orderNo)
        batch.push(
          api.post(`/api/v1/question-sets/${currentSetId.value}/questions`, payload).then(() => {
            quickAddProgress.done++
          })
        )
      }
      await Promise.all(batch)
    }

    await loadQuestions()
    resetQuestionForm(true)
    successMessage.value = `Berhasil menambahkan ${count} soal (${type.replace('_', ' ')})`
    setTimeout(() => { successMessage.value = '' }, 5000)
    isQuickAddOpen.value = false
  } catch (err) {
    errorMessage.value = 'Gagal quick-add: ' + (err.response?.data?.error?.message || 'Terjadi kesalahan')
  } finally {
    isQuickAddProcessing.value = false
  }
}


const buildQuestionPayload = () => {
  const payload = {
    type: questionForm.type,
    stem: questionForm.stem,
    order_no: Number(questionForm.order_no) || 1,
    weight: Number(questionForm.weight) > 0 ? Number(questionForm.weight) : 1,
  }
  if (['mc_single', 'mc_multiple'].includes(questionForm.type)) {
     if (questionForm.type === 'mc_single') {
       payload.options = mcOptions.map(o => ({
          label: o.label,
          content: o.content,
          is_correct: o.is_correct
       }))
     } else {
       // mc_multiple uses the dedicated mcMultipleOptions reactive array
       payload.options = mcMultipleOptions.map(o => ({
          label: o.label,
          content: o.content,
          is_correct: o.is_correct
       }))
     }
  }

  if (questionForm.type === 'short_answer') {
     payload.answers = shortAnswers
       .map((a, i) => ({ answer_text: extractAcceptedAnswerText(a.text), order_no: i + 1 }))
       .filter(a => a.answer_text)
  }
  if (questionForm.type === 'matching') {
     payload.pairs = matchingPairs
       .filter(p => p.left_content.trim() || p.right_content.trim())
       .map((p, i) => ({ left_content: p.left_content.trim(), right_content: p.right_content.trim(), order_no: i + 1 }))
  }
  if (questionForm.type === 'true_false') {
    payload.statements = trueFalseStatements
      .filter(s => s.content.trim())
      .map((s, i) => ({
        content: s.content.trim(),
        correct: s.correct,
        order_no: i + 1
      }))
    payload.correct = trueFalseStatements[0]?.correct ?? true
  }
  if (questionForm.type === 'essay') {
     payload.rubric_text = questionForm.rubric_text
     payload.max_score = Number(questionForm.max_score) || 0
  }
  return payload
}

const saveQuestion = async () => {
  successMessage.value = ''
  errorMessage.value = ''
  try {
    if (editingQuestionId.value) {
      await api.patch(`/api/v1/questions/${editingQuestionId.value}`, buildQuestionPayload())
      successMessage.value = 'Soal #' + questionForm.order_no + ' berhasil diperbarui'
      // Don't reset form on edit so user can continue tweaking
    } else {
      await api.post(`/api/v1/question-sets/${currentSetId.value}/questions`, buildQuestionPayload())
      successMessage.value = 'Soal baru berhasil ditambahkan'
      resetQuestionForm()
    }
    await loadQuestions()
    // Auto clear success message after 5 seconds
    setTimeout(() => { successMessage.value = '' }, 5000)
  } catch (err) {
    errorMessage.value = 'Gagal menyimpan soal: ' + (err.response?.data?.error?.message || 'Terjadi kesalahan')
  }
}

const deleteQuestion = async (id) => {
  if (!confirm('Hapus soal ini?')) return
  try {
    await api.delete(`/api/v1/questions/${id}`)
    await loadQuestions()
  } catch {
    errorMessage.value = 'Gagal menghapus'
  }
}

const applyQuestionToForm = (item) => {
  if (!item) return
  const nextType = String(item.type || '').trim() || 'mc_single'
  editingQuestionId.value = item.id
  questionForm.stem = item.stem || ''
  questionForm.order_no = item.order_no
  questionForm.weight = Number(item.weight) > 0 ? Number(item.weight) : 1
  questionForm.options_text = (item.options || []).map(o => `${o.label}|${o.content}|${o.is_correct}`).join('\n')
  questionForm.answers_text = (item.answers || []).map(a => a.answer_text).join('\n')
  questionForm.pairs_text = (item.pairs || []).map(p => `${p.left_content}|${p.right_content}`).join('\n')
  questionForm.correct = item.true_false?.correct ?? true
  questionForm.rubric_text = item.essay?.rubric_text || ''
  questionForm.max_score = item.essay?.max_score ?? 100
  
  if (nextType === 'mc_single' && item.options?.length) {
    mcOptions.splice(0, mcOptions.length)
    // Map whatever labels come from backend (A, B, C, D, E, etc.)
    item.options.forEach(o => {
      mcOptions.push({
        label: o.label,
        content: o.content || '',
        is_correct: o.is_correct
      })
    })
    // Ensure at least 4 for UI consistency if fewer
    const labels = ['A', 'B', 'C', 'D', 'E']
    while (mcOptions.length < 4) {
      const nextLabel = labels[mcOptions.length] || String.fromCharCode(65 + mcOptions.length)
      mcOptions.push({ label: nextLabel, content: '', is_correct: false })
    }
  }
  if (nextType === 'mc_multiple' && item.options?.length) {
    mcMultipleOptions.splice(0, mcMultipleOptions.length)
    item.options.forEach(o => {
      mcMultipleOptions.push({ label: o.label, content: o.content, is_correct: o.is_correct })
    })
  }
  if (nextType === 'short_answer' && item.answers?.length) {
    shortAnswers.splice(0, shortAnswers.length)
    item.answers.forEach(a => {
      shortAnswers.push({ text: a.answer_text || '' })
    })
  } else if (nextType === 'short_answer') {
    shortAnswers.splice(0, shortAnswers.length)
    shortAnswers.push({ text: '' })
  }
  if (nextType === 'true_false' && item.statements?.length) {
    trueFalseStatements.splice(0, trueFalseStatements.length)
    item.statements.forEach(s => {
      trueFalseStatements.push({ content: s.content || '', correct: s.correct ?? true })
    })
    if (trueFalseStatements.length === 0) trueFalseStatements.push({ content: '', correct: true })
  }
  if (nextType === 'matching' && item.pairs?.length) {
    matchingPairs.splice(0, matchingPairs.length)
    item.pairs.forEach(p => {
      matchingPairs.push({ left_content: p.left_content || '', right_content: p.right_content || '' })
    })
    if (matchingPairs.length === 0) { matchingPairs.push({ left_content: '', right_content: '' }); matchingPairs.push({ left_content: '', right_content: '' }) }
  }

  // Set type last so branch editors mount with already-hydrated data.
  questionForm.type = nextType

  // Force TinyMCE instances to remount after reactive model updates are flushed.
  nextTick(() => {
    editorRenderVersion.value += 1
  })
}

const populateQuestionForm = async (item) => {
  if (!item?.id) {
    applyQuestionToForm(item)
    return
  }

  const seq = ++questionPopulateSeq.value
  // Apply list payload immediately so editor never appears empty while waiting detail API.
  applyQuestionToForm(item)
  let hydrated = item

  try {
    const { data } = await api.get(`/api/v1/questions/${item.id}`)
    if (data?.data) hydrated = data.data
  } catch {
    // Fallback to list payload when detail request fails.
  }

  if (seq !== questionPopulateSeq.value) return
  applyQuestionToForm(hydrated)
}


const importDocx = async () => {
  if (!docxFile.value) return
  isImporting.value = true
  const formData = new FormData()
  formData.append('file', docxFile.value)
  try {
    await api.post(`/api/v1/question-sets/${currentSetId.value}/import-docx`, formData)
    docxFile.value = null
    successMessage.value = 'Berhasil import DOCX'
    await loadQuestions()
  } catch {
    errorMessage.value = 'Gagal import DOCX'
  } finally {
    isImporting.value = false
  }
}

const goToPreview = () => {
  const role = route.path.startsWith('/admin') ? 'admin' : 'teacher'
  router.push(`/${role}/bank-soal/preview/${currentSetId.value}`)
}

const syncEditorTypeRoute = (nextType) => {
  const normalized = normalizeEditorType(nextType)
  const current = String(route.params.editorType || '').trim()
  if (current === normalized) return
  router.replace({
    path: `${routeRolePrefix.value}/bank-soal/new/${normalized}`,
    query: route.query,
  })
}

onMounted(async () => {
  questionForm.type = normalizeEditorType(route.params.editorType)
  syncEditorTypeRoute(questionForm.type)
  loadSubjects()
  loadLevels()
  if (currentSetId.value) {
    await loadSetDetail()
    await loadQuestions()
    await initEditorFromFirstQuestion()
  }
})

watch(
  () => route.params.editorType,
  (nextType) => {
    const normalized = normalizeEditorType(nextType)
    if (questionForm.type !== normalized) questionForm.type = normalized
  },
)

watch(
  () => questionForm.type,
  (nextType) => {
    syncEditorTypeRoute(nextType)
  },
)

// Watch for query changes (if user clicks "Buat Soal" again)
watch(() => route.query.id, async (newId) => {
  currentSetId.value = newId || ''
  initializedSetId.value = ''
  if (newId) {
    await loadSetDetail()
    await loadQuestions()
    await initEditorFromFirstQuestion()
  } else {
    currentSet.value = null
    questions.value = []
    createForm.title = ''
  }
})

// Re-init only when crossing route namespace (admin <-> teacher), because Vue
// can reuse this component instance for both paths.
watch(
  () => route.path,
  async (nextPath, prevPath) => {
    const prevRole = String(prevPath || '').startsWith('/admin') ? 'admin' : 'teacher'
    const nextRole = String(nextPath || '').startsWith('/admin') ? 'admin' : 'teacher'
    if (prevRole === nextRole) return

    const nextId = route.query.id || ''
    if (!nextId) return
    currentSetId.value = nextId
    initializedSetId.value = ''
    await loadSetDetail()
    await loadQuestions()
    await initEditorFromFirstQuestion()
  },
)

// Soal: richer toolbar (2 lines expected if needed)
const stemToolbar = 'undo redo | bold italic underline | fontfamily fontsize forecolor | alignleft aligncenter alignright | bullist numlist | table image media | math charmap code | fullscreen'
// Jawaban: compact 1-line toolbar
const optionToolbar = 'bold italic underline | fontsize | alignleft | bullist numlist | image math | charmap code'

</script>


<template>
  <LayoutAuthenticated>
    <SectionMain>
      <SectionTitleLineWithButton
        :icon="isStepEditor ? mdiBookOpenOutline : mdiDatabasePlus"
        :title="isStepEditor ? `Editor: ${currentSet?.title || 'Loading...'}` : 'Buat Bank Soal Baru'"
        main
      >
        <div class="flex items-center gap-2">
          <BaseButton
            v-if="isStepEditor"
            :icon="mdiPencil"
            :label="isEditingMetadata ? 'Batal Edit Judul' : 'Edit Info Bank Soal'"
            color="purple"
            small
            @click="isEditingMetadata = !isEditingMetadata"
          />
          <BaseButton
            v-if="isStepEditor"
            :icon="mdiEye"
            label="Pratinjau"
            color="info"
            small
            @click="goToPreview"
          />
          <BaseButton
            v-if="isStepEditor"
            :icon="mdiPlus"
            label="Tambah Banyak"
            color="info"
            small
            @click="openQuickAddModal"
          />
          <BaseButton
            :icon="mdiArrowLeft"
            :label="isStepEditor ? 'Tutup Editor' : 'Batal'"
            color="success"
            small
            @click="router.push(route.path.startsWith('/admin') ? '/admin/bank-soal' : '/teacher/bank-soal')"
          />
        </div>
      </SectionTitleLineWithButton>

      <CardBoxModal
        v-model="isQuickAddOpen"
        title="Tambah Soal (Batch)"
        has-cancel
        is-form
        :is-processing="isQuickAddProcessing"
        :button-label="isQuickAddProcessing ? 'Memproses...' : `Buat ${Math.max(1, Math.min(100, Number(quickAddForm.count) || 1))} Soal`"
        @confirm="confirmQuickAdd"
      >
        <div class="rounded-xl border border-slate-200 bg-slate-50 px-4 py-3 text-xs text-slate-600 dark:border-slate-800 dark:bg-slate-900/40 dark:text-slate-300">
          Soal akan dibuat sebagai template yang valid (bisa diedit setelahnya). Maksimum 100 soal per sekali proses.
        </div>

        <FormField label="Tipe Soal">
          <FormControl v-model="quickAddForm.type" :options="questionTypeOptions" />
        </FormField>

        <FormField label="Jumlah Soal">
          <FormControl v-model="quickAddForm.count" type="number" inputmode="numeric" placeholder="10" />
          <p class="mt-1 text-[11px] text-slate-400">Urutan soal otomatis dimulai dari nomor terbesar + 1.</p>
        </FormField>

        <div
          v-if="quickAddProgress.total"
          class="rounded-xl border border-blue-200 bg-blue-50 px-4 py-3 text-xs text-blue-700 dark:border-blue-900/50 dark:bg-blue-900/20 dark:text-blue-200"
        >
          Membuat soal: {{ quickAddProgress.done }} / {{ quickAddProgress.total }}
        </div>
      </CardBoxModal>

      <!-- Edit Metadata Section (Toggleable when in Editor) -->
      <div v-if="isStepEditor && isEditingMetadata" class="mb-6 animate-fade-in-down">
         <div class="rounded-3xl bg-purple-600 shadow-2xl shadow-purple-300/40 p-6 text-white">
            <div class="flex items-center justify-between mb-6">
               <h3 class="text-xl font-bold flex items-center gap-3">
                  <div class="h-10 w-10 bg-white/20 rounded-full flex items-center justify-center">
                    <BaseIcon :path="mdiPencil" />
                  </div>
                  Edit Informasi Bank Soal
               </h3>
               <BaseButton :icon="mdiContentSave" color="white" label="Simpan Perubahan" @click="updateMetadata" />
            </div>
            <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
               <FormField label="Judul Bank Soal" class="white-label">
                  <FormControl v-model="createForm.title" />
               </FormField>
               <FormField label="Jenjang" class="white-label">
                  <FormControl
                    v-model="createForm.jenjang"
                    :options="[{ value: '', label: 'Pilih Jenjang' }, { value: 'SD', label: 'SD' }, { value: 'SMP', label: 'SMP' }, { value: 'SMA', label: 'SMA' }, { value: 'SMK', label: 'SMK' }]"
                  />
               </FormField>
               <FormField label="Level / Kelas" class="white-label">
                  <FormControl
                    v-model="createForm.level_id"
                    :options="[{ value: '', label: 'Pilih Level' }, ...levels.map(l => ({ value: l.id, label: l.name }))]"
                  />
               </FormField>
            </div>
         </div>
      </div>

      <!-- STEP 1: CREATION FORM -->
      <div v-if="!isStepEditor" class="max-w-2xl mx-auto py-12 animate-fade-in-up">
        <CardBox class="shadow-2xl p-10 border-t-8 border-purple-600">
          <div class="text-center mb-8">
            <div class="h-16 w-16 bg-purple-50 dark:bg-purple-900/30 rounded-full flex items-center justify-center mx-auto mb-4">
              <BaseIcon :path="mdiDatabasePlus" size="32" class="text-purple-600" />
            </div>
            <h2 class="text-2xl font-black dark:text-slate-100 uppercase tracking-tight">Buat Bank Soal</h2>
            <p class="text-slate-500 dark:text-slate-400 mt-2">Isi data dasar untuk memulai pembuatan soal.</p>
          </div>

          <div v-if="errorMessage" class="mb-6 rounded-lg bg-red-50 dark:bg-red-900/20 px-4 py-3 text-sm text-red-700 border border-red-200">{{ errorMessage }}</div>

          <div class="grid gap-6">
            <FormField label="Mata Pelajaran">
              <FormControl
                v-model="createForm.subject_id"
                :options="[{ value: '', label: 'Pilih mata pelajaran' }, ...subjects.map(s => ({ value: s.id, label: `${s.code} - ${s.name}` }))]"
                size="large"
              />
            </FormField>

            <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
              <FormField label="Jenjang">
                <FormControl
                  v-model="createForm.jenjang"
                  :options="[{ value: '', label: 'Pilih Jenjang' }, { value: 'SD', label: 'SD' }, { value: 'SMP', label: 'SMP' }, { value: 'SMA', label: 'SMA' }, { value: 'SMK', label: 'SMK' }]"
                  size="large"
                />
              </FormField>

              <FormField label="Level / Kelas">
                <FormControl
                  v-model="createForm.level_id"
                  :options="[{ value: '', label: 'Pilih Level' }, ...levels.map(l => ({ value: l.id, label: l.name }))]"
                  size="large"
                />
              </FormField>
            </div>

            <FormField label="Judul Bank Soal" help="Gunakan judul yang deskriptif, contoh: PAT Matematika Kelas X">
              <FormControl
                v-model="createForm.title"
                placeholder="Masukkan judul bank soal..."
                size="large"
              />
            </FormField>

            <div class="pt-4">
              <BaseButton
                :icon="mdiContentSave"
                color="purple"
                label="Buat & Lanjut ke Editor Soal"
                class="w-full py-4 text-lg font-bold shadow-lg"
                :loading="isLoading"
                @click="createAndContinue"
              />
            </div>
          </div>
        </CardBox>
      </div>

        <!-- STEP 2: PREMIUM PG EDITOR -->
        <div v-else-if="questionForm.type === 'mc_single'" class="animate-fade-in-up">
          <div v-if="errorMessage" class="mb-4 rounded-xl bg-red-50 px-5 py-4 text-sm text-red-700 border border-red-200">{{ errorMessage }}</div>
          <div v-if="successMessage" class="mb-4 rounded-xl bg-emerald-50 px-5 py-4 text-sm text-emerald-700 border border-emerald-200">{{ successMessage }}</div>

          <div class="mb-8 overflow-hidden rounded-2xl border border-slate-200 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-900/10 shadow-xs" is-form>
            <div class="bg-white dark:bg-slate-900 px-8 py-6 border-b flex flex-wrap items-center justify-between gap-4">
              <div class="flex items-center gap-6">
                <h4 class="font-extrabold text-2xl dark:text-slate-100 flex items-center gap-3">
                   <span class="text-slate-800 dark:text-slate-200">No:</span>
                  <span class="bg-purple-600 text-white px-3 py-1 rounded-xl text-lg min-w-[40px] text-center shadow-lg shadow-purple-200">{{ questionForm.order_no }}</span>
                </h4>
                <div class="h-10 w-[1px] bg-slate-200 dark:bg-slate-800"></div>
                <div class="flex items-center gap-3">
                   <span class="text-xs font-black uppercase tracking-widest text-slate-400">Tipe Soal </span>
                   <FormControl v-model="questionForm.type" :options="questionTypeOptions" transparent class="min-w-[180px]" />
                   <div class="flex items-center gap-2">
                     <span class="text-xs font-black uppercase tracking-widest text-slate-400">Bobot</span>
                     <input
                       v-model.number="questionForm.weight"
                       type="number"
                       min="1"
                       step="1"
                       class="w-20 rounded-lg border border-slate-200 bg-white px-2 py-1 text-sm font-bold text-slate-700 outline-none focus:border-indigo-500 dark:border-slate-700 dark:bg-slate-900 dark:text-slate-100"
                     />
                   </div>
                </div>
              </div>
              <div class="flex items-center gap-3">
                 <button type="button" 
                   class="flex items-center gap-2 px-6 py-2.5 rounded-xl border-2 border-purple-600 text-purple-600 font-bold hover:bg-purple-50 transition-all active:scale-95"
                   @click="resetQuestionForm"
                 >
                   <BaseIcon :path="mdiRefresh" size="20" />
                   <span>Reset</span>
                 </button>
                 <button type="button" 
                   class="flex items-center gap-2 px-8 py-2.5 rounded-xl bg-purple-600 text-white font-bold hover:bg-purple-700 shadow-lg shadow-purple-200 transition-all active:scale-95"
                   @click="saveQuestion"
                 >
                   <BaseIcon :path="mdiPlus" size="20" />
                   <span>Simpan</span>
                 </button>
              </div>
            </div>


            <div class="p-8 grid gap-8 xl:grid-cols-2 bg-white dark:bg-slate-950">
              <!-- Left: Question Stem -->
              <div class="space-y-4">
                <label class="block font-black text-[11px] uppercase tracking-[0.2em] text-slate-400">Soal</label>
                <RichEditor :key="`${activeEditorScopeKey}-stem`" v-model="questionForm.stem" :height="500" :toolbar="stemToolbar" placeholder="Tulis soal disini..." />
              </div>


              <!-- Right: Options -->
              <div class="space-y-4">
                <div
                  v-for="(opt, idx) in mcOptions"
                  :key="`${activeEditorScopeKey}-mc-single-${opt.label}`"
                   class="rounded-xl border overflow-hidden shadow-sm bg-white dark:bg-slate-900 transition-all"
                   :class="opt.is_correct ? 'border-purple-400 ring-4 ring-purple-400/20' : 'border-slate-200 dark:border-slate-800'"
                >
                  <!-- Card Header -->
                   <div class="px-4 py-3 border-b border-purple-100 dark:border-purple-800 bg-purple-50/50 dark:bg-purple-900/60 flex items-center justify-between">
                     <div class="flex items-center gap-2">
                        <div class="h-7 w-7 rounded-lg bg-purple-100 dark:bg-purple-900/50 flex items-center justify-center">
                          <BaseIcon :path="mdiPencil" size="14" class="text-purple-600 dark:text-purple-400" />
                        </div>
                        <span class="font-black text-sm text-purple-700 dark:text-purple-200">Jawaban {{ opt.label }}</span>
                     </div>
                    <div class="flex items-center gap-3">
                      <!-- Single-correct radio indicator -->
                       <label class="flex items-center gap-2 cursor-pointer select-none">
                         <span class="text-sm font-semibold" :class="opt.is_correct ? 'text-purple-600' : 'text-slate-400'">Jawaban benar</span>
                         <input
                           type="radio"
                           :name="'mc-correct-' + currentSetId"
                           :checked="opt.is_correct"
                           @change="mcCorrectLabel = opt.label"
                           class="h-5 w-5 accent-purple-600 cursor-pointer"
                         />
                       </label>
                      <button type="button"
                        class="h-7 w-7 flex items-center justify-center rounded-lg transition-colors"
                        :disabled="mcOptions.length <= 2"
                        :class="mcOptions.length <= 2 ? 'text-slate-300 cursor-not-allowed' : 'text-red-400 hover:bg-red-50 hover:text-red-600 cursor-pointer'"
                        @click="removeMcOption(idx)"
                        title="Hapus jawaban ini"
                      >
                        <BaseIcon :path="mdiDelete" size="16" />
                      </button>
                    </div>
                  </div>
                  <!-- Card Editor -->
                  <div class="p-3">
                    <RichEditor
                      :key="`${activeEditorScopeKey}-mc-single-editor-${opt.label}`"
                      v-model="opt.content"
                      :height="160"
                      :toolbar="optionToolbar"
                      :placeholder="`Tulis jawaban ${opt.label} disini...`"
                    />
                  </div>
                </div>

                <!-- Tambah Jawaban Button -->
                <button type="button"
                  class="w-full flex items-center justify-center gap-2 py-3.5 rounded-xl border-2 border-dashed border-emerald-400 text-emerald-600 font-bold bg-emerald-50/50 hover:bg-emerald-50 transition-all active:scale-95"
                  @click="addMcOption"
                >
                  <BaseIcon :path="mdiPlus" size="18" />
                  Tambah Jawaban
                </button>
              </div>
            </div>
          </div>


          <!-- List of other questions below -->
          <div class="space-y-4">
             <h4 class="text-sm font-black text-slate-500 uppercase tracking-widest px-2">Daftar Soal Lainya ({{ questions.length }})</h4>
             <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
               <QuestionQuickAddCard @click="openQuickAddModal" />
               <div v-for="item in questions" :key="item.id" class="rounded-xl border border-slate-200 dark:border-slate-800 p-4 bg-white dark:bg-slate-900/40 hover:shadow-lg transition-all" :class="{ 'ring-2 ring-blue-500': editingQuestionId === item.id }">
                  <div class="flex items-center justify-between mb-2">
                    <span class="text-xs font-black text-slate-400">#{{ item.order_no }}</span>
                    <div class="flex gap-3">
                      <BaseIcon :path="mdiPencil" size="16" class="text-blue-500 cursor-pointer" @click="populateQuestionForm(item)" />
                      <BaseIcon :path="mdiDelete" size="16" class="text-red-500 cursor-pointer" @click="deleteQuestion(item.id)" />
                    </div>
                  </div>
                  <div class="mb-2 text-[10px] font-mono text-slate-400 truncate" :title="item.id">ID: {{ formatQuestionId(item.id) }}</div>
                  <div class="text-xs text-slate-600 dark:text-slate-300 line-clamp-3" v-html="item.stem"></div>
                  <div v-if="buildQuestionCardAnswerPreviewHtml(item)" class="mt-2 rounded-lg border border-slate-100 bg-slate-50/60 px-2.5 py-2 text-[11px] text-slate-500 dark:border-slate-800 dark:bg-slate-900/40 dark:text-slate-300" v-html="buildQuestionCardAnswerPreviewHtml(item)"></div>
               </div>
             </div>
          </div>
        </div>

        <!-- STEP 2: mc_multiple (PG KOMPLEKS) EDITOR -->
        <div v-else-if="questionForm.type === 'mc_multiple'" class="animate-fade-in-up">
          <div v-if="errorMessage" class="mb-4 rounded-xl bg-red-50 px-5 py-4 text-sm text-red-700 border border-red-200">{{ errorMessage }}</div>
          <div v-if="successMessage" class="mb-4 rounded-xl bg-emerald-50 px-5 py-4 text-sm text-emerald-700 border border-emerald-200">{{ successMessage }}</div>

          <!-- Header Bar -->
          <div class="mb-6 flex flex-wrap items-center justify-between gap-4">
            <div class="flex items-center gap-4">
              <div class="text-xs font-black uppercase tracking-widest text-slate-400">SOAL NOMOR:</div>
              <span class="h-10 w-10 flex items-center justify-center bg-slate-900 dark:bg-slate-100 text-white dark:text-slate-900 font-black rounded-full text-lg shadow">{{ questionForm.order_no }}</span>
              <div class="flex items-center gap-2">
                <span class="text-xs font-black uppercase tracking-widest text-slate-400">Tipe</span>
                <FormControl v-model="questionForm.type" :options="questionTypeOptions" transparent class="min-w-[180px]" />
                <div class="flex items-center gap-2">
                  <span class="text-xs font-black uppercase tracking-widest text-slate-400">Bobot</span>
                  <input
                    v-model.number="questionForm.weight"
                    type="number"
                    min="1"
                    step="1"
                    class="w-20 rounded-lg border border-slate-200 bg-white px-2 py-1 text-sm font-bold text-slate-700 outline-none focus:border-indigo-500 dark:border-slate-700 dark:bg-slate-900 dark:text-slate-100"
                  />
                </div>
              </div>
            </div>
            <div class="flex items-center gap-3">
              <button type="button"
                class="flex items-center gap-2 px-5 py-2.5 rounded-xl border-2 border-slate-300 dark:border-slate-700 text-slate-600 dark:text-slate-300 font-bold hover:bg-slate-50 transition-all active:scale-95"
                @click="resetQuestionForm"
              >
                <BaseIcon :path="mdiRefresh" size="18" />
                Reset
              </button>
              <button type="button"
                class="flex items-center gap-2 px-7 py-2.5 rounded-xl bg-blue-600 text-white font-bold hover:bg-blue-700 shadow-lg shadow-blue-200 transition-all active:scale-95"
                @click="saveQuestion"
              >
                <BaseIcon :path="mdiPlus" size="18" />
                Simpan
              </button>
            </div>
          </div>

          <!-- Title -->
          <div class="mb-5">
            <h3 class="text-xl font-extrabold text-slate-800 dark:text-slate-100">
              Soal Kompleks Nomor: <span class="text-indigo-600">{{ questionForm.order_no }}</span>
            </h3>
          </div>

          <!-- Two-column layout: Stem | Answers -->
          <div class="grid gap-6 xl:grid-cols-2 items-start">
            <!-- Left: Question Stem -->
            <div class="rounded-xl border border-slate-200 dark:border-slate-800 overflow-hidden shadow-sm bg-white dark:bg-slate-900">
              <div class="px-5 py-3.5 border-b border-slate-100 dark:border-slate-800 bg-slate-50 dark:bg-slate-900/60">
                <label class="font-black text-[11px] uppercase tracking-[0.18em] text-slate-400">Soal</label>
              </div>
              <div class="p-4">
                <RichEditor :key="`${activeEditorScopeKey}-stem`" v-model="questionForm.stem" :height="480" :toolbar="stemToolbar" placeholder="Tulis soal disini..." />
              </div>
            </div>

            <!-- Right: Answer Cards -->
            <div class="space-y-4">
              <div
                v-for="(opt, idx) in mcMultipleOptions"
                :key="`${activeEditorScopeKey}-mc-multiple-${opt.label || idx}`"
                class="rounded-xl border overflow-hidden shadow-sm bg-white dark:bg-slate-900 transition-all"
                :class="opt.is_correct ? 'border-blue-400 ring-2 ring-blue-400/40' : 'border-slate-200 dark:border-slate-800'"
              >
                <!-- Card Header -->
                <div class="px-4 py-3 border-b border-slate-100 dark:border-slate-800 bg-slate-50 dark:bg-slate-900/60 flex items-center justify-between">
                  <div class="flex items-center gap-2">
                    <div class="h-7 w-7 rounded-lg bg-blue-100 dark:bg-blue-900/50 flex items-center justify-center">
                      <BaseIcon :path="mdiPencil" size="14" class="text-blue-600 dark:text-blue-400" />
                    </div>
                    <span class="font-black text-sm text-slate-700 dark:text-slate-200">Jawaban {{ opt.label }}</span>
                  </div>
                  <div class="flex items-center gap-3">
                    <label class="flex items-center gap-2 cursor-pointer select-none">
                      <span class="text-sm font-semibold" :class="opt.is_correct ? 'text-blue-600' : 'text-slate-400'">Jawaban benar</span>
                      <input
                        type="checkbox"
                        :checked="opt.is_correct"
                        @change="opt.is_correct = $event.target.checked"
                        class="h-5 w-5 rounded accent-blue-600 cursor-pointer"
                      />
                    </label>
                    <button type="button"
                      class="h-7 w-7 flex items-center justify-center rounded-lg transition-colors"
                      :disabled="mcMultipleOptions.length <= 2"
                      :class="mcMultipleOptions.length <= 2 ? 'text-slate-300 cursor-not-allowed' : 'text-red-400 hover:bg-red-50 hover:text-red-600 cursor-pointer'"
                      @click="removeMcMultipleOption(idx)"
                      title="Hapus jawaban ini"
                    >
                      <BaseIcon :path="mdiDelete" size="16" />
                    </button>
                  </div>
                </div>
                <!-- Card Editor -->
                <div class="p-3">
                  <RichEditor
                    :key="`${activeEditorScopeKey}-mc-multiple-editor-${opt.label || idx}`"
                    v-model="opt.content"
                    :height="130"
                    :toolbar="optionToolbar"
                    :placeholder="`Tulis jawaban ${opt.label} disini...`"
                  />
                </div>
              </div>

              <!-- Tambah Jawaban Button -->
              <button type="button"
                class="w-full flex items-center justify-center gap-2 py-3.5 rounded-xl border-2 border-dashed border-emerald-400 text-emerald-600 font-bold bg-emerald-50/50 hover:bg-emerald-50 transition-all active:scale-95"
                @click="addMcMultipleOption"
              >
                <BaseIcon :path="mdiPlus" size="18" />
                Tambah Jawaban
              </button>
            </div>
          </div>

          <!-- List of other questions below -->
          <div class="space-y-4 mt-10">
             <h4 class="text-sm font-black text-slate-500 uppercase tracking-widest px-2">Daftar Soal Lainya ({{ questions.length }})</h4>
             <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
               <QuestionQuickAddCard @click="openQuickAddModal" />
               <div v-for="item in questions" :key="item.id" class="rounded-xl border border-slate-200 dark:border-slate-800 p-4 bg-white dark:bg-slate-900/40 hover:shadow-lg transition-all" :class="{ 'ring-2 ring-blue-500': editingQuestionId === item.id }">
                  <div class="flex items-center justify-between mb-2">
                    <span class="text-xs font-black text-slate-400">#{{ item.order_no }}</span>
                    <div class="flex gap-3">
                      <BaseIcon :path="mdiPencil" size="16" class="text-blue-500 cursor-pointer" @click="populateQuestionForm(item)" />
                      <BaseIcon :path="mdiDelete" size="16" class="text-red-500 cursor-pointer" @click="deleteQuestion(item.id)" />
                    </div>
                  </div>
                  <div class="mb-2 text-[10px] font-mono text-slate-400 truncate" :title="item.id">ID: {{ formatQuestionId(item.id) }}</div>
                  <span class="inline-block mb-1 px-2 py-0.5 rounded text-[9px] font-black uppercase bg-slate-100 dark:bg-slate-800 text-slate-500">{{ item.type.replace('_', ' ') }}</span>
                  <div class="text-xs text-slate-600 dark:text-slate-300 line-clamp-3" v-html="item.stem"></div>
                  <div v-if="buildQuestionCardAnswerPreviewHtml(item)" class="mt-2 rounded-lg border border-slate-100 bg-slate-50/60 px-2.5 py-2 text-[11px] text-slate-500 dark:border-slate-800 dark:bg-slate-900/40 dark:text-slate-300" v-html="buildQuestionCardAnswerPreviewHtml(item)"></div>
               </div>
             </div>
          </div>
        </div>

        <!-- STEP 2: short_answer (ISIAN SINGKAT) EDITOR -->
        <div v-else-if="questionForm.type === 'short_answer'" class="animate-fade-in-up">
          <div v-if="errorMessage" class="mb-4 rounded-xl bg-red-50 px-5 py-4 text-sm text-red-700 border border-red-200">{{ errorMessage }}</div>
          <div v-if="successMessage" class="mb-4 rounded-xl bg-emerald-50 px-5 py-4 text-sm text-emerald-700 border border-emerald-200">{{ successMessage }}</div>

          <!-- Header Bar -->
          <div class="mb-6 flex flex-wrap items-center justify-between gap-4">
            <div class="flex items-center gap-4">
              <div class="text-xs font-black uppercase tracking-widest text-slate-400">SOAL NOMOR:</div>
              <span class="h-10 w-10 flex items-center justify-center bg-slate-900 dark:bg-slate-100 text-white dark:text-slate-900 font-black rounded-full text-lg shadow">{{ questionForm.order_no }}</span>
              <div class="flex items-center gap-2">
                <span class="text-xs font-black uppercase tracking-widest text-slate-400">Tipe</span>
                <FormControl v-model="questionForm.type" :options="questionTypeOptions" transparent class="min-w-[180px]" />
                <div class="flex items-center gap-2">
                  <span class="text-xs font-black uppercase tracking-widest text-slate-400">Bobot</span>
                  <input
                    v-model.number="questionForm.weight"
                    type="number"
                    min="1"
                    step="1"
                    class="w-20 rounded-lg border border-slate-200 bg-white px-2 py-1 text-sm font-bold text-slate-700 outline-none focus:border-indigo-500 dark:border-slate-700 dark:bg-slate-900 dark:text-slate-100"
                  />
                </div>
              </div>
            </div>
            <div class="flex items-center gap-3">
              <button type="button"
                class="flex items-center gap-2 px-5 py-2.5 rounded-xl border-2 border-slate-300 dark:border-slate-700 text-slate-600 dark:text-slate-300 font-bold hover:bg-slate-50 transition-all active:scale-95"
                @click="resetQuestionForm"
              >
                <BaseIcon :path="mdiRefresh" size="18" />
                Reset
              </button>
              <button type="button"
                class="flex items-center gap-2 px-7 py-2.5 rounded-xl bg-teal-600 text-white font-bold hover:bg-teal-700 shadow-lg shadow-teal-200 transition-all active:scale-95"
                @click="saveQuestion"
              >
                <BaseIcon :path="mdiContentSave" size="18" />
                Simpan
              </button>
            </div>
          </div>

          <!-- Title -->
          <div class="mb-5">
            <h3 class="text-xl font-extrabold text-slate-800 dark:text-slate-100">
              Soal Isian Singkat Nomor: <span class="text-teal-600">{{ questionForm.order_no }}</span>
            </h3>
          </div>

          <!-- Two-column layout: Stem | Answers -->
          <div class="grid gap-6 xl:grid-cols-2 items-start">
            <!-- Left: Question Stem Card -->
            <div class="rounded-2xl border border-slate-200 dark:border-slate-700 overflow-hidden shadow-sm bg-white dark:bg-slate-900">
              <!-- Card Header -->
              <div class="px-5 py-3.5 border-b border-slate-100 dark:border-slate-800 bg-slate-50 dark:bg-slate-900/70 flex items-center gap-3">
                <div class="h-7 w-7 rounded-lg bg-teal-100 dark:bg-teal-900/40 flex items-center justify-center">
                  <BaseIcon :path="mdiFileDocumentOutline" size="16" class="text-teal-600 dark:text-teal-400" />
                </div>
                <label class="font-black text-[11px] uppercase tracking-[0.18em] text-slate-500 dark:text-slate-400">Soal</label>
              </div>
              <!-- Card Body -->
              <div class="p-4">
                <RichEditor :key="`${activeEditorScopeKey}-stem`" v-model="questionForm.stem" :height="480" :toolbar="stemToolbar" placeholder="Tulis soal isian singkat disini..." />
              </div>
            </div>

            <!-- Right: Answer Cards -->
            <div class="space-y-4">
              <!-- Answer Card -->
              <div
                v-for="(ans, idx) in shortAnswers"
                :key="`${activeEditorScopeKey}-short-answer-${idx}`"
                class="rounded-2xl border border-teal-300 dark:border-teal-700 overflow-hidden shadow-sm bg-white dark:bg-slate-900 ring-2 ring-teal-300/40 dark:ring-teal-700/40 transition-all"
              >
                <!-- Card Header -->
                <div class="px-4 py-3 border-b border-teal-100 dark:border-teal-900/50 bg-teal-50 dark:bg-teal-900/20 flex items-center justify-between">
                  <div class="flex items-center gap-2">
                    <div class="h-7 w-7 rounded-lg bg-teal-100 dark:bg-teal-900/50 flex items-center justify-center">
                      <BaseIcon :path="mdiPencil" size="14" class="text-teal-600 dark:text-teal-400" />
                    </div>
                    <span class="font-black text-sm text-teal-700 dark:text-teal-300">Jawaban Benar {{ shortAnswers.length > 1 ? idx + 1 : '' }}</span>
                  </div>
                  <div class="flex items-center gap-3">
                    <span class="inline-flex items-center gap-1 px-2.5 py-1 rounded-full bg-teal-100 dark:bg-teal-900/50 text-teal-700 dark:text-teal-300 text-[10px] font-black uppercase tracking-wider">
                      ✓ Diterima
                    </span>
                    <button type="button"
                      class="h-7 w-7 flex items-center justify-center rounded-lg transition-colors"
                      :disabled="shortAnswers.length <= 1"
                      :class="shortAnswers.length <= 1 ? 'text-slate-300 cursor-not-allowed' : 'text-red-400 hover:bg-red-50 hover:text-red-600 cursor-pointer'"
                      @click="removeShortAnswer(idx)"
                      title="Hapus jawaban ini"
                    >
                      <BaseIcon :path="mdiDelete" size="16" />
                    </button>
                  </div>
                </div>
                <!-- Card Body: RichEditor -->
                <div class="p-3">
                  <RichEditor
                    :key="`${activeEditorScopeKey}-short-answer-editor-${idx}`"
                    v-model="ans.text"
                    :height="160"
                    :toolbar="optionToolbar"
                    :placeholder="`Tulis jawaban yang diterima ${idx + 1}...`"
                  />
                  <p class="mt-2 text-[11px] text-slate-400 px-1">Jawaban siswa akan dicocokkan dengan teks ini. Anda bisa menambahkan rumus atau simbol jika diperlukan.</p>
                </div>
              </div>

              <!-- Tambah Jawaban Button -->
              <button type="button"
                class="w-full flex items-center justify-center gap-2 py-3.5 rounded-xl border-2 border-dashed border-teal-400 text-teal-600 font-bold bg-teal-50/50 hover:bg-teal-50 transition-all active:scale-95"
                @click="addShortAnswer"
              >
                <BaseIcon :path="mdiPlus" size="18" />
                Tambah Jawaban Alternatif
              </button>

              <!-- Help Note -->
              <div class="rounded-xl bg-amber-50 dark:bg-amber-900/20 border border-amber-200 dark:border-amber-800 px-4 py-3">
                <p class="text-xs text-amber-700 dark:text-amber-300 font-semibold">💡 Tip: Anda dapat menambahkan beberapa jawaban alternatif yang dianggap benar. Misalnya: "6", "enam", "6,0"</p>
              </div>
            </div>
          </div>

          <!-- List of other questions below -->
          <div class="space-y-4 mt-10">
             <h4 class="text-sm font-black text-slate-500 uppercase tracking-widest px-2">Daftar Soal Lainya ({{ questions.length }})</h4>
             <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
               <QuestionQuickAddCard @click="openQuickAddModal" />
               <div v-for="item in questions" :key="item.id" class="rounded-xl border border-slate-200 dark:border-slate-800 p-4 bg-white dark:bg-slate-900/40 hover:shadow-lg transition-all" :class="{ 'ring-2 ring-teal-500': editingQuestionId === item.id }">
                  <div class="flex items-center justify-between mb-2">
                    <span class="text-xs font-black text-slate-400">#{{ item.order_no }}</span>
                    <div class="flex gap-3">
                      <BaseIcon :path="mdiPencil" size="16" class="text-blue-500 cursor-pointer" @click="populateQuestionForm(item)" />
                      <BaseIcon :path="mdiDelete" size="16" class="text-red-500 cursor-pointer" @click="deleteQuestion(item.id)" />
                    </div>
                  </div>
                  <div class="mb-2 text-[10px] font-mono text-slate-400 truncate" :title="item.id">ID: {{ formatQuestionId(item.id) }}</div>
                  <span class="inline-block mb-1 px-2 py-0.5 rounded text-[9px] font-black uppercase bg-teal-100 dark:bg-teal-900/40 text-teal-600">{{ item.type.replace('_', ' ') }}</span>
                  <div class="text-xs text-slate-600 dark:text-slate-300 line-clamp-3" v-html="item.stem"></div>
                  <div v-if="buildQuestionCardAnswerPreviewHtml(item)" class="mt-2 rounded-lg border border-slate-100 bg-slate-50/60 px-2.5 py-2 text-[11px] text-slate-500 dark:border-slate-800 dark:bg-slate-900/40 dark:text-slate-300" v-html="buildQuestionCardAnswerPreviewHtml(item)"></div>
               </div>
             </div>
          </div>
        </div>

        <!-- STEP 2: essay (URAIAN) EDITOR -->
        <div v-else-if="questionForm.type === 'essay'" class="animate-fade-in-up">
          <div v-if="errorMessage" class="mb-4 rounded-xl bg-red-50 px-5 py-4 text-sm text-red-700 border border-red-200">{{ errorMessage }}</div>
          <div v-if="successMessage" class="mb-4 rounded-xl bg-emerald-50 px-5 py-4 text-sm text-emerald-700 border border-emerald-200">{{ successMessage }}</div>

          <!-- Header Bar -->
          <div class="mb-6 flex flex-wrap items-center justify-between gap-4">
            <div class="flex items-center gap-4">
              <div class="text-xs font-black uppercase tracking-widest text-slate-400">SOAL NOMOR:</div>
              <span class="h-10 w-10 flex items-center justify-center bg-slate-900 dark:bg-slate-100 text-white dark:text-slate-900 font-black rounded-full text-lg shadow">{{ questionForm.order_no }}</span>
              <div class="flex items-center gap-2">
                <span class="text-xs font-black uppercase tracking-widest text-slate-400">Tipe</span>
                <FormControl v-model="questionForm.type" :options="questionTypeOptions" transparent class="min-w-[180px]" />
                <div class="flex items-center gap-2">
                  <span class="text-xs font-black uppercase tracking-widest text-slate-400">Bobot</span>
                  <input
                    v-model.number="questionForm.weight"
                    type="number"
                    min="1"
                    step="1"
                    class="w-20 rounded-lg border border-slate-200 bg-white px-2 py-1 text-sm font-bold text-slate-700 outline-none focus:border-indigo-500 dark:border-slate-700 dark:bg-slate-900 dark:text-slate-100"
                  />
                </div>
              </div>
            </div>
            <div class="flex items-center gap-3">
              <button type="button"
                class="flex items-center gap-2 px-5 py-2.5 rounded-xl border-2 border-slate-300 dark:border-slate-700 text-slate-600 dark:text-slate-300 font-bold hover:bg-slate-50 transition-all active:scale-95"
                @click="resetQuestionForm"
              >
                <BaseIcon :path="mdiRefresh" size="18" />
                Reset
              </button>
              <button type="button"
                class="flex items-center gap-2 px-7 py-2.5 rounded-xl bg-violet-600 text-white font-bold hover:bg-violet-700 shadow-lg shadow-violet-200 transition-all active:scale-95"
                @click="saveQuestion"
              >
                <BaseIcon :path="mdiContentSave" size="18" />
                Simpan
              </button>
            </div>
          </div>

          <!-- Title -->
          <div class="mb-5">
            <h3 class="text-xl font-extrabold text-slate-800 dark:text-slate-100">
              Soal Uraian Nomor: <span class="text-violet-600">{{ questionForm.order_no }}</span>
            </h3>
          </div>

          <!-- Two-column layout: Stem | Rubric -->
          <div class="grid gap-6 xl:grid-cols-2 items-start">
            <!-- Left: Question Stem Card -->
            <div class="rounded-2xl border border-slate-200 dark:border-slate-700 overflow-hidden shadow-sm bg-white dark:bg-slate-900">
              <!-- Card Header -->
              <div class="px-5 py-3.5 border-b border-slate-100 dark:border-slate-800 bg-slate-50 dark:bg-slate-900/70 flex items-center gap-3">
                <div class="h-7 w-7 rounded-lg bg-violet-100 dark:bg-violet-900/40 flex items-center justify-center">
                  <BaseIcon :path="mdiFileDocumentOutline" size="16" class="text-violet-600 dark:text-violet-400" />
                </div>
                <label class="font-black text-[11px] uppercase tracking-[0.18em] text-slate-500 dark:text-slate-400">Soal</label>
              </div>
              <!-- Card Body -->
              <div class="p-4">
                <RichEditor :key="`${activeEditorScopeKey}-stem`" v-model="questionForm.stem" :height="480" :toolbar="stemToolbar" placeholder="Tulis soal uraian disini..." />
              </div>
            </div>

            <!-- Right: Rubric & Score Card -->
            <div class="space-y-4">
              <!-- Rubric Card -->
              <div class="rounded-2xl border border-violet-300 dark:border-violet-700 overflow-hidden shadow-sm bg-white dark:bg-slate-900 ring-2 ring-violet-300/40 dark:ring-violet-700/40 transition-all">
                <!-- Card Header -->
                <div class="px-4 py-3 border-b border-violet-100 dark:border-violet-900/50 bg-violet-50 dark:bg-violet-900/20 flex items-center justify-between">
                  <div class="flex items-center gap-2">
                    <div class="h-7 w-7 rounded-lg bg-violet-100 dark:bg-violet-900/50 flex items-center justify-center">
                      <BaseIcon :path="mdiPencil" size="14" class="text-violet-600 dark:text-violet-400" />
                    </div>
                    <span class="font-black text-sm text-violet-700 dark:text-violet-300">Rubrik / Kunci Jawaban</span>
                  </div>
                  <span class="inline-flex items-center gap-1 px-2.5 py-1 rounded-full bg-violet-100 dark:bg-violet-900/50 text-violet-700 dark:text-violet-300 text-[10px] font-black uppercase tracking-wider">
                    Panduan Guru
                  </span>
                </div>
                <!-- Card Body: RichEditor for rubric -->
                <div class="p-4">
                  <RichEditor
                    :key="`${activeEditorScopeKey}-essay-rubric`"
                    v-model="questionForm.rubric_text"
                    :height="300"
                    :toolbar="optionToolbar"
                    placeholder="Tulis rubrik penilaian atau kunci jawaban disini..."
                  />
                  <p class="mt-2 text-[11px] text-slate-400">Rubrik hanya terlihat oleh guru saat mengoreksi jawaban siswa.</p>
                </div>
              </div>

              <!-- Max Score Card -->
              <div class="rounded-2xl border border-violet-200 dark:border-violet-800 overflow-hidden shadow-sm bg-white dark:bg-slate-900">
                <!-- Card Header -->
                <div class="px-4 py-3 border-b border-violet-100 dark:border-violet-900/50 bg-violet-50/60 dark:bg-violet-900/10 flex items-center gap-2">
                  <span class="font-black text-sm text-violet-700 dark:text-violet-300">Skor Maksimal</span>
                </div>
                <!-- Card Body -->
                <div class="p-4">
                  <div class="flex items-center gap-4">
                    <input
                      v-model.number="questionForm.max_score"
                      type="number"
                      min="0"
                      max="1000"
                      class="w-36 rounded-xl border border-slate-200 dark:border-slate-700 bg-slate-50 dark:bg-slate-800 px-4 py-3 text-slate-800 dark:text-slate-100 text-2xl font-black text-center placeholder-slate-400 focus:outline-none focus:ring-2 focus:ring-violet-400 focus:border-violet-400 transition-all"
                      placeholder="100"
                    />
                    <div>
                      <p class="text-sm font-semibold text-slate-600 dark:text-slate-300">poin</p>
                      <p class="text-[11px] text-slate-400 mt-0.5">Nilai maksimum yang bisa diperoleh siswa</p>
                    </div>
                  </div>
                </div>
              </div>

              <!-- Help Note -->
              <div class="rounded-xl bg-amber-50 dark:bg-amber-900/20 border border-amber-200 dark:border-amber-800 px-4 py-3">
                <p class="text-xs text-amber-700 dark:text-amber-300 font-semibold">💡 Tip: Uraian dinilai secara manual oleh guru. Rubrik membantu konsistensi penilaian antar jawaban siswa.</p>
              </div>
            </div>
          </div>

          <!-- List of other questions below -->
          <div class="space-y-4 mt-10">
             <h4 class="text-sm font-black text-slate-500 uppercase tracking-widest px-2">Daftar Soal Lainya ({{ questions.length }})</h4>
             <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
               <QuestionQuickAddCard @click="openQuickAddModal" />
               <div v-for="item in questions" :key="item.id" class="rounded-xl border border-slate-200 dark:border-slate-800 p-4 bg-white dark:bg-slate-900/40 hover:shadow-lg transition-all" :class="{ 'ring-2 ring-violet-500': editingQuestionId === item.id }">
                  <div class="flex items-center justify-between mb-2">
                    <span class="text-xs font-black text-slate-400">#{{ item.order_no }}</span>
                    <div class="flex gap-3">
                      <BaseIcon :path="mdiPencil" size="16" class="text-blue-500 cursor-pointer" @click="populateQuestionForm(item)" />
                      <BaseIcon :path="mdiDelete" size="16" class="text-red-500 cursor-pointer" @click="deleteQuestion(item.id)" />
                    </div>
                  </div>
                  <div class="mb-2 text-[10px] font-mono text-slate-400 truncate" :title="item.id">ID: {{ formatQuestionId(item.id) }}</div>
                  <span class="inline-block mb-1 px-2 py-0.5 rounded text-[9px] font-black uppercase bg-violet-100 dark:bg-violet-900/40 text-violet-600">{{ item.type.replace('_', ' ') }}</span>
                  <div class="text-xs text-slate-600 dark:text-slate-300 line-clamp-3" v-html="item.stem"></div>
                  <div v-if="buildQuestionCardAnswerPreviewHtml(item)" class="mt-2 rounded-lg border border-slate-100 bg-slate-50/60 px-2.5 py-2 text-[11px] text-slate-500 dark:border-slate-800 dark:bg-slate-900/40 dark:text-slate-300" v-html="buildQuestionCardAnswerPreviewHtml(item)"></div>
               </div>
             </div>
          </div>
        </div>

        <!-- STEP 2: true_false (BENAR/SALAH) EDITOR -->
        <div v-else-if="questionForm.type === 'true_false'" class="animate-fade-in-up">
          <div v-if="errorMessage" class="mb-4 rounded-xl bg-red-50 px-5 py-4 text-sm text-red-700 border border-red-200">{{ errorMessage }}</div>
          <div v-if="successMessage" class="mb-4 rounded-xl bg-emerald-50 px-5 py-4 text-sm text-emerald-700 border border-emerald-200">{{ successMessage }}</div>

          <!-- Premium Header Bar -->
          <div class="mb-8 overflow-hidden rounded-2xl border border-slate-200 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-900/10 shadow-xs">
            <div class="bg-white dark:bg-slate-900 px-8 py-6 border-b flex flex-wrap items-center justify-between gap-4">
              <div class="flex items-center gap-6">
                <h4 class="font-extrabold text-2xl dark:text-slate-100 flex items-center gap-3">
                   <span class="text-slate-800 dark:text-slate-200">No:</span>
                   <span class="bg-emerald-600 text-white px-3 py-1 rounded-xl text-lg min-w-[40px] text-center shadow-lg shadow-emerald-200">{{ questionForm.order_no }}</span>
                </h4>
                <div class="h-10 w-[1px] bg-slate-200 dark:bg-slate-800"></div>
                <div class="flex items-center gap-3">
                   <span class="text-xs font-black uppercase tracking-widest text-slate-400">Tipe Soal </span>
                   <FormControl v-model="questionForm.type" :options="questionTypeOptions" transparent class="min-w-[180px]" />
                   <div class="flex items-center gap-2">
                     <span class="text-xs font-black uppercase tracking-widest text-slate-400">Bobot</span>
                     <input
                       v-model.number="questionForm.weight"
                       type="number"
                       min="1"
                       step="1"
                       class="w-20 rounded-lg border border-slate-200 bg-white px-2 py-1 text-sm font-bold text-slate-700 outline-none focus:border-indigo-500 dark:border-slate-700 dark:bg-slate-900 dark:text-slate-100"
                     />
                   </div>
                 </div>
              </div>
              <div class="flex items-center gap-3">
                <button type="button" 
                  class="flex items-center gap-2 px-6 py-2.5 rounded-xl border-2 border-emerald-600 text-emerald-600 font-bold hover:bg-emerald-50 transition-all active:scale-95"
                  @click="resetQuestionForm"
                >
                  <BaseIcon :path="mdiRefresh" size="20" />
                  <span>Reset</span>
                </button>
                <button type="button" 
                  class="flex items-center gap-2 px-8 py-2.5 rounded-xl bg-emerald-600 text-white font-bold hover:bg-emerald-700 shadow-lg shadow-emerald-200 transition-all active:scale-95"
                  @click="saveQuestion"
                >
                  <BaseIcon :path="mdiPlus" size="20" />
                  <span>Simpan</span>
                </button>
              </div>
            </div>

            <div class="p-8 grid gap-8 xl:grid-cols-2 bg-white dark:bg-slate-950 items-start">
              <!-- Left: Stimulus -->
              <div class="space-y-4">
                <div class="flex items-center gap-2">
                  <div class="h-6 w-6 rounded-lg bg-emerald-100 dark:bg-emerald-900/50 flex items-center justify-center">
                    <BaseIcon :path="mdiFileDocumentOutline" size="14" class="text-emerald-600 dark:text-emerald-400" />
                  </div>
                  <label class="font-black text-[11px] uppercase tracking-[0.2em] text-slate-400">Stimulus / Soal Utama</label>
                </div>
                <RichEditor :key="`${activeEditorScopeKey}-stem`" v-model="questionForm.stem" :height="450" :toolbar="stemToolbar" placeholder="Tulis stimulus soal disini..." />
              </div>

              <!-- Right: Statements Management -->
              <div class="space-y-4">
                 <div class="flex items-center justify-between px-2">
                    <label class="font-black text-[11px] uppercase tracking-[0.2em] text-slate-400">Pernyataan Benar / Salah</label>
                    <span class="px-2 py-0.5 rounded-full bg-emerald-100 text-emerald-700 text-[10px] font-black uppercase">{{ trueFalseStatements.length }} Item</span>
                 </div>

                 <div class="space-y-3">
                    <div 
                      v-for="(st, idx) in trueFalseStatements" 
                      :key="idx"
                      class="rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 overflow-hidden shadow-sm"
                    >
                       <div class="px-3 py-2 border-b border-slate-100 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-900/50 flex items-center justify-between">
                          <span class="text-[10px] font-black text-slate-400 uppercase tracking-widest">Pernyataan {{ idx + 1 }}</span>
                          <div class="flex items-center gap-2">
                            <!-- Toggle Button Style for True/False -->
                            <div class="flex bg-slate-200 dark:bg-slate-800 rounded-lg p-0.5">
                               <button type="button" 
                                 @click="st.correct = true"
                                 class="px-3 py-1 rounded-md text-[10px] font-black transition-all"
                                 :class="st.correct ? 'bg-emerald-500 text-white shadow-sm' : 'text-slate-500 hover:text-slate-700'"
                               >BENAR</button>
                               <button type="button" 
                                 @click="st.correct = false"
                                 class="px-3 py-1 rounded-md text-[10px] font-black transition-all"
                                 :class="!st.correct ? 'bg-rose-500 text-white shadow-sm' : 'text-slate-500 hover:text-slate-700'"
                               >SALAH</button>
                            </div>
                            <button type="button" 
                              v-if="trueFalseStatements.length > 1"
                              @click="removeTrueFalseStatement(idx)"
                              class="text-red-400 hover:text-red-600 transition-colors p-1"
                            >
                              <BaseIcon :path="mdiDelete" size="14" />
                            </button>
                          </div>
                       </div>
                       <div class="p-2">
                          <textarea 
                             v-model="st.content" 
                             class="w-full bg-transparent border-none focus:ring-0 text-sm placeholder-slate-400 resize-none min-h-[60px]" 
                             placeholder="Ketikkan pernyataan di sini..."
                          ></textarea>
                       </div>
                    </div>
                 </div>

                 <button type="button"
                    class="w-full flex items-center justify-center gap-2 py-3 rounded-xl border-2 border-dashed border-emerald-400 text-emerald-600 font-bold bg-emerald-50/30 hover:bg-emerald-50 transition-all active:scale-95 mt-4"
                    @click="addTrueFalseStatement"
                  >
                    <BaseIcon :path="mdiPlus" size="16" />
                    Tambah Pernyataan
                  </button>
              </div>
            </div>
          </div>

          <!-- List of other questions below -->
          <div class="space-y-4 mt-12">
             <h4 class="text-[10px] font-black text-slate-400 uppercase tracking-widest px-2">Daftar Soal Lainya ({{ questions.length }})</h4>
             <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
               <QuestionQuickAddCard @click="openQuickAddModal" />
               <div v-for="item in questions" :key="item.id" class="relative overflow-hidden rounded-2xl border border-slate-200 dark:border-slate-800 p-5 bg-white dark:bg-slate-900/40 hover:shadow-xl transition-all group" :class="{ 'ring-2 ring-emerald-500 bg-emerald-50/20': editingQuestionId === item.id }">
                  <div class="flex items-center justify-between mb-3">
                    <span class="text-[10px] font-black text-slate-400 px-2 py-0.5 bg-slate-100 rounded">#{{ item.order_no }}</span>
                    <div class="flex gap-2 opacity-0 group-hover:opacity-100 transition-opacity">
                      <button type="button" class="p-1.5 rounded-lg bg-blue-50 text-blue-600 hover:bg-blue-100" @click="populateQuestionForm(item)"><BaseIcon :path="mdiPencil" size="14" /></button>
                      <button type="button" class="p-1.5 rounded-lg bg-red-50 text-red-600 hover:bg-red-100" @click="deleteQuestion(item.id)"><BaseIcon :path="mdiDelete" size="14" /></button>
                    </div>
                  </div>
                  <div class="mb-2 text-[10px] font-mono text-slate-400 truncate" :title="item.id">ID: {{ formatQuestionId(item.id) }}</div>
                  <div class="flex gap-2 mb-2">
                     <span class="px-2 py-0.5 rounded text-[9px] font-black uppercase tracking-tighter" :class="item.type === 'mc_single' ? 'bg-indigo-100 text-indigo-600' : 'bg-emerald-100 text-emerald-600'">{{ item.type.replace('_', ' ') }}</span>
                  </div>
                  <div class="text-xs text-slate-600 dark:text-slate-300 font-medium line-clamp-2 leading-relaxed" v-html="item.stem"></div>
                  <div v-if="buildQuestionCardAnswerPreviewHtml(item)" class="mt-2 rounded-lg border border-slate-100 bg-slate-50/60 px-2.5 py-2 text-[11px] text-slate-500 dark:border-slate-800 dark:bg-slate-900/40 dark:text-slate-300" v-html="buildQuestionCardAnswerPreviewHtml(item)"></div>
               </div>
             </div>
          </div>
        </div>

        <!-- STEP 2: matching (MENJODOHKAN) EDITOR -->
        <div v-else-if="questionForm.type === 'matching'" class="animate-fade-in-up">
          <div v-if="errorMessage" class="mb-4 rounded-xl bg-red-50 px-5 py-4 text-sm text-red-700 border border-red-200">{{ errorMessage }}</div>
          <div v-if="successMessage" class="mb-4 rounded-xl bg-emerald-50 px-5 py-4 text-sm text-emerald-700 border border-emerald-200">{{ successMessage }}</div>

          <!-- Premium Header Bar -->
          <div class="mb-8 overflow-hidden rounded-2xl border border-slate-200 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-900/10 shadow-xs">
            <div class="bg-white dark:bg-slate-900 px-8 py-6 border-b flex flex-wrap items-center justify-between gap-4">
              <div class="flex items-center gap-6">
                <h4 class="font-extrabold text-2xl dark:text-slate-100 flex items-center gap-3">
                   <span class="text-slate-800 dark:text-slate-200">No:</span>
                   <span class="bg-amber-500 text-white px-3 py-1 rounded-xl text-lg min-w-[40px] text-center shadow-lg shadow-amber-200">{{ questionForm.order_no }}</span>
                </h4>
                <div class="h-10 w-[1px] bg-slate-200 dark:bg-slate-800"></div>
                <div class="flex items-center gap-3">
                   <span class="text-xs font-black uppercase tracking-widest text-slate-400">Tipe Soal </span>
                   <FormControl v-model="questionForm.type" :options="questionTypeOptions" transparent class="min-w-[180px]" />
                   <div class="flex items-center gap-2">
                     <span class="text-xs font-black uppercase tracking-widest text-slate-400">Bobot</span>
                     <input
                       v-model.number="questionForm.weight"
                       type="number"
                       min="1"
                       step="1"
                       class="w-20 rounded-lg border border-slate-200 bg-white px-2 py-1 text-sm font-bold text-slate-700 outline-none focus:border-indigo-500 dark:border-slate-700 dark:bg-slate-900 dark:text-slate-100"
                     />
                   </div>
                 </div>
              </div>
              <div class="flex items-center gap-3">
                <button type="button" 
                  class="flex items-center gap-2 px-6 py-2.5 rounded-xl border-2 border-amber-500 text-amber-500 font-bold hover:bg-amber-50 transition-all active:scale-95"
                  @click="resetQuestionForm"
                >
                  <BaseIcon :path="mdiRefresh" size="20" />
                  <span>Reset</span>
                </button>
                <button type="button" 
                  class="flex items-center gap-2 px-8 py-2.5 rounded-xl bg-amber-500 text-white font-bold hover:bg-amber-600 shadow-lg shadow-amber-200 transition-all active:scale-95"
                  @click="saveQuestion"
                >
                  <BaseIcon :path="mdiPlus" size="20" />
                  <span>Simpan</span>
                </button>
              </div>
            </div>

            <div class="p-8 grid gap-8 xl:grid-cols-2 bg-white dark:bg-slate-950 items-start">
              <!-- Left: Stem -->
              <div class="space-y-4">
                <div class="flex items-center gap-2">
                  <div class="h-6 w-6 rounded-lg bg-amber-100 dark:bg-amber-900/50 flex items-center justify-center">
                    <BaseIcon :path="mdiFileDocumentOutline" size="14" class="text-amber-600 dark:text-amber-400" />
                  </div>
                  <label class="font-black text-[11px] uppercase tracking-[0.2em] text-slate-400">Stimulus / Soal Utama</label>
                </div>
                <RichEditor :key="`${activeEditorScopeKey}-stem`" v-model="questionForm.stem" :height="450" :toolbar="stemToolbar" placeholder="Tulis stimulus soal disini..." />
              </div>

              <!-- Right: Pairs Management -->
              <div class="space-y-4">
                 <div class="flex items-center justify-between px-2">
                    <label class="font-black text-[11px] uppercase tracking-[0.2em] text-slate-400">Pasangan Jawaban</label>
                    <span class="px-2 py-0.5 rounded-full bg-amber-100 text-amber-700 text-[10px] font-black uppercase">{{ matchingPairs.length }} Pasangan</span>
                 </div>

                 <!-- List of Pairs in Cards -->
                 <div class="space-y-4">
                    <div 
                      v-for="(pair, idx) in matchingPairs" 
                      :key="`${activeEditorScopeKey}-matching-${idx}`"
                      class="rounded-xl border border-slate-200 dark:border-slate-800 bg-slate-50/30 dark:bg-slate-900/30 overflow-hidden shadow-sm hover:shadow-md transition-all"
                    >
                       <div class="px-4 py-3 border-b border-slate-100 dark:border-slate-800 bg-white dark:bg-slate-900 flex items-center justify-between">
                          <div class="flex items-center gap-2">
                            <span class="h-6 w-6 flex items-center justify-center rounded-lg bg-amber-500 text-white text-[10px] font-black font-mono shadow-sm">{{ idx + 1 }}</span>
                            <span class="text-xs font-black text-slate-500 uppercase tracking-widest">Pasangan #{{ idx + 1 }}</span>
                          </div>
                          <button type="button" 
                            v-if="matchingPairs.length > 1"
                            @click="removeMatchingPair(idx)" 
                            class="text-red-400 hover:text-red-600 transition-colors"
                          >
                            <BaseIcon :path="mdiDelete" size="16" />
                          </button>
                       </div>
                       
                       <div class="p-4 space-y-4">
                          <!-- Ruas Kiri -->
                          <div class="space-y-1.5">
                            <label class="text-[9px] font-black text-amber-600 dark:text-amber-400 uppercase tracking-widest ml-1">Ruas Kiri (Akan Diacak)</label>
                            <div class="bg-white dark:bg-slate-900 rounded-lg border border-slate-200 dark:border-slate-800 overflow-hidden">
                              <RichEditor
                                :key="`${activeEditorScopeKey}-matching-left-${idx}`"
                                v-model="pair.left_content"
                                :height="120"
                                :toolbar="optionToolbar"
                                :placeholder="`Isi ruas kiri ${idx + 1}...`"
                              />
                            </div>
                          </div>
                          
                          <!-- Arrow -->
                          <div class="flex justify-center -my-2 relative z-10">
                             <div class="bg-white dark:bg-slate-800 border dark:border-slate-700 p-1 rounded-full shadow-sm">
                                <BaseIcon :path="mdiArrowDown" size="14" class="text-slate-400" />
                             </div>
                          </div>

                          <!-- Ruas Kanan -->
                          <div class="space-y-1.5">
                            <label class="text-[9px] font-black text-slate-500 dark:text-slate-400 uppercase tracking-widest ml-1">Ruas Kanan (Jawaban Benar)</label>
                            <div class="bg-white dark:bg-slate-900 rounded-lg border border-slate-200 dark:border-slate-800 overflow-hidden">
                              <RichEditor
                                :key="`${activeEditorScopeKey}-matching-right-${idx}`"
                                v-model="pair.right_content"
                                :height="120"
                                :toolbar="optionToolbar"
                                :placeholder="`Isi ruas kanan (jawaban) ${idx + 1}...`"
                              />
                            </div>
                          </div>
                       </div>
                    </div>
                 </div>

                 <button type="button"
                    class="w-full flex items-center justify-center gap-2 py-3.5 rounded-xl border-2 border-dashed border-amber-400 text-amber-600 font-bold bg-amber-50/30 hover:bg-amber-50 transition-all active:scale-95 mt-4"
                    @click="addMatchingPair"
                  >
                    <BaseIcon :path="mdiPlus" size="18" />
                    Tambah Pasangan Baru
                  </button>

                  <div class="rounded-xl bg-slate-50 dark:bg-slate-900/50 border border-slate-200 dark:border-slate-800 p-4 mt-6">
                    <p class="text-[11px] text-slate-500 italic">💡 Ruas kiri akan diacak saat ujian berlangsung. Pastikan ruas kanan berisi pasangan yang secara logika tepat untuk ruas kiri di kartu yang sama.</p>
                  </div>
              </div>
            </div>
          </div>

          <!-- List of other questions below -->
          <div class="space-y-4 mt-12">
             <h4 class="text-[10px] font-black text-slate-400 uppercase tracking-widest px-2">Daftar Soal Lainya ({{ questions.length }})</h4>
             <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
               <QuestionQuickAddCard @click="openQuickAddModal" />
               <div v-for="item in questions" :key="item.id" class="rounded-xl border border-slate-200 dark:border-slate-800 p-4 bg-white dark:bg-slate-900/40 hover:shadow-lg transition-all" :class="{ 'ring-2 ring-amber-500': editingQuestionId === item.id }">
                  <div class="flex items-center justify-between mb-2">
                    <span class="text-xs font-black text-slate-400">#{{ item.order_no }}</span>
                    <div class="flex gap-3">
                      <BaseIcon :path="mdiPencil" size="16" class="text-blue-500 cursor-pointer" @click="populateQuestionForm(item)" />
                      <BaseIcon :path="mdiDelete" size="16" class="text-red-500 cursor-pointer" @click="deleteQuestion(item.id)" />
                    </div>
                  </div>
                  <div class="mb-2 text-[10px] font-mono text-slate-400 truncate" :title="item.id">ID: {{ formatQuestionId(item.id) }}</div>
                  <span class="inline-block mb-1 px-2 py-0.5 rounded text-[9px] font-black uppercase bg-amber-100 dark:bg-amber-900/40 text-amber-600">{{ item.type.replace('_', ' ') }}</span>
                  <div class="text-xs text-slate-600 dark:text-slate-300 line-clamp-3" v-html="item.stem"></div>
                  <div v-if="buildQuestionCardAnswerPreviewHtml(item)" class="mt-2 rounded-lg border border-slate-100 bg-slate-50/60 px-2.5 py-2 text-[11px] text-slate-500 dark:border-slate-800 dark:bg-slate-900/40 dark:text-slate-300" v-html="buildQuestionCardAnswerPreviewHtml(item)"></div>
               </div>
             </div>
          </div>
        </div>



        <!-- STEP 2: FOCUSED EDITOR (Fallback for other types) -->
        <div v-else class="animate-fade-in-up">

        <div v-if="errorMessage" class="mb-4 rounded-xl bg-red-50 px-5 py-4 text-sm text-red-700 border border-red-200">{{ errorMessage }}</div>
        <div v-if="successMessage" class="mb-4 rounded-xl bg-emerald-50 px-5 py-4 text-sm text-emerald-700 border border-emerald-200">{{ successMessage }}</div>

        <div class="grid gap-8 xl:grid-cols-12 items-start">
          <!-- Left: Questions -->
          <div class="xl:col-span-7 space-y-6">
            <div class="flex items-center justify-between px-2">
              <h4 class="text-sm font-black text-slate-500 uppercase tracking-widest">Pertanyaan ({{ questions.length }})</h4>
              <BaseButton :icon="mdiRefresh" small outline @click="loadQuestions" />
            </div>

            <QuestionQuickAddCard
              title="Tambah Soal (Batch)"
              subtitle="Tambah banyak soal sekaligus dari template"
              @click="openQuickAddModal"
            />

            <div v-if="questions.length" class="space-y-4">
              <div v-for="item in questions" :key="item.id" class="relative rounded-2xl border border-slate-200 dark:border-slate-800 p-6 bg-white dark:bg-slate-900/40 hover:shadow-xl transition-all" :class="{ 'ring-2 ring-blue-500 bg-blue-50/10': editingQuestionId === item.id }">
                <div class="absolute -left-3 top-6 h-8 w-8 flex items-center justify-center rounded-lg bg-slate-900 text-sm font-black text-white shadow-lg">{{ item.order_no }}</div>
                <div class="mb-4 flex items-center justify-between pl-6">
                  <span class="rounded-lg bg-slate-100 dark:bg-slate-800 px-3 py-1 text-[10px] font-black uppercase text-slate-500 tracking-wider">{{ item.type.replace('_', ' ') }}</span>
                  <div class="flex items-center gap-5">
                    <BaseIcon :path="mdiPencil" size="18" class="text-blue-500 hover:scale-125 transition-transform cursor-pointer" @click="populateQuestionForm(item)" />
                    <BaseIcon :path="mdiDelete" size="18" class="text-red-500 hover:scale-125 transition-transform cursor-pointer" @click="deleteQuestion(item.id)" />
                  </div>
                </div>
                <div class="mb-3 pl-6 text-[10px] font-mono text-slate-400 truncate" :title="item.id">ID: {{ formatQuestionId(item.id) }}</div>
                <div class="mb-6 pl-6 text-base text-slate-700 dark:text-slate-200 leading-relaxed">{{ item.stem }}</div>
                
                <!-- Answer Preview -->
                <div v-if="item.options?.length" class="pl-6 space-y-1 mb-2">
                   <div v-for="opt in item.options" :key="opt.label" class="text-sm flex items-start gap-2" :class="opt.is_correct ? 'text-emerald-600 font-bold' : 'text-slate-500'">
                      <span>{{ opt.label }}.</span>
                      <span>{{ opt.content }}</span>
                   </div>
                </div>
              </div>
            </div>
            <div v-else class="rounded-3xl border-2 border-dashed border-slate-200 dark:border-slate-800 p-24 text-center">
              <p class="text-slate-400">Belum ada pertanyaan. Mulai tambahkan di panel kanan.</p>
            </div>
          </div>

          <!-- Right: Form -->
          <div class="xl:col-span-5">
            <CardBox class="sticky top-24 shadow-2xl p-8">
               <h4 class="text-xl font-black mb-6 flex items-center justify-between">
                  <div class="flex items-center gap-2">
                    <span class="text-blue-500">{{ editingQuestionId ? '📝' : '✨' }}</span>
                    {{ editingQuestionId ? 'Edit Soal' : 'Tambah Soal' }}
                  </div>
                  <span class="bg-indigo-600 text-white px-3 py-1 rounded-lg text-sm">No: {{ questionForm.order_no }}</span>
               </h4>

               <div class="grid gap-5">
                 <FormField label="Tipe Soal">
                   <FormControl v-model="questionForm.type" :options="questionTypeOptions" />
                 </FormField>
                 <FormField label="Bobot Soal">
                   <FormControl v-model.number="questionForm.weight" type="number" min="1" />
                 </FormField>
                 <FormField label="Isi Pertanyaan">
                   <FormControl v-model="questionForm.stem" type="textarea" rows="4" />
                 </FormField>
                 
                 <div v-if="['mc_single'].includes(questionForm.type)">
                    <FormField label="Opsi (Format: Label|Teks|true)">
                      <FormControl v-model="questionForm.options_text" type="textarea" rows="4" class="font-mono text-xs" />
                    </FormField>
                 </div>

                 <BaseButton :icon="editingQuestionId ? mdiContentSave : mdiPlus" color="info" :label="editingQuestionId ? 'Update Soal' : 'Simpan Soal'" class="w-full py-4 text-lg font-bold" @click="saveQuestion" />
                 <BaseButton v-if="editingQuestionId" small color="whiteDark" label="Batal Edit" outline class="w-full mt-2" @click="resetQuestionForm" />

                 <!-- DOCX -->
                 <div class="mt-8 pt-6 border-t dark:border-slate-800">
                    <h5 class="text-[10px] font-black uppercase text-slate-400 mb-3 tracking-widest italic">Punya file DOCX? Import di sini:</h5>
                    <div class="flex gap-2">
                       <FormFilePicker v-model="docxFile" label="Pilih File" small class="flex-1" />
                       <BaseButton :icon="mdiFileDocumentOutline" color="info" label="Import" small :loading="isImporting" :disabled="!docxFile" @click="importDocx" />
                    </div>
                    <div class="mt-3 rounded-2xl border border-purple-200 bg-purple-50 p-4 text-purple-900 dark:border-purple-500/20 dark:bg-purple-500/10 dark:text-purple-100">
                      <div class="flex items-center justify-between gap-3">
                      <p class="text-xs leading-relaxed text-purple-800 dark:text-purple-100/80">
                        Gunakan template dari submenu <span class="font-semibold">Impor Soal</span>, lalu simpan sebagai <span class="font-mono">.docx</span> sebelum import di sini.
                      </p>
                      </div>
                    </div>
                 </div>
               </div>
            </CardBox>
          </div>
        </div>
      </div>
    </SectionMain>
  </LayoutAuthenticated>
</template>

<style scoped>
.animate-fade-in-up {
  animation: fadeInUp 0.5s ease-out;
}
@keyframes fadeInUp {
  from { opacity: 0; transform: translateY(20px); }
  to { opacity: 1; transform: translateY(0); }
}
.white-label :deep(label) {
  color: white !important;
}
.white-label :deep(.text-slate-500),
.white-label :deep(.text-gray-500) {
  color: rgba(255, 255, 255, 0.7) !important;
}
</style>
