import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { useExamStore } from '@/stores/exam.js'
import { createPersistedStatePlugin } from '@/stores/plugins/persistedstate.js'

const postMock = vi.fn()

vi.mock('@/services/api.js', () => ({
  api: {
    post: (...args) => postMock(...args),
    get: vi.fn(),
  },
}))

const makePinia = () => {
  const pinia = createPinia()
  pinia.use(createPersistedStatePlugin())
  setActivePinia(pinia)
  return pinia
}

describe('examStore', () => {
  beforeEach(() => {
    localStorage.clear()
    postMock.mockReset().mockResolvedValue({ data: {} })
    makePinia()
  })

  it('startExam sets sessionId, startTime, and questions', () => {
    const store = useExamStore()
    const startedAt = '2026-04-24T00:00:00Z'
    const qs = [{ id: 'q1' }, { id: 'q2' }]
    store.startExam({ sessionId: 'sess-1', startTime: startedAt, questions: qs })

    expect(store.sessionId).toBe('sess-1')
    expect(store.startTime).toBe(startedAt)
    expect(store.questions).toEqual(qs)
  })

  it('saveAnswer updates answers map and persists to localStorage', async () => {
    const store = useExamStore()
    store.startExam({ sessionId: 'sess-1', questions: [{ id: 'q1' }] })

    await store.saveAnswer('q1', { selected_option_id: 'A' })

    expect(store.answers.q1).toEqual({ selected_option_id: 'A' })
    const persisted = JSON.parse(localStorage.getItem('mycbt:exam-store'))
    expect(persisted.answers.q1).toEqual({ selected_option_id: 'A' })
    expect(postMock).toHaveBeenCalledWith('/api/v1/student/sessions/sess-1/answers', {
      question_id: 'q1',
      answer_json: JSON.stringify({ selected_option_id: 'A' }),
    })
  })

  it('finishExam posts to API and clears state', async () => {
    const store = useExamStore()
    store.startExam({ sessionId: 'sess-finish', questions: [{ id: 'q1' }] })
    store.answers = { q1: { text: 'x' } }

    await store.finishExam()

    expect(postMock).toHaveBeenCalledWith('/api/v1/student/sessions/sess-finish/submit')
    expect(store.sessionId).toBe(null)
    expect(store.questions).toEqual([])
    expect(store.answers).toEqual({})
  })

  it('rehydrates state from localStorage via persisted-state plugin', () => {
    localStorage.setItem(
      'mycbt:exam-store',
      JSON.stringify({
        sessionId: 'sess-rehydrate',
        startTime: '2026-04-24T01:00:00Z',
        questions: [{ id: 'q10' }],
        answers: { q10: { text: 'persisted' } },
      }),
    )

    makePinia()
    const store = useExamStore()

    expect(store.sessionId).toBe('sess-rehydrate')
    expect(store.startTime).toBe('2026-04-24T01:00:00Z')
    expect(store.questions).toEqual([{ id: 'q10' }])
    expect(store.answers).toEqual({ q10: { text: 'persisted' } })
  })
})

