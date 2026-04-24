import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import QuestionNavigator from '@/components/student/QuestionNavigator.vue'

const makeQuestions = (n = 40) =>
  Array.from({ length: n }, (_, i) => ({ id: `q-${i + 1}` }))

describe('QuestionNavigator', () => {
  it('renders 40 questions and marks answered/unanswered status', () => {
    const wrapper = mount(QuestionNavigator, {
      props: {
        questions: makeQuestions(40),
        answers: { 'q-1': { selected_option_id: 'a' } },
        flagged: {},
        currentIndex: 0,
      },
    })

    expect(wrapper.findAll('button').length).toBe(40)
    expect(wrapper.get('[data-testid="qnav-0"]').attributes('data-status')).toBe('answered')
    expect(wrapper.get('[data-testid="qnav-1"]').attributes('data-status')).toBe('unanswered')
  })

  it('emits jump-to with correct index when question number clicked', async () => {
    const wrapper = mount(QuestionNavigator, {
      props: {
        questions: makeQuestions(40),
        answers: {},
        flagged: {},
        currentIndex: 0,
      },
    })

    await wrapper.get('[data-testid="qnav-9"]').trigger('click')
    expect(wrapper.emitted('jump-to')).toBeTruthy()
    expect(wrapper.emitted('jump-to')[0]).toEqual([9])
  })

  it('shows flagged state and icon', () => {
    const wrapper = mount(QuestionNavigator, {
      props: {
        questions: makeQuestions(40),
        answers: {},
        flagged: { 'q-3': true },
        currentIndex: 0,
      },
    })

    expect(wrapper.get('[data-testid="qnav-2"]').attributes('data-status')).toBe('flagged')
    expect(wrapper.findAll('[data-testid="flag-icon"]').length).toBe(1)
  })
})

