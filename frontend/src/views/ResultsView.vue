<script setup>
import { computed, onMounted, ref } from 'vue'
import { mdiChartBar, mdiTrophyOutline } from '@mdi/js'
import LayoutAuthenticated from '@/layouts/LayoutAuthenticated.vue'
import SectionMain from '@/components/SectionMain.vue'
import SectionTitleLineWithButton from '@/components/SectionTitleLineWithButton.vue'
import CardBox from '@/components/CardBox.vue'
import CardBoxWidget from '@/components/CardBoxWidget.vue'
import { api } from '@/services/api.js'

const results = ref([])
const isLoading = ref(false)
const errorMessage = ref('')

const loadResults = async () => {
  isLoading.value = true
  errorMessage.value = ''
  try {
    const { data } = await api.get('/api/v1/student/results', { params: { limit: 50, offset: 0 } })
    results.value = data?.data || []
  } catch (error) {
    results.value = []
    errorMessage.value = error?.response?.data?.error?.message || 'Gagal memuat hasil ujian dari backend'
  } finally {
    isLoading.value = false
  }
}

onMounted(loadResults)

const averageScore = computed(() => {
  if (!results.value.length) return 0
  const total = results.value.reduce((sum, item) => sum + Number(item.score || 0), 0)
  return Math.round(total / results.value.length)
})

const formatDateTime = (value) => {
  if (!value) return '-'
  const parsed = new Date(value)
  if (Number.isNaN(parsed.getTime())) return value
  const formatted = parsed.toLocaleString('id-ID', {
    dateStyle: 'medium',
    timeStyle: 'short',
    hour12: false,
  }).replace(/\./g, ':')
  const offset = -parsed.getTimezoneOffset() / 60
  let tz = ''
  if (offset === 7) tz = 'WIB'
  else if (offset === 8) tz = 'WITA'
  else if (offset === 9) tz = 'WIT'
  else tz = offset >= 0 ? `GMT+${offset}` : `GMT${offset}`
  return `${formatted} ${tz}`
}

const statusLabel = (value) => {
  const status = String(value || '').toLowerCase()
  if (status === 'submitted') return 'Submitted'
  if (status === 'expired') return 'Expired'
  if (status === 'forced') return 'Forced'
  return value || '-'
}
</script>

<template>
  <LayoutAuthenticated>
    <SectionMain>
      <SectionTitleLineWithButton :icon="mdiChartBar" title="Hasil Ujian" main />

      <div class="mb-6 grid gap-6 md:grid-cols-2">
        <CardBoxWidget
          :icon="mdiTrophyOutline"
          color="text-amber-500"
          label="Rata-rata Nilai"
          :number="averageScore"
        />
        <CardBoxWidget
          :icon="mdiChartBar"
          color="text-sky-500"
          label="Riwayat Ujian"
          :number="results.length"
        />
      </div>

      <CardBox>
        <div v-if="isLoading" class="mb-4 text-sm text-slate-500 dark:text-slate-400 italic">Memuat hasil ujian...</div>
        <div v-else-if="errorMessage" class="mb-4 rounded-lg bg-red-50 dark:bg-red-900/20 px-4 py-3 text-sm text-red-700 dark:text-red-400 border border-red-100 dark:border-red-900/40">{{ errorMessage }}</div>

        <div class="overflow-x-auto">
          <table class="w-full text-left text-sm">
            <thead class="border-b dark:border-slate-800 bg-slate-50 dark:bg-slate-800/50 text-slate-600 dark:text-slate-300 uppercase text-xs tracking-wider font-bold">
              <tr>
                <th class="px-3 py-3">Ujian</th>
                <th class="px-3 py-3">Mapel</th>
                <th class="px-3 py-3">Dikumpulkan</th>
                <th class="px-3 py-3 text-center">Nilai</th>
                <th class="px-3 py-3 text-center">Benar</th>
                <th class="px-3 py-3 text-center">Status</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="item in results" :key="item.session_id || item.id" class="border-b dark:border-slate-800 last:border-b-0 hover:bg-slate-50/50 dark:hover:bg-slate-800/30 transition-colors">
                <td class="px-3 py-3 font-medium dark:text-slate-100">{{ item.exam_title }}</td>
                <td class="px-3 py-3 text-slate-500 dark:text-slate-400">{{ item.subject }}</td>
                <td class="px-3 py-3 text-xs text-slate-500 dark:text-slate-400 font-mono italic">{{ formatDateTime(item.submitted_at) }}</td>
                <td class="px-3 py-3 text-center font-bold text-lg text-info dark:text-sky-400">
                  {{ item.score }}
                </td>
                <td class="px-3 py-3 text-center text-slate-600 dark:text-slate-300 font-mono">
                  {{ item.correct_count }}/{{ item.auto_scorable_questions ?? item.total_questions }}
                </td>
                <td class="px-3 py-3 text-center flex flex-col items-center gap-1 justify-center">
                  <span
                    class="rounded-full px-2 py-0.5 text-[10px] font-bold uppercase tracking-tight"
                    :class="
                      String(item.session_status || item.status) === 'submitted' || item.status === 'Tuntas'
                        ? 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400'
                        : 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400'
                    "
                  >
                    {{ statusLabel(item.session_status || item.status) }}
                  </span>
                  <span v-if="item.pending_grading_count > 0" class="text-[9px] font-bold text-amber-600 dark:text-amber-400 uppercase tracking-tighter animate-pulse">
                    Proses Koreksi
                  </span>
                </td>
              </tr>
              <tr v-if="!results.length">
                <td colspan="6" class="px-3 py-10 text-center text-slate-400 dark:text-slate-500 italic">Belum ada hasil ujian.</td>
              </tr>
            </tbody>
          </table>
        </div>
      </CardBox>
    </SectionMain>
  </LayoutAuthenticated>
</template>
