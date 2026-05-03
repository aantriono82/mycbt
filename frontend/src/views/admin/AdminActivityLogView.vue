<script setup>
import { computed, onMounted, ref } from 'vue'
import { mdiHistory, mdiRefresh, mdiDeleteSweepOutline, mdiDelete, mdiFilterOutline, mdiCalendarClockOutline } from '@mdi/js'
import LayoutAuthenticated from '@/layouts/LayoutAuthenticated.vue'
import SectionMain from '@/components/SectionMain.vue'
import SectionTitleLineWithButton from '@/components/SectionTitleLineWithButton.vue'
import CardBox from '@/components/CardBox.vue'
import BaseButton from '@/components/BaseButton.vue'
import BaseIcon from '@/components/BaseIcon.vue'
import FormField from '@/components/FormField.vue'
import FormControl from '@/components/FormControl.vue'
import { api } from '@/services/api.js'
import { shortCode2 } from '@/utils/shortCode.js'

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
  ip: '',
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

const shortId = (value) => shortCode2(value)

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
      ip: filters.value.ip?.trim() || undefined,
      from: filters.value.from || undefined,
      to: filters.value.to || undefined,
    }
    const { data } = await api.get('/api/v1/admin/login-logs', {
      params,
    })
    items.value = data?.data?.items || []
    total.value = data?.data?.total || 0
  } catch (err) {
    errorMessage.value = err?.response?.data?.error?.message || 'Gagal memuat log aktivitas'
  } finally {
    isLoading.value = false
  }
}

const applyFilters = async () => {
  offset.value = 0
  await load()
}

const resetFilters = async () => {
  filters.value = { q: '', role: 'all', ip: '', from: '', to: '' }
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
  if (!confirm('Hapus log ini?')) return
  try {
    await api.delete(`/api/v1/admin/login-logs/${id}`)
    successMessage.value = 'Log dihapus.'
    await load()
  } catch (err) {
    errorMessage.value = err?.response?.data?.error?.message || 'Gagal menghapus log'
  }
}

const clearAll = async () => {
  if (!confirm('Hapus semua log aktivitas login?')) return
  try {
    const { data } = await api.delete('/api/v1/admin/login-logs')
    const deleted = data?.data?.deleted ?? 0
    successMessage.value = `Berhasil menghapus ${deleted} log.`
    offset.value = 0
    await load()
  } catch (err) {
    errorMessage.value = err?.response?.data?.error?.message || 'Gagal menghapus semua log'
  }
}

const pruneOlderThan = async (days = 30) => {
  if (!confirm(`Hapus semua log yang lebih lama dari ${days} hari?`)) return
  try {
    const { data } = await api.delete('/api/v1/admin/login-logs/prune', { params: { days } })
    const deleted = data?.data?.deleted ?? 0
    successMessage.value = `Berhasil menghapus ${deleted} log (>${days} hari).`
    offset.value = 0
    await load()
  } catch (err) {
    errorMessage.value = err?.response?.data?.error?.message || 'Gagal prune log'
  }
}

onMounted(load)
</script>

<template>
  <LayoutAuthenticated>
    <SectionMain>
      <SectionTitleLineWithButton :icon="mdiHistory" title="Log Aktivitas Login" main>
        <div class="flex items-center gap-2">
          <BaseButton :icon="mdiRefresh" color="info" label="Refresh" @click="load" />
          <BaseButton :icon="mdiCalendarClockOutline" color="purple" label="Hapus >30 Hari" @click="pruneOlderThan(30)" />
          <BaseButton :icon="mdiDeleteSweepOutline" color="danger" label="Hapus Semua" @click="clearAll" />
        </div>
      </SectionTitleLineWithButton>

      <CardBox class="shadow-md" color="blue">
        <div v-if="errorMessage" class="mb-4 rounded-lg bg-red-50 p-3 text-xs text-red-700 border border-red-100">
          {{ errorMessage }}
        </div>
        <div v-if="successMessage" class="mb-4 rounded-lg bg-emerald-50 p-3 text-xs text-emerald-700 border border-emerald-100">
          {{ successMessage }}
        </div>

        <div class="mb-5 grid gap-4 md:grid-cols-12 border border-emerald-400/60 dark:border-emerald-800/80 p-4 rounded-2xl bg-emerald-50/20 dark:bg-emerald-900/10">
          <div class="md:col-span-12 flex items-center gap-2 text-xs font-black uppercase tracking-widest text-emerald-600 dark:text-emerald-400">
            <BaseIcon :path="mdiFilterOutline" size="16" />
            Filter
          </div>
          <div class="md:col-span-4">
            <FormField label="Cari Username">
              <FormControl v-model="filters.q" placeholder="mis. admin / gurutest" />
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
                  { value: 'student', label: 'Student' },
                ]"
              />
            </FormField>
          </div>
          <div class="md:col-span-3">
            <FormField label="IP">
              <FormControl v-model="filters.ip" placeholder="127.0.0.1" />
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
          <div class="md:col-span-12 flex flex-wrap items-center justify-between gap-3">
            <div class="flex items-center gap-2">
              <BaseButton color="info" label="Apply" @click="applyFilters" />
              <BaseButton color="success" label="Reset" @click="resetFilters" />
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

        <div class="overflow-x-auto rounded-xl border border-purple-400/60 dark:border-purple-800/80">
          <table class="w-full text-left text-sm">
            <thead class="border-b dark:border-slate-800 bg-slate-50 dark:bg-slate-800/50 text-slate-600 dark:text-slate-300 uppercase text-[10px] tracking-widest font-black">
              <tr>
                <th class="px-3 py-3">Waktu</th>
                <th class="px-3 py-3">User</th>
                <th class="px-3 py-3">Role</th>
                <th class="px-3 py-3">IP</th>
                <th class="px-3 py-3">User Agent</th>
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
                  {{ fmtTime(it.logged_in_at) }}
                </td>
                <td class="px-3 py-3">
                  <div class="font-bold text-slate-800 dark:text-slate-200">{{ it.username }}</div>
                  <div v-if="it.user_id" class="text-[10px] text-slate-400 font-mono">{{ shortId(it.user_id) }}</div>
                </td>
                <td class="px-3 py-3">
                  <span class="inline-flex items-center rounded-lg bg-slate-100 dark:bg-slate-800 px-2 py-1 text-[10px] font-black uppercase text-slate-500">
                    {{ it.role }}
                  </span>
                </td>
                <td class="px-3 py-3 text-xs font-mono text-slate-500">{{ it.ip }}</td>
                <td class="px-3 py-3 text-xs text-slate-500 max-w-[520px]">
                  <div class="truncate" :title="it.user_agent">{{ it.user_agent }}</div>
                </td>
                <td class="px-3 py-3 text-right">
                  <button
                    class="inline-flex items-center gap-2 rounded-xl px-3 py-2 text-xs font-bold text-red-600 hover:bg-red-50 dark:hover:bg-red-900/20"
                    @click="deleteRow(it.id)"
                    title="Hapus log"
                  >
                    <BaseIcon :path="mdiDelete" size="16" />
                    Hapus
                  </button>
                </td>
              </tr>
              <tr v-if="!items.length && !isLoading">
                <td class="px-3 py-6 text-center text-slate-400" colspan="6">Belum ada log login.</td>
              </tr>
            </tbody>
          </table>
        </div>
      </CardBox>
    </SectionMain>
  </LayoutAuthenticated>
</template>
