const ALPHABET = '0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ'

export const shortCode2 = (value) => {
  const raw = String(value || '')
    .toUpperCase()
    .replace(/[^0-9A-F]/g, '')

  if (!raw) return '--'

  const tail = raw.slice(-6)
  const n = Number.parseInt(tail, 16)
  if (Number.isNaN(n)) return '--'

  const mod = n % (36 * 36 * 36 * 36)
  const c1 = Math.floor(mod / (36 * 36 * 36)) % 36
  const c2 = Math.floor(mod / (36 * 36)) % 36
  const c3 = Math.floor(mod / 36) % 36
  const c4 = mod % 36
  return `${ALPHABET[c1]}${ALPHABET[c2]}${ALPHABET[c3]}${ALPHABET[c4]}`
}
