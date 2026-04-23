import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { api } from '@/services/api.js'

export const useExamStore = defineStore('exam', () => {
    const sessionId = ref(null)
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

    const saveAnswer = async (questionId) => {
        if (!sessionId.value || !questionId) return
        const answer = answers.value[questionId] ?? {}
        try {
            await api.post(`/api/v1/student/sessions/${sessionId.value}/answers`, {
                question_id: questionId,
                answer_json: JSON.stringify(answer)
            })
        } catch (err) {
            console.error('Failed to save answer:', err)
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
        questions.value = []
        answers.value = {}
        timeLeft.value = 0
        submitDone.value = false
        errorMessage.value = ''
        currentQuestionIdx.value = 0
    }

    return {
        sessionId,
        examTitle,
        questions,
        answers,
        timeLeft,
        isLoading,
        errorMessage,
        submitDone,
        currentQuestion,
        currentQuestionIdx,
        loadExamData,
        saveAnswer,
        submitExam,
        reset,
        startTimer,
        stopTimer
    }
})
