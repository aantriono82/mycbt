<script setup>
import { computed, nextTick, onErrorCaptured, onMounted, ref, watch, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  mdiClockOutline,
  mdiCheckCircleOutline,
  mdiInformationOutline,
  mdiAlert
} from '@mdi/js'
import BaseIcon from '@/components/BaseIcon.vue'
import { api } from '@/services/api.js'

const route = useRoute()
const router = useRouter()

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
	const flagged = ref({})
const examTitle = ref('AtigaCBT Workspace')

const currentQuestion = computed(() => questions.value[currentIndex.value])

const cardScrollEl = ref(null)

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
      if (q.type === 'mc_multiple') {
        processed[q.id] = []
      } else {
        processed[q.id] = ''
      }
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
  try {
    const answer = answers.value[question.id]
    await api.post(`/api/v1/student/sessions/${sessionId.value}/answers`, {
      question_id: question.id,
      answer_json: typeof answer === 'object' ? JSON.stringify(answer) : String(answer)
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
const renderMath = (rootEl) => {
  if (!rootEl) return
  if (window.renderMathInElement) {
    window.renderMathInElement(rootEl, {
      delimiters: [
        { left: '$$', right: '$$', display: true },
        { left: '$', right: '$', display: false }
      ],
      throwOnError: false
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
   const ans = answers.value[q.id]
   if (ans === undefined || ans === null || ans === '') return false
   if (Array.isArray(ans)) return ans.length > 0
   if (typeof ans === 'object') return Object.keys(ans).length > 0
   return true
}

const goPrev = () => {
  if (currentIndex.value > 0) {
    currentIndex.value--
  }
}

const goNext = () => {
  if (questions.value.length > 0 && currentIndex.value < questions.value.length - 1) {
    currentIndex.value++
  }
}

const scrollToQuestion = (idx) => {
  if (idx >= 0 && idx < questions.value.length) {
    currentIndex.value = idx
  }
}

	const toggleMulti = (questionId, optId) => {
		const qid = String(questionId || '')
		if (!qid) return
		const id = String(optId || '')
		if (!id) return
		if (!Array.isArray(answers.value[qid])) answers.value[qid] = []
		const arr = answers.value[qid]
		const i = arr.indexOf(id)
		if (i >= 0) arr.splice(i, 1)
		else arr.push(id)
		saveAnswer(currentQuestion.value)
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
  scheduleRenderMath()
  if (cardScrollEl.value && typeof cardScrollEl.value.scrollTo === 'function') {
    cardScrollEl.value.scrollTo({ top: 0, behavior: 'smooth' })
  } else if (cardScrollEl.value) {
    cardScrollEl.value.scrollTop = 0
  }
  window.scrollTo({ top: 0, behavior: 'smooth' })
}, { flush: 'post' })

onMounted(() => {
  loadExamData()
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
          <div class="text-xs font-semibold">SISWA TERDAFTAR</div>
        </div>
        <div class="tka-timer bg-white text-slate-900 px-4 py-1.5 rounded-full font-semibold flex items-center gap-2">
          <span class="text-xs">SISA WAKTU:</span>
          <span class="font-mono">{{ formatTime(timeLeft) }}</span>
        </div>
      </div>
    </header>

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
    <main v-if="sessionId && !submitDone && currentQuestion" class="max-w-[1500px] mx-auto px-6 py-6 pb-24 animate-fade-in">
       <div class="grid lg:grid-cols-[1fr_340px] gap-6 items-start">
          
          <!-- LEFT COLUMN: Main Card -->
          <div class="bg-white rounded-lg border border-slate-200 shadow-sm min-h-[75vh] max-h-[calc(100vh-180px)] flex flex-col">
             <!-- Card Header -->
             <div class="px-6 py-4 border-b border-slate-200 flex items-center justify-between bg-white">
                <span class="text-[#0B7EA1] font-bold uppercase text-sm tracking-wide">SOAL NOMOR: {{ currentIndex + 1 }}</span>
                <div class="flex items-center gap-4 text-[11px] font-bold text-slate-500">
                   <span class="uppercase opacity-60">Ukuran Font:</span>
                   <button class="h-6 w-8 flex items-center justify-center border border-slate-300 rounded bg-slate-50 text-slate-700 font-bold leading-none">A</button>
                   <button class="h-7 w-9 flex items-center justify-center border-2 border-[#0B7EA1] rounded bg-white text-[#0B7EA1] font-black leading-none">A</button>
                   <button class="h-7 w-10 flex items-center justify-center border border-slate-300 rounded bg-slate-50 text-slate-700 font-bold leading-none text-base">A</button>
                </div>
             </div>

             <div ref="cardScrollEl" class="flex-1 overflow-auto">
                <!-- STIMULUS / CONTENT -->
                <div class="p-6 prose prose-slate max-w-none text-base leading-relaxed text-slate-900">
                   <div v-html="renderHtml(currentQuestion.stem)"></div>
                </div>

                <!-- INTERACTION / ANSWERS -->
                <div class="px-6 pb-10 pt-6 border-t border-slate-200 bg-white">
                <div class="mb-4 flex items-center justify-between">
                   <div class="text-xs font-semibold uppercase text-slate-400 tracking-widest">OPSI JAWABAN</div>
                   <div class="hidden sm:block text-xs text-slate-400">Klik untuk memilih jawaban</div>
                </div>

                <!-- MC Single -->
                <div v-if="currentQuestion.type === 'mc_single'" class="space-y-4">
                   <label
                     v-for="opt in currentQuestion.options" 
                     :key="opt.id"
                     class="flex items-center gap-4 p-4 rounded border transition-colors cursor-pointer bg-white text-left w-full"
                     :class="answers[currentQuestion.id] === String(opt.id) ? 'border-[#0B7EA1] bg-[#0B7EA1]/[0.03]' : 'border-slate-200 hover:border-slate-300'"
                   >
                      <input
                        type="radio"
                        :name="'mc-'+currentQuestion.id"
                        class="h-4 w-4 accent-slate-600"
                        :checked="answers[currentQuestion.id] === String(opt.id)"
                        @change="() => { answers[currentQuestion.id] = String(opt.id); saveAnswer(currentQuestion) }"
                      />
                      <div class="text-slate-800 font-medium text-sm leading-relaxed" v-html="renderHtml(opt.content)"></div>
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
                     :aria-pressed="answers[currentQuestion.id]?.includes(String(opt.id)) ? 'true' : 'false'"
                     :class="answers[currentQuestion.id]?.includes(String(opt.id)) ? 'border-[#0B7EA1] bg-[#0B7EA1]/[0.03]' : 'border-slate-200 hover:border-slate-300'"
                   >
                      <div class="pt-0.5">
                        <div
                          class="h-5 w-5 rounded border flex items-center justify-center"
                          :class="answers[currentQuestion.id]?.includes(String(opt.id)) ? 'bg-[#0B7EA1] border-[#0B7EA1] text-white' : 'bg-white border-slate-300 text-transparent'"
                        >
                          <span class="text-xs font-black leading-none">✓</span>
                        </div>
                      </div>
                      <div class="text-slate-800 font-medium text-sm leading-relaxed" v-html="renderHtml(opt.content)"></div>
                   </button>
                </div>

	                <!-- True/False (ANBK-style) + Legacy Fallback -->
	                <div v-else-if="currentQuestion.type === 'true_false'" class="overflow-hidden rounded border border-slate-300 bg-white">
	                   <table class="w-full table-fixed border-collapse text-sm border-2 border-slate-300">
	                      <colgroup>
	                        <col class="w-[76%]" />
	                        <col class="w-[12%]" />
	                        <col class="w-[12%]" />
	                      </colgroup>
	                      <thead class="bg-white font-semibold text-slate-800">
	                         <tr>
	                            <th class="py-3 px-4 text-left border-2 border-slate-300">Pernyataan / Pertanyaan</th>
	                            <th class="py-3 px-3 text-center border-2 border-slate-300">Benar</th>
	                            <th class="py-3 px-3 text-center border-2 border-slate-300">Salah</th>
	                         </tr>
	                      </thead>
	                      <tbody>
	                         <template v-if="currentQuestion.statements?.length">
	                           <tr v-for="st in currentQuestion.statements" :key="st.id">
	                              <td class="py-4 px-4 text-slate-800 align-top border-2 border-slate-300" v-html="renderHtml(st.content)"></td>
	                              <td class="py-4 px-3 text-center align-top border-2 border-slate-300">
	                                 <input
	                                   type="radio"
	                                   :name="'st-'+st.id"
	                                   :checked="answers[currentQuestion.id]?.[st.id] === true"
                                   @change="() => { if(!answers[currentQuestion.id] || Array.isArray(answers[currentQuestion.id])) answers[currentQuestion.id]={}; answers[currentQuestion.id][st.id]=true; saveAnswer(currentQuestion) }"
	                                   class="h-4 w-4 accent-slate-600"
	                                 />
	                              </td>
	                              <td class="py-4 px-3 text-center align-top border-2 border-slate-300">
	                                 <input
	                                   type="radio"
	                                   :name="'st-'+st.id"
	                                   :checked="answers[currentQuestion.id]?.[st.id] === false"
                                   @change="() => { if(!answers[currentQuestion.id] || Array.isArray(answers[currentQuestion.id])) answers[currentQuestion.id]={}; answers[currentQuestion.id][st.id]=false; saveAnswer(currentQuestion) }"
	                                   class="h-4 w-4 accent-slate-600"
	                                 />
	                              </td>
	                           </tr>
	                         </template>
	                         <tr v-else>
	                            <!-- Don't duplicate stimulus/stem here; T/F statements should come from `statements`. -->
	                            <td class="py-4 px-4 text-slate-600 align-top border-2 border-slate-300 italic">
	                              Lihat stimulus / soal utama di atas.
	                            </td>
	                            <td class="py-4 px-3 text-center align-top border-2 border-slate-300">
	                               <input
	                                 type="radio"
	                                 :name="'tf-legacy-'+currentQuestion.id"
                                 :checked="answers[currentQuestion.id] === true"
                                 @change="() => { answers[currentQuestion.id] = true; saveAnswer(currentQuestion) }"
	                                 class="h-4 w-4 accent-slate-600"
	                               />
	                            </td>
	                            <td class="py-4 px-3 text-center align-top border-2 border-slate-300">
	                               <input
	                                 type="radio"
	                                 :name="'tf-legacy-'+currentQuestion.id"
                                 :checked="answers[currentQuestion.id] === false"
                                 @change="() => { answers[currentQuestion.id] = false; saveAnswer(currentQuestion) }"
                                 class="h-4 w-4 accent-slate-600"
                               />
                            </td>
                         </tr>
                      </tbody>
                   </table>
                </div>

                <!-- Essay / Short Answer -->
                <div v-else-if="['essay', 'short_answer'].includes(currentQuestion.type)">
                   <textarea 
                      v-model="answers[currentQuestion.id]" 
                      class="w-full p-10 rounded-3xl border-2 border-slate-100 bg-white focus:border-blue-500 outline-none transition-all shadow-inner text-xl font-bold italic text-slate-800" 
                      rows="8" 
                      placeholder="Masukkan jawaban Anda secara detail di sini..."
                      @blur="saveAnswer(currentQuestion)"
                    ></textarea>
                </div>
                </div>
             </div>
          </div>

          <!-- RIGHT COLUMN: Sidebar -->
          <aside class="hidden lg:block bg-white rounded-lg border border-slate-200 shadow-sm overflow-hidden sticky top-[74px]">
             <div class="px-6 py-4 border-b border-slate-200 bg-white">
                <span class="text-slate-800 font-bold uppercase text-sm tracking-wide select-none">DAFTAR SOAL</span>
             </div>
             <div class="p-5">
                <div class="grid grid-cols-5 gap-2">
                   <button
                      v-for="(q, idx) in questions"
                      :key="q.id"
                      class="h-10 w-10 flex items-center justify-center rounded border font-semibold text-sm transition-colors"
                      :class="currentIndex === idx
                        ? 'border-[#0B7EA1] ring-1 ring-[#0B7EA1] text-[#0B7EA1] bg-white'
                        : (isFlagged(q.id)
                          ? 'bg-[#F4C20D] border-[#F4C20D] text-white'
                          : (isAnswered(q)
                            ? 'bg-emerald-500 border-emerald-500 text-white'
                            : 'bg-white border-slate-200 text-slate-600 hover:border-slate-300'))"
                      @click="scrollToQuestion(idx)"
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
             </div>
          </aside>

       </div>
    </main>

    <div v-else-if="sessionId && !submitDone && !isLoading && !errorMessage" class="max-w-xl mx-auto mt-20 p-12 bg-white rounded-3xl border-2 border-slate-100 shadow-2xl text-center">
      <BaseIcon :path="mdiAlert" size="64" class="text-amber-500 mb-6 mx-auto" />
      <h2 class="text-2xl font-black text-slate-800 mb-4 uppercase">Soal Tidak Ditemukan</h2>
      <p class="text-slate-600 mb-8 leading-relaxed">
        Index: {{ currentIndex + 1 }} / {{ questions.length }}. Klik tombol di bawah untuk kembali ke soal pertama.
      </p>
      <button
        @click="scrollToQuestion(0)"
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
</style>
