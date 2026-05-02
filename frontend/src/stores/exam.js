import { defineStore } from 'pinia'
import { ref, computed, watch } from 'vue'
import { api } from '@/services/api.js'
import { debounce } from '@/utils/debounce.js'

export const useExamStore = defineStore('exam', () => {
    const PERSIST_KEY = 'atigacbt:exam-store'
    const sessionId = ref(null)
    const startTime = ref(null)
    const examTitle = ref('AtigaCBT Workspace')
    const questions = ref([])
    const answers = ref({})
    const timeLeft = ref(0)
    const timerInterval = ref(null)
    const isLoading = ref(false)
    const errorMessage = ref('')
    const submitDone = ref(false)

    const currentQuestionIdx = ref(0)
    const currentQuestion = computed(() => questions.value[currentQuestionIdx.value])

    const persistNow = () => {
        try {
            localStorage.setItem(PERSIST_KEY, JSON.stringify({
                sessionId: sessionId.value,
                startTime: startTime.value,
                questions: questions.value,
                answers: answers.value,
                currentQuestionIdx: currentQuestionIdx.value,
                examTitle: examTitle.value,
                timeLeft: timeLeft.value,
            }))
        } catch {
            // ignore storage errors
        }
    }

    try {
        const raw = localStorage.getItem(PERSIST_KEY)
        if (raw) {
            const persisted = JSON.parse(raw)
            if (persisted && typeof persisted === 'object') {
                sessionId.value = persisted.sessionId ?? null
                startTime.value = persisted.startTime ?? null
                questions.value = Array.isArray(persisted.questions) ? persisted.questions : []
                answers.value = persisted.answers && typeof persisted.answers === 'object' ? persisted.answers : {}
                currentQuestionIdx.value = Number.isInteger(persisted.currentQuestionIdx) ? persisted.currentQuestionIdx : 0
                examTitle.value = persisted.examTitle || examTitle.value
                timeLeft.value = Number.isFinite(persisted.timeLeft) ? persisted.timeLeft : 0
            }
        }
    } catch {
        // ignore malformed persisted payload
    }

    const debouncedPersist = debounce(() => {
        persistNow()
    }, 500)

    watch(
        [sessionId, startTime, questions, answers, currentQuestionIdx, examTitle, timeLeft],
        debouncedPersist,
        { deep: true },
    )

    const startExam = ({ sessionId: sid, startTime: startedAt, questions: qs = [] } = {}) => {
        sessionId.value = sid || null
        startTime.value = startedAt || new Date().toISOString()
        questions.value = Array.isArray(qs) ? qs : []
        if (!answers.value || typeof answers.value !== 'object') {
            answers.value = {}
        }
        persistNow()
    }

    const loadExamData = async (sid) => {
        if (!sid) return
        sessionId.value = sid
        isLoading.value = true
        errorMessage.value = ''

        try {
            const [sessResp, questionsResp, answersResp] = await Promise.all([
                api.get(`/api/v1/student/sessions/${sid}`),
                api.get(`/api/v1/student/sessions/${sid}/questions`),
                api.get(`/api/v1/student/sessions/${sid}/answers`)
            ])

            const state = sessResp.data.data
            const questionsData = questionsResp.data.data

            if (state.exam?.title) {
                examTitle.value = state.exam.title
            }

            questions.value = Array.isArray(questionsData.questions) ? questionsData.questions : []

            // Initial defaults
            const processed = {}
            questions.value.forEach(q => {
                if (q.type === 'mc_single') processed[q.id] = { selected_option_id: '' }
                else if (q.type === 'mc_multiple') processed[q.id] = { selected_option_ids: [] }
                else if (q.type === 'true_false') processed[q.id] = q.statements?.length ? { values: {} } : { value: null }
                else if (q.type === 'matching') processed[q.id] = { pairs: {} }
                else if (q.type === 'essay' || q.type === 'short_answer') processed[q.id] = { text: '' }
                else processed[q.id] = {}
            })

            // Overlay with actual answers
            const existingAnswersList = Array.isArray(answersResp.data.data) ? answersResp.data.data : []
            existingAnswersList.forEach(ans => {
                const qid = ans.question_id
                if (!qid) return
                try {
                    processed[qid] = typeof ans.answer_json === 'string' ? JSON.parse(ans.answer_json) : ans.answer_json
                } catch {
                    processed[qid] = ans.answer_json
                }
            })

            answers.value = processed
            timeLeft.value = state.remaining_seconds || 0
            startTimer()
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
                stopTimer()
                submitExam()
            }
        }, 1000)
    }

    const stopTimer = () => {
        if (timerInterval.value) {
            clearInterval(timerInterval.value)
            timerInterval.value = null
        }
    }

    const isSaving = ref(false)
    const lastSavedAt = ref(null)

    const saveAnswer = async (questionId, answer = undefined) => {
        if (!sessionId.value || !questionId) return
        if (answer !== undefined) {
            answers.value[questionId] = answer
        }
        const payload = answers.value[questionId] ?? {}
        persistNow()

        isSaving.value = true
        try {
            await api.post(`/api/v1/student/sessions/${sessionId.value}/answers`, {
                question_id: questionId,
                answer_json: JSON.stringify(payload)
            })
            lastSavedAt.value = new Date()
        } catch (err) {
            console.error('Failed to save answer:', err)
        } finally {
            // Add a small delay so the 'Saving' state is visible for micro-interaction
            setTimeout(() => {
                isSaving.value = false
            }, 800)
        }
    }

    const submitExam = async () => {
        if (!sessionId.value) return
        try {
            isLoading.value = true
            await api.post(`/api/v1/student/sessions/${sessionId.value}/submit`)
            submitDone.value = true
            stopTimer()
        } catch (err) {
            throw err
        } finally {
            isLoading.value = false
        }
    }

    const reset = () => {
        stopTimer()
        sessionId.value = null
        startTime.value = null
        questions.value = []
        answers.value = {}
        timeLeft.value = 0
        submitDone.value = false
        errorMessage.value = ''
        currentQuestionIdx.value = 0
        isSaving.value = false
        lastSavedAt.value = null
        persistNow()
    }

    const finishExam = async () => {
        if (!sessionId.value) return
        await api.post(`/api/v1/student/sessions/${sessionId.value}/submit`)
        reset()
    }

    return {
        sessionId,
        startTime,
        examTitle,
        questions,
        answers,
        timeLeft,
        isLoading,
        isSaving,
        lastSavedAt,
        errorMessage,
        submitDone,
        currentQuestion,
        currentQuestionIdx,
        startExam,
        loadExamData,
        saveAnswer,
        submitExam,
        finishExam,
        reset,
        startTimer,
        stopTimer
    }
}, {
    persist: {
        key: 'atigacbt:exam-store',
        paths: ['sessionId', 'startTime', 'questions', 'answers', 'currentQuestionIdx', 'examTitle', 'timeLeft'],
    },
})
