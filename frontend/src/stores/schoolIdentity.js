import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { api } from '@/services/api.js'

export const useSchoolIdentityStore = defineStore('schoolIdentity', () => {
  const schoolIdentity = ref({
    school_name: '',
    logo_url: '',
  })
  const logoVersion = ref(Date.now())
  const isLoading = ref(false)

  const extractLogoUrl = (payload = {}) => {
    const candidates = [
      payload.logo_url,
      payload.logoUrl,
      payload.url,
      payload.file_url,
      payload.path,
      payload.location,
    ]
    const found = candidates.find((item) => String(item || '').trim())
    return found ? String(found).trim() : ''
  }

  const pickIdentityPayload = (raw = {}) => {
    if (!raw || typeof raw !== 'object') return {}
    if (raw.data && typeof raw.data === 'object') return raw.data
    if (raw.school_identity && typeof raw.school_identity === 'object') return raw.school_identity
    if (raw.schoolIdentity && typeof raw.schoolIdentity === 'object') return raw.schoolIdentity
    return raw
  }

  const normalizeSchoolIdentity = (raw = {}) => {
    const identity = pickIdentityPayload(raw)
    return {
      ...identity,
      school_name: String(identity.school_name || identity.name || '').trim(),
      logo_url: extractLogoUrl(identity),
    }
  }

  const apiOrigin = computed(() => {
    const baseURL = String(api.defaults.baseURL || '').trim()
    if (!baseURL) return ''
    try {
      return new URL(baseURL, window.location.origin).origin
    } catch {
      return ''
    }
  })

  const schoolBrandName = computed(() => schoolIdentity.value.school_name?.trim() || 'ATIGACBT')
  
  const schoolBrandLogo = computed(() => {
    const logo = extractLogoUrl(normalizeSchoolIdentity(schoolIdentity.value))
    if (!logo) return '/logo_atiga.png'
    if (/^(data:|blob:|https?:\/\/|\/\/)/i.test(logo)) return logo
    
    const withVersion = (value) => `${value}${String(value).includes('?') ? '&' : '?'}v=${logoVersion.value}`
    
    if (apiOrigin.value) {
      if (logo.startsWith('/')) return withVersion(`${apiOrigin.value}${logo}`)
      return withVersion(`${apiOrigin.value}/${logo.replace(/^\/+/, '')}`)
    }
    if (logo.startsWith('/')) return withVersion(logo)
    return withVersion(`/${logo.replace(/^\/+/, '')}`)
  })

  const schoolBrandTagline = computed(() => (schoolIdentity.value.school_name?.trim() ? 'School CBT' : 'Professional CBT'))

  const loadSchoolIdentity = async (forceAdmin = false) => {
    if (isLoading.value) return
    isLoading.value = true
    try {
      // Selalu coba public dulu
      const { data } = await api.get('/api/v1/public/school-identity')
      const normalized = normalizeSchoolIdentity(data || {})
      schoolIdentity.value = { ...schoolIdentity.value, ...normalized }
      logoVersion.value = Date.now()
    } catch (e) {
      // Jika public gagal dan kita punya akses admin (atau dipaksa), coba endpoint settings
      if (forceAdmin) {
        try {
          const { data } = await api.get('/api/v1/settings/school-identity')
          const normalized = normalizeSchoolIdentity(data || {})
          schoolIdentity.value = { ...schoolIdentity.value, ...normalized }
          logoVersion.value = Date.now()
        } catch (err) {
          console.error('Failed to load school identity via settings:', err)
        }
      } else {
        console.error('Failed to load school identity via public:', e)
      }
    } finally {
      isLoading.value = false
    }
  }

  const updateIdentity = (data) => {
    const normalized = normalizeSchoolIdentity(data)
    schoolIdentity.value = { ...schoolIdentity.value, ...normalized }
    logoVersion.value = Date.now()
  }

  return {
    schoolIdentity,
    logoVersion,
    isLoading,
    schoolBrandName,
    schoolBrandLogo,
    schoolBrandTagline,
    loadSchoolIdentity,
    updateIdentity,
  }
})
