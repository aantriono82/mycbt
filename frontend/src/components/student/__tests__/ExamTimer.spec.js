import { mount } from '@vue/test-utils'
import { describe, expect, it, vi, beforeEach, afterEach } from 'vitest'
import ExamTimer from '@/components/student/ExamTimer.vue'

describe('ExamTimer', () => {
  beforeEach(() => {
    vi.useFakeTimers()
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('renders initial duration=3600 as 01:00:00', () => {
    const wrapper = mount(ExamTimer, { props: { duration: 3600 } })
    expect(wrapper.get('[data-testid="exam-timer"]').text()).toBe('01:00:00')
  })

  it('after 60 seconds shows 00:59:00', async () => {
    const wrapper = mount(ExamTimer, { props: { duration: 3600 } })
    vi.advanceTimersByTime(60000)
    await wrapper.vm.$nextTick()
    expect(wrapper.get('[data-testid="exam-timer"]').text()).toBe('00:59:00')
  })

  it('emits time-up and calls submitExam when remaining time hits zero', async () => {
    const submitExam = vi.fn()
    const wrapper = mount(ExamTimer, { props: { duration: 60, submitExam } })

    vi.advanceTimersByTime(60000)
    await wrapper.vm.$nextTick()

    expect(wrapper.emitted('time-up')).toBeTruthy()
    expect(submitExam).toHaveBeenCalledTimes(1)
    expect(wrapper.get('[data-testid="exam-timer"]').text()).toBe('00:00:00')
  })

  it('clears interval on unmount to avoid timer leak', () => {
    const clearSpy = vi.spyOn(globalThis, 'clearInterval')
    const wrapper = mount(ExamTimer, { props: { duration: 120 } })
    const beforeUnmount = clearSpy.mock.calls.length
    wrapper.unmount()
    expect(clearSpy.mock.calls.length).toBeGreaterThan(beforeUnmount)
  })
})

