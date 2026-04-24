import { describe, expect, it, vi } from 'vitest'
import { handleApiError } from '@/services/api.js'

const makeStore = () => ({
  pushWarning: vi.fn(),
  pushError: vi.fn(),
})

describe('api error handling', () => {
  it('redirects to /login when response status is 401', async () => {
    const notificationStore = makeStore()
    const redirectToLogin = vi.fn()
    const err = {
      response: {
        status: 401,
        data: { error: { message: 'unauthorized' } },
      },
    }

    await expect(
      handleApiError(err, { notificationStore, redirectToLogin }),
    ).rejects.toBe(err)

    expect(redirectToLogin).toHaveBeenCalledTimes(1)
    expect(notificationStore.pushWarning).toHaveBeenCalled()
  })

  it('handles 422 exam time expired by calling callback', async () => {
    const notificationStore = makeStore()
    const onExamTimeExpired = vi.fn()
    const err = {
      response: {
        status: 422,
        data: { error: { message: 'exam time expired' } },
      },
    }

    await expect(
      handleApiError(err, { notificationStore, onExamTimeExpired }),
    ).rejects.toBe(err)

    expect(onExamTimeExpired).toHaveBeenCalledTimes(1)
    expect(notificationStore.pushWarning).toHaveBeenCalled()
  })

  it('handles offline/network error with local-save toast message', async () => {
    const notificationStore = makeStore()
    const onOfflineSaved = vi.fn()
    const err = new Error('Network Error')

    await expect(
      handleApiError(err, { notificationStore, onOfflineSaved }),
    ).rejects.toBe(err)

    expect(notificationStore.pushWarning).toHaveBeenCalledWith(
      'koneksi terputus, jawaban disimpan lokal',
    )
    expect(onOfflineSaved).toHaveBeenCalledTimes(1)
  })
})

