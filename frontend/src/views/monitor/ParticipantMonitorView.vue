<script setup>
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { mdiAccountSearchOutline, mdiRefresh, mdiAccountCheck, mdiAccountAlert } from '@mdi/js'
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

const authStore = useAuthStore()

const exams = ref([])
const selectedExamId = ref('')
const isLoading = ref(false)
const isLoadingParticipants = ref(false)
const errorMessage = ref('')
const successMessage = ref('')
const actionSessionId = ref('')
const batchAction = ref('')

const q = ref('')
const appliedQuery = ref('')
const participantFilter = ref('all')
const participants = ref([])
const liveStatus = ref('idle')
const lastUpdatedAt = ref('')
let pollHandle = 0
let es = null

const stats = computed(() => ({
  total: participants.value.length,
  online: participants.value.filter((item) => item.connection_status === 'online').length,
  blocked: participants.value.filter((item) => item.connection_status === 'blocked').length,
}))

const filteredParticipants = computed(() => {
  if (participantFilter.value === 'all') return participants.value
  if (participantFilter.value === 'online') return participants.value.filter((item) => item.connection_status === 'online')
  if (participantFilter.value === 'blocked') return participants.value.filter((item) => item.connection_status === 'blocked')
  if (participantFilter.value === 'in_progress') return participants.value.filter((item) => item.session_status === 'in_progress')
  if (participantFilter.value === 'forced') return participants.value.filter((item) => item.session_status === 'forced')
  if (participantFilter.value === 'submitted') return participants.value.filter((item) => item.session_status === 'submitted')
  if (participantFilter.value === 'not_joined') return participants.value.filter((item) => !item.session_status)
  return participants.value
})

const bulkForceSubmitTargets = computed(() =>
  filteredParticipants.value.filter((item) => item.session_id && item.session_status === 'in_progress'),
)

const bulkResetTargets = computed(() =>
  filteredParticipants.value.filter(
    (item) =>
      item.session_id &&
      item.session_status !== 'submitted' &&
      item.session_status !== 'forced',
  ),
)

const loadExams = async () => {
  if (!authStore.isAuthenticated) return
  isLoading.value = true
  errorMessage.value = ''
  try {
    const { data } = await api.get('/api/v1/exams', {
      params: { limit: 20, offset: 0 },
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

const loadParticipants = async () => {
  if (!authStore.isAuthenticated || !selectedExamId.value) {
    participants.value = []
    return
  }
  isLoadingParticipants.value = true
  try {
    const { data } = await api.get(`/api/v1/exams/${selectedExamId.value}/monitor/participants`, {
      params: { q: appliedQuery.value, limit: 300, offset: 0 },
    })
    participants.value = data?.data || []
    lastUpdatedAt.value = data?.meta?.server_time || ''
    errorMessage.value = ''
  } catch (error) {
    participants.value = []
    errorMessage.value = error?.response?.data?.error?.message || 'Gagal memuat peserta'
  } finally {
    isLoadingParticipants.value = false
  }
}

const startPolling = () => {
  stopPolling()
  liveStatus.value = 'polling'
  pollHandle = window.setInterval(loadParticipants, 10_000)
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
    view: 'participants',
    access_token: token,
  })
  if (appliedQuery.value) {
    query.set('q', appliedQuery.value)
  }
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
      participants.value = payload?.data || []
      lastUpdatedAt.value = payload?.meta?.server_time || ''
      errorMessage.value = ''
      liveStatus.value = 'live'
    } catch {
      // ignore
    }
  })

  es.addEventListener('error', () => {
    liveStatus.value = 'polling'
    stopStream()
    startPolling()
  })
}

const stopStream = () => {
  if (es) es.close()
  es = null
}

const forceSubmitSession = async (item) => {
  if (!item?.session_id || item.session_status !== 'in_progress') return
  errorMessage.value = ''
  successMessage.value = ''
  const ok = window.confirm(
    `Force submit untuk ${item.student_name}?\nSesi aktif akan ditutup dan hasil otomatis dihitung dari jawaban yang sudah tersimpan.`,
  )
  if (!ok) return

  try {
    actionSessionId.value = item.session_id
    await api.post(`/api/v1/exams/${selectedExamId.value}/sessions/${item.session_id}/force-submit`, {})
    successMessage.value = 'Force submit berhasil'
    await loadParticipants()
    startStream()
    if (!es) startPolling()
  } catch (error) {
    errorMessage.value = error?.response?.data?.error?.message || 'Gagal force submit'
  } finally {
    actionSessionId.value = ''
  }
}

const resetSession = async (item) => {
  if (!item?.session_id) return
  errorMessage.value = ''
  successMessage.value = ''
  const ok = window.confirm(`Reset login untuk ${item.student_name}?\nSession akan dihapus agar siswa bisa join ulang.`)
  if (!ok) return

  try {
    actionSessionId.value = item.session_id
    await api.post(`/api/v1/exams/${selectedExamId.value}/sessions/${item.session_id}/reset`, {})
    successMessage.value = 'Reset login berhasil'
    await loadParticipants()
    startStream()
    if (!es) startPolling()
  } catch (error) {
    errorMessage.value = error?.response?.data?.error?.message || 'Gagal reset login'
  } finally {
    actionSessionId.value = ''
  }
}

const runBatchAction = async (kind) => {
  const targets = kind === 'force-submit' ? bulkForceSubmitTargets.value : bulkResetTargets.value
  if (!targets.length) return

  errorMessage.value = ''
  successMessage.value = ''

  const label = kind === 'force-submit' ? 'force submit' : 'reset login'
  const ok = window.confirm(
    `${label.toUpperCase()} untuk ${targets.length} peserta hasil filter saat ini?\nAksi ini akan dijalankan satu per satu.`,
  )
  if (!ok) return

  batchAction.value = kind
  let successCount = 0
  let failedCount = 0

  try {
    for (const item of targets) {
      try {
        const path =
          kind === 'force-submit'
            ? `/api/v1/exams/${selectedExamId.value}/sessions/${item.session_id}/force-submit`
            : `/api/v1/exams/${selectedExamId.value}/sessions/${item.session_id}/reset`
        await api.post(path, {})
        successCount += 1
      } catch {
        failedCount += 1
      }
    }

    if (failedCount > 0) {
      errorMessage.value = `${label} selesai sebagian: ${successCount} berhasil, ${failedCount} gagal`
    } else {
      successMessage.value = `${label} berhasil untuk ${successCount} peserta`
    }
    await loadParticipants()
    startStream()
    if (!es) startPolling()
  } finally {
    batchAction.value = ''
  }
}

const applyFilter = async () => {
  appliedQuery.value = q.value.trim()
  await loadParticipants()
  startStream()
  if (!es) startPolling()
}

const refreshMonitor = async () => {
  await loadExams()
  await loadParticipants()
  startStream()
  if (!es) startPolling()
}

watch(selectedExamId, async () => {
  await loadParticipants()
  startStream()
  if (!es) startPolling()
})

onMounted(async () => {
  await loadExams()
  await loadParticipants()
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
      <SectionTitleLineWithButton :icon="mdiAccountSearchOutline" title="Monitor Peserta" main>
        <BaseButton :icon="mdiRefresh" color="info" label="Refresh Monitor" @click="refreshMonitor" />
      </SectionTitleLineWithButton>

      <div v-if="!authStore.isAuthenticated" class="mb-6 rounded-xl bg-amber-50 dark:bg-amber-900/20 px-4 py-3 text-sm text-amber-700 dark:text-amber-400 border border-amber-100 dark:border-amber-900/40">
        Login terlebih dulu agar monitor peserta dapat memuat data backend.
      </div>
      <div v-else-if="errorMessage" class="mb-6 rounded-xl bg-red-50 dark:bg-red-900/20 px-4 py-3 text-sm text-red-700 dark:text-red-400 border border-red-100 dark:border-red-900/40">
        {{ errorMessage }}
      </div>
      <div v-if="successMessage" class="mb-6 rounded-xl bg-emerald-50 dark:bg-emerald-900/20 px-4 py-3 text-sm text-emerald-700 dark:text-emerald-400 border border-emerald-100 dark:border-emerald-900/40">
        {{ successMessage }}
      </div>

      <div class="mb-6 grid gap-6 md:grid-cols-3">
        <CardBoxWidget :icon="mdiAccountSearchOutline" color="text-sky-500" label="Total Peserta" :number="stats.total" />
        <CardBoxWidget :icon="mdiAccountCheck" color="text-emerald-500" label="Online" :number="stats.online" />
        <CardBoxWidget :icon="mdiAccountAlert" color="text-amber-500" label="Perlu Tindakan" :number="stats.blocked" />
      </div>

      <div class="mb-6 grid gap-6 md:grid-cols-3 md:items-stretch">
        <CardBox class="mb-0">
          <FormField label="Pilih Ujian" class="mb-0">
            <FormControl
              v-model="selectedExamId"
              :options="exams.map((item) => ({ value: item.id, label: item.title }))"
            />
          </FormField>
        </CardBox>
        <CardBox class="mb-0">
          <FormField label="Filter Operasional" class="mb-0">
            <FormControl
              v-model="participantFilter"
              :options="[
                { value: 'all', label: 'Semua' },
                { value: 'online', label: 'Online' },
                { value: 'in_progress', label: 'In Progress' },
                { value: 'blocked', label: 'Belum Join' },
                { value: 'submitted', label: 'Submitted' },
                { value: 'forced', label: 'Forced' },
                { value: 'not_joined', label: 'Not Joined' },
              ]"
            />
          </FormField>
        </CardBox>
        <CardBox class="mb-0">
          <FormField label="Cari (nama/username/nis)" class="mb-0">
            <FormControl v-model="q" placeholder="Ketik lalu klik Terapkan" />
          </FormField>
        </CardBox>
      </div>

      <CardBox class="mb-6">
        <div class="flex flex-wrap items-center gap-3">
          <BaseButton color="whiteDark" outline label="Terapkan" :disabled="isLoadingParticipants" @click="applyFilter" />
          <BaseButton
            color="warning"
            outline
            label="Force Submit Terfilter"
            :disabled="isLoadingParticipants || !!batchAction || !bulkForceSubmitTargets.length"
            @click="runBatchAction('force-submit')"
          />
          <BaseButton
            color="info"
            outline
            label="Reset Terfilter"
            :disabled="isLoadingParticipants || !!batchAction || !bulkResetTargets.length"
            @click="runBatchAction('reset')"
          />
          <div class="ml-auto text-sm text-slate-500 dark:text-slate-500 italic">
            Status: <span class="uppercase font-bold tracking-tighter" :class="liveStatus === 'live' ? 'text-emerald-600 dark:text-emerald-400' : 'text-slate-400'">{{ liveStatus === 'live' ? 'live SSE' : liveStatus === 'connecting' ? 'menghubungkan...' : liveStatus === 'polling' ? 'polling fallback' : 'idle' }}</span>
            <span v-if="lastUpdatedAt"> · {{ lastUpdatedAt }}</span>
            <span class="dark:text-slate-400"> · <span class="font-bold">{{ filteredParticipants.length }}</span>/{{ participants.length }} peserta</span>
          </div>
        </div>
        <div v-if="isLoadingParticipants" class="mt-2 text-xs text-info italic">Memuat data peserta terbaru...</div>
      </CardBox>

      <div class="grid gap-6 xl:grid-cols-[minmax(0,1fr)_340px]">
        <CardBox>
          <h3 class="mb-4 text-lg font-semibold dark:text-slate-100">Status Peserta</h3>
          <div class="overflow-x-auto">
            <table class="w-full text-left text-sm">
              <thead class="border-b dark:border-slate-800 bg-slate-50 dark:bg-slate-800/50 text-slate-600 dark:text-slate-300 uppercase text-xs tracking-wider font-bold">
                <tr>
                  <th class="px-3 py-3">Nama</th>
                  <th class="px-3 py-3 text-center">Status</th>
                  <th class="px-3 py-3 text-center">Progress</th>
                  <th class="px-3 py-3">Last Seen</th>
                  <th class="px-3 py-3 text-center">Aksi</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="item in filteredParticipants" :key="item.student_id" class="border-b dark:border-slate-800 last:border-b-0 hover:bg-slate-50/50 dark:hover:bg-slate-800/30 transition-colors">
                  <td class="px-3 py-3">
                    <div class="font-medium dark:text-slate-100">{{ item.student_name }}</div>
                    <div class="text-[10px] font-mono text-slate-500 dark:text-slate-400 italic">{{ item.student_username }} · {{ item.student_nis }}</div>
                  </td>
                  <td class="px-3 py-3 text-center">
                    <span
                      class="rounded-full px-2 py-0.5 text-[10px] font-bold uppercase tracking-tight"
                      :class="
                        item.connection_status === 'online'
                          ? 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400'
                          : item.connection_status === 'blocked'
                            ? 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400'
                            : 'bg-slate-100 text-slate-600 dark:bg-slate-800 dark:text-slate-400'
                      "
                    >
                      {{ item.connection_status }}
                    </span>
                    <div class="mt-1 text-[10px] font-bold uppercase text-slate-400 dark:text-slate-500">
                      {{ item.session_status || 'NOT JOINED' }}
                    </div>
                  </td>
                  <td class="px-3 py-3 text-center">
                    <div class="font-bold text-info dark:text-sky-400">{{ item.progress_percent }}%</div>
                    <div class="text-[10px] text-slate-500 dark:text-slate-500 font-mono">({{ item.answered_questions }}/{{ item.total_questions }})</div>
                  </td>
                  <td class="px-3 py-3 text-xs text-slate-500 dark:text-slate-500 italic">{{ item.last_seen_at || '-' }}</td>
                  <td class="px-3 py-3">
                    <div class="flex flex-wrap gap-2">
                      <BaseButton
                        color="warning"
                        small
                        outline
                        label="Force Submit"
                        :disabled="!item.session_id || item.session_status !== 'in_progress' || actionSessionId === item.session_id || !!batchAction"
                        @click="forceSubmitSession(item)"
                      />
                      <BaseButton
                        color="info"
                        small
                        outline
                        label="Reset"
                        :disabled="!item.session_id || item.session_status === 'submitted' || item.session_status === 'forced' || actionSessionId === item.session_id || !!batchAction"
                        @click="resetSession(item)"
                      />
                    </div>
                  </td>
                </tr>
                <tr v-if="!filteredParticipants.length">
                  <td colspan="5" class="px-3 py-10 text-center text-slate-400 dark:text-slate-500 italic">Belum ada peserta target untuk ujian ini.</td>
                </tr>
              </tbody>
            </table>
          </div>
        </CardBox>

        <CardBox>
          <h3 class="mb-4 text-lg font-semibold dark:text-slate-100">Ujian Aktif</h3>
          <div v-if="isLoading" class="text-sm text-slate-500 dark:text-slate-400 italic">Memuat daftar ujian...</div>
          <div v-else class="space-y-3">
            <div
              v-for="exam in exams.slice(0, 6)"
              :key="exam.id"
              class="rounded-xl border border-slate-200 dark:border-slate-800 px-4 py-3 bg-slate-50/30 dark:bg-slate-800/20"
            >
              <div class="font-semibold dark:text-slate-200">{{ exam.title }}</div>
              <div class="text-[10px] font-mono text-slate-500 dark:text-slate-400 italic break-all leading-tight">{{ exam.starts_at }}</div>
            </div>
            <div v-if="!exams.length" class="text-sm text-slate-400 dark:text-slate-500 italic">
              Belum ada ujian yang bisa dimonitor.
            </div>
          </div>
        </CardBox>
      </div>
    </SectionMain>
  </LayoutAuthenticated>
</template>
