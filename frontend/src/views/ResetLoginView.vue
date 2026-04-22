<script setup>
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { mdiAccountSwitchOutline, mdiRefresh } from '@mdi/js'
import LayoutAuthenticated from '@/layouts/LayoutAuthenticated.vue'
import SectionMain from '@/components/SectionMain.vue'
import SectionTitleLineWithButton from '@/components/SectionTitleLineWithButton.vue'
import CardBox from '@/components/CardBox.vue'
import BaseButton from '@/components/BaseButton.vue'
import FormField from '@/components/FormField.vue'
import FormControl from '@/components/FormControl.vue'
import { api } from '@/services/api.js'
import { useAuthStore } from '@/stores/auth.js'

const authStore = useAuthStore()

const exams = ref([])
const selectedExamId = ref('')
const q = ref('')

const participants = ref([])
const isLoading = ref(false)
const isLoadingParticipants = ref(false)
const errorMessage = ref('')
const successMessage = ref('')
const actionSessionId = ref('')

let pollHandle = 0

const canLoad = computed(() => authStore.isAuthenticated)

const stats = computed(() => {
  const total = participants.value.length
  const blocked = participants.value.filter((x) => x.connection_status === 'blocked').length
  const online = participants.value.filter((x) => x.connection_status === 'online').length
  const inProgress = participants.value.filter((x) => x.session_status === 'in_progress').length
  return { total, blocked, online, inProgress }
})

const loadExams = async () => {
  if (!canLoad.value) return
  isLoading.value = true
  errorMessage.value = ''
  try {
    const { data } = await api.get('/api/v1/exams', { params: { limit: 200, offset: 0 } })
    exams.value = data?.data || []
    if (!selectedExamId.value && exams.value.length) selectedExamId.value = exams.value[0].id
  } catch (error) {
    exams.value = []
    errorMessage.value = error?.response?.data?.error?.message || 'Gagal memuat daftar ujian'
  } finally {
    isLoading.value = false
  }
}

const loadParticipants = async () => {
  if (!canLoad.value || !selectedExamId.value) {
    participants.value = []
    return
  }
  isLoadingParticipants.value = true
  errorMessage.value = ''
  try {
    const { data } = await api.get(`/api/v1/exams/${selectedExamId.value}/monitor/participants`, {
      params: { q: q.value, limit: 500, offset: 0 },
    })
    participants.value = data?.data || []
  } catch (error) {
    participants.value = []
    errorMessage.value = error?.response?.data?.error?.message || 'Gagal memuat peserta'
  } finally {
    isLoadingParticipants.value = false
  }
}

const startPolling = () => {
  stopPolling()
  pollHandle = window.setInterval(loadParticipants, 10_000)
}

const stopPolling = () => {
  if (pollHandle) window.clearInterval(pollHandle)
  pollHandle = 0
}

const resetSession = async (row) => {
  if (!row.session_id) return
  successMessage.value = ''
  errorMessage.value = ''
  const ok = window.confirm(`Reset login untuk ${row.student_name}?\nSession akan dihapus agar siswa bisa join ulang.`)
  if (!ok) return
  try {
    actionSessionId.value = row.session_id
    await api.post(`/api/v1/exams/${selectedExamId.value}/sessions/${row.session_id}/reset`, {})
    successMessage.value = 'Reset login berhasil'
    await loadParticipants()
  } catch (error) {
    errorMessage.value = error?.response?.data?.error?.message || 'Gagal reset login'
  } finally {
    actionSessionId.value = ''
  }
}

const forceSubmitSession = async (row) => {
  if (!row.session_id) return
  successMessage.value = ''
  errorMessage.value = ''
  const ok = window.confirm(
    `Force submit untuk ${row.student_name}?\nSesi aktif akan ditutup dan hasil otomatis dihitung dari jawaban yang sudah tersimpan.`,
  )
  if (!ok) return
  try {
    actionSessionId.value = row.session_id
    await api.post(`/api/v1/exams/${selectedExamId.value}/sessions/${row.session_id}/force-submit`, {})
    successMessage.value = 'Force submit berhasil'
    await loadParticipants()
  } catch (error) {
    errorMessage.value = error?.response?.data?.error?.message || 'Gagal force submit'
  } finally {
    actionSessionId.value = ''
  }
}

watch(selectedExamId, loadParticipants)

onMounted(async () => {
  await loadExams()
  await loadParticipants()
  startPolling()
})

onBeforeUnmount(() => {
  stopPolling()
})
</script>

<template>
  <LayoutAuthenticated>
    <SectionMain>
      <SectionTitleLineWithButton :icon="mdiAccountSwitchOutline" title="Reset Login" main>
        <BaseButton :icon="mdiRefresh" color="info" label="Refresh" @click="loadExams(); loadParticipants()" />
      </SectionTitleLineWithButton>

      <div v-if="!authStore.isAuthenticated" class="mb-6 rounded-xl bg-amber-50 dark:bg-amber-900/20 px-4 py-3 text-sm text-amber-700 dark:text-amber-400 border border-amber-100 dark:border-amber-900/40">
        Login terlebih dulu agar reset login dapat memuat data backend.
      </div>
      <div v-else-if="errorMessage" class="mb-6 rounded-xl bg-red-50 dark:bg-red-900/20 px-4 py-3 text-sm text-red-700 dark:text-red-400 border border-red-100 dark:border-red-900/40">
        {{ errorMessage }}
      </div>
      <div v-if="successMessage" class="mb-6 rounded-xl bg-emerald-50 dark:bg-emerald-900/20 px-4 py-3 text-sm text-emerald-700 dark:text-emerald-400 border border-emerald-100 dark:border-emerald-900/40">
        {{ successMessage }}
      </div>

      <CardBox class="mb-6">
        <div class="rounded-xl bg-sky-50 dark:bg-sky-900/20 px-4 py-4 text-sm text-sky-900 dark:text-sky-300 border border-sky-100 dark:border-sky-900/40">
          Operasi cepat sekarang tersedia langsung di menu <code class="font-bold dark:text-sky-400">Monitor Peserta</code>.
          Halaman ini tetap dipertahankan untuk operator yang ingin fokus khusus pada reset login dan force submit tanpa panel monitor lain.
        </div>
      </CardBox>

      <div class="mb-6 grid gap-6 md:grid-cols-2 md:items-stretch">
        <CardBox class="mb-0">
          <FormField label="Pilih Ujian" class="mb-0">
            <FormControl
              v-model="selectedExamId"
              :options="exams.map((item) => ({ value: item.id, label: item.title }))"
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
          <BaseButton color="whiteDark" outline label="Terapkan" :disabled="isLoadingParticipants" @click="loadParticipants" />
          <div v-if="isLoading || isLoadingParticipants" class="text-sm text-slate-500 dark:text-slate-400 italic font-mono animate-pulse">Memuat data...</div>
          <div class="ml-auto text-xs uppercase font-bold tracking-widest text-slate-400 dark:text-slate-500">
            TOTAL {{ stats.total }} · ONLINE {{ stats.online }} · PROGRESS {{ stats.inProgress }} · BELUM JOIN {{ stats.blocked }}
          </div>
        </div>
      </CardBox>

      <CardBox>
        <div class="overflow-x-auto">
          <table class="w-full text-left text-sm">
            <thead class="border-b dark:border-slate-800 bg-slate-50 dark:bg-slate-800/50 text-slate-600 dark:text-slate-300 uppercase text-xs tracking-wider font-bold">
              <tr>
                <th class="px-3 py-3">Siswa</th>
                <th class="px-3 py-3 text-center">Status</th>
                <th class="px-3 py-3 text-center">Progress</th>
                <th class="px-3 py-3">Last Seen</th>
                <th class="px-3 py-3 text-center">Aksi Operator</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="row in participants" :key="row.student_id" class="border-b dark:border-slate-800 last:border-b-0 hover:bg-slate-50/50 dark:hover:bg-slate-800/30 transition-colors">
                <td class="px-3 py-3 font-medium">
                  <div class="dark:text-slate-100">{{ row.student_name }}</div>
                  <div class="text-[10px] font-mono text-slate-500 dark:text-slate-400 italic">{{ row.student_username }} · {{ row.student_nis }}</div>
                </td>
                <td class="px-3 py-3 text-center">
                  <span
                    class="rounded-full px-2 py-0.5 text-[10px] font-bold uppercase tracking-tight"
                    :class="
                      row.connection_status === 'online'
                        ? 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400'
                        : row.connection_status === 'blocked'
                          ? 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400'
                          : 'bg-slate-100 text-slate-600 dark:bg-slate-800 dark:text-slate-400'
                    "
                  >
                    {{ row.connection_status }}
                  </span>
                  <div class="mt-1 text-[10px] font-bold uppercase text-slate-400 dark:text-slate-500">{{ row.session_status || 'NOT JOINED' }}</div>
                </td>
                <td class="px-3 py-3 text-center">
                  <div class="font-bold text-info dark:text-sky-400">{{ row.progress_percent }}%</div>
                  <div class="text-[10px] text-slate-500 dark:text-slate-500 font-mono">({{ row.answered_questions }}/{{ row.total_questions }})</div>
                </td>
                <td class="px-3 py-3 text-xs text-slate-500 dark:text-slate-500 italic">{{ row.last_seen_at || '-' }}</td>
                <td class="px-3 py-3">
                  <div class="flex flex-wrap gap-2">
                    <BaseButton
                      color="warning"
                      small
                      outline
                      label="Force Submit"
                      :disabled="!row.session_id || row.session_status !== 'in_progress' || actionSessionId === row.session_id"
                      @click="forceSubmitSession(row)"
                    />
                    <BaseButton
                      color="success"
                      small
                      label="Reset"
                      :disabled="!row.session_id || row.session_status === 'submitted' || row.session_status === 'forced' || actionSessionId === row.session_id"
                      @click="resetSession(row)"
                    />
                  </div>
                </td>
              </tr>
              <tr v-if="!participants.length && !isLoadingParticipants">
                <td colspan="5" class="px-3 py-10 text-center text-slate-400 dark:text-slate-500 italic">Belum ada peserta target untuk ujian ini.</td>
              </tr>
            </tbody>
          </table>
        </div>
      </CardBox>
    </SectionMain>
  </LayoutAuthenticated>
</template>
