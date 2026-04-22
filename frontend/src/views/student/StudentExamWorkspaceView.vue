<script setup>
import { computed, nextTick, onErrorCaptured, onMounted, ref, watch, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  mdiCheckCircleOutline,
  mdiInformationOutline,
  mdiAlert
} from '@mdi/js'
import BaseIcon from '@/components/BaseIcon.vue'
import { api } from '@/services/api.js'
import { useAuthStore } from '@/stores/auth.js'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()

const sessionId = computed(() => route.params.sessionId)
const isLoading = ref(true)
const errorMessage = ref('')
const questions = ref([])
const currentIndex = ref(0)
const answers = ref({})
const timeLeft = ref(0)
const timerInterval = ref(null)
const submitDone = ref(false)
const showSubmitModal = ref(false)
const showQuestionListModal = ref(false)
const flagged = ref({})
const examTitle = ref('AtigaCBT Workspace')

const currentQuestion = computed(() => questions.value[currentIndex.value])
const participantName = computed(() => authStore.userDisplayName)

const cardScrollEl = ref(null)

const FONT_KEY = 'mycbt_exam_font_size'
const fontSize = ref('md') // sm | md | lg
try {
  const saved = localStorage.getItem(FONT_KEY)
  if (saved === 'sm' || saved === 'md' || saved === 'lg') fontSize.value = saved
} catch {
  // ignore
}
watch(
  fontSize,
  (v) => {
    try {
      localStorage.setItem(FONT_KEY, String(v || 'md'))
    } catch {
      // ignore
    }
  },
  { flush: 'sync' },
)
const setFontSize = (v) => {
  if (v !== 'sm' && v !== 'md' && v !== 'lg') return
  fontSize.value = v
}
const fontClass = computed(() => {
  if (fontSize.value === 'sm') return 'tka-font-sm'
  if (fontSize.value === 'lg') return 'tka-font-lg'
  return 'tka-font-md'
})

const tokenVerified = ref(false)
const tokenValue = ref('')
const tokenError = ref('')
const tokenChecking = ref(false)
const TOKEN_OK_KEY = computed(() => `mycbt_session_token_ok_${String(sessionId.value || '')}`)

const verifyToken = async () => {
  if (!sessionId.value) return
  tokenError.value = ''
  const token = String(tokenValue.value || '').trim()
  if (!token) {
    tokenError.value = 'Token wajib diisi.'
    return
  }
  tokenChecking.value = true
  try {
    await api.post(`/api/v1/student/sessions/${sessionId.value}/verify-token`, { token })
    tokenVerified.value = true
    try {
      localStorage.setItem(TOKEN_OK_KEY.value, '1')
    } catch {
      // ignore
    }
    await loadExamData()
  } catch (err) {
    tokenError.value = err?.response?.data?.error?.message || 'Token tidak valid'
  } finally {
    tokenChecking.value = false
  }
}

const ensureAnswerShapeForQuestion = (q) => {
  const qid = String(q?.id ?? '')
  if (!qid) return
  const t = String(q?.type ?? '')
  const cur = answers.value[qid]

  if (t === 'mc_single') {
    // Expected schema: { selected_option_id: "opt-id" }
    if (cur && typeof cur === 'object' && !Array.isArray(cur)) {
      answers.value[qid] = { selected_option_id: String(cur.selected_option_id || '') }
      return
    }
    // Back-compat: stored as raw option id string
    answers.value[qid] = { selected_option_id: cur ? String(cur) : '' }
    return
  }

  if (t === 'mc_multiple') {
    // Expected schema: { selected_option_ids: ["a","b"] }
    if (cur && typeof cur === 'object' && !Array.isArray(cur)) {
      const arr = Array.isArray(cur.selected_option_ids) ? cur.selected_option_ids : []
      answers.value[qid] = { selected_option_ids: arr.map(v => String(v)) }
      return
    }
    // Back-compat: stored as string[] or single string
    if (Array.isArray(cur)) {
      answers.value[qid] = { selected_option_ids: cur.map(v => String(v)) }
      return
    }
    answers.value[qid] = { selected_option_ids: cur ? [String(cur)] : [] }
    return
  }

  if (t === 'true_false') {
    // Expected schema:
    // legacy: { value: true/false }
    // statements: { values: { "<statement_id>": true/false } }
    if (q?.statements?.length) {
      if (cur && typeof cur === 'object' && !Array.isArray(cur)) {
        const next = { values: {} }
        const rawValues = cur.values && typeof cur.values === 'object' && !Array.isArray(cur.values) ? cur.values : {}
        for (const st of q.statements) {
          const sid = String(st?.id ?? '')
          if (!sid) continue
          const v = rawValues[sid]
          if (v === true || v === false) next.values[sid] = v
          else if (v === 'true' || v === 1) next.values[sid] = true
          else if (v === 'false' || v === 0) next.values[sid] = false
        }
        answers.value[qid] = next
        return
      }
      // Back-compat: map like { [sid]: bool }
      if (cur && typeof cur === 'object' && !Array.isArray(cur)) {
        const next = { values: {} }
        for (const st of q.statements) {
          const sid = String(st?.id ?? '')
          if (!sid) continue
          const v = cur[sid]
          if (v === true || v === false) next.values[sid] = v
          else if (v === 'true' || v === 1) next.values[sid] = true
          else if (v === 'false' || v === 0) next.values[sid] = false
        }
        answers.value[qid] = next
        return
      }
      answers.value[qid] = { values: {} }
      return
    }

    if (cur && typeof cur === 'object' && !Array.isArray(cur) && typeof cur.value === 'boolean') {
      answers.value[qid] = { value: !!cur.value }
      return
    }
    if (cur === true || cur === false) {
      answers.value[qid] = { value: cur }
      return
    }
    if (cur === 'true' || cur === 1) {
      answers.value[qid] = { value: true }
      return
    }
    if (cur === 'false' || cur === 0) {
      answers.value[qid] = { value: false }
      return
    }
    answers.value[qid] = { value: null }
    return
  }

  if (t === 'matching') {
    // Expected schema: { pairs: { "<pairId:L>": "<pairId:R>" } }
    if (cur && typeof cur === 'object' && !Array.isArray(cur)) {
      const pairs = cur.pairs && typeof cur.pairs === 'object' && !Array.isArray(cur.pairs) ? cur.pairs : {}
      const nextPairs = {}
      for (const [k, v] of Object.entries(pairs)) {
        const kk = String(k || '').trim()
        const vv = String(v || '').trim()
        if (!kk || !vv) continue
        nextPairs[kk] = vv
      }
      answers.value[qid] = { pairs: nextPairs }
      return
    }
    answers.value[qid] = { pairs: {} }
    return
  }

  if (t === 'essay' || t === 'short_answer') {
    // Expected schema: { text: "..." }
    if (cur && typeof cur === 'object' && !Array.isArray(cur)) {
      answers.value[qid] = { text: String(cur.text || '') }
      return
    }
    // Back-compat: raw text
    answers.value[qid] = { text: cur ? String(cur) : '' }
    return
  }
}

const renderHtml = (html) => {
  const raw = String(html || '')
  if (!raw.trim()) return '<p><em>(Konten soal kosong)</em></p>'
  try {
    const doc = new DOMParser().parseFromString(raw, 'text/html')
    doc.querySelectorAll('script, iframe, object, embed').forEach(n => n.remove())
    doc.querySelectorAll('*').forEach(el => {
      for (const attr of Array.from(el.attributes || [])) {
        if (/^on/i.test(attr.name)) el.removeAttribute(attr.name)
      }
      if (el.tagName === 'IMG') {
        el.setAttribute('loading', 'lazy')
      }
      const style = el.getAttribute && el.getAttribute('style')
      if (style) {
        const parts = style.split(';').map(s => s.trim()).filter(Boolean)
        const filtered = parts.filter(p => {
          const name = (p.split(':')[0] || '').trim().toLowerCase()
          return name && name !== 'color' && name !== 'background' && name !== 'background-color'
        })
        if (filtered.length) el.setAttribute('style', filtered.join('; '))
        else el.removeAttribute('style')
      }
    })
    return doc.body.innerHTML
  } catch {
    return raw
  }
}

const formatTime = (seconds) => {
  const h = Math.floor(seconds / 3600)
  const m = Math.floor((seconds % 3600) / 60)
  const s = seconds % 60
  return [h, m, s].map(v => v < 10 ? '0' + v : v).join(':')
}

const loadExamData = async () => {
  if (!sessionId.value) {
    isLoading.value = false
    errorMessage.value = 'ID Sesi tidak ditemukan'
    return
  }
  isLoading.value = true
  errorMessage.value = ''
  try {
    const [sessResp, questionsResp, answersResp] = await Promise.all([
      api.get(`/api/v1/student/sessions/${sessionId.value}`),
      api.get(`/api/v1/student/sessions/${sessionId.value}/questions`),
      api.get(`/api/v1/student/sessions/${sessionId.value}/answers`)
    ])

    const state = sessResp.data.data
    const questionsData = questionsResp.data.data
    
    // Update topbar title
    if (state.exam?.title) {
       examTitle.value = state.exam.title
       document.title = state.exam.title + ' - AtigaCBT'
    }

    questions.value = Array.isArray(questionsData.questions) ? questionsData.questions : []
    
    // Process existing answers (mapped from array to object map)
    const existingAnswersList = Array.isArray(answersResp.data.data) ? answersResp.data.data : []
    const processed = {}
    
    // Fill with empty defaults based on question type
    questions.value.forEach(q => {
      if (q.type === 'mc_single') processed[q.id] = { selected_option_id: '' }
      else if (q.type === 'mc_multiple') processed[q.id] = { selected_option_ids: [] }
      else if (q.type === 'true_false') processed[q.id] = q.statements?.length ? { values: {} } : { value: null }
      else if (q.type === 'matching') processed[q.id] = { pairs: {} }
      else if (q.type === 'essay' || q.type === 'short_answer') processed[q.id] = { text: '' }
      else processed[q.id] = {}
    })

    // Overlay with actual answers from backend
    existingAnswersList.forEach(ans => {
      const qid = ans.question_id
      if (!qid) return
      try {
        // answer_json is already a JSON string from backend scan if it was raw bytes,
        // but let's check if it's already an object/array from the JSON response.
        const val = ans.answer_json
        if (typeof val === 'string') {
          processed[qid] = JSON.parse(val)
        } else {
          processed[qid] = val
        }
      } catch {
        processed[qid] = ans.answer_json
      }
    })
    answers.value = processed
    questions.value.forEach(q => ensureAnswerShapeForQuestion(q))

    // Timer calculation from session state
    timeLeft.value = state.remaining_seconds || 0
    
    startTimer()
    scheduleRenderMath()
  } catch (err) {
    errorMessage.value = err.response?.data?.error?.message || 'Gagal memuat data ujian'
  } finally {
    isLoading.value = false
  }
}

const startTimer = () => {
  if (timerInterval.value) clearInterval(timerInterval.value)
  timerInterval.value = setInterval(() => {
    if (timeLeft.value > 0) {
      timeLeft.value--
    } else {
      clearInterval(timerInterval.value)
      // Auto submit if time is out
      submitExam()
    }
  }, 1000)
}

const saveAnswer = async (question) => {
  if (!question?.id) return
  try {
    const answer = answers.value[question.id] ?? {}
    await api.post(`/api/v1/student/sessions/${sessionId.value}/answers`, {
      question_id: question.id,
      answer_json: JSON.stringify(answer)
    })
  } catch (err) {
    console.error('Failed to save answer:', err)
  }
}

const submitExam = async () => {
  try {
    isLoading.value = true
    await api.post(`/api/v1/student/sessions/${sessionId.value}/submit`)
    submitDone.value = true
    showSubmitModal.value = false
  } catch (err) {
    alert(err.response?.data?.error?.message || 'Gagal mengirim jawaban')
  } finally {
    isLoading.value = false
  }
}

let mathRaf = 0
const fixCommonLatexCommands = (input) => {
  const s = String(input || '')
  // Auto-fix common TeX commands when author forgot leading backslash, e.g. "frac{1}{2}".
  return s.replace(
    /(^|[^\\a-zA-Z])(frac|sqrt|times|cdot|pm|mp|div|leq|geq|neq|approx|sum|prod|int|lim|infty|pi|alpha|beta|gamma|theta|lambda|mu|sigma|omega|sin|cos|tan|log|ln)\b/g,
    '$1\\\\$2',
  )
}
const renderMath = (rootEl) => {
  if (!rootEl) return
  if (window.renderMathInElement) {
    window.renderMathInElement(rootEl, {
      delimiters: [
        { left: '$$', right: '$$', display: true },
        { left: '$', right: '$', display: false },
        { left: '\\(', right: '\\)', display: false },
        { left: '\\[', right: '\\]', display: true },
      ],
      throwOnError: false,
      preProcess: fixCommonLatexCommands,
    })
  }
}
const scheduleRenderMath = async () => {
  if (mathRaf) cancelAnimationFrame(mathRaf)
  await nextTick()
  mathRaf = requestAnimationFrame(() => {
    try {
      renderMath(cardScrollEl.value)
    } catch (e) {
      console.warn('renderMath failed:', e)
    }
  })
}

const isAnswered = (q) => {
  const id = String(q?.id ?? '')
  if (!id) return false
  const t = String(q?.type ?? '')
  const ans = answers.value[id]
  if (!ans) return false

  if (t === 'mc_single') return !!String(ans?.selected_option_id || '').trim()
  if (t === 'mc_multiple') return Array.isArray(ans?.selected_option_ids) && ans.selected_option_ids.length > 0
  if (t === 'true_false') {
    if (q?.statements?.length) return ans?.values && Object.keys(ans.values).length > 0
    return typeof ans?.value === 'boolean'
  }
  if (t === 'short_answer' || t === 'essay') return !!String(ans?.text || '').trim()
  if (t === 'matching') return ans?.pairs && Object.keys(ans.pairs).length > 0
  return true
}

const goPrev = () => {
  setIndex(currentIndex.value - 1)
}

const goNext = () => {
  setIndex(currentIndex.value + 1)
}

const setIndex = (idx) => {
  const n = Number(idx)
  if (!Number.isFinite(n)) return
  if (n < 0 || n >= questions.value.length) return
  currentIndex.value = n
}

const onNavClick = (idx, ev) => {
  if (ev?.preventDefault) ev.preventDefault()
  if (ev?.stopPropagation) ev.stopPropagation()
  setIndex(idx)
}

const toggleMulti = (questionId, optId) => {
  const qid = String(questionId || '')
  if (!qid) return
  const id = String(optId || '')
  if (!id) return
  ensureAnswerShapeForQuestion({ id: qid, type: 'mc_multiple' })
  const cur = answers.value[qid]
  const arr = Array.isArray(cur?.selected_option_ids) ? cur.selected_option_ids : []
  const i = arr.indexOf(id)
  if (i >= 0) arr.splice(i, 1)
  else arr.push(id)
  answers.value[qid] = { selected_option_ids: arr }
  saveAnswer(currentQuestion.value)
}

const setTrueFalseLegacy = (val) => {
  const q = currentQuestion.value
  if (!q?.id) return
  answers.value[q.id] = { value: !!val }
  saveAnswer(q)
}

const setTrueFalseStatement = (statementId, val) => {
  const q = currentQuestion.value
  if (!q?.id) return
  ensureAnswerShapeForQuestion(q)
  const cur = answers.value[q.id]
  const values = cur?.values && typeof cur.values === 'object' && !Array.isArray(cur.values) ? { ...cur.values } : {}
  values[String(statementId)] = !!val
  answers.value[q.id] = { values }
  saveAnswer(q)
}

const setMatchingPair = (leftId, rightId) => {
  const q = currentQuestion.value
  if (!q?.id) return
  ensureAnswerShapeForQuestion(q)
  const cur = answers.value[q.id]
  const pairs = cur?.pairs && typeof cur.pairs === 'object' && !Array.isArray(cur.pairs) ? { ...cur.pairs } : {}
  const l = String(leftId || '').trim()
  const r = String(rightId || '').trim()
  if (!l) return
  if (!r) delete pairs[l]
  else pairs[l] = r
  answers.value[q.id] = { pairs }
  saveAnswer(q)
}

const isFlagged = (qid) => !!flagged.value[String(qid || '')]

const toggleFlagged = () => {
  const q = currentQuestion.value
  if (!q) return
  const id = String(q.id || '')
  if (!id) return
  flagged.value[id] = !flagged.value[id]
}

watch(currentIndex, () => {
  const q = questions.value?.[currentIndex.value]
  if (q) ensureAnswerShapeForQuestion(q)
  scheduleRenderMath()
  if (cardScrollEl.value && typeof cardScrollEl.value.scrollTo === 'function') {
    cardScrollEl.value.scrollTo({ top: 0, behavior: 'smooth' })
  } else if (cardScrollEl.value) {
    cardScrollEl.value.scrollTop = 0
  }
  window.scrollTo({ top: 0, behavior: 'smooth' })

  // Keep the active number visible in the sidebar navigator when list is long.
  setTimeout(() => {
    const el = document.querySelector(`[data-qnav-idx="${currentIndex.value}"]`)
    if (el && typeof el.scrollIntoView === 'function') {
      el.scrollIntoView({ block: 'nearest', inline: 'nearest' })
    }
  }, 50)
}, { flush: 'post' })

onMounted(() => {
  // Gate: require token entry before showing questions.
  isLoading.value = false
  try {
    if (localStorage.getItem(TOKEN_OK_KEY.value) === '1') {
      tokenVerified.value = true
      loadExamData()
    }
  } catch {
    // ignore
  }
  document.addEventListener('visibilitychange', handleVisibilityChange)
})
onUnmounted(() => {
  if (timerInterval.value) clearInterval(timerInterval.value)
  if (mathRaf) cancelAnimationFrame(mathRaf)
  document.removeEventListener('visibilitychange', handleVisibilityChange)
})

const handleVisibilityChange = () => {
  if (document.visibilityState === 'hidden' && sessionId.value && !submitDone.value) {
    api.post(`/api/v1/student/sessions/${sessionId.value}/heartbeat`, {
      type: 'focus_loss',
      timestamp: new Date().toISOString()
    }).catch(() => {})
  }
}

onErrorCaptured((err) => {
  errorMessage.value = err?.message ? String(err.message) : 'Terjadi error saat merender soal'
  return false
})

const currentTextAnswer = computed({
  get() {
    const q = currentQuestion.value
    if (!q?.id) return ''
    ensureAnswerShapeForQuestion(q)
    return String(answers.value[q.id]?.text || '')
  },
  set(v) {
    const q = currentQuestion.value
    if (!q?.id) return
    ensureAnswerShapeForQuestion(q)
    answers.value[q.id].text = String(v ?? '')
  },
})

const stripHtml = (html) => String(html || '').replace(/<[^>]*>/g, '').trim()

const matchingRightOptions = computed(() => {
  const q = currentQuestion.value
  if (!q || q.type !== 'matching') return []
  const items = Array.isArray(q.matching_right) ? q.matching_right : []
  return items.map(it => ({
    id: String(it.id || ''),
    text: stripHtml(it.content || ''),
    orderNo: it.order_no ?? null,
  }))
})

</script>

<template>
  <div class="tka-theme min-h-screen bg-slate-100">
    <!-- TOP NAVBAR -->
    <header class="tka-topbar text-white px-6 py-3 flex items-center justify-between shadow sticky top-0 z-50">
      <div class="font-bold tracking-tight text-sm md:text-base uppercase select-none">
        {{ examTitle }}
      </div>
      <div class="flex items-center gap-4">
        <div class="hidden sm:block text-right leading-tight">
          <div class="text-xs font-semibold uppercase opacity-90">Peserta</div>
          <div class="text-sm font-black">{{ participantName }}</div>
        </div>
        <div class="tka-timer bg-white text-slate-900 px-4 py-1.5 rounded-full font-semibold flex items-center gap-2">
          <span class="text-xs">SISA WAKTU:</span>
          <span class="font-mono">{{ formatTime(timeLeft) }}</span>
        </div>
        <button
          type="button"
          class="lg:hidden bg-white/10 border border-white/20 hover:bg-white/15 active:bg-white/20 px-3 py-2 rounded font-black uppercase text-[11px] tracking-widest"
          @click="showQuestionListModal = true"
        >
          Daftar Soal
        </button>
      </div>
    </header>

    <!-- TOKEN GATE -->
    <div
      v-if="sessionId && !submitDone && !tokenVerified"
      class="max-w-xl mx-auto mt-20 p-12 bg-white rounded-3xl border-2 border-slate-100 shadow-2xl text-center animate-fade-in"
    >
      <h2 class="text-2xl font-black text-slate-800 mb-3 uppercase">Masukkan Token Ujian</h2>
      <p class="text-slate-500 mb-8 leading-relaxed">
        Token diperlukan sebelum soal ditampilkan. Silakan masukkan token yang diberikan pengawas/guru.
      </p>

      <div class="max-w-sm mx-auto space-y-3">
        <input
          v-model="tokenValue"
          class="w-full px-5 py-4 rounded-2xl border-2 border-slate-100 bg-white focus:border-[#0B7EA1] outline-none transition-all shadow-inner text-lg font-black tracking-widest text-slate-800 text-center"
          placeholder="TOKEN"
          autocomplete="off"
          spellcheck="false"
          @keydown.enter.prevent="verifyToken"
        />
        <div v-if="tokenError" class="text-sm text-rose-600 font-semibold">{{ tokenError }}</div>
        <button
          type="button"
          class="w-full bg-[#0D47A1] text-white py-4 rounded-2xl font-black uppercase tracking-widest shadow-xl transition-all hover:scale-105 active:scale-95 disabled:opacity-60 disabled:hover:scale-100"
          :disabled="tokenChecking"
          @click="verifyToken"
        >
          {{ tokenChecking ? 'Memeriksa...' : 'Mulai Ujian' }}
        </button>
        <button
          type="button"
          class="w-full bg-slate-100 text-slate-700 py-4 rounded-2xl font-black uppercase tracking-widest transition-all hover:bg-slate-200"
          @click="router.push('/student/ujian')"
        >
          Kembali
        </button>
      </div>
    </div>

    <!-- LOADING / ERROR -->
    <div v-if="isLoading && !submitDone" class="flex flex-col items-center justify-center min-h-[60vh]">
       <div class="h-16 w-16 border-4 border-[#0D47A1] border-t-transparent rounded-full animate-spin mb-6"></div>
       <p class="text-[#0D47A1] font-black uppercase tracking-[0.3em] animate-pulse">Menyiapkan Lembar Ujian...</p>
    </div>

    <div v-else-if="errorMessage" class="max-w-xl mx-auto mt-20 p-12 bg-white rounded-3xl border-2 border-red-50 shadow-2xl text-center">
        <BaseIcon :path="mdiAlert" size="64" class="text-red-500 mb-6 mx-auto" />
        <h2 class="text-2xl font-black text-slate-800 mb-4 uppercase">Terjadi Kesalahan</h2>
        <p class="text-slate-600 mb-8 leading-relaxed">{{ errorMessage }}</p>
        <button @click="router.push('/student/exams')" class="bg-[#0D47A1] text-white px-8 py-3 rounded-xl font-bold uppercase transition-all hover:scale-105 active:scale-95">Kembali</button>
    </div>

    <!-- SUBMIT DONE -->
    <div v-else-if="submitDone" class="max-w-2xl mx-auto mt-20 p-16 bg-white rounded-3xl border-2 border-emerald-50 shadow-2xl text-center animate-fade-in">
        <div class="w-24 h-24 bg-emerald-100 rounded-full flex items-center justify-center mx-auto mb-8 shadow-inner">
           <BaseIcon :path="mdiCheckCircleOutline" size="48" class="text-emerald-600" />
        </div>
        <h2 class="text-3xl font-black text-slate-800 mb-4 uppercase tracking-tight">Ujian Selesai!</h2>
        <p class="text-slate-500 mb-12 text-lg leading-relaxed">Jawaban Anda telah berhasil terkirim ke server. Terima kasih telah mengikuti simulasi ujian dengan jujur.</p>
        <button 
           @click="router.push('/student/exams')" 
           class="bg-[#0D47A1] text-white px-12 py-4 rounded-2xl font-black uppercase tracking-widest shadow-xl hover:shadow-2xl hover:scale-105 active:scale-95 transition-all"
        >
           KEMBALI KE BERANDA
        </button>
    </div>

    <!-- MAIN EXAM AREA -->
    <Transition name="qswap" mode="out-in">
    <main v-if="sessionId && !submitDone && currentQuestion" :key="currentQuestion.id" class="max-w-[1500px] mx-auto px-6 py-6 pb-24">
       <div class="grid lg:grid-cols-[1fr_340px] gap-6 items-start">
          
          <!-- LEFT COLUMN: Main Card -->
	          <div class="bg-white rounded-lg border border-slate-200 shadow-sm min-h-[75vh] max-h-[calc(100vh-180px)] flex flex-col">
	             <!-- Card Header -->
	             <div class="px-6 py-4 border-b border-slate-200 flex items-center justify-between bg-white">
	                <span class="text-[#0B7EA1] font-bold uppercase text-sm tracking-wide">SOAL NOMOR: {{ currentIndex + 1 }}</span>
	                <div class="flex items-center gap-4 text-[11px] font-bold text-slate-500">
	                   <span class="uppercase opacity-60">Ukuran Font:</span>
	                   <button
	                     type="button"
	                     class="h-6 w-8 flex items-center justify-center border rounded font-bold leading-none transition-colors"
	                     :class="fontSize === 'sm' ? 'border-2 border-[#0B7EA1] bg-white text-[#0B7EA1]' : 'border-slate-300 bg-slate-50 text-slate-700 hover:border-slate-400'"
	                     @click="setFontSize('sm')"
	                   >
	                     A
	                   </button>
	                   <button
	                     type="button"
	                     class="h-7 w-9 flex items-center justify-center border rounded font-black leading-none transition-colors"
	                     :class="fontSize === 'md' ? 'border-2 border-[#0B7EA1] bg-white text-[#0B7EA1]' : 'border-slate-300 bg-slate-50 text-slate-700 hover:border-slate-400'"
	                     @click="setFontSize('md')"
	                   >
	                     A
	                   </button>
	                   <button
	                     type="button"
	                     class="h-7 w-10 flex items-center justify-center border rounded font-bold leading-none transition-colors text-base"
	                     :class="fontSize === 'lg' ? 'border-2 border-[#0B7EA1] bg-white text-[#0B7EA1]' : 'border-slate-300 bg-slate-50 text-slate-700 hover:border-slate-400'"
	                     @click="setFontSize('lg')"
	                   >
	                     A
	                   </button>
	                </div>
	             </div>

		               <div ref="cardScrollEl" class="flex-1 overflow-auto" :class="fontClass">
	                <!-- STIMULUS / CONTENT -->
	                <div class="p-6 prose prose-slate max-w-none text-base leading-relaxed text-slate-900">
	                   <div class="flex justify-between items-start mb-4">
	                      <div class="text-[10px] font-black uppercase text-slate-400 tracking-[0.2em]">STIMULUS / SOAL UTAMA</div>
	                   </div>
	                   <div v-html="renderHtml(currentQuestion.stem)"></div>
	                </div>

	                <!-- INTERACTION / ANSWERS -->
		                <div :key="currentQuestion.id + ':' + currentQuestion.type" class="px-6 pb-10 pt-6 border-t border-slate-200 bg-white">
	                <div class="mb-8 flex items-center justify-between">
	                   <div class="text-[10px] font-black uppercase text-slate-400 tracking-[0.2em]">
	                      {{ currentQuestion.type === 'true_false' ? 'DAFTAR PERNYATAAN & JAWABAN' : 'OPSI JAWABAN' }}
	                   </div>
	                   <div class="hidden sm:block text-xs text-slate-400">Klik untuk memilih jawaban</div>
	                </div>

                <!-- MC Single -->
                <div v-if="currentQuestion.type === 'mc_single'" class="space-y-4">
                   <label
                     v-for="opt in currentQuestion.options" 
                     :key="opt.id"
                     class="flex items-center gap-4 p-4 rounded border transition-colors cursor-pointer bg-white text-left w-full"
                     :class="answers[currentQuestion.id]?.selected_option_id === String(opt.id) ? 'border-[#0B7EA1] bg-[#0B7EA1]/[0.03]' : 'border-slate-200 hover:border-slate-300'"
                   >
                      <input
                        type="radio"
                        :name="'mc-'+currentQuestion.id"
                        class="h-4 w-4 accent-slate-600"
                        :checked="answers[currentQuestion.id]?.selected_option_id === String(opt.id)"
                        @change="() => { answers[currentQuestion.id] = { selected_option_id: String(opt.id) }; saveAnswer(currentQuestion) }"
                      />
	                      <div class="text-slate-800 font-medium text-sm leading-relaxed tka-option-text" v-html="renderHtml(opt.content)"></div>
	                   </label>
	                </div>

                <!-- MC Multiple (PG Kompleks) -->
                <div v-else-if="currentQuestion.type === 'mc_multiple'" class="space-y-4">
                   <button
                     type="button"
                     v-for="opt in currentQuestion.options" 
                     :key="opt.id"
                     @click="toggleMulti(currentQuestion.id, opt.id)"
                     class="flex items-start gap-4 p-4 rounded border transition-colors cursor-pointer bg-white text-left w-full"
                     :aria-pressed="answers[currentQuestion.id]?.selected_option_ids?.includes(String(opt.id)) ? 'true' : 'false'"
                     :class="answers[currentQuestion.id]?.selected_option_ids?.includes(String(opt.id)) ? 'border-[#0B7EA1] bg-[#0B7EA1]/[0.03]' : 'border-slate-200 hover:border-slate-300'"
                   >
                      <div class="pt-0.5">
                        <div
                          class="h-5 w-5 rounded border flex items-center justify-center"
                          :class="answers[currentQuestion.id]?.selected_option_ids?.includes(String(opt.id)) ? 'bg-[#0B7EA1] border-[#0B7EA1] text-white' : 'bg-white border-slate-300 text-transparent'"
                        >
                          <span class="text-xs font-black leading-none">✓</span>
                        </div>
                      </div>
	                      <div class="text-slate-800 font-medium text-sm leading-relaxed tka-option-text" v-html="renderHtml(opt.content)"></div>
	                   </button>
	                </div>

                <!-- True/False (ANBK-style) + Legacy Fallback -->
                <div v-else-if="currentQuestion.type === 'true_false'" class="space-y-4">
                   <div v-if="currentQuestion.statements?.length" class="bg-white rounded-2xl border-2 border-slate-200 overflow-hidden shadow-sm">
                      <table class="w-full table-fixed border-collapse text-sm border-2 border-slate-300">
                         <colgroup>
                           <col class="w-[76%]" />
                           <col class="w-[12%]" />
                           <col class="w-[12%]" />
                         </colgroup>
                         <thead class="bg-slate-50/80 font-black text-[10px] text-slate-500 uppercase tracking-widest">
                            <tr>
                               <th class="py-4 px-6 text-left border-2 border-slate-300">Pernyataan / Pertanyaan</th>
                               <th class="py-4 px-3 text-center border-2 border-slate-300">Benar</th>
                               <th class="py-4 px-3 text-center border-2 border-slate-300">Salah</th>
                            </tr>
                         </thead>
                         <tbody>
                            <tr v-for="st in currentQuestion.statements" :key="st.id" class="hover:bg-slate-50/30 transition-colors">
                               <td class="py-6 px-6 text-slate-900 font-medium align-middle border-2 border-slate-300 leading-relaxed" v-html="renderHtml(st.content)"></td>
                               <td class="py-6 px-3 text-center align-middle border-2 border-slate-300">
                                  <input
                                    type="radio"
                                    :name="'st-'+st.id"
                                    :checked="answers[currentQuestion.id]?.values?.[st.id] === true"
                                    @change="() => setTrueFalseStatement(st.id, true)"
                                    class="h-6 w-6 accent-[#0B7EA1]"
                                  />
                               </td>
                               <td class="py-6 px-3 text-center align-middle border-2 border-slate-300">
                                  <input
                                    type="radio"
                                    :name="'st-'+st.id"
                                    :checked="answers[currentQuestion.id]?.values?.[st.id] === false"
                                    @change="() => setTrueFalseStatement(st.id, false)"
                                    class="h-6 w-6 accent-[#0B7EA1]"
                                  />
                               </td>
                            </tr>
                         </tbody>
                      </table>
                   </div>

                   <!-- Legacy Single Statement -->
                   <div v-else class="bg-white rounded-2xl border-2 border-slate-200 overflow-hidden shadow-sm">
                      <table class="w-full table-fixed border-collapse text-sm border-2 border-slate-300">
                         <colgroup>
                           <col class="w-[76%]" />
                           <col class="w-[12%]" />
                           <col class="w-[12%]" />
                         </colgroup>
                         <thead class="bg-slate-50/80 font-black text-[10px] text-slate-500 uppercase tracking-widest">
                            <tr>
                               <th class="py-4 px-6 text-left border-2 border-slate-300">Pernyataan / Pertanyaan</th>
                               <th class="py-4 px-3 text-center border-2 border-slate-300">Benar</th>
                               <th class="py-4 px-3 text-center border-2 border-slate-300">Salah</th>
                            </tr>
                         </thead>
                         <tbody>
                            <tr>
                               <td class="py-8 px-6 text-slate-500 italic font-semibold align-middle border-2 border-slate-300 leading-relaxed">
                                  Lihat stimulus / soal utama di atas.
                               </td>
                               <td class="py-8 px-3 text-center align-middle border-2 border-slate-300">
                                  <input
                                    type="radio"
                                    :name="'tf-legacy-'+currentQuestion.id"
                                    :checked="answers[currentQuestion.id]?.value === true"
                                    @change="() => setTrueFalseLegacy(true)"
                                    class="h-6 w-6 accent-[#0B7EA1]"
                                  />
                               </td>
                               <td class="py-8 px-3 text-center align-middle border-2 border-slate-300">
                                  <input
                                    type="radio"
                                    :name="'tf-legacy-'+currentQuestion.id"
                                    :checked="answers[currentQuestion.id]?.value === false"
                                    @change="() => setTrueFalseLegacy(false)"
                                    class="h-6 w-6 accent-[#0B7EA1]"
                                  />
                               </td>
                            </tr>
                         </tbody>
                      </table>
                   </div>
                </div>

                <!-- Matching -->
                <div v-else-if="currentQuestion.type === 'matching'" class="space-y-6">
                   <div v-if="!currentQuestion.matching_left?.length" class="p-10 bg-white rounded-2xl border-2 border-dashed border-slate-200 text-center text-slate-400 italic">
                      Tidak ada pasangan untuk soal menjodohkan ini.
                   </div>

                   <div v-else class="bg-white rounded-2xl border-2 border-slate-100 shadow-inner overflow-hidden">
                      <div class="px-8 py-6 border-b border-slate-100 bg-slate-50/70 flex items-center justify-between">
                         <div class="text-[10px] font-black uppercase tracking-[0.3em] text-slate-500">MENJODOHKAN</div>
                         <div class="text-[10px] font-black uppercase tracking-widest text-slate-400">Pilih pasangan yang sesuai</div>
                      </div>

                      <div class="p-8 space-y-4">
                         <div
                           v-for="left in currentQuestion.matching_left"
                           :key="left.id"
                           class="grid md:grid-cols-[1fr_24px_1fr] gap-4 items-center p-6 rounded-2xl border-2 border-slate-100 bg-white"
                         >
                            <div class="font-black text-slate-900 text-lg leading-snug" v-html="renderHtml(left.content)"></div>
                            <div class="text-center text-slate-300 font-black">→</div>

                            <div class="space-y-2">
                               <select
                                 class="w-full p-4 rounded-xl border-2 border-slate-100 bg-white focus:border-blue-500 outline-none transition-all font-bold text-slate-900"
                                 :value="answers[currentQuestion.id]?.pairs?.[left.id] || ''"
                                 @change="setMatchingPair(left.id, $event.target.value)"
                               >
                                  <option value="">Pilih pasangan...</option>
                                  <option
                                    v-for="opt in matchingRightOptions"
                                    :key="opt.id"
                                    :value="opt.id"
                                  >
                                    {{ opt.text || '(kosong)' }}
                                  </option>
                               </select>
                            </div>
                         </div>
                      </div>
                   </div>
                </div>

                <!-- Essay / Short Answer -->
                <div v-else-if="currentQuestion.type === 'short_answer'" class="space-y-6">
                   <div class="relative">
                      <input
                         v-model="currentTextAnswer"
                         class="w-full p-6 rounded-2xl border-2 border-slate-100 bg-white focus:border-blue-500 outline-none transition-all shadow-inner text-lg font-medium text-slate-900"
                         placeholder="Ketik jawaban isian singkat..."
                         @blur="saveAnswer(currentQuestion)"
                      />
                   </div>
                </div>

                <div v-else-if="currentQuestion.type === 'essay'" class="space-y-6">
                   <textarea
                      v-model="currentTextAnswer"
                      class="w-full p-10 rounded-3xl border-2 border-slate-100 bg-white focus:border-blue-500 outline-none transition-all shadow-inner text-xl font-medium text-slate-800"
                      rows="8"
                      placeholder="Masukkan jawaban uraian secara detail..."
                      @blur="saveAnswer(currentQuestion)"
                   ></textarea>
                </div>
                </div>
             </div>
          </div>

          <!-- RIGHT COLUMN: Sidebar -->
	          <aside class="hidden lg:flex bg-white rounded-lg border border-slate-200 shadow-sm overflow-hidden sticky top-[74px] max-h-[calc(100vh-96px)] flex-col">
	            <div class="px-6 py-4 border-b border-slate-200 bg-white shrink-0">
	              <span class="text-slate-800 font-bold uppercase text-sm tracking-wide select-none">DAFTAR SOAL</span>
	            </div>
	            <div class="p-5 overflow-auto">
	              <div class="grid grid-cols-5 gap-2">
	                <button
	                  v-for="(q, idx) in questions"
	                  :key="q.id"
	                  :data-qnav-idx="idx"
	                  type="button"
	                  class="h-10 w-10 flex items-center justify-center rounded border font-semibold text-sm transition-colors"
	                  :class="currentIndex === idx
	                    ? 'border-[#0B7EA1] ring-1 ring-[#0B7EA1] text-[#0B7EA1] bg-white'
	                    : (isFlagged(q.id)
	                      ? 'bg-[#F4C20D] border-[#F4C20D] text-white'
	                      : (isAnswered(q)
	                        ? 'bg-emerald-500 border-emerald-500 text-white'
	                        : 'bg-white border-slate-200 text-slate-600 hover:border-slate-300'))"
	                  @click="(e) => onNavClick(idx, e)"
	                >
	                  {{ idx + 1 }}
	                </button>
	              </div>

              <div class="mt-6 pt-4 border-t border-slate-200 grid grid-cols-2 gap-3 text-xs text-slate-600">
                <div class="flex items-center gap-2"><span class="h-3 w-3 bg-emerald-500 rounded-sm"></span>Sudah</div>
                <div class="flex items-center gap-2"><span class="h-3 w-3 bg-[#F4C20D] rounded-sm"></span>Ragu</div>
                <div class="flex items-center gap-2"><span class="h-3 w-3 border-2 border-[#0B7EA1] rounded-sm"></span>Dibuka</div>
                <div class="flex items-center gap-2"><span class="h-3 w-3 border border-slate-300 rounded-sm bg-white"></span>Belum</div>
              </div>

              <div class="mt-6 pt-6 border-t border-slate-200">
                <button
                  type="button"
                  class="w-full bg-[#0D47A1] text-white py-3 rounded font-black uppercase tracking-widest text-xs shadow-sm hover:brightness-95 active:scale-[0.99] transition-all"
                  @click="showSubmitModal = true"
                >
                  Selesai Ujian
                </button>
              </div>
            </div>
          </aside>

       </div>
    </main>
    </Transition>

    <div v-if="tokenVerified && sessionId && !submitDone && !isLoading && !errorMessage && !currentQuestion" class="max-w-xl mx-auto mt-20 p-12 bg-white rounded-3xl border-2 border-slate-100 shadow-2xl text-center">
      <BaseIcon :path="mdiAlert" size="64" class="text-amber-500 mb-6 mx-auto" />
      <h2 class="text-2xl font-black text-slate-800 mb-4 uppercase">Soal Tidak Ditemukan</h2>
      <p class="text-slate-600 mb-8 leading-relaxed">
        Index: {{ currentIndex + 1 }} / {{ questions.length }}. Klik tombol di bawah untuk kembali ke soal pertama.
      </p>
      <button
        @click="setIndex(0)"
        class="bg-[#0D47A1] text-white px-8 py-3 rounded-xl font-bold uppercase transition-all hover:scale-105 active:scale-95"
      >
        Reset Ke Soal 1
      </button>
    </div>

    <!-- NAVIGATION FOOTER -->
    <footer v-if="sessionId && !submitDone" class="bg-white border-t border-slate-200 p-4 fixed bottom-0 left-0 right-0 z-[100] shadow-[0_-12px_30px_-18px_rgba(0,0,0,0.3)]">
       <div class="max-w-[1500px] mx-auto w-full flex items-center justify-between">
          <button 
             class="bg-[#E74C3C] text-white px-8 py-2.5 rounded font-bold uppercase text-xs shadow-sm hover:brightness-95 active:scale-[0.99] transition-all disabled:opacity-50"
             :disabled="currentIndex === 0"
             @click="goPrev()"
          >
             Soal Sebelumnya
          </button>

          <button
            type="button"
            class="flex items-center gap-3 bg-[#F4C20D] text-white px-10 py-2.5 rounded font-bold uppercase text-xs shadow-sm hover:brightness-95 active:scale-[0.99] transition-all"
            @click="toggleFlagged"
          >
            <span class="h-4 w-4 rounded border border-white/70 bg-white/10 flex items-center justify-center">
              <span v-if="isFlagged(currentQuestion?.id)" class="text-[10px] font-black leading-none">✓</span>
            </span>
            Ragu-Ragu
          </button>

          <button 
             class="bg-[#0B7EA1] text-white px-8 py-2.5 rounded font-bold uppercase text-xs shadow-sm hover:brightness-95 active:scale-[0.99] transition-all"
             @click="currentIndex === questions.length - 1 ? (showSubmitModal = true) : goNext()"
          >
             {{ currentIndex === questions.length - 1 ? 'Selesai' : 'Soal Berikutnya' }}
          </button>
       </div>
    </footer>

    <!-- QUESTION LIST MODAL (Mobile) -->
    <div
      v-if="showQuestionListModal"
      class="fixed inset-0 bg-slate-900/60 backdrop-blur-md z-[180] flex items-center justify-center p-4 animate-fade-in"
      @click.self="showQuestionListModal = false"
    >
      <div class="bg-white rounded-2xl w-full max-w-lg shadow-2xl border border-slate-100 overflow-hidden max-h-[82vh] flex flex-col">
        <div class="px-5 py-4 border-b border-slate-200 flex items-center justify-between shrink-0">
          <div class="font-black uppercase tracking-widest text-xs text-slate-800">Daftar Soal</div>
          <button
            type="button"
            class="px-3 py-2 rounded bg-slate-100 hover:bg-slate-200 text-slate-700 font-black uppercase text-[11px] tracking-widest"
            @click="showQuestionListModal = false"
          >
            Tutup
          </button>
        </div>
        <div class="p-5 overflow-auto">
          <div class="grid grid-cols-6 gap-2">
            <button
              v-for="(q, idx) in questions"
              :key="q.id"
              type="button"
              class="h-10 flex items-center justify-center rounded border font-semibold text-sm transition-colors"
              :class="currentIndex === idx
                ? 'border-[#0B7EA1] ring-1 ring-[#0B7EA1] text-[#0B7EA1] bg-white'
                : (isFlagged(q.id)
                  ? 'bg-[#F4C20D] border-[#F4C20D] text-white'
                  : (isAnswered(q)
                    ? 'bg-emerald-500 border-emerald-500 text-white'
                    : 'bg-white border-slate-200 text-slate-600 hover:border-slate-300'))"
              @click="(e) => { onNavClick(idx, e); showQuestionListModal = false }"
            >
              {{ idx + 1 }}
            </button>
          </div>

          <div class="mt-5 pt-4 border-t border-slate-200 grid grid-cols-2 gap-3 text-xs text-slate-600">
            <div class="flex items-center gap-2"><span class="h-3 w-3 bg-emerald-500 rounded-sm"></span>Sudah</div>
            <div class="flex items-center gap-2"><span class="h-3 w-3 bg-[#F4C20D] rounded-sm"></span>Ragu</div>
            <div class="flex items-center gap-2"><span class="h-3 w-3 border-2 border-[#0B7EA1] rounded-sm"></span>Dibuka</div>
            <div class="flex items-center gap-2"><span class="h-3 w-3 border border-slate-300 rounded-sm bg-white"></span>Belum</div>
          </div>

          <div class="mt-5 pt-5 border-t border-slate-200">
            <button
              type="button"
              class="w-full bg-[#0D47A1] text-white py-3 rounded font-black uppercase tracking-widest text-xs shadow-sm hover:brightness-95 active:scale-[0.99] transition-all"
              @click="() => { showQuestionListModal = false; showSubmitModal = true }"
            >
              Selesai Ujian
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- SUBMIT MODAL -->
    <div v-if="showSubmitModal" class="fixed inset-0 bg-slate-900/60 backdrop-blur-md z-[200] flex items-center justify-center p-6 animate-fade-in shadow-2xl">
       <div class="bg-white rounded-3xl max-w-md w-full p-10 shadow-2xl border-2 border-blue-50">
          <div class="w-20 h-20 bg-blue-50 rounded-full flex items-center justify-center mx-auto mb-6 shadow-inner">
             <BaseIcon :path="mdiInformationOutline" size="48" class="text-[#0D47A1]" />
          </div>
          <h3 class="text-2xl font-black text-slate-800 text-center uppercase mb-4 tracking-tight">Konfirmasi Selesai</h3>
          <p class="text-slate-500 text-center mb-10 text-lg leading-relaxed">Apakah Anda yakin ingin mengakhiri ujian ini? Pastikan semua soal telah terjawab dengan benar.</p>
          <div class="flex flex-col gap-4">
             <button @click="submitExam" class="w-full bg-[#0D47A1] text-white py-4 rounded-xl font-black uppercase tracking-widest shadow-xl transition-all hover:scale-105 active:scale-95">YA, SAYA YAKIN</button>
             <button @click="showSubmitModal = false" class="w-full bg-slate-100 text-slate-600 py-4 rounded-xl font-black uppercase tracking-widest transition-all hover:bg-slate-200">TIDAK, KEMBALI</button>
          </div>
       </div>
    </div>
  </div>
</template>

<style scoped>
.tka-theme { font-family: system-ui, -apple-system, Segoe UI, Roboto, Arial, sans-serif; }
.tka-topbar { background: #0b7ea1; }
.tka-timer { box-shadow: inset 0 0 0 2px #f59e0b; }
.animate-fade-in { animation: fadeIn 0.5s cubic-bezier(0.16, 1, 0.3, 1); }
@keyframes fadeIn { from { opacity: 0; transform: translateY(30px) scale(0.98); } to { opacity: 1; transform: translateY(0) scale(1); } }
:deep(.prose) { max-width: none; }
:deep(.prose img) { border-radius: 20px; display: block; margin: 40px auto; box-shadow: 0 20px 25px -5px rgb(0 0 0 / 0.1); border: 8px solid #f8fafc; }
:deep(.prose h1) { font-weight: 900; color: #0D47A1; }
:deep(.prose table) { border-collapse: collapse; border: 2px solid #e2e8f0; width: 100%; border-radius: 12px; overflow: hidden; }
:deep(.prose th) { background: #f8fafc; padding: 12px; text-transform: uppercase; font-size: 0.75rem; letter-spacing: 0.1em; }
:deep(.prose td) { padding: 12px; border: 1px solid #f1f5f9; }

/* Font size presets for student exam view (applied to the scrollable card area). */
.tka-font-sm :deep(.prose) { font-size: 0.95rem; line-height: 1.65; }
.tka-font-md :deep(.prose) { font-size: 1rem; line-height: 1.65; }
.tka-font-lg :deep(.prose) { font-size: 1.125rem; line-height: 1.65; }

.tka-font-sm .tka-option-text { font-size: 0.95rem; }
.tka-font-md .tka-option-text { font-size: 1rem; }
.tka-font-lg .tka-option-text { font-size: 1.125rem; }

/* Question swap transition */
.qswap-enter-active,
.qswap-leave-active {
  transition: opacity 180ms cubic-bezier(0.16, 1, 0.3, 1), transform 220ms cubic-bezier(0.16, 1, 0.3, 1);
}
.qswap-enter-from {
  opacity: 0;
  transform: translateY(10px);
}
.qswap-leave-to {
  opacity: 0;
  transform: translateY(-10px);
}
@media (prefers-reduced-motion: reduce) {
  .qswap-enter-active,
  .qswap-leave-active {
    transition: none;
  }
}
</style>
