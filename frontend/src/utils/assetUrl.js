const getApiBaseUrl = () => String(import.meta.env.VITE_API_BASE_URL || '').trim()
const DEFAULT_API_ORIGIN = 'http://localhost:8080'

const inferApiOriginFromWindow = () => {
  if (typeof window === 'undefined' || !window.location?.hostname) return ''
  const { protocol, hostname } = window.location
  if (hostname === 'localhost' || hostname === '127.0.0.1') return ''
  if (hostname.startsWith('api.')) return `${protocol}//${hostname}`
  return `${protocol}//api.${hostname}`
}

export const getApiOrigin = () => {
  const baseUrl = getApiBaseUrl()
  if (baseUrl) {
    try {
      return new URL(baseUrl).origin
    } catch {
      // ignore malformed env and fall back below
    }
  }
  const inferred = inferApiOriginFromWindow()
  if (inferred) {
    return inferred
  }
  if (typeof window !== 'undefined' && window.location?.origin) {
    return window.location.origin
  }
  return DEFAULT_API_ORIGIN
}

export const resolveBackendAssetUrl = (value) => {
  const raw = String(value || '').trim()
  if (!raw) return ''
  if (raw.startsWith('data:') || raw.startsWith('blob:')) return raw

  const origin = getApiOrigin().replace(/\/+$/, '')
  if (/^(https?:)?\/\//i.test(raw)) {
    try {
      const url = new URL(raw)
      const isLoopbackStorage = ['127.0.0.1', 'localhost'].includes(url.hostname) && ['9000', '9001'].includes(url.port)
      const parts = url.pathname.split('/').filter(Boolean)
      if (isLoopbackStorage && parts.length >= 2) {
        return `${origin}/uploads/${parts.slice(1).join('/')}`
      }
      return raw
    } catch {
      return raw
    }
  }
  if (raw.startsWith('/')) return `${origin}${raw}`
  return `${origin}/${raw.replace(/^\/+/, '')}`
}
