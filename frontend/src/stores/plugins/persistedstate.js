const isObject = (v) => v && typeof v === 'object' && !Array.isArray(v)

const pickPaths = (store, paths = []) => {
  if (!paths.length) return { ...store.$state }
  const out = {}
  for (const p of paths) {
    if (Object.prototype.hasOwnProperty.call(store.$state, p)) {
      out[p] = store.$state[p]
    }
  }
  return out
}

export const createPersistedStatePlugin = () => {
  return ({ store, options }) => {
    const persist = options?.persist
    if (!persist) return

    const key = persist.key || `pinia:${store.$id}`
    const paths = Array.isArray(persist.paths) ? persist.paths : []
    const storage = persist.storage || localStorage

    try {
      const raw = storage.getItem(key)
      if (raw) {
        const parsed = JSON.parse(raw)
        if (isObject(parsed)) {
          store.$patch(parsed)
        }
      }
    } catch {
      // ignore malformed persisted state
    }

    store.$subscribe(
      () => {
        try {
          storage.setItem(key, JSON.stringify(pickPaths(store, paths)))
        } catch {
          // ignore storage errors
        }
      },
      { detached: true },
    )
  }
}

