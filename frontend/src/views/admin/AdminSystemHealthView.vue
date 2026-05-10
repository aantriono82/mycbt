<script setup>
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { mdiDeleteSweepOutline, mdiHeartPulse, mdiRefresh } from '@mdi/js'
import LayoutAuthenticated from '@/layouts/LayoutAuthenticated.vue'
import SectionMain from '@/components/SectionMain.vue'
import SectionTitleLineWithButton from '@/components/SectionTitleLineWithButton.vue'
import CardBox from '@/components/CardBox.vue'
import BaseButton from '@/components/BaseButton.vue'
import { api } from '@/services/api.js'

const isLoading = ref(false)
const errorMessage = ref('')
const payload = ref(null)
const selectedBatchType = ref('all')
const selectedHistoryLimit = ref(20)
let timer = null

const formatBytes = (value) => {
  const num = Number(value || 0)
  if (!Number.isFinite(num) || num <= 0) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  const exp = Math.min(Math.floor(Math.log(num) / Math.log(1024)), units.length - 1)
  const result = num / 1024 ** exp
  return `${result.toFixed(exp === 0 ? 0 : 2)} ${units[exp]}`
}

const formatTime = (iso) => {
  if (!iso) return '-'
  try {
    return new Date(iso).toLocaleString('id-ID', { dateStyle: 'medium', timeStyle: 'medium', hour12: false }).replace(/\./g, ':')
  } catch {
    return iso
  }
}

const metrics = computed(() => payload.value?.metrics || {})
const queue = computed(() => payload.value?.queue || {})
const queueStatus = computed(() => queue.value?.status || 'idle')
const queueHistory = computed(() => payload.value?.queue_history || [])
const filteredQueueHistory = computed(() => {
  if (selectedBatchType.value === 'all') return queueHistory.value
  return queueHistory.value.filter((item) => item?.batch_type === selectedBatchType.value)
})

const refresh = async () => {
  isLoading.value = true
  errorMessage.value = ''
  try {
    const { data } = await api.get('/api/v1/admin/system/health', {
      params: {
        history_limit: selectedHistoryLimit.value,
      },
    })
    payload.value = data?.data || null
  } catch (err) {
    errorMessage.value = err?.response?.data?.error?.message || 'Gagal memuat system health'
  } finally {
    isLoading.value = false
  }
}

const clearHistory = async () => {
  if (!confirm('Hapus semua riwayat batch queue?')) return
  try {
    await api.delete('/api/v1/admin/system/health/queue-history')
    await refresh()
  } catch (err) {
    errorMessage.value = err?.response?.data?.error?.message || 'Gagal menghapus riwayat queue'
  }
}

onMounted(async () => {
  await refresh()
  timer = setInterval(refresh, 5000)
})

onUnmounted(() => {
  if (timer) {
    clearInterval(timer)
    timer = null
  }
})
</script>

<template>
  <LayoutAuthenticated>
    <SectionMain>
      <SectionTitleLineWithButton :icon="mdiHeartPulse" title="System Health" main>
        <BaseButton :icon="mdiRefresh" color="info" label="Refresh" :disabled="isLoading" @click="refresh" />
      </SectionTitleLineWithButton>

      <div v-if="errorMessage" class="mb-4 rounded-lg bg-red-50 p-3 text-xs text-red-700 border border-red-100">
        {{ errorMessage }}
      </div>

      <div class="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
        <CardBox>
          <div class="text-xs uppercase text-slate-500">SSE Aktif</div>
          <div class="mt-1 text-2xl font-bold">{{ metrics.active_sse_connections ?? 0 }}</div>
        </CardBox>
        <CardBox>
          <div class="text-xs uppercase text-slate-500">Avg API Latency</div>
          <div class="mt-1 text-2xl font-bold">{{ metrics.average_api_latency_ms ?? 0 }} ms</div>
        </CardBox>
        <CardBox>
          <div class="text-xs uppercase text-slate-500">CPU Usage (Process)</div>
          <div class="mt-1 text-2xl font-bold">{{ metrics.process_cpu_usage_pct ?? 0 }}%</div>
        </CardBox>
        <CardBox>
          <div class="text-xs uppercase text-slate-500">Queue Status</div>
          <div class="mt-1 text-xl font-bold">{{ queueStatus }}</div>
        </CardBox>
      </div>

      <CardBox class="mt-4">
        <div class="text-xs uppercase text-slate-500 mb-2">Queue</div>
        <div class="grid gap-2 text-sm md:grid-cols-2">
          <div>Batch Type: <strong>{{ queue.batch_type || '-' }}</strong></div>
          <div>Last Updated: <strong>{{ formatTime(queue.last_updated_utc) }}</strong></div>
          <div>Pending: <strong>{{ queue.pending ?? 0 }}</strong></div>
          <div>Processing: <strong>{{ queue.processing ?? 0 }}</strong></div>
          <div>Success: <strong>{{ queue.success ?? 0 }}</strong></div>
          <div>Failed: <strong>{{ queue.failed ?? 0 }}</strong></div>
          <div class="md:col-span-2">Last Error: <strong>{{ queue.last_error || '-' }}</strong></div>
        </div>
      </CardBox>

      <CardBox class="mt-4">
        <div class="mb-3 flex flex-wrap items-end justify-between gap-3">
          <div>
            <div class="text-xs uppercase text-slate-500 mb-2">Riwayat Batch Terakhir</div>
            <label class="text-xs text-slate-500">Filter Type</label>
            <select v-model="selectedBatchType" class="mt-1 block rounded border border-slate-300 bg-white px-2 py-1 text-sm dark:border-slate-700 dark:bg-slate-900">
              <option value="all">Semua</option>
              <option value="exam_results_blast">exam_results_blast</option>
              <option value="announcement_blast">announcement_blast</option>
            </select>
            <label class="mt-2 block text-xs text-slate-500">Jumlah Riwayat</label>
            <select v-model.number="selectedHistoryLimit" class="mt-1 block rounded border border-slate-300 bg-white px-2 py-1 text-sm dark:border-slate-700 dark:bg-slate-900" @change="refresh">
              <option :value="20">20</option>
              <option :value="50">50</option>
              <option :value="100">100</option>
            </select>
          </div>
          <BaseButton :icon="mdiDeleteSweepOutline" color="danger" label="Clear History" :disabled="isLoading" @click="clearHistory" />
        </div>
        <div class="overflow-x-auto">
          <table class="w-full text-left text-sm">
            <thead class="border-b dark:border-slate-800 text-slate-500">
              <tr>
                <th class="py-2 pr-3">Batch</th>
                <th class="py-2 pr-3">Type</th>
                <th class="py-2 pr-3">Status</th>
                <th class="py-2 pr-3">Jobs</th>
                <th class="py-2 pr-3">Sukses</th>
                <th class="py-2 pr-3">Gagal</th>
                <th class="py-2 pr-3">Mulai</th>
                <th class="py-2 pr-3">Selesai</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="item in filteredQueueHistory" :key="item.batch_id" class="border-b dark:border-slate-800 last:border-0">
                <td class="py-2 pr-3 font-mono">#{{ item.batch_id }}</td>
                <td class="py-2 pr-3">{{ item.batch_type || '-' }}</td>
                <td class="py-2 pr-3">{{ item.status || '-' }}</td>
                <td class="py-2 pr-3">{{ item.total_jobs ?? 0 }}</td>
                <td class="py-2 pr-3">{{ item.success ?? 0 }}</td>
                <td class="py-2 pr-3">{{ item.failed ?? 0 }}</td>
                <td class="py-2 pr-3">{{ formatTime(item.started_at_utc) }}</td>
                <td class="py-2 pr-3">{{ formatTime(item.completed_at_utc) }}</td>
              </tr>
              <tr v-if="!filteredQueueHistory.length">
                <td colspan="8" class="py-3 text-slate-500">Belum ada riwayat batch.</td>
              </tr>
            </tbody>
          </table>
        </div>
      </CardBox>

      <CardBox class="mt-4">
        <div class="text-xs uppercase text-slate-500 mb-2">Runtime</div>
        <div class="grid gap-2 text-sm md:grid-cols-2">
          <div>Server Time: <strong>{{ formatTime(metrics.server_time_utc) }}</strong></div>
          <div>Last GC: <strong>{{ formatTime(metrics.last_gc_time_utc) }}</strong></div>
          <div>Goroutines: <strong>{{ metrics.goroutines ?? 0 }}</strong></div>
          <div>GOMAXPROCS / CPU Core: <strong>{{ metrics.gomaxprocs ?? 0 }} / {{ metrics.num_cpu ?? 0 }}</strong></div>
          <div>Heap Alloc: <strong>{{ formatBytes(metrics.heap_alloc_bytes) }}</strong></div>
          <div>Heap In Use: <strong>{{ formatBytes(metrics.heap_inuse_bytes) }}</strong></div>
          <div>Heap Sys: <strong>{{ formatBytes(metrics.heap_sys_bytes) }}</strong></div>
          <div>Stack In Use: <strong>{{ formatBytes(metrics.stack_inuse_bytes) }}</strong></div>
          <div>Total Alloc: <strong>{{ formatBytes(metrics.total_alloc_bytes) }}</strong></div>
          <div>System Memory: <strong>{{ formatBytes(metrics.sys_bytes) }}</strong></div>
        </div>
      </CardBox>
    </SectionMain>
  </LayoutAuthenticated>
</template>
