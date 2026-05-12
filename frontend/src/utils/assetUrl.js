const getApiBaseUrl = () => String(import.meta.env.VITE_API_BASE_URL || '').trim()
const DEFAULT_API_ORIGIN = 'http://localhost:8080'

const getApiBasePath = () => {
  const baseUrl = getApiBaseUrl()
  if (!baseUrl) return ''
  try {
    const parsed = new URL(baseUrl, 'http://localhost')
    const path = String(parsed.pathname || '').replace(/\/+$/, '')
    return path === '/' ? '' : path
  } catch {
    return ''
  }
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
  const apiBasePath = getApiBasePath()
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
  if (raw.startsWith('/')) {
    if (apiBasePath && raw.startsWith('/uploads/')) {
      return `${origin}${apiBasePath}${raw}`
    }
    return `${origin}${raw}`
  }
  return `${origin}/${raw.replace(/^\/+/, '')}`
}
