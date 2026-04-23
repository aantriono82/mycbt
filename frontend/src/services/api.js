import axios from 'axios'

const TOKEN_KEY = 'mycbt_token'

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

    if (!error.response) {
      notificationStore.pushError('Koneksi internet terputus atau server tidak merespons.')
      return Promise.reject(error)
    }

    const { status, data } = error.response
    const message = data?.error?.message || data?.message || 'Terjadi kesalahan pada sistem.'

    if (status === 401) {
      notificationStore.pushWarning('Sesi Anda telah berakhir. Silakan login kembali.')
      setStoredToken('')
      // Delay redirect slightly so user sees the message
      setTimeout(() => {
        window.location.href = '/#/login'
      }, 1500)
    } else if (status === 403) {
      notificationStore.pushError('Akses ditolak: Anda tidak memiliki izin untuk aksi ini.')
    } else if (status >= 500) {
      notificationStore.pushError(`Server Error: ${message}`)
    } else if (status === 404) {
      // notificationStore.pushWarning('Data tidak ditemukan.')
    } else {
      notificationStore.pushError(message)
    }

    return Promise.reject(error)
  },
)
