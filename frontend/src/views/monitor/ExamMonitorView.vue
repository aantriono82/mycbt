<script setup>
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { mdiContentCopy, mdiMonitorEye, mdiRefresh, mdiToggleSwitch, mdiToggleSwitchOffOutline, mdiWifi } from '@mdi/js'
import LayoutAuthenticated from '@/layouts/LayoutAuthenticated.vue'
import SectionMain from '@/components/SectionMain.vue'
import SectionTitleLineWithButton from '@/components/SectionTitleLineWithButton.vue'
import CardBox from '@/components/CardBox.vue'
import CardBoxWidget from '@/components/CardBoxWidget.vue'
import BaseButton from '@/components/BaseButton.vue'
import FormField from '@/components/FormField.vue'
import FormControl from '@/components/FormControl.vue'
import { api } from '@/services/api.js'
import { useAuthStore } from '@/stores/auth.js'

const route = useRoute()
const authStore = useAuthStore()

const vibrate = (pattern = 10) => {
  if (typeof navigator !== 'undefined' && navigator.vibrate) {
    navigator.vibrate(pattern)
  }
}

const exams = ref([])
const tokens = ref([])
const selectedExamId = ref('')
const isLoading = ref(false)
const errorMessage = ref('')
const successMessage = ref('')

const sessions = ref([])
const sessionFilter = ref('all')
const liveStatus = ref('idle')
const lastUpdatedAt = ref('')
const tokenActionId = ref('')
let pollHandle = 0
let es = null

const title = computed(() => route.meta?.title || 'Monitor Ujian')

const stats = computed(() => {
  const online = sessions.value.filter((item) => item.connection_status === 'online').length
  const working = sessions.value.filter((item) => item.status === 'in_progress').length
  const submitted = sessions.value.filter((item) => item.status === 'submitted').length
  return { online, working, submitted }
})

const filteredSessions = computed(() => {
  if (sessionFilter.value === 'all') return sessions.value
  if (sessionFilter.value === 'online') return sessions.value.filter((item) => item.connection_status === 'online')
  if (sessionFilter.value === 'in_progress') return sessions.value.filter((item) => item.status === 'in_progress')
  if (sessionFilter.value === 'submitted') return sessions.value.filter((item) => item.status === 'submitted')
  if (sessionFilter.value === 'forced') return sessions.value.filter((item) => item.status === 'forced')
  return sessions.value
})

const activeTokens = computed(() => tokens.value.filter((item) => item.is_active))

const loadSessions = async () => {
  if (!authStore.isAuthenticated || !selectedExamId.value) {
    sessions.value = []
    return
  }
  try {
    const { data } = await api.get(`/api/v1/exams/${selectedExamId.value}/monitor/sessions`, {
      params: { limit: 200, offset: 0 },
    })
    sessions.value = data?.data || []
    lastUpdatedAt.value = data?.meta?.server_time || ''
    errorMessage.value = ''
  } catch (error) {
    sessions.value = []
    errorMessage.value = error?.response?.data?.error?.message || 'Gagal memuat sesi peserta'
  }
}

const startPolling = () => {
  stopPolling()
  liveStatus.value = 'polling'
  pollHandle = window.setInterval(loadSessions, 10_000)
}

const stopPolling = () => {
  if (pollHandle) window.clearInterval(pollHandle)
  pollHandle = 0
}

const startStream = () => {
  stopStream()
  if (!authStore.isAuthenticated || !selectedExamId.value) return
  const token = authStore.token || ''
  if (!token) return

  const base = api.defaults?.baseURL || ''
  const query = new URLSearchParams({
    view: 'sessions',
    access_token: token,
  })
  const url = `${base}/api/v1/exams/${selectedExamId.value}/monitor/stream?${query.toString()}`
  try {
    liveStatus.value = 'connecting'
    es = new EventSource(url)
  } catch {
    es = null
    liveStatus.value = 'polling'
    return
  }

  es.addEventListener('hello', () => {
    liveStatus.value = 'live'
  })

  es.addEventListener('snapshot', (evt) => {
    try {
      const payload = JSON.parse(evt.data || '{}')
      sessions.value = payload?.data || []
      lastUpdatedAt.value = payload?.meta?.server_time || ''
      errorMessage.value = ''
      liveStatus.value = 'live'
    } catch {
      // ignore
    }
  })

  es.addEventListener('error', () => {
    // Fallback to polling.
    liveStatus.value = 'polling'
    stopStream()
    startPolling()
  })
}

const stopStream = () => {
  if (es) es.close()
  es = null
}

const refreshMonitor = async () => {
  await loadExams()
  await loadTokens()
  startStream()
  if (!es) startPolling()
}

const toggleToken = async (token) => {
  if (!token?.id) return
  errorMessage.value = ''
  successMessage.value = ''
  try {
    tokenActionId.value = token.id
    await api.patch(`/api/v1/tokens/${token.id}`, {
      is_active: !token.is_active,
    })
    vibrate(10)
    successMessage.value = `Token ${!token.is_active ? 'diaktifkan' : 'dinonaktifkan'}`
    await loadTokens()
  } catch (error) {
    errorMessage.value = error?.response?.data?.error?.message || 'Gagal mengubah status token'
  } finally {
    tokenActionId.value = ''
  }
}

const copyToken = async (token) => {
  if (!token?.token) return
  errorMessage.value = ''
  successMessage.value = ''
  try {
    await navigator.clipboard.writeText(token.token)
    vibrate(5)
    successMessage.value = `Token ${token.token} disalin`
  } catch {
    errorMessage.value = 'Gagal menyalin token ke clipboard'
  }
}

const loadExams = async () => {
  if (!authStore.isAuthenticated) return
  isLoading.value = true
  errorMessage.value = ''
  try {
    const { data } = await api.get('/api/v1/exams', {
      params: { limit: 100, offset: 0 },
    })
    exams.value = data?.data || []
    if (!selectedExamId.value && exams.value.length) {
      selectedExamId.value = exams.value[0].id
    }
  } catch (error) {
    errorMessage.value = error?.response?.data?.error?.message || 'Gagal memuat daftar ujian'
  } finally {
    isLoading.value = false
  }
}

const loadTokens = async () => {
  if (!authStore.isAuthenticated || !selectedExamId.value) {
    tokens.value = []
    sessions.value = []
    return
  }
  try {
    const { data } = await api.get(`/api/v1/exams/${selectedExamId.value}/tokens`, {
      params: { limit: 20, offset: 0 },
    })
    tokens.value = data?.data || []
    await loadSessions()
  } catch {
    tokens.value = []
    sessions.value = []
  }
}

watch(selectedExamId, async () => {
  await loadTokens()
  startStream()
  if (!es) startPolling()
})

onMounted(async () => {
  await loadExams()
  await loadTokens()
  startStream()
  if (!es) startPolling()
})

onBeforeUnmount(() => {
  stopPolling()
  stopStream()
})
</script>

<template>
  <LayoutAuthenticated>
    <SectionMain>
      <SectionTitleLineWithButton :icon="mdiMonitorEye" :title="title" main>
        <BaseButton :icon="mdiRefresh" color="info" label="Refresh Monitor" @click="refreshMonitor" />
      </SectionTitleLineWithButton>

      <CardBox class="mb-6" color="blue">
        <div class="grid gap-4 md:grid-cols-[280px_220px_1fr] md:items-start">
          <FormField label="Pilih Ujian">
            <FormControl
              v-model="selectedExamId"
              :options="exams.map((item) => ({ value: item.id, label: item.title }))"
            />
          </FormField>
          <FormField label="Filter Sesi">
            <FormControl
              v-model="sessionFilter"
              :options="[
                { value: 'all', label: 'Semua' },
                { value: 'online', label: 'Online' },
                { value: 'in_progress', label: 'In Progress' },
                { value: 'submitted', label: 'Submitted' },
                { value: 'forced', label: 'Forced' },
              ]"
            />
          </FormField>
          <div class="text-sm text-slate-600 dark:text-slate-400 md:pt-9">
            Monitor memakai <span class="text-sky-600 dark:text-sky-400 font-semibold">SSE</span> saat tersedia, lalu fallback ke polling 10 detik jika koneksi live gagal.
          </div>
        </div>
        <div class="mt-3 text-sm text-slate-500 dark:text-slate-500 italic">
          Status koneksi: <span class="font-bold uppercase tracking-tighter" :class="liveStatus === 'live' ? 'text-emerald-600 dark:text-emerald-400' : 'text-slate-500'">{{ liveStatus === 'live' ? 'live SSE' : liveStatus === 'connecting' ? 'menghubungkan...' : liveStatus === 'polling' ? 'polling fallback' : 'idle' }}</span>
          <span v-if="lastUpdatedAt"> · Update terakhir: {{ lastUpdatedAt }}</span>
          <span class="dark:text-slate-400"> · Menampilkan <span class="font-bold">{{ filteredSessions.length }}</span> dari <span class="font-bold">{{ sessions.length }}</span> sesi</span>
        </div>
      </CardBox>

      <div v-if="!authStore.isAuthenticated" class="mb-6 rounded-xl bg-amber-50 dark:bg-amber-900/20 px-4 py-3 text-sm text-amber-700 dark:text-amber-400 border border-amber-100 dark:border-amber-900/40">
        Login terlebih dulu agar monitor ujian dapat memuat data backend.
      </div>
      <div v-else-if="errorMessage" class="mb-6 rounded-xl bg-red-50 dark:bg-red-900/20 px-4 py-3 text-sm text-red-700 dark:text-red-400 border border-red-100 dark:border-red-900/40">
        {{ errorMessage }}
      </div>
      <div v-if="successMessage" class="mb-6 rounded-xl bg-emerald-50 dark:bg-emerald-900/20 px-4 py-3 text-sm text-emerald-700 dark:text-emerald-400 border border-emerald-100 dark:border-emerald-900/40">
        {{ successMessage }}
      </div>

      <div class="mb-6 grid gap-6 md:grid-cols-3">
        <CardBoxWidget :icon="mdiWifi" color="emerald" label="Peserta Online" :number="stats.online" />
        <CardBoxWidget :icon="mdiMonitorEye" color="sky" label="Sedang Mengerjakan" :number="stats.working" />
        <CardBoxWidget :icon="mdiRefresh" color="amber" label="Sudah Submit" :number="stats.submitted" />
      </div>

      <div class="grid gap-6 xl:grid-cols-[minmax(0,1fr)_340px]">
        <CardBox color="purple">
          <h3 class="mb-4 text-lg font-semibold dark:text-slate-100">Aktivitas Peserta</h3>
          <div class="overflow-x-auto -mx-6 sm:mx-0">
            <!-- Desktop Table -->
            <table class="hidden sm:table w-full text-left text-sm">
              <thead class="border-b dark:border-slate-800 bg-slate-50 dark:bg-slate-800/50 text-slate-600 dark:text-slate-300 uppercase text-xs tracking-wider font-bold">
	                <tr>
	                  <th class="px-6 py-4">Peserta</th>
	                  <th class="px-3 py-4 text-center">Koneksi</th>
	                  <th class="px-3 py-4 text-center">Status</th>
	                  <th class="px-3 py-4 text-center">Progress</th>
	                  <th class="px-6 py-4">Last Seen</th>
	                </tr>
              </thead>
              <tbody>
                <tr v-for="item in filteredSessions" :key="item.session_id" class="border-b dark:border-slate-800 last:border-b-0 hover:bg-slate-50/50 dark:hover:bg-slate-800/30 transition-colors">
	                  <td class="px-6 py-4">
	                    <div class="font-bold dark:text-slate-100">{{ item.student_name }}</div>
	                    <div class="text-slate-500 dark:text-slate-400 text-xs">{{ item.student_username }}</div>
	                  </td>
	                  <td class="px-3 py-4 text-center">
	                    <div class="flex flex-col items-center gap-1">
	                      <span
	                        class="rounded-full px-2 py-0.5 text-[10px] font-black uppercase tracking-widest"
	                        :class="item.connection_status === 'online' ? 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400' : 'bg-slate-100 text-slate-600 dark:bg-slate-800 dark:text-slate-400'"
	                      >
	                        {{ item.connection_status }}
	                      </span>
	                      <span v-if="item.warning_count > 0" class="flex items-center gap-0.5 px-1.5 py-0.5 rounded bg-red-100 text-red-700 dark:bg-red-900/40 dark:text-red-400 text-[9px] font-black animate-pulse">
	                         ⚠️ {{ item.warning_count }}
	                      </span>
	                    </div>
	                  </td>
	                  <td class="px-3 py-4 text-center">
                      <span class="text-[10px] font-black uppercase tracking-widest px-2 py-0.5 rounded-lg border dark:border-slate-800">
                        {{ item.status }}
                      </span>
                    </td>
	                  <td class="px-3 py-4 text-center font-mono font-black text-blue-600 dark:text-sky-400 text-lg">{{ item.progress_percent }}%</td>
	                  <td class="px-6 py-4 text-xs text-slate-500 dark:text-slate-500">{{ item.last_seen_at || '-' }}</td>
	                </tr>
              </tbody>
            </table>

            <!-- Mobile Card View -->
            <div class="sm:hidden divide-y divide-slate-100 dark:divide-slate-800">
              <div v-for="item in filteredSessions" :key="item.session_id" class="p-4 flex flex-col gap-3">
                <div class="flex items-start justify-between gap-3">
                  <div>
                    <div class="font-black text-slate-900 dark:text-white uppercase tracking-tighter">{{ item.student_name }}</div>
                    <div class="text-slate-400 text-[10px] font-bold uppercase tracking-widest">{{ item.student_username }}</div>
                  </div>
                  <div class="flex flex-col items-end gap-1.5">
                    <span
                      class="rounded-lg px-2 py-1 text-[9px] font-black uppercase tracking-widest shadow-sm"
                      :class="item.connection_status === 'online' ? 'bg-emerald-500 text-white' : 'bg-slate-200 text-slate-600 dark:bg-slate-800 dark:text-slate-400'"
                    >
                      {{ item.connection_status }}
                    </span>
                    <span v-if="item.warning_count > 0" class="px-1.5 py-0.5 rounded bg-red-500 text-white text-[9px] font-black animate-pulse">
                       ⚠️ {{ item.warning_count }} WARN
                    </span>
                  </div>
                </div>
                
                <div class="flex items-center justify-between bg-slate-50 dark:bg-slate-800/50 rounded-xl p-3 border border-slate-100 dark:border-slate-800">
                  <div class="flex flex-col">
                    <span class="text-[9px] font-black text-slate-400 uppercase tracking-widest leading-none mb-1">Status</span>
                    <span class="text-[10px] font-black uppercase tracking-widest text-slate-600 dark:text-slate-300">{{ item.status }}</span>
                  </div>
                  <div class="flex flex-col items-center">
                    <span class="text-[9px] font-black text-slate-400 uppercase tracking-widest leading-none mb-1">Progress</span>
                    <span class="font-mono font-black text-blue-600 dark:text-sky-400 text-lg">{{ item.progress_percent }}%</span>
                  </div>
                  <div class="flex flex-col items-end">
                    <span class="text-[9px] font-black text-slate-400 uppercase tracking-widest leading-none mb-1">Last Seen</span>
                    <span class="text-[10px] font-bold text-slate-500">{{ item.last_seen_at?.split(' ')[1] || '-' }}</span>
                  </div>
                </div>
              </div>
              <div v-if="!filteredSessions.length" class="py-10 text-center text-slate-400 italic text-sm">
                Belum ada sesi peserta.
              </div>
            </div>
          </div>
        </CardBox>

        <CardBox color="blue">
          <h3 class="mb-4 text-lg font-semibold dark:text-slate-100">Token Ujian</h3>
          <div v-if="activeTokens.length" class="mb-4 rounded-xl bg-emerald-50 dark:bg-emerald-900/20 px-4 py-3 border border-emerald-100 dark:border-emerald-900/40">
            <div class="text-[10px] font-bold uppercase tracking-widest text-emerald-700 dark:text-emerald-400">Token Aktif</div>
            <div class="mt-1 flex items-center gap-3">
              <div class="font-mono text-2xl font-black text-emerald-900 dark:text-emerald-200">{{ activeTokens[0].token }}</div>
              <BaseButton
                :icon="mdiContentCopy"
                color="success"
                small
                outline
                label="Salin"
                @click="copyToken(activeTokens[0])"
              />
            </div>
          </div>
          <div class="space-y-3">
            <div
              v-for="token in tokens"
              :key="token.id"
              class="rounded-xl border border-slate-200 dark:border-slate-800 px-4 py-3 bg-slate-50/30 dark:bg-slate-800/20"
            >
              <div class="mb-1 font-mono font-bold text-slate-700 dark:text-slate-200">{{ token.token }}</div>
              <div class="text-xs text-slate-600 dark:text-slate-400">
                Aktif: <span :class="token.is_active ? 'text-emerald-600 dark:text-emerald-400' : 'text-red-500'">{{ token.is_active ? 'Ya' : 'Tidak' }}</span>
              </div>
              <div class="text-xs text-slate-500 dark:text-slate-500 italic mt-1 leading-tight">
                Window: {{ token.valid_from || '-' }} s/d {{ token.valid_to || '-' }}
              </div>
              <div class="mt-3 flex flex-wrap gap-2">
                <BaseButton
                  :icon="mdiContentCopy"
                  color="success"
                  small
                  outline
                  label="Copy"
                  @click="copyToken(token)"
                />
                <BaseButton
                  :icon="token.is_active ? mdiToggleSwitchOffOutline : mdiToggleSwitch"
                  color="info"
                  small
                  outline
                  :label="token.is_active ? 'Off' : 'On'"
                  :disabled="tokenActionId === token.id"
                  @click="toggleToken(token)"
                />
              </div>
            </div>
            <div v-if="!tokens.length" class="text-sm text-slate-500 dark:text-slate-500 italic">
              Belum ada token untuk ujian ini.
            </div>
          </div>
        </CardBox>
      </div>
    </SectionMain>
  </LayoutAuthenticated>
</template>
