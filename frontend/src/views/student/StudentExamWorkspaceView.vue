<script setup>
import { computed, nextTick, onErrorCaptured, onMounted, ref, watch, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  mdiCheckCircleOutline,
  mdiInformationOutline,
  mdiAlert,
  mdiFullscreen,
  mdiFullscreenExit
} from '@mdi/js'
import BaseIcon from '@/components/BaseIcon.vue'
import BaseButton from '@/components/BaseButton.vue'
import QuillEditor from '@/components/QuillEditor.vue'
import { api } from '@/services/api.js'
import { useAuthStore } from '@/stores/auth.js'
import { useExamStore } from '@/stores/exam.js'
import { storeToRefs } from 'pinia'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const examStore = useExamStore()

const { 
  sessionId, 
  questions, 
  answers, 
  timeLeft, 
  isLoading, 
  errorMessage, 
  submitDone, 
  examTitle, 
  currentQuestionIdx: currentIndex, 
  currentQuestion 
} = storeToRefs(examStore)

const showSubmitModal = ref(false)
const showQuestionListModal = ref(false)
const flagged = ref({})
const participantName = computed(() => authStore.userDisplayName)

const isFullscreen = ref(false)
const sidebarHidden = ref(false)
const toggleFullscreen = () => {
  if (!document.fullscreenElement) {
    document.documentElement.requestFullscreen().catch(err => {
      console.error(`Error attempting to enable full-screen mode: ${err.message}`)
    })
  } else {
    document.exitFullscreen()
  }
}

const handleFullscreenChange = () => {
  isFullscreen.value = !!document.fullscreenElement
}

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
  if (!route.params.sessionId) return
  tokenError.value = ''
  const token = String(tokenValue.value || '').trim()
  if (!token) {
    tokenError.value = 'Token wajib diisi.'
    return
  }
  tokenChecking.value = true
  try {
    await api.post(`/api/v1/student/sessions/${route.params.sessionId}/verify-token`, { token })
    tokenVerified.value = true
    try {
      localStorage.setItem(TOKEN_OK_KEY.value, '1')
    } catch {
      // ignore
    }
    await examStore.loadExamData(route.params.sessionId)
    scheduleRenderMath()
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
    if (cur && typeof cur === 'object' && !Array.isArray(cur)) {
      answers.value[qid] = { selected_option_id: String(cur.selected_option_id || '') }
      return
    }
    answers.value[qid] = { selected_option_id: cur ? String(cur) : '' }
    return
  }

  if (t === 'mc_multiple') {
    if (cur && typeof cur === 'object' && !Array.isArray(cur)) {
      const arr = Array.isArray(cur.selected_option_ids) ? cur.selected_option_ids : []
      answers.value[qid] = { selected_option_ids: arr.map(v => String(v)) }
      return
    }
    if (Array.isArray(cur)) {
      answers.value[qid] = { selected_option_ids: cur.map(v => String(v)) }
      return
    }
    answers.value[qid] = { selected_option_ids: cur ? [String(cur)] : [] }
    return
  }

  if (t === 'true_false') {
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
    answers.value[qid] = { value: null }
    return
  }

  if (t === 'matching') {
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
    if (cur && typeof cur === 'object' && !Array.isArray(cur)) {
      answers.value[qid] = { text: String(cur.text || '') }
      return
    }
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
      if (el.tagName === 'IMG') el.setAttribute('loading', 'lazy')
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

const submitExam = async () => {
  try {
    await examStore.submitExam()
    showSubmitModal.value = false
    await router.push('/student/hasil')
  } catch (err) {
    alert(err.response?.data?.error?.message || 'Gagal mengirim jawaban')
  }
}

let mathRaf = 0
const fixCommonLatexCommands = (input) => {
  const s = String(input || '')
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

const answeredQuestionMap = computed(() => {
  const map = {}
  for (const q of questions.value || []) {
    const id = String(q?.id ?? '')
    if (!id) continue
    const t = String(q?.type ?? '')
    const ans = answers.value[id]
    let answered = false
    if (ans) {
      if (t === 'mc_single') answered = !!String(ans?.selected_option_id || '').trim()
      else if (t === 'mc_multiple') answered = Array.isArray(ans?.selected_option_ids) && ans.selected_option_ids.length > 0
      else if (t === 'true_false') {
        answered = q?.statements?.length ? !!(ans?.values && Object.keys(ans.values).length > 0) : typeof ans?.value === 'boolean'
      } else if (t === 'short_answer' || t === 'essay') answered = !!String(ans?.text || '').trim()
      else if (t === 'matching') answered = !!(ans?.pairs && Object.keys(ans.pairs).length > 0)
      else answered = true
    }
    map[id] = answered
  }
  return map
})

const unansweredCount = computed(() => {
  const total = Array.isArray(questions.value) ? questions.value.length : 0
  if (total <= 0) return 0
  let answered = 0
  for (const q of questions.value) {
    const id = String(q?.id ?? '')
    if (!id) continue
    if (answeredQuestionMap.value[id]) answered += 1
  }
  return Math.max(0, total - answered)
})

const navStatusById = computed(() => {
  const out = {}
  for (const [idx, q] of (questions.value || []).entries()) {
    const id = String(q?.id ?? '')
    if (!id) continue
    if (idx === currentIndex.value) out[id] = 'current'
    else if (flagged.value[id]) out[id] = 'flagged'
    else if (answeredQuestionMap.value[id]) out[id] = 'answered'
    else out[id] = 'unanswered'
  }
  return out
})

const navButtonClass = (q, idx) => {
  const id = String(q?.id ?? '')
  const status = navStatusById.value[id] || (idx === currentIndex.value ? 'current' : 'unanswered')
  if (status === 'current') return 'border-blue-600 ring-2 ring-blue-500/20 text-blue-700 bg-blue-50 shadow-sm'
  if (status === 'flagged') return 'bg-amber-400 border-amber-400 text-white shadow-md shadow-amber-500/20'
  if (status === 'answered') return 'bg-emerald-500 border-emerald-500 text-white'
  return 'bg-white dark:bg-slate-800 border-slate-100 dark:border-slate-800 text-slate-400 hover:border-blue-300 dark:hover:border-blue-800 transition-all font-medium'
}

const goPrev = () => setIndex(currentIndex.value - 1)
const goNext = () => setIndex(currentIndex.value + 1)
const setIndex = (idx) => {
  const n = Number(idx)
  if (!Number.isFinite(n) || n < 0 || n >= questions.value.length) return
  currentIndex.value = n
}

const onNavClick = (idx, ev) => {
  if (ev?.preventDefault) ev.preventDefault()
  if (ev?.stopPropagation) ev.stopPropagation()
  setIndex(idx)
}

const onNavClickAndClose = (idx, ev) => {
  onNavClick(idx, ev)
  showQuestionListModal.value = false
}

const saveAnswer = (q) => {
  const qid = String(q?.id ?? '')
  if (!qid) return
  ensureAnswerShapeForQuestion(q)
  examStore.saveAnswer(qid)
}

const toggleMulti = (questionId, optId) => {
  const qid = String(questionId || '')
  const id = String(optId || '')
  if (!qid || !id) return
  ensureAnswerShapeForQuestion({ id: qid, type: 'mc_multiple' })
  const cur = answers.value[qid]
  const arr = Array.isArray(cur?.selected_option_ids) ? [...cur.selected_option_ids] : []
  const i = arr.indexOf(id)
  if (i >= 0) arr.splice(i, 1)
  else arr.push(id)
  answers.value[qid].selected_option_ids = arr
  examStore.saveAnswer(qid)
}

const setTrueFalseLegacy = (val) => {
  const q = currentQuestion.value
  if (!q?.id) return
  answers.value[q.id] = { value: !!val }
  examStore.saveAnswer(q.id)
}

const setTrueFalseStatement = (statementId, val) => {
  const q = currentQuestion.value
  if (!q?.id) return
  ensureAnswerShapeForQuestion(q)
  const cur = answers.value[q.id]
  const values = cur?.values && typeof cur.values === 'object' && !Array.isArray(cur.values) ? { ...cur.values } : {}
  values[String(statementId)] = !!val
  answers.value[q.id].values = values
  examStore.saveAnswer(q.id)
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
  answers.value[q.id].pairs = pairs
  examStore.saveAnswer(q.id)
}

const isFlagged = (qid) => !!flagged.value[String(qid || '')]
const toggleFlagged = () => {
  const q = currentQuestion.value
  if (!q?.id) return
  flagged.value[q.id] = !flagged.value[q.id]
}

watch(currentIndex, () => {
  const q = questions.value?.[currentIndex.value]
  if (q) ensureAnswerShapeForQuestion(q)
  syncEditorModelByQuestion()
  scheduleRenderMath()
  if (cardScrollEl.value) cardScrollEl.value.scrollTop = 0
  window.scrollTo({ top: 0, behavior: 'smooth' })
  setTimeout(() => {
    const el = document.querySelector(`[data-qnav-idx="${currentIndex.value}"]`)
    if (el) el.scrollIntoView({ block: 'nearest', inline: 'nearest' })
  }, 50)
}, { flush: 'post' })

watch(
  () => currentQuestion.value?.id,
  () => {
    syncEditorModelByQuestion()
  },
  { immediate: true },
)

onMounted(() => {
  const sid = route.params.sessionId
  if (sid) {
    try {
      if (localStorage.getItem(TOKEN_OK_KEY.value) === '1') {
        tokenVerified.value = true
        examStore.loadExamData(sid).then(() => scheduleRenderMath())
      }
    } catch {}
  }
  document.addEventListener('visibilitychange', handleVisibilityChange)
  document.addEventListener('fullscreenchange', handleFullscreenChange)
})

onUnmounted(() => {
  examStore.stopTimer()
  if (mathRaf) cancelAnimationFrame(mathRaf)
  if (textAnswerSaveTimer) clearTimeout(textAnswerSaveTimer)
  document.removeEventListener('visibilitychange', handleVisibilityChange)
  document.removeEventListener('fullscreenchange', handleFullscreenChange)
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
  get: () => {
    const q = currentQuestion.value
    if (!q?.id) return ''
    ensureAnswerShapeForQuestion(q)
    return String(answers.value[q.id]?.text || '')
  },
  set: (v) => {
    const q = currentQuestion.value
    if (!q?.id) return
    ensureAnswerShapeForQuestion(q)
    answers.value[q.id].text = String(v ?? '')
  },
})

const stripHtml = (html) => String(html || '').replace(/<[^>]*>/g, '').trim()
const shortAnswerEditorHtml = ref('')
const essayEditorHtml = ref('')
let textAnswerSaveTimer = 0

const textToQuillHtml = (text) => {
  const raw = String(text || '')
  if (!raw.trim()) return ''
  return `<p>${raw.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;')}</p>`
}

const quillHtmlToPlainText = (html) => {
  const raw = String(html || '')
  if (!raw.trim()) return ''
  try {
    const doc = new DOMParser().parseFromString(raw, 'text/html')
    doc.querySelectorAll('.ql-formula[data-value]').forEach((node) => {
      const latex = String(node.getAttribute('data-value') || '').trim()
      node.replaceWith(doc.createTextNode(latex))
    })
    const text = String(doc.body.textContent || '')
    return text.replace(/\s+/g, ' ').trim()
  } catch {
    return stripHtml(raw)
  }
}

const syncEditorModelByQuestion = () => {
  const q = currentQuestion.value
  if (!q?.id) {
    shortAnswerEditorHtml.value = ''
    essayEditorHtml.value = ''
    return
  }
  ensureAnswerShapeForQuestion(q)
  const current = String(answers.value[q.id]?.text || '')
  if (q.type === 'short_answer') {
    shortAnswerEditorHtml.value = textToQuillHtml(current)
    essayEditorHtml.value = ''
    return
  }
  if (q.type === 'essay') {
    essayEditorHtml.value = current
    shortAnswerEditorHtml.value = ''
    return
  }
  shortAnswerEditorHtml.value = ''
  essayEditorHtml.value = ''
}

const queueSaveCurrentTextAnswer = () => {
  if (textAnswerSaveTimer) clearTimeout(textAnswerSaveTimer)
  textAnswerSaveTimer = setTimeout(() => {
    if (!currentQuestion.value?.id) return
    saveAnswer(currentQuestion.value)
  }, 450)
}

const onShortAnswerEditorUpdate = (html) => {
  shortAnswerEditorHtml.value = String(html || '')
  currentTextAnswer.value = quillHtmlToPlainText(shortAnswerEditorHtml.value)
  queueSaveCurrentTextAnswer()
}

const onEssayEditorUpdate = (html) => {
  essayEditorHtml.value = String(html || '')
  currentTextAnswer.value = essayEditorHtml.value
  queueSaveCurrentTextAnswer()
}

const matchingRightOptions = computed(() => {
  const q = currentQuestion.value
  if (!q || q.type !== 'matching') return []
  return (q.matching_right || []).map(it => ({
    id: String(it.id || ''),
    text: stripHtml(it.content || ''),
    orderNo: it.order_no ?? null,
  }))
})

</script>

<template>
  <div class="tka-theme min-h-screen bg-slate-100">
    <!-- TOP NAVBAR -->
    <header class="bg-blue-600 dark:bg-slate-900 border-b border-blue-700 dark:border-slate-800 text-white px-6 py-4 flex items-center justify-between shadow-lg sticky top-0 z-50">
      <div class="flex items-center gap-3">
         <div class="p-2 bg-white/10 rounded-lg hidden sm:block">
            <BaseIcon :path="mdiInformationOutline" size="20" />
         </div>
         <div class="font-black tracking-tighter text-sm md:text-lg uppercase select-none">
           {{ examTitle }}
         </div>
      </div>
      <div class="flex items-center gap-3 md:gap-6">
        <div class="hidden md:block text-right leading-tight">
          <div class="text-[10px] font-black uppercase opacity-70 tracking-widest">Peserta</div>
          <div class="text-xs font-bold">{{ participantName }}</div>
        </div>
        <div :class="[
          'px-4 py-2 rounded-2xl font-black flex items-center gap-2 transition-all duration-500 shadow-inner',
          timeLeft < 300 ? 'bg-red-500 text-white animate-pulse' : 'bg-blue-700/50 text-white border border-blue-400/20'
        ]">
          <span class="text-[10px] hidden sm:inline tracking-widest">SISA WAKTU:</span>
          <span class="font-mono text-sm md:text-base">{{ formatTime(timeLeft) }}</span>
        </div>
        <button
          type="button"
          class="hidden sm:flex items-center justify-center p-2.5 rounded-xl bg-white/10 hover:bg-white/20 transition-all active:scale-95"
          title="Toggle Fullscreen"
          @click="toggleFullscreen"
        >
          <BaseIcon :path="isFullscreen ? mdiFullscreenExit : mdiFullscreen" size="20" />
        </button>
        <button
          type="button"
          class="hidden lg:flex items-center justify-center p-2.5 rounded-xl bg-white/10 hover:bg-white/20 transition-all active:scale-95"
          :title="sidebarHidden ? 'Show Sidebar' : 'Hide Sidebar (Focus Mode)'"
          @click="sidebarHidden = !sidebarHidden"
        >
          <BaseIcon :path="mdiViewGridOutline" size="20" :class="{ 'opacity-50': sidebarHidden }" />
        </button>
        <button
          type="button"
          class="lg:hidden bg-white/15 border border-white/20 hover:bg-white/25 active:scale-95 px-4 py-2.5 rounded-xl font-black uppercase text-[10px] tracking-widest transition-all"
          @click="showQuestionListModal = true"
        >
          Menu
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
    <main v-if="sessionId && !submitDone && currentQuestion" :key="currentQuestion.id" class="max-w-[1600px] mx-auto px-4 md:px-6 py-6 pb-28">
       <div class="flex flex-col lg:flex-row gap-6 items-start">
          
          <!-- LEFT COLUMN: Main Card -->
	          <div :class="['bg-white dark:bg-slate-900 rounded-[2rem] border border-slate-200 dark:border-slate-800 shadow-xl flex-1 flex flex-col transition-all duration-500', sidebarHidden ? 'w-full' : 'lg:w-3/4']">
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

		               <div ref="cardScrollEl" class="flex-1 overflow-auto scrollbar-styled-light dark:scrollbar-styled-dark" :class="fontClass">
	                <!-- STIMULUS / CONTENT -->
	                <div class="p-6 md:p-10 prose prose-slate dark:prose-invert max-w-none text-base leading-relaxed text-slate-800 dark:text-slate-200">
	                   <div class="flex justify-between items-start mb-6">
	                      <div class="px-3 py-1 bg-slate-100 dark:bg-slate-800 rounded-lg text-[10px] font-black uppercase text-slate-500 tracking-[0.2em]">STIMULUS / SOAL UTAMA</div>
	                   </div>
	                   <div class="tka-question-content" v-html="renderHtml(currentQuestion.stem)"></div>
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
                <div v-if="currentQuestion.type === 'mc_single'" class="grid grid-cols-1 gap-3">
                   <label
                     v-for="(opt, idx) in currentQuestion.options" 
                     :key="opt.id"
                     class="group flex items-center gap-5 p-5 rounded-2xl border-2 transition-all cursor-pointer bg-white dark:bg-slate-800 text-left w-full hover:shadow-md active:scale-[0.99]"
                     :class="answers[currentQuestion.id]?.selected_option_id === String(opt.id) ? 'border-blue-600 bg-blue-50/50 dark:bg-blue-900/20' : 'border-slate-100 dark:border-slate-800 hover:border-blue-200'"
                   >
                      <div class="flex-none h-8 w-8 rounded-xl border-2 flex items-center justify-center font-black text-sm transition-all"
                        :class="answers[currentQuestion.id]?.selected_option_id === String(opt.id) ? 'bg-blue-600 border-blue-600 text-white' : 'bg-slate-50 dark:bg-slate-900 border-slate-200 dark:border-slate-700 text-slate-500 group-hover:border-blue-400 group-hover:text-blue-600'"
                      >
                         {{ String.fromCharCode(65 + idx) }}
                      </div>
                      <input
                        type="radio"
                        :name="'mc-'+currentQuestion.id"
                        class="hidden"
                        :checked="answers[currentQuestion.id]?.selected_option_id === String(opt.id)"
                        @change="() => { answers[currentQuestion.id] = { selected_option_id: String(opt.id) }; saveAnswer(currentQuestion) }"
                      />
	                      <div class="grow text-slate-700 dark:text-slate-300 font-medium text-sm md:text-base leading-relaxed tka-option-text" v-html="renderHtml(opt.content)"></div>
	                   </label>
	                </div>

                <!-- MC Multiple (PG Kompleks) -->
                <div v-else-if="currentQuestion.type === 'mc_multiple'" class="grid grid-cols-1 gap-3">
                   <button
                     type="button"
                     v-for="(opt, idx) in currentQuestion.options" 
                     :key="opt.id"
                     @click="toggleMulti(currentQuestion.id, opt.id)"
                     class="group flex items-center gap-5 p-5 rounded-2xl border-2 transition-all cursor-pointer bg-white dark:bg-slate-800 text-left w-full hover:shadow-md active:scale-[0.99]"
                     :aria-pressed="answers[currentQuestion.id]?.selected_option_ids?.includes(String(opt.id)) ? 'true' : 'false'"
                     :class="answers[currentQuestion.id]?.selected_option_ids?.includes(String(opt.id)) ? 'border-blue-600 bg-blue-50/50 dark:bg-blue-900/20' : 'border-slate-100 dark:border-slate-800 hover:border-blue-200'"
                   >
                      <div class="flex-none h-8 w-8 rounded-xl border-2 flex items-center justify-center font-black text-sm transition-all"
                        :class="answers[currentQuestion.id]?.selected_option_ids?.includes(String(opt.id)) ? 'bg-blue-600 border-blue-600 text-white shadow-lg shadow-blue-500/30' : 'bg-slate-50 dark:bg-slate-900 border-slate-200 dark:border-slate-700 text-slate-500 group-hover:border-blue-400 group-hover:text-blue-600'"
                      >
                         {{ String.fromCharCode(65 + idx) }}
                      </div>
	                      <div class="grow text-slate-700 dark:text-slate-300 font-medium text-sm md:text-base leading-relaxed tka-option-text" v-html="renderHtml(opt.content)"></div>
                       <div class="flex-none">
                          <div
                            class="h-5 w-5 rounded-lg border-2 flex items-center justify-center transition-all"
                            :class="answers[currentQuestion.id]?.selected_option_ids?.includes(String(opt.id)) ? 'bg-blue-600 border-blue-600 text-white' : 'border-slate-200 dark:border-slate-700'"
                          >
                            <span v-if="answers[currentQuestion.id]?.selected_option_ids?.includes(String(opt.id))" class="text-[10px] font-black leading-none">✓</span>
                          </div>
                      </div>
	                   </button>
	                </div>

                <!-- True/False (ANBK-style) + Legacy Fallback -->
                <div v-else-if="currentQuestion.type === 'true_false'" class="space-y-4">
                   <div v-if="currentQuestion.statements?.length" class="bg-white dark:bg-slate-900 rounded-[2rem] border border-slate-100 dark:border-slate-800 overflow-hidden shadow-xl">
                      <table class="w-full table-fixed border-collapse text-sm">
                         <colgroup>
                           <col class="w-[70%]" />
                           <col class="w-[15%]" />
                           <col class="w-[15%]" />
                         </colgroup>
                         <thead class="bg-slate-50/80 dark:bg-slate-800/80 font-black text-[10px] text-slate-500 uppercase tracking-widest">
                            <tr>
                               <th class="py-5 px-6 text-left border-b border-slate-100 dark:border-slate-800">Pernyataan / Pertanyaan</th>
                               <th class="py-5 px-3 text-center border-b border-slate-100 dark:border-slate-800">Benar</th>
                               <th class="py-5 px-3 text-center border-b border-slate-100 dark:border-slate-800">Salah</th>
                            </tr>
                         </thead>
                          <tbody class="divide-y divide-slate-50 dark:divide-slate-800">
                             <tr v-for="st in currentQuestion.statements" :key="st.id" class="group hover:bg-blue-50/30 transition-colors">
                                <td class="py-6 px-6 text-slate-700 dark:text-slate-300 font-medium align-middle leading-relaxed" v-html="renderHtml(st.content)"></td>
                                <td class="py-6 px-3 text-center align-middle">
                                   <div class="flex justify-center">
                                     <input
                                       type="radio"
                                       :name="'st-'+st.id"
                                       :checked="answers[currentQuestion.id]?.values?.[st.id] === true"
                                       @change="() => setTrueFalseStatement(st.id, true)"
                                       class="h-6 w-6 cursor-pointer accent-blue-600"
                                     />
                                   </div>
                                </td>
                                <td class="py-6 px-3 text-center align-middle">
                                   <div class="flex justify-center">
                                     <input
                                       type="radio"
                                       :name="'st-'+st.id"
                                       :checked="answers[currentQuestion.id]?.values?.[st.id] === false"
                                       @change="() => setTrueFalseStatement(st.id, false)"
                                       class="h-6 w-6 cursor-pointer accent-red-500"
                                     />
                                   </div>
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
                  <QuillEditor
                    v-model="shortAnswerEditorHtml"
                    :height="120"
                    :enable-math="true"
                    placeholder="Tulis jawaban isian singkat (bisa notasi matematika)..."
                    @update:model-value="onShortAnswerEditorUpdate"
                    @blur="queueSaveCurrentTextAnswer"
                  />
                  <div class="text-[11px] font-semibold text-slate-500">
                    Gunakan tombol formula (`fx`) untuk notasi matematika. Jawaban isian singkat tetap disimpan sebagai teks agar auto-scoring tetap valid.
                  </div>
                </div>

                <div v-else-if="currentQuestion.type === 'essay'" class="space-y-6">
                  <QuillEditor
                    v-model="essayEditorHtml"
                    :height="260"
                    :enable-math="true"
                    placeholder="Masukkan jawaban uraian secara detail (termasuk notasi matematika)..."
                    @update:model-value="onEssayEditorUpdate"
                    @blur="queueSaveCurrentTextAnswer"
                  />
                </div>
                </div>
             </div>
          </div>

          <!-- RIGHT COLUMN: Sidebar -->
	          <aside v-if="!sidebarHidden" class="hidden lg:flex bg-white dark:bg-slate-900 rounded-[2rem] border border-slate-200 dark:border-slate-800 shadow-xl overflow-hidden sticky top-[90px] max-h-[calc(100vh-120px)] flex-col w-[340px] animate-fade-in">
	            <div class="px-8 py-5 border-b border-slate-100 dark:border-slate-800 bg-white dark:bg-slate-900 shrink-0">
	              <span class="text-slate-800 dark:text-slate-200 font-black uppercase text-xs tracking-[0.2em] select-none">Navigasi Soal</span>
	            </div>
	            <div class="p-6 overflow-auto scrollbar-styled-light dark:scrollbar-styled-dark">
	              <div class="grid grid-cols-5 gap-3">
	                <button
	                  v-for="(q, idx) in questions"
	                  :key="q.id"
	                  :data-qnav-idx="idx"
	                  type="button"
	                  class="h-10 w-10 flex items-center justify-center rounded-xl border-2 font-bold text-xs transition-all active:scale-90"
	                  :class="navButtonClass(q, idx)"
	                  @click="(e) => onNavClick(idx, e)"
	                >
	                  {{ idx + 1 }}
	                </button>
	              </div>

              <div class="mt-8 pt-6 border-t border-slate-100 dark:border-slate-800 space-y-3">
                <div class="flex items-center gap-3 text-[10px] font-black uppercase tracking-widest text-slate-400">
                   <span class="h-3 w-3 bg-emerald-500 rounded-full"></span> Terjawab
                </div>
                <div class="flex items-center gap-3 text-[10px] font-black uppercase tracking-widest text-slate-400">
                   <span class="h-3 w-3 bg-amber-400 rounded-full"></span> Ragu-Ragu
                </div>
                <div class="flex items-center gap-3 text-[10px] font-black uppercase tracking-widest text-slate-400">
                   <span class="h-3 w-3 border-2 border-blue-600 rounded-full"></span> Aktif
                </div>
                 <div class="flex items-center gap-3 text-[10px] font-black uppercase tracking-widest text-slate-400">
                   <span class="h-3 w-3 border-2 border-slate-200 dark:border-slate-700 rounded-full"></span> Belum
                </div>
              </div>

              <div class="mt-8">
                <button
                  type="button"
                  class="w-full bg-blue-600 hover:bg-blue-700 text-white py-4 rounded-2xl font-black uppercase tracking-widest text-[10px] shadow-lg shadow-blue-500/20 active:scale-95 transition-all"
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
    <footer v-if="sessionId && !submitDone" class="bg-white/80 dark:bg-slate-900/80 backdrop-blur-md border-t border-slate-200 dark:border-slate-800 p-5 px-6 fixed bottom-0 left-0 right-0 z-[100] shadow-[0_-12px_40px_-15px_rgba(0,0,0,0.1)]">
       <div class="max-w-[1600px] mx-auto w-full flex items-center justify-between">
          <button 
             class="group flex items-center gap-2 bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-300 px-8 py-3 rounded-2xl font-black uppercase text-[10px] tracking-widest shadow-sm hover:bg-slate-200 dark:hover:bg-slate-700 active:scale-95 transition-all disabled:opacity-30 disabled:cursor-not-allowed"
             :disabled="currentIndex === 0"
             @click="goPrev()"
          >
             <span class="group-hover:-translate-x-1 transition-transform">←</span>
             Soal Sebelumnya
          </button>

          <button
            type="button"
            class="hidden sm:flex items-center gap-3 bg-amber-50 dark:bg-amber-900/20 text-amber-600 dark:text-amber-400 px-10 py-3 rounded-2xl font-black uppercase text-[10px] tracking-widest border border-amber-200 dark:border-amber-700/50 hover:bg-amber-100 transition-all active:scale-95"
            @click="toggleFlagged"
          >
            <div class="h-5 w-5 rounded-lg border-2 flex items-center justify-center transition-all"
              :class="isFlagged(currentQuestion?.id) ? 'bg-amber-400 border-amber-400 text-white' : 'border-amber-400/50'"
            >
              <span v-if="isFlagged(currentQuestion?.id)" class="text-[10px] font-black leading-none">✓</span>
            </div>
            Ragu-Ragu
          </button>

          <button 
             class="group flex items-center gap-2 bg-blue-600 text-white px-8 py-3 rounded-2xl font-black uppercase text-[10px] tracking-widest shadow-lg shadow-blue-500/20 hover:bg-blue-700 active:scale-95 transition-all"
             @click="currentIndex === questions.length - 1 ? (showSubmitModal = true) : goNext()"
          >
             {{ currentIndex === questions.length - 1 ? 'Selesai' : 'Soal Berikutnya' }}
             <span v-if="currentIndex !== questions.length - 1" class="group-hover:translate-x-1 transition-transform">→</span>
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
              :class="navButtonClass(q, idx)"
              @click="(e) => onNavClickAndClose(idx, e)"
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
          <p
            v-if="unansweredCount > 0"
            class="mb-6 rounded-xl border border-amber-200 bg-amber-50 px-4 py-3 text-center text-sm font-bold text-amber-700"
          >
            Masih ada {{ unansweredCount }} soal belum dijawab.
          </p>
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
