import { defineStore } from 'pinia'
import { computed, ref, watch } from 'vue'
import { api, getStoredToken, setStoredToken } from '@/services/api.js'

const USER_KEY = 'atigacbt_user'

export const ROLES = /** @type {const} */ ({
  ADMIN: 'admin',
  TEACHER: 'teacher',
  STUDENT: 'student',
})

const isValidRole = (role) => Object.values(ROLES).includes(role)

export const homeRouteForRole = (role) => {
  if (role === ROLES.TEACHER) return '/teacher/dashboard'
  if (role === ROLES.STUDENT) return '/student/dashboard'
  return '/admin/dashboard'
}

export const profileRouteForRole = (role) => {
  if (role === ROLES.TEACHER) return '/teacher/profil'
  if (role === ROLES.STUDENT) return '/student/profil'
  return '/profile'
}

export const routeAllowedForRole = (role, path) => {
  if (!path || path === '/' || path === '/dashboard') return true
  if (path === '/login' || path === '/error') return true
  if (path === '/profile') return role === ROLES.ADMIN
  if (path.startsWith('/admin/')) return role === ROLES.ADMIN
  if (path.startsWith('/teacher/')) return role === ROLES.TEACHER
  if (path.startsWith('/student/')) return role === ROLES.STUDENT
  return true
}

export const useAuthStore = defineStore('auth', () => {
  const role = ref(ROLES.ADMIN)
  const token = ref(getStoredToken())
  const user = ref(null)
  const isLoading = ref(false)
  const errorMessage = ref('')

  try {
    const savedUser = localStorage.getItem(USER_KEY)
    if (savedUser) {
      const parsed = JSON.parse(savedUser)
      if (parsed?.role && isValidRole(parsed.role)) {
        user.value = parsed
        role.value = parsed.role
      }
    }
  } catch {
    // ignore
  }

  watch(
    user,
    (v) => {
      try {
        if (v) {
          localStorage.setItem(USER_KEY, JSON.stringify(v))
        } else {
          localStorage.removeItem(USER_KEY)
        }
      } catch {
        // ignore
      }
    },
    { deep: true, flush: 'sync' },
  )

  const roleLabel = computed(() => {
    if (role.value === ROLES.ADMIN) return 'Admin'
    if (role.value === ROLES.TEACHER) return 'Guru'
    if (role.value === ROLES.STUDENT) return 'Siswa'
    return role.value
  })

  const userDisplayName = computed(
    () => user.value?.name || user.value?.full_name || user.value?.username || 'Pengguna',
  )

  const userEmail = computed(() => user.value?.email || '')

  const isAuthenticated = computed(() => !!token.value)

  const setRole = (nextRole) => {
    if (!isValidRole(nextRole)) return
    role.value = nextRole
  }

  const login = async ({ username, password }) => {
    isLoading.value = true
    errorMessage.value = ''
    try {
      const { data } = await api.post('/api/v1/auth/login', { username, password })
      const accessToken = data?.data?.token || data?.data?.access_token || ''
      if (!accessToken) {
        throw new Error('Token tidak ditemukan pada response login')
      }
      token.value = accessToken
      setStoredToken(accessToken)

      const me = await api.get('/api/v1/me')
      user.value = me?.data?.data || null
      if (user.value?.role && isValidRole(user.value.role)) {
        role.value = user.value.role
      }
      return true
    } catch (error) {
      token.value = ''
      user.value = null
      setStoredToken('')
      errorMessage.value =
        error?.response?.data?.error?.message || error?.message || 'Login gagal'
      return false
    } finally {
      isLoading.value = false
    }
  }

  const loadMe = async () => {
    if (!token.value) return false
    try {
      const { data } = await api.get('/api/v1/me')
      user.value = data?.data || null
      if (user.value?.role && isValidRole(user.value.role)) {
        role.value = user.value.role
      }
      return true
    } catch {
      token.value = ''
      user.value = null
      setStoredToken('')
      return false
    }
  }

  const loginWithToken = async (newToken) => {
    isLoading.value = true
    errorMessage.value = ''
    try {
      token.value = newToken
      setStoredToken(newToken)
      const me = await api.get('/api/v1/me')
      user.value = me?.data?.data || null
      if (user.value?.role && isValidRole(user.value.role)) {
        role.value = user.value.role
      }
      return true
    } catch (error) {
      logout()
      errorMessage.value = error?.response?.data?.error?.message || error?.message || 'Token login gagal'
      return false
    } finally {
      isLoading.value = false
    }
  }

  const logout = () => {
    token.value = ''
    user.value = null
    errorMessage.value = ''
    setStoredToken('')
    role.value = ROLES.ADMIN
  }

  return {
    role,
    roleLabel,
    token,
    user,
    userDisplayName,
    userEmail,
    isAuthenticated,
    isLoading,
    errorMessage,
    setRole,
    login,
    loginWithToken,
    loadMe,
    logout,
  }
})
