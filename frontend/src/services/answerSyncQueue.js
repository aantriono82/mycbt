const DB_NAME = 'atigacbt-offline-sync'
const DB_VERSION = 1
const STORE_NAME = 'answer_queue'

const hasIndexedDB = () => typeof indexedDB !== 'undefined'

const openDb = () => new Promise((resolve, reject) => {
  if (!hasIndexedDB()) {
    reject(new Error('IndexedDB not available'))
    return
  }
  const request = indexedDB.open(DB_NAME, DB_VERSION)
  request.onupgradeneeded = () => {
    const db = request.result
    if (!db.objectStoreNames.contains(STORE_NAME)) {
      const store = db.createObjectStore(STORE_NAME, { keyPath: 'id' })
      store.createIndex('sessionId', 'sessionId', { unique: false })
      store.createIndex('status', 'status', { unique: false })
      store.createIndex('updatedAt', 'updatedAt', { unique: false })
      store.createIndex('sessionQuestionKey', 'sessionQuestionKey', { unique: true })
    }
  }
  request.onsuccess = () => resolve(request.result)
  request.onerror = () => reject(request.error || new Error('Failed to open IndexedDB'))
})

const withStore = async (mode, fn) => {
  const db = await openDb()
  return new Promise((resolve, reject) => {
    const tx = db.transaction(STORE_NAME, mode)
    const store = tx.objectStore(STORE_NAME)
    let settled = false

    const done = (value) => {
      if (settled) return
      settled = true
      resolve(value)
    }
    const fail = (err) => {
      if (settled) return
      settled = true
      reject(err)
    }

    tx.oncomplete = () => done(undefined)
    tx.onerror = () => fail(tx.error || new Error('IndexedDB transaction failed'))
    tx.onabort = () => fail(tx.error || new Error('IndexedDB transaction aborted'))

    fn(store, done, fail)
  }).finally(() => db.close())
}

const requestToPromise = (request) => new Promise((resolve, reject) => {
  request.onsuccess = () => resolve(request.result)
  request.onerror = () => reject(request.error || new Error('IndexedDB request failed'))
})

const makeId = () => {
  try {
    return crypto.randomUUID()
  } catch {
    return `${Date.now()}-${Math.random().toString(16).slice(2)}`
  }
}

const makeSessionQuestionKey = (sessionId, questionId) => `${String(sessionId)}::${String(questionId)}`

export const enqueueAnswer = async ({ sessionId, questionId, answerJson }) => {
  if (!sessionId || !questionId) return
  const now = new Date().toISOString()
  await withStore('readwrite', async (store, done, fail) => {
    try {
      const key = makeSessionQuestionKey(sessionId, questionId)
      const existing = await requestToPromise(store.index('sessionQuestionKey').get(key))
      const payload = {
        id: existing?.id || makeId(),
        sessionId: String(sessionId),
        questionId: String(questionId),
        sessionQuestionKey: key,
        answerJson: String(answerJson || '{}'),
        status: 'pending',
        retryCount: Number(existing?.retryCount || 0),
        createdAt: existing?.createdAt || now,
        updatedAt: now,
      }
      store.put(payload)
      done(payload)
    } catch (err) {
      fail(err)
    }
  })
}

export const listPendingAnswers = async (sessionId) => {
  if (!sessionId) return []
  return withStore('readonly', async (store, done, fail) => {
    try {
      const rows = await requestToPromise(store.getAll())
      const items = Array.isArray(rows) ? rows : []
      const out = items
        .filter((it) => it?.sessionId === String(sessionId) && it?.status === 'pending')
        .sort((a, b) => String(a.updatedAt || '').localeCompare(String(b.updatedAt || '')))
      done(out)
    } catch (err) {
      fail(err)
    }
  })
}

export const removeQueueItem = async (id) => {
  if (!id) return
  await withStore('readwrite', (store) => {
    store.delete(id)
  })
}

export const bumpRetryCount = async (id) => {
  if (!id) return
  const now = new Date().toISOString()
  await withStore('readwrite', async (store, done, fail) => {
    try {
      const row = await requestToPromise(store.get(id))
      if (!row) {
        done()
        return
      }
      store.put({
        ...row,
        retryCount: Number(row.retryCount || 0) + 1,
        updatedAt: now,
      })
      done()
    } catch (err) {
      fail(err)
    }
  })
}

export const countPendingAnswers = async (sessionId = null) => {
  return withStore('readonly', async (store, done, fail) => {
    try {
      const rows = await requestToPromise(store.getAll())
      const items = Array.isArray(rows) ? rows : []
      const filtered = items.filter((it) => {
        if (it?.status !== 'pending') return false
        if (!sessionId) return true
        return it?.sessionId === String(sessionId)
      })
      done(filtered.length)
    } catch (err) {
      fail(err)
    }
  })
}

export const clearSessionQueue = async (sessionId) => {
  if (!sessionId) return
  await withStore('readwrite', async (store, done, fail) => {
    try {
      const rows = await requestToPromise(store.getAll())
      for (const row of rows || []) {
        if (row?.sessionId === String(sessionId)) {
          store.delete(row.id)
        }
      }
      done()
    } catch (err) {
      fail(err)
    }
  })
}
