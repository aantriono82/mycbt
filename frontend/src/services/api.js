import axios from 'axios'

const TOKEN_KEY = 'atigacbt_token'

export const api = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080',
})

export const getStoredToken = () => {
  try {
    return localStorage.getItem(TOKEN_KEY) || ''
  } catch {
    return ''
  }
}

export const setStoredToken = (token) => {
  try {
    if (token) {
      localStorage.setItem(TOKEN_KEY, token)
    } else {
      localStorage.removeItem(TOKEN_KEY)
    }
  } catch {
    // ignore
  }
}

export const handleApiError = (
  error,
  {
    notificationStore,
    redirectToLogin = () => {
      window.location.href = '/#/login'
    },
    onExamTimeExpired = null,
    onOfflineSaved = null,
  } = {},
) => {
  if (!error.response) {
    notificationStore?.pushWarning?.('koneksi terputus, jawaban disimpan lokal')
    if (typeof onOfflineSaved === 'function') onOfflineSaved(error)
    return Promise.reject(error)
  }

  const { status, data } = error.response
  const message = data?.error?.message || data?.message || 'Terjadi kesalahan pada sistem.'

  if (status === 401) {
    notificationStore?.pushWarning?.('Sesi Anda telah berakhir. Silakan login kembali.')
    setStoredToken('')
    redirectToLogin()
  } else if (status === 422 && String(message).toLowerCase().includes('exam time expired')) {
    notificationStore?.pushWarning?.('Waktu ujian telah habis.')
    if (typeof onExamTimeExpired === 'function') onExamTimeExpired(error)
    if (typeof window !== 'undefined') {
      window.dispatchEvent(
        new CustomEvent('exam-time-expired', {
          detail: { message },
        }),
      )
    }
  } else if (status === 403) {
    notificationStore?.pushError?.('Akses ditolak: Anda tidak memiliki izin untuk aksi ini.')
  } else if (status >= 500) {
    notificationStore?.pushError?.(`Server Error: ${message}`)
  } else if (status !== 404) {
    notificationStore?.pushError?.(message)
  }

  return Promise.reject(error)
}

api.interceptors.request.use((config) => {
  const token = getStoredToken()
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

import { useNotificationStore } from '@/stores/notification.js'

api.interceptors.response.use(
  (response) => response,
  (error) => {
    const notificationStore = useNotificationStore()
    return handleApiError(error, { notificationStore })
  },
)
