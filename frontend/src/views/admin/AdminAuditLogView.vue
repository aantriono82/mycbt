<script setup>
import { computed, onMounted, ref } from 'vue'
import {
  mdiClipboardTextSearchOutline,
  mdiRefresh,
  mdiDeleteSweepOutline,
  mdiDelete,
  mdiFilterOutline,
  mdiInformationOutline,
  mdiCalendarClockOutline,
  mdiDownload,
} from '@mdi/js'
import LayoutAuthenticated from '@/layouts/LayoutAuthenticated.vue'
import SectionMain from '@/components/SectionMain.vue'
import SectionTitleLineWithButton from '@/components/SectionTitleLineWithButton.vue'
import CardBox from '@/components/CardBox.vue'
import BaseButton from '@/components/BaseButton.vue'
import BaseIcon from '@/components/BaseIcon.vue'
import FormField from '@/components/FormField.vue'
import FormControl from '@/components/FormControl.vue'
import CardBoxModal from '@/components/CardBoxModal.vue'
import { api } from '@/services/api.js'

const isLoading = ref(false)
const errorMessage = ref('')
const successMessage = ref('')

const items = ref([])
const total = ref(0)
const limit = ref(100)
const offset = ref(0)

const filters = ref({
  q: '',
  role: 'all',
  method: 'all',
  path: '',
  status: '',
  from: '',
  to: '',
})

const pageInfo = computed(() => {
  const start = total.value === 0 ? 0 : offset.value + 1
  const end = Math.min(offset.value + limit.value, total.value)
  return `${start}-${end} dari ${total.value}`
})

const fmtTime = (iso) => {
  try {
    const d = new Date(iso)
    const formatted = d.toLocaleString('id-ID', {
      dateStyle: 'medium',
      timeStyle: 'short',
      hour12: false,
    }).replace(/\./g, ':')
    const offset = -d.getTimezoneOffset() / 60
    let tz = ''
    if (offset === 7) tz = 'WIB'
    else if (offset === 8) tz = 'WITA'
    else if (offset === 9) tz = 'WIT'
    else tz = offset >= 0 ? `GMT+${offset}` : `GMT${offset}`
    return `${formatted} ${tz}`
  } catch {
    return iso
  }
}

const load = async () => {
  isLoading.value = true
  errorMessage.value = ''
  successMessage.value = ''
  try {
    const params = {
      limit: limit.value,
      offset: offset.value,
      q: filters.value.q?.trim() || undefined,
      role: filters.value.role || undefined,
      method: filters.value.method || undefined,
      path: filters.value.path?.trim() || undefined,
      status: filters.value.status?.trim() || undefined,
      from: filters.value.from || undefined,
      to: filters.value.to || undefined,
    }
    const { data } = await api.get('/api/v1/admin/audit-logs', { params })
    items.value = data?.data?.items || []
    total.value = data?.data?.total || 0
  } catch (err) {
    errorMessage.value = err?.response?.data?.error?.message || 'Gagal memuat audit log'
  } finally {
    isLoading.value = false
  }
}

const applyFilters = async () => {
  offset.value = 0
  await load()
}

const resetFilters = async () => {
  filters.value = { q: '', role: 'all', method: 'all', path: '', status: '', from: '', to: '' }
  offset.value = 0
  await load()
}

const nextPage = async () => {
  if (offset.value + limit.value >= total.value) return
  offset.value += limit.value
  await load()
}

const prevPage = async () => {
  offset.value = Math.max(0, offset.value - limit.value)
  await load()
}

const deleteRow = async (id) => {
  if (!confirm('Hapus audit log ini?')) return
  try {
    await api.delete(`/api/v1/admin/audit-logs/${id}`)
    successMessage.value = 'Audit log dihapus.'
    await load()
  } catch (err) {
    errorMessage.value = err?.response?.data?.error?.message || 'Gagal menghapus audit log'
  }
}

const clearAll = async () => {
  if (!confirm('Hapus semua audit log?')) return
  try {
    const { data } = await api.delete('/api/v1/admin/audit-logs')
    const deleted = data?.data?.deleted ?? 0
    successMessage.value = `Berhasil menghapus ${deleted} audit log.`
    offset.value = 0
    await load()
  } catch (err) {
    errorMessage.value = err?.response?.data?.error?.message || 'Gagal menghapus semua audit log'
  }
}

const pruneOlderThan = async (days = 30) => {
  if (!confirm(`Hapus semua audit log yang lebih lama dari ${days} hari?`)) return
  try {
    const { data } = await api.delete('/api/v1/admin/audit-logs/prune', { params: { days } })
    const deleted = data?.data?.deleted ?? 0
    successMessage.value = `Berhasil menghapus ${deleted} audit log (>${days} hari).`
    offset.value = 0
    await load()
  } catch (err) {
    errorMessage.value = err?.response?.data?.error?.message || 'Gagal prune audit log'
  }
}

const exportCSV = async () => {
  errorMessage.value = ''
  successMessage.value = ''
  try {
    const params = {
      q: filters.value.q?.trim() || undefined,
      role: filters.value.role || undefined,
      method: filters.value.method || undefined,
      path: filters.value.path?.trim() || undefined,
      status: filters.value.status?.trim() || undefined,
      from: filters.value.from || undefined,
      to: filters.value.to || undefined,
    }
    const response = await api.get('/api/v1/admin/audit-logs/export', {
      params,
      responseType: 'blob',
    })

    const url = window.URL.createObjectURL(new Blob([response.data]))
    const link = document.createElement('a')
    link.href = url

    const contentDisposition = response.headers?.['content-disposition']
    let filename = `audit_logs_${new Date().toISOString().slice(0, 10)}.csv`
    if (contentDisposition) {
      const filenameMatch = String(contentDisposition).match(/filename="?([^"]+)"?/)
      if (filenameMatch && filenameMatch[1]) filename = filenameMatch[1]
    }

    link.setAttribute('download', filename)
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    window.URL.revokeObjectURL(url)

    successMessage.value = 'Export CSV dimulai.'
  } catch (err) {
    errorMessage.value = err?.response?.data?.error?.message || 'Gagal export CSV'
  }
}

const isDetailOpen = ref(false)
const detail = ref(null)

const openDetail = (it) => {
  detail.value = it
  isDetailOpen.value = true
}

const payloadPretty = computed(() => {
  if (!detail.value?.payload_json) return '{}'
  try {
    return JSON.stringify(detail.value.payload_json, null, 2)
  } catch {
    return String(detail.value.payload_json || '')
  }
})

onMounted(load)
</script>

<template>
  <LayoutAuthenticated>
    <SectionMain>
      <SectionTitleLineWithButton :icon="mdiClipboardTextSearchOutline" title="Audit Log" main>
        <div class="flex items-center gap-2">
          <BaseButton :icon="mdiRefresh" color="info" label="Refresh" @click="load" />
          <BaseButton :icon="mdiDownload" color="info" outline label="Export CSV" @click="exportCSV" />
          <BaseButton :icon="mdiCalendarClockOutline" color="warning" label="Hapus >30 Hari" @click="pruneOlderThan(30)" />
          <BaseButton :icon="mdiDeleteSweepOutline" color="danger" label="Hapus Semua" @click="clearAll" />
        </div>
      </SectionTitleLineWithButton>

      <CardBoxModal v-model="isDetailOpen" title="Detail Audit Log" has-cancel>
        <div v-if="detail" class="space-y-3">
          <div class="grid gap-3 md:grid-cols-2 text-xs">
            <div class="rounded-xl border border-slate-200 dark:border-slate-800 p-3">
              <div class="text-[10px] font-black uppercase tracking-widest text-slate-400">Actor</div>
              <div class="mt-1 font-bold text-slate-800 dark:text-slate-200">{{ detail.username || '-' }}</div>
              <div class="text-[11px] text-slate-500">{{ detail.name || '' }}</div>
              <div class="mt-1 text-[11px] font-mono text-slate-400">{{ detail.user_id || '' }}</div>
            </div>
            <div class="rounded-xl border border-slate-200 dark:border-slate-800 p-3">
              <div class="text-[10px] font-black uppercase tracking-widest text-slate-400">Request</div>
              <div class="mt-1 font-mono text-[11px] text-slate-600 dark:text-slate-300">{{ detail.request_id }}</div>
              <div class="mt-1 text-slate-600 dark:text-slate-300">
                <span class="font-black">{{ detail.method }}</span>
                <span class="ml-2">{{ detail.path }}</span>
              </div>
              <div class="mt-1 text-[11px] text-slate-500">Status: {{ detail.status_code }}</div>
            </div>
          </div>

          <div class="rounded-xl border border-slate-200 dark:border-slate-800 p-3 text-xs">
            <div class="text-[10px] font-black uppercase tracking-widest text-slate-400">Meta</div>
            <div class="mt-1 text-slate-600 dark:text-slate-300">Waktu: {{ fmtTime(detail.created_at) }}</div>
            <div class="mt-1 text-slate-600 dark:text-slate-300">IP: {{ detail.ip || '-' }}</div>
            <div class="mt-1 text-slate-600 dark:text-slate-300 break-words">UA: {{ detail.user_agent || '-' }}</div>
            <div v-if="detail.query" class="mt-1 text-slate-600 dark:text-slate-300 break-words">Query: {{ detail.query }}</div>
          </div>

          <div class="rounded-xl border border-slate-200 dark:border-slate-800 p-3">
            <div class="text-[10px] font-black uppercase tracking-widest text-slate-400">Payload</div>
            <pre class="mt-2 max-h-72 overflow-auto rounded-lg bg-slate-950 p-3 text-[11px] text-slate-100">{{ payloadPretty }}</pre>
          </div>
        </div>
      </CardBoxModal>

      <CardBox class="shadow-md">
        <div v-if="errorMessage" class="mb-4 rounded-lg bg-red-50 p-3 text-xs text-red-700 border border-red-100">
          {{ errorMessage }}
        </div>
        <div v-if="successMessage" class="mb-4 rounded-lg bg-emerald-50 p-3 text-xs text-emerald-700 border border-emerald-100">
          {{ successMessage }}
        </div>

        <div class="mb-5 grid gap-4 md:grid-cols-12">
          <div class="md:col-span-12 flex items-center gap-2 text-xs font-black uppercase tracking-widest text-slate-400">
            <BaseIcon :path="mdiFilterOutline" size="16" />
            Filter
          </div>
          <div class="md:col-span-4">
            <FormField label="Cari (user/path/request_id)">
              <FormControl v-model="filters.q" placeholder="mis. admin / /api/v1/admin" />
            </FormField>
          </div>
          <div class="md:col-span-2">
            <FormField label="Role">
              <FormControl
                v-model="filters.role"
                :options="[
                  { value: 'all', label: 'Semua' },
                  { value: 'admin', label: 'Admin' },
                  { value: 'teacher', label: 'Teacher' },
                ]"
              />
            </FormField>
          </div>
          <div class="md:col-span-2">
            <FormField label="Method">
              <FormControl
                v-model="filters.method"
                :options="[
                  { value: 'all', label: 'Semua' },
                  { value: 'POST', label: 'POST' },
                  { value: 'PUT', label: 'PUT' },
                  { value: 'PATCH', label: 'PATCH' },
                  { value: 'DELETE', label: 'DELETE' },
                ]"
              />
            </FormField>
          </div>
          <div class="md:col-span-2">
            <FormField label="Status">
              <FormControl v-model="filters.status" placeholder="200" inputmode="numeric" />
            </FormField>
          </div>
          <div class="md:col-span-2">
            <FormField label="Path contains">
              <FormControl v-model="filters.path" placeholder="/api/v1/admin" />
            </FormField>
          </div>
          <div class="md:col-span-3 grid grid-cols-2 gap-3">
            <FormField label="From (tanggal)">
              <FormControl v-model="filters.from" type="date" />
            </FormField>
            <FormField label="To (tanggal)">
              <FormControl v-model="filters.to" type="date" />
            </FormField>
          </div>

          <div class="md:col-span-9 flex flex-wrap items-end justify-between gap-3">
            <div class="flex items-center gap-2">
              <BaseButton color="info" label="Apply" @click="applyFilters" />
              <BaseButton outline color="whiteDark" label="Reset" @click="resetFilters" />
            </div>
            <div class="flex items-center gap-3">
              <FormField label="Limit">
                <FormControl v-model="limit" type="number" inputmode="numeric" />
              </FormField>
              <div class="text-xs text-slate-500 dark:text-slate-400">
                {{ pageInfo }}
                <span v-if="isLoading" class="ml-2">Memuat...</span>
              </div>
              <div class="flex items-center gap-2">
                <BaseButton small outline label="Prev" color="whiteDark" @click="prevPage" />
                <BaseButton small outline label="Next" color="whiteDark" @click="nextPage" />
              </div>
            </div>
          </div>
        </div>

        <div class="overflow-x-auto">
          <table class="w-full text-left text-sm">
            <thead class="border-b dark:border-slate-800 bg-slate-50 dark:bg-slate-800/50 text-slate-600 dark:text-slate-300 uppercase text-[10px] tracking-widest font-black">
              <tr>
                <th class="px-3 py-3">Waktu</th>
                <th class="px-3 py-3">Actor</th>
                <th class="px-3 py-3">Req</th>
                <th class="px-3 py-3">Method</th>
                <th class="px-3 py-3">Path</th>
                <th class="px-3 py-3 text-center">Status</th>
                <th class="px-3 py-3 text-right">Aksi</th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="it in items"
                :key="it.id"
                class="border-b dark:border-slate-800 last:border-b-0 hover:bg-slate-50 dark:hover:bg-slate-800/50 transition-colors"
              >
                <td class="px-3 py-3 text-xs font-mono text-slate-600 dark:text-slate-300 whitespace-nowrap">
                  {{ fmtTime(it.created_at) }}
                </td>
                <td class="px-3 py-3">
                  <div class="font-bold text-slate-800 dark:text-slate-200">{{ it.username || '-' }}</div>
                  <div class="text-[10px] text-slate-400 uppercase font-black">{{ it.role || '-' }}</div>
                </td>
                <td class="px-3 py-3 text-[11px] font-mono text-slate-500 max-w-[220px]">
                  <div class="truncate" :title="it.request_id">{{ it.request_id }}</div>
                </td>
                <td class="px-3 py-3">
                  <span class="inline-flex items-center rounded-lg bg-slate-100 dark:bg-slate-800 px-2 py-1 text-[10px] font-black uppercase text-slate-600">
                    {{ it.method }}
                  </span>
                </td>
                <td class="px-3 py-3 text-xs text-slate-600 dark:text-slate-300 max-w-[520px]">
                  <div class="truncate" :title="it.path">{{ it.path }}</div>
                </td>
                <td class="px-3 py-3 text-center">
                  <span
                    class="inline-flex items-center rounded-lg px-2 py-1 text-[10px] font-black"
                    :class="it.status_code >= 400 ? 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-200' : 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-200'"
                  >
                    {{ it.status_code }}
                  </span>
                </td>
                <td class="px-3 py-3 text-right whitespace-nowrap">
                  <button
                    class="inline-flex items-center gap-2 rounded-xl px-3 py-2 text-xs font-bold text-slate-700 hover:bg-slate-100 dark:text-slate-200 dark:hover:bg-slate-800"
                    @click="openDetail(it)"
                    title="Detail"
                  >
                    <BaseIcon :path="mdiInformationOutline" size="16" />
                    Detail
                  </button>
                  <button
                    class="ml-2 inline-flex items-center gap-2 rounded-xl px-3 py-2 text-xs font-bold text-red-600 hover:bg-red-50 dark:hover:bg-red-900/20"
                    @click="deleteRow(it.id)"
                    title="Hapus"
                  >
                    <BaseIcon :path="mdiDelete" size="16" />
                    Hapus
                  </button>
                </td>
              </tr>
              <tr v-if="!items.length && !isLoading">
                <td class="px-3 py-6 text-center text-slate-400" colspan="7">Belum ada audit log.</td>
              </tr>
            </tbody>
          </table>
        </div>
      </CardBox>
    </SectionMain>
  </LayoutAuthenticated>
</template>
