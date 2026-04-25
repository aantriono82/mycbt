<script setup>
import { computed, onMounted, ref, watch } from 'vue'
import { mdiChartBoxOutline, mdiDownload, mdiRefresh, mdiAccountGroup, mdiCheckCircleOutline, mdiClockAlertOutline, mdiChartLine, mdiSend, mdiCloudUploadOutline } from '@mdi/js'
import LayoutAuthenticated from '@/layouts/LayoutAuthenticated.vue'
import SectionMain from '@/components/SectionMain.vue'
import SectionTitleLineWithButton from '@/components/SectionTitleLineWithButton.vue'
import CardBox from '@/components/CardBox.vue'
import DashboardCard from '@/components/DashboardCard.vue'
import BaseButton from '@/components/BaseButton.vue'
import FormField from '@/components/FormField.vue'
import FormControl from '@/components/FormControl.vue'
import { api } from '@/services/api.js'
import { useAuthStore } from '@/stores/auth.js'
import CardBoxModal from '@/components/CardBoxModal.vue'
import ItemAnalysisAI from '@/components/ItemAnalysisAI.vue'
import BaseSkeleton from '@/components/BaseSkeleton.vue'
import { mdiPencil, mdiClose } from '@mdi/js'

const formatDateTime = (value) => {
  if (!value) return '-'
  const d = new Date(value)
  if (Number.isNaN(d.getTime())) return value
  const formatted = d.toLocaleString('id-ID', {
    dateStyle: 'medium',
    timeStyle: 'short',
    hour12: false
  }).replace(/\./g, ':')

  const offset = -d.getTimezoneOffset() / 60
  let tz = ''
  if (offset === 7) tz = 'WIB'
  else if (offset === 8) tz = 'WITA'
  else if (offset === 9) tz = 'WIT'
  else tz = offset >= 0 ? `GMT+${offset}` : `GMT${offset}`

  return `${formatted} ${tz}`
}

const authStore = useAuthStore()

const exams = ref([])
const selectedExamId = ref('')
const q = ref('')

const results = ref([])
const meta = ref(null)
const itemAnalysis = ref([])
const itemMeta = ref(null)
const scoreDistribution = ref(null)

const isLoadingExams = ref(false)
const isLoadingResults = ref(false)
const isLoadingItems = ref(false)
const isLoadingDistribution = ref(false)
const isLoadingEssays = ref(false)
const isSavingEssay = ref(false)
const errorMessage = ref('')
const isBlastingResults = ref(false)
const isSyncingLTI = ref(false)
const blastChannels = ref(['email', 'whatsapp'])
const showScoreDistribution = ref(true)
const showItemAnalysis = ref(true)

const isEssayModalActive = ref(false)
const selectedSessionForEssay = ref(null)
const essayAttempts = ref([])

const canLoad = computed(() => authStore.isAuthenticated)

const stats = computed(() => {
  const items = results.value || []
  const total = items.length
  const submitted = items.filter((x) => x.status === 'submitted').length
  const expired = items.filter((x) => x.status === 'expired').length
  const avg = total ? Math.round(items.reduce((s, x) => s + Number(x.score || 0), 0) / total) : 0
  return { total, submitted, expired, avg }
})

const filteredExamsForSelect = computed(() => {
  const now = new Date()
  return exams.value
    .filter((ex) => {
      const startsAt = new Date(ex.starts_at)
      // Only show exams that have started and are not drafts
      return startsAt <= now && ex.status !== 'draft'
    })
    .map((item) => ({ id: item.id, label: item.title }))
})

const loadExams = async () => {
  if (!canLoad.value) return
  isLoadingExams.value = true
  errorMessage.value = ''
  try {
    const { data } = await api.get('/api/v1/exams', { params: { limit: 200, offset: 0 } })
    exams.value = data?.data || []
    if (!selectedExamId.value && filteredExamsForSelect.value.length) {
      selectedExamId.value = filteredExamsForSelect.value[0].id
    }
  } catch (error) {
    exams.value = []
    errorMessage.value = error?.response?.data?.error?.message || 'Gagal memuat daftar ujian'
  } finally {
    isLoadingExams.value = false
  }
}

const loadResults = async () => {
  if (!canLoad.value || !selectedExamId.value) {
    results.value = []
    meta.value = null
    return
  }
  isLoadingResults.value = true
  errorMessage.value = ''
  try {
    const { data } = await api.get(`/api/v1/exams/${selectedExamId.value}/results`, {
      params: { q: q.value, limit: 100, offset: 0 },
    })
    results.value = data?.data || []
    meta.value = data?.meta || null
  } catch (error) {
    results.value = []
    meta.value = null
    errorMessage.value = error?.response?.data?.error?.message || 'Gagal memuat rekap hasil ujian'
  } finally {
    isLoadingResults.value = false
  }
}

const loadItemAnalysis = async () => {
  if (!canLoad.value || !selectedExamId.value) {
    itemAnalysis.value = []
    itemMeta.value = null
    return
  }
  isLoadingItems.value = true
  try {
    const { data } = await api.get(`/api/v1/exams/${selectedExamId.value}/item-analysis`)
    itemAnalysis.value = data?.data || []
    itemMeta.value = data?.meta || null
  } catch (error) {
    itemAnalysis.value = []
    itemMeta.value = null
    errorMessage.value = error?.response?.data?.error?.message || 'Gagal memuat analisis butir'
  } finally {
    isLoadingItems.value = false
  }
}

const loadScoreDistribution = async () => {
  if (!canLoad.value || !selectedExamId.value) {
    scoreDistribution.value = null
    return
  }
  isLoadingDistribution.value = true
  try {
    const { data } = await api.get(`/api/v1/exams/${selectedExamId.value}/score-distribution`)
    scoreDistribution.value = data?.data || null
  } catch (error) {
    scoreDistribution.value = null
    errorMessage.value = error?.response?.data?.error?.message || 'Gagal memuat distribusi nilai'
  } finally {
    isLoadingDistribution.value = false
  }
}

const openEssayModal = async (row) => {
  selectedSessionForEssay.value = row
  essayAttempts.value = []
  isEssayModalActive.value = true
  await loadEssays()
}

const loadEssays = async () => {
  if (!selectedSessionForEssay.value) return
  isLoadingEssays.value = true
  try {
    const { data } = await api.get(`/api/v1/exams/${selectedExamId.value}/sessions/${selectedSessionForEssay.value.session_id}/essays`)
    essayAttempts.value = data?.data || []
  } catch (error) {
    errorMessage.value = error?.response?.data?.error?.message || 'Gagal memuat jawaban essay'
  } finally {
    isLoadingEssays.value = false
  }
}

const saveEssayScore = async (attempt) => {
  if (isSavingEssay.value) return
  isSavingEssay.value = true
  try {
    await api.post(`/api/v1/exams/${selectedExamId.value}/sessions/${selectedSessionForEssay.value.session_id}/essays/score`, {
      question_id: attempt.question_id,
      score: attempt.manual_score || 0,
      feedback: attempt.manual_feedback || '',
    })
    // No need to reload everything, just a subtle confirmation
  } catch (error) {
    alert(error?.response?.data?.error?.message || 'Gagal menyimpan nilai')
  } finally {
    isSavingEssay.value = false
  }
}

const handleEssayModalDone = async () => {
  isEssayModalActive.value = false
  // Reload results to see updated total score
  await loadResults()
}

const exportResults = async () => {
  if (!canLoad.value || !selectedExamId.value) return
  errorMessage.value = ''
  try {
    const response = await api.get(`/api/v1/exams/${selectedExamId.value}/export`, {
      responseType: 'blob',
    })
    const blob = new Blob([response.data], {
      type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
    })
    const href = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = href
    a.download = `exam-results-${selectedExamId.value}.xlsx`
    document.body.appendChild(a)
    a.click()
    a.remove()
    URL.revokeObjectURL(href)
  } catch (error) {
    errorMessage.value = error?.response?.data?.error?.message || 'Gagal export hasil ujian'
  }
}

const blastResults = async () => {
  if (!selectedExamId.value) return
  if (!confirm('Kirim hasil nilai ke semua siswa (Email/WA) sesuai pengaturan?')) return
  
  isBlastingResults.value = true
  try {
    const { data } = await api.post(`/api/v1/exams/${selectedExamId.value}/results/blast`, {
      channels: blastChannels.value,
    })
    alert(`Blast selesai. Terkirim: ${data.data.sent_count}, Gagal: ${data.data.failed_count}`)
  } catch (error) {
    alert(error?.response?.data?.error?.message || 'Gagal mengirim nilai')
  } finally {
    isBlastingResults.value = false
  }
}

const syncLTIScores = async () => {
  if (!selectedExamId.value) return
  if (!confirm('Sinkronkan nilai final yang sudah selesai dikoreksi ke LMS via LTI AGS?')) return

  isSyncingLTI.value = true
  try {
    const { data } = await api.post(`/api/v1/exams/${selectedExamId.value}/lti/sync-scores`)
    const result = data?.data || {}
    const detailText = Array.isArray(result.details) && result.details.length ? `\n\nDetail:\n- ${result.details.join('\n- ')}` : ''
    alert(`Sync LMS selesai.\nTarget: ${result.target_count || 0}\nBerhasil: ${result.synced_count || 0}\nSkip: ${result.skipped_count || 0}\nGagal: ${result.failed_count || 0}${detailText}`)
  } catch (error) {
    alert(error?.response?.data?.error?.message || 'Gagal sinkronkan nilai ke LMS')
  } finally {
    isSyncingLTI.value = false
  }
}

const exportItemAnalysis = async () => {
  if (!canLoad.value || !selectedExamId.value) return
  errorMessage.value = ''
  try {
    const response = await api.get(`/api/v1/exams/${selectedExamId.value}/item-analysis/export`, {
      responseType: 'blob',
    })
    const blob = new Blob([response.data], {
      type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
    })
    const href = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = href
    a.download = `item-analysis-${selectedExamId.value}.xlsx`
    document.body.appendChild(a)
    a.click()
    a.remove()
    URL.revokeObjectURL(href)
  } catch (error) {
    errorMessage.value = error?.response?.data?.error?.message || 'Gagal export item analysis'
  }
}

watch(selectedExamId, async () => {
  await loadResults()
  await loadItemAnalysis()
  await loadScoreDistribution()
})

onMounted(async () => {
  await loadExams()
  await loadResults()
  await loadItemAnalysis()
  await loadScoreDistribution()
})
</script>

<template>
  <LayoutAuthenticated>
    <SectionMain>
      <SectionTitleLineWithButton :icon="mdiChartBoxOutline" title="Evaluasi / Hasil Nilai" main>
        <div class="flex items-center gap-1.5 overflow-x-auto pb-1 lg:pb-0">
          <BaseButton 
            :icon="mdiRefresh" 
            color="info" 
            small
            label="Refresh" 
            @click="loadExams(); loadResults(); loadItemAnalysis(); loadScoreDistribution()" 
          />
          <BaseButton 
            :icon="mdiCloudUploadOutline" 
            color="success" 
            small
            label="LMS" 
            :disabled="isSyncingLTI || !selectedExamId" 
            @click="syncLTIScores" 
          />
          <BaseButton 
            :icon="mdiDownload" 
            color="purple" 
            small
            label="Analisis" 
            @click="exportItemAnalysis" 
          />
          
          <div class="flex items-center gap-1.5 bg-slate-100/50 dark:bg-slate-800/20 p-1 rounded-xl border border-slate-100 dark:border-slate-700 shadow-inner">
            <label 
              class="flex items-center gap-1.5 px-2.5 py-1.5 rounded-lg text-[10px] font-black uppercase tracking-tighter cursor-pointer transition-all shadow-sm border"
              :class="blastChannels.includes('email') 
                ? 'bg-blue-600 border-blue-600 text-white' 
                : 'bg-white border-slate-200 text-slate-400 dark:bg-slate-900 dark:border-slate-800'"
            >
              <input type="checkbox" v-model="blastChannels" value="email" class="rounded-sm border-none bg-slate-100 text-blue-600 focus:ring-0 w-3 h-3" /> 
              Email
            </label>
            
            <label 
              class="flex items-center gap-1.5 px-2.5 py-1.5 rounded-lg text-[10px] font-black uppercase tracking-tighter cursor-pointer transition-all shadow-sm border"
              :class="blastChannels.includes('whatsapp') 
                ? 'bg-emerald-600 border-emerald-600 text-white' 
                : 'bg-white border-slate-200 text-slate-400 dark:bg-slate-900 dark:border-slate-800'"
            >
              <input type="checkbox" v-model="blastChannels" value="whatsapp" class="rounded-sm border-none bg-slate-100 text-emerald-600 focus:ring-0 w-3 h-3" /> 
              WA
            </label>
            
            <BaseButton 
              :icon="mdiSend" 
              color="purple" 
              small
              label="Blast" 
              :disabled="isBlastingResults || !selectedExamId" 
              @click="blastResults" 
            />
          </div>
        </div>
      </SectionTitleLineWithButton>

      <div v-if="!authStore.isAuthenticated" class="mb-6 rounded-xl bg-amber-50 dark:bg-amber-900/20 px-4 py-3 text-sm text-amber-700 dark:text-amber-400 border border-amber-100 dark:border-amber-900/40">
        Login terlebih dulu agar evaluasi dapat memuat data backend.
      </div>
      <div v-else-if="errorMessage" class="mb-6 rounded-xl bg-red-50 dark:bg-red-900/20 px-4 py-3 text-sm text-red-700 dark:text-red-400 border border-red-100 dark:border-red-900/40">
        {{ errorMessage }}
      </div>

      <CardBox class="mb-6">
        <div class="flex flex-col gap-4 md:flex-row md:items-end">
          <div class="md:w-80 shrink-0">
            <label class="block mb-1 text-sm font-semibold text-slate-600 dark:text-slate-300">Pilih Ujian</label>
            <FormControl
              v-model="selectedExamId"
              :options="filteredExamsForSelect"
            />
          </div>
          <div class="flex-1">
            <label class="block mb-1 text-sm font-semibold text-slate-600 dark:text-slate-300">Cari Siswa <span class="font-normal text-slate-400 dark:text-slate-500">(nama/username/nis)</span></label>
            <FormControl v-model="q" placeholder="Ketik lalu klik Terapkan" />
          </div>
          <div class="shrink-0">
            <BaseButton color="info" label="Terapkan" :disabled="isLoadingResults" @click="loadResults" />
          </div>
        </div>
        <div class="mt-3 flex gap-3 h-5 items-center">
          <template v-if="isLoadingExams || isLoadingResults">
            <BaseSkeleton width="w-48" height="h-4" />
          </template>
          <div class="text-sm text-slate-500 dark:text-slate-400" v-else-if="meta?.exam">
            Hasil Ujian: <span class="font-bold dark:text-slate-100">{{ meta.exam.title }}</span>
          </div>
        </div>
      </CardBox>

      <div class="mb-8 grid gap-4 grid-cols-1 sm:grid-cols-2 lg:grid-cols-4">
        <template v-if="isLoadingResults">
          <CardBox v-for="i in 4" :key="i" class="h-32 flex flex-col justify-center">
            <div class="flex items-center gap-4">
              <BaseSkeleton width="w-12" height="h-12" rounded="rounded-2xl" />
              <div class="space-y-2 flex-1">
                <BaseSkeleton width="w-24" height="h-3" />
                <BaseSkeleton width="w-12" height="h-8" />
              </div>
            </div>
          </CardBox>
        </template>
        <template v-else>
          <DashboardCard label="Total Peserta" :number="stats.total" :icon="mdiAccountGroup" color="blue" />
          <DashboardCard label="Selesai (Submitted)" :number="stats.submitted" :icon="mdiCheckCircleOutline" color="emerald" />
          <DashboardCard label="Terlambat (Expired)" :number="stats.expired" :icon="mdiClockAlertOutline" color="amber" />
          <DashboardCard label="Rata-rata Nilai" :number="stats.avg" :icon="mdiChartLine" color="indigo" />
        </template>
      </div>

      <CardBox>
        <div class="overflow-x-auto">
          <table class="w-full text-left text-sm">
            <thead class="border-b dark:border-slate-800 bg-slate-50 dark:bg-slate-800/50 text-slate-600 dark:text-slate-300 uppercase text-xs tracking-wider font-bold">
              <tr>
                <th class="px-3 py-3">Siswa</th>
                <th class="px-3 py-3">Username</th>
                <th class="px-3 py-3 text-center">NIS</th>
                <th class="px-3 py-3 text-center">Status</th>
                <th class="px-3 py-3">Dikumpulkan</th>
                <th class="px-3 py-3 text-center">Benar</th>
                <th class="px-3 py-3 text-center">Nilai</th>
                <th class="px-3 py-3 text-center">Grading</th>
              </tr>
            </thead>
            <tbody>
              <template v-if="isLoadingResults">
                <tr v-for="i in 5" :key="i">
                  <td v-for="j in 8" :key="j" class="px-3 py-4">
                    <BaseSkeleton width="w-full" height="h-4" />
                  </td>
                </tr>
              </template>
              <tr v-else v-for="row in results" :key="row.session_id" class="border-b dark:border-slate-800 last:border-b-0 hover:bg-slate-50/50 dark:hover:bg-slate-800/30 transition-colors">
                <td class="px-3 py-3 font-medium dark:text-slate-100">{{ row.student_name }}</td>
                <td class="px-3 py-3 text-slate-500 dark:text-slate-400">{{ row.student_username }}</td>
                <td class="px-3 py-3 text-center text-slate-500 dark:text-slate-400">{{ row.student_nis }}</td>
                <td class="px-3 py-3 text-center">
                  <span
                    class="rounded-full px-2 py-0.5 text-[10px] font-bold uppercase tracking-tight"
                    :class="row.status === 'submitted' ? 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400' : row.status === 'expired' ? 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400' : 'bg-slate-100 text-slate-600 dark:bg-slate-800 dark:text-slate-400'"
                  >
                    {{ row.status }}
                  </span>
                </td>
                <td class="px-3 py-3 text-slate-500 dark:text-slate-400 text-xs">{{ formatDateTime(row.finished_at) }}</td>
                <td class="px-3 py-3 text-center text-slate-600 dark:text-slate-300 font-mono">{{ row.correct_count }}/{{ row.auto_scorable_questions }}</td>
                <td class="px-3 py-3 text-center font-bold text-lg text-info dark:text-sky-400">{{ row.score }}</td>
                <td class="px-3 py-3 text-center">
                  <div class="flex flex-col items-center gap-1">
                    <BaseButton
                      :icon="mdiPencil"
                      color="whiteDark"
                      smaller
                      @click="openEssayModal(row)"
                      label="Koreksi"
                    />
                    <div v-if="row.pending_grading_count > 0" class="text-[9px] font-bold text-amber-600 dark:text-amber-400 uppercase tracking-tighter">
                      {{ row.pending_grading_count }} pending
                    </div>
                    <div v-else-if="row.manual_scored_count > 0" class="text-[9px] font-bold text-emerald-600 dark:text-emerald-400 uppercase tracking-tighter">
                      Selesai
                    </div>
                  </div>
                </td>
              </tr>
              <tr v-if="!results.length && !isLoadingResults">
                <td colspan="7" class="px-3 py-10 text-center text-slate-400 dark:text-slate-500 italic">
                  Belum ada data peserta untuk ujian ini.
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </CardBox>

      <CardBox class="mt-6" v-if="showScoreDistribution">
        <div class="flex items-center justify-between mb-4">
          <h3 class="text-lg font-semibold dark:text-slate-100">Distribusi Nilai</h3>
          <BaseButton 
            :icon="mdiClose" 
            color="whiteDark" 
            small 
            rounded-full
            @click="showScoreDistribution = false" 
          />
        </div>
        <div v-if="scoreDistribution" class="mb-6 grid gap-4 md:grid-cols-4">
          <CardBox class="text-center bg-slate-50/50 dark:bg-slate-800/30">
            <div class="text-[10px] uppercase font-bold text-slate-400">Min</div>
            <div class="mt-1 text-xl font-bold dark:text-slate-100">{{ scoreDistribution.min_score }}</div>
          </CardBox>
          <CardBox class="text-center bg-slate-50/50 dark:bg-slate-800/30">
            <div class="text-[10px] uppercase font-bold text-slate-400">Median</div>
            <div class="mt-1 text-xl font-bold dark:text-slate-100">{{ scoreDistribution.median_score }}</div>
          </CardBox>
          <CardBox class="text-center bg-slate-50/50 dark:bg-slate-800/30">
            <div class="text-[10px] uppercase font-bold text-slate-400">Rata-rata</div>
            <div class="mt-1 text-xl font-bold text-sky-600 dark:text-sky-400">{{ scoreDistribution.average_score }}</div>
          </CardBox>
          <CardBox class="text-center bg-slate-50/50 dark:bg-slate-800/30">
            <div class="text-[10px] uppercase font-bold text-slate-400">Max</div>
            <div class="mt-1 text-xl font-bold dark:text-slate-100">{{ scoreDistribution.max_score }}</div>
          </CardBox>
        </div>
        <div v-if="scoreDistribution" class="space-y-3">
          <div
            v-for="bin in scoreDistribution.distribution_bins || []"
            :key="bin.label"
            class="grid items-center gap-3 md:grid-cols-[100px_1fr_140px]"
          >
            <div class="text-xs font-bold text-slate-500 dark:text-slate-400 tracking-tight">{{ bin.label }}</div>
            <div class="h-2 rounded-full bg-slate-100 dark:bg-slate-800 shadow-inner">
              <div
                class="h-2 rounded-full bg-gradient-to-r from-sky-400 to-blue-500 shadow-sm transition-all duration-1000"
                :style="{ width: `${Math.min(100, Number(bin.percent || 0))}%` }"
              />
            </div>
            <div class="text-xs text-slate-600 dark:text-slate-400 font-medium"><span class="dark:text-slate-200">{{ bin.count }}</span> siswa <span class="bg-slate-100 dark:bg-slate-800 px-1.5 py-0.5 rounded ml-1">({{ bin.percent }}%)</span></div>
          </div>
        </div>
        <div v-else-if="isLoadingDistribution" class="mb-4 text-sm text-slate-500 dark:text-slate-400 italic">Memuat distribusi nilai...</div>
        <div v-else class="mb-4 text-sm text-slate-400 dark:text-slate-500 italic">Belum ada data distribusi nilai untuk ujian ini.</div>
      </CardBox>
      <div v-else class="mt-4 flex justify-end">
        <button @click="showScoreDistribution = true" class="text-xs font-bold text-blue-600 hover:underline">
          + Tampilkan Distribusi Nilai
        </button>
      </div>

      <CardBox class="mt-6" v-if="showItemAnalysis">
        <div class="flex items-center justify-between mb-4">
          <h3 class="text-lg font-semibold dark:text-slate-100 flex items-center gap-2">
            Analisis Butir Soal
            <span class="text-[10px] bg-slate-100 dark:bg-slate-800 text-slate-500 px-2 py-0.5 rounded-full uppercase tracking-widest font-black">Psikometri Standard</span>
          </h3>
          <BaseButton 
            :icon="mdiClose" 
            color="whiteDark" 
            small 
            rounded-full
            @click="showItemAnalysis = false" 
          />
        </div>

        <!-- Legend / Guide for Teachers -->
        <div class="mb-6 grid gap-4 md:grid-cols-3">
          <div class="p-4 rounded-2xl border border-slate-100 dark:border-slate-800 bg-white/50 dark:bg-slate-900/30">
            <h4 class="text-xs font-black uppercase text-slate-400 mb-3 tracking-widest">Tingkat Kesulitan (P-Value)</h4>
            <div class="space-y-2">
              <div class="flex justify-between items-center text-xs">
                <span class="text-emerald-600 font-black">31% - 70% (Sedang)</span>
                <span class="text-slate-400 font-medium">Ideal / Informatif</span>
              </div>
              <div class="flex justify-between items-center text-xs">
                <span class="text-rose-600 font-black">≤ 30% (Sukar)</span>
                <span class="text-slate-400 font-medium whitespace-nowrap">Hanya dijawab sedikit siswa</span>
              </div>
              <div class="flex justify-between items-center text-xs">
                <span class="text-sky-600 font-black">> 70% (Mudah)</span>
                <span class="text-slate-400 font-medium">Hampir semua benar</span>
              </div>
            </div>
          </div>
          
          <div class="p-4 rounded-2xl border border-slate-100 dark:border-slate-800 bg-white/50 dark:bg-slate-900/30">
            <h4 class="text-xs font-black uppercase text-slate-400 mb-3 tracking-widest">Daya Pembeda (Akurasi)</h4>
            <div class="space-y-2">
              <div class="flex justify-between items-center text-xs">
                <span class="bg-emerald-100 text-emerald-700 px-2 py-0.5 rounded-lg font-black tracking-tight">≥ 40% (Sangat Baik)</span>
                <BaseIcon :path="mdiCheckCircleOutline" size="16" class="text-emerald-500" />
              </div>
              <div class="flex justify-between items-center text-xs">
                <span class="bg-amber-100 text-amber-700 px-2 py-0.5 rounded-lg font-black tracking-tight">20% - 29% (Cukup)</span>
                <span class="text-slate-400 font-medium">Perlu Revisi</span>
              </div>
              <div class="flex justify-between items-center text-xs">
                <span class="bg-red-500 text-white px-2 py-0.5 rounded-lg font-black tracking-tight">Negatif / < 20%</span>
                <span class="text-red-500 font-black italic">Wajib Buang!</span>
              </div>
            </div>
          </div>

          <div class="p-4 rounded-2xl border border-slate-100 dark:border-slate-800 bg-white/50 dark:bg-slate-900/30">
            <h4 class="text-xs font-black uppercase text-slate-400 mb-3 tracking-widest">Istilah Data</h4>
            <div class="space-y-2 text-xs text-slate-700 dark:text-slate-300">
               <p><strong class="text-slate-900 dark:text-white">N:</strong> Total siswa yang ikut tes.</p>
               <p><strong class="text-slate-900 dark:text-white">Benar:</strong> Jumlah real jawaban benar.</p>
               <p><strong class="text-slate-900 dark:text-white">Distraktor:</strong> Sebaran pilihan (A,B,C,D).</p>
            </div>
          </div>
        </div>

        <div class="mb-3 text-[10px] uppercase font-bold tracking-widest text-slate-400 dark:text-slate-500" v-if="itemMeta?.total_sessions">
          Sesi: {{ itemMeta.total_sessions }} · Soal: {{ itemMeta.total_items }}
        </div>
        <div class="overflow-x-auto">
          <table class="w-full text-left text-sm">
            <thead class="border-b dark:border-slate-800 bg-slate-50 dark:bg-slate-800/50 text-slate-600 dark:text-slate-300 uppercase text-xs tracking-wider font-bold">
              <tr>
                <th class="px-3 py-3">No</th>
                <th class="px-3 py-3">Tipe</th>
                <th class="px-3 py-3">Stem</th>
                <th class="px-3 py-3 text-center">N</th>
                <th class="px-3 py-3 text-center">Benar</th>
                <th class="px-3 py-3 text-center">P-Value</th>
                <th class="px-3 py-3 text-center">Kategori</th>
                <th class="px-3 py-3">Akurasi</th>
                <th class="px-3 py-3">Distraktor</th>
              </tr>
            </thead>
            <tbody>
              <template v-if="isLoadingItems">
                <tr v-for="i in 5" :key="i">
                  <td v-for="j in 9" :key="j" class="px-3 py-4">
                    <BaseSkeleton width="w-full" height="h-4" />
                  </td>
                </tr>
              </template>
              <tr v-else v-for="row in itemAnalysis" :key="row.question_id" class="border-b dark:border-slate-800 last:border-b-0 hover:bg-slate-50/30 dark:hover:bg-slate-800/30 transition-colors">
                <td class="px-3 py-3 font-bold dark:text-slate-300">#{{ row.order_no }}</td>
                <td class="px-3 py-3">
                  <span class="text-[10px] uppercase font-bold text-slate-400 tracking-tighter">{{ row.question_type }}</span>
                </td>
                <td class="px-3 py-3">
                  <div class="max-w-xl truncate text-slate-700 dark:text-slate-300 italic">{{ row.stem }}</div>
                </td>
                <td class="px-3 py-3 text-center text-slate-500 dark:text-slate-400 font-mono">{{ row.answered_count }}</td>
                <td class="px-3 py-3 text-center text-emerald-600 dark:text-emerald-400 font-bold font-mono">{{ row.correct_count }}</td>
                <td class="px-3 py-3 text-center">
                  <div 
                    class="font-black text-sm"
                    :class="
                      row.p_value_percent <= 30 
                        ? 'text-rose-600 dark:text-rose-400' 
                        : row.p_value_percent <= 70 
                          ? 'text-emerald-600 dark:text-emerald-400' 
                          : 'text-sky-600 dark:text-sky-400'
                    "
                  >
                    {{ row.p_value_percent }}%
                  </div>
                  <div class="text-[9px] uppercase font-bold text-slate-400 tracking-tighter">
                    {{ row.difficulty_label }}
                  </div>
                </td>
                <td class="px-3 py-3">
                  <div class="flex items-center gap-1.5">
                    <div 
                      class="text-xs font-black px-1.5 py-0.5 rounded"
                      :class="
                        row.discrimination_index < 0
                          ? 'bg-red-500 text-white animate-pulse'
                          : row.discrimination_index < 20
                            ? 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400'
                            : row.discrimination_index < 30
                              ? 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400'
                              : row.discrimination_index < 40
                                ? 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400'
                                : 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400'
                      "
                    >
                      D: {{ row.discrimination_index }}%
                    </div>
                    <BaseIcon 
                      v-if="row.discrimination_index < 20" 
                      :path="mdiClockAlertOutline" 
                      size="14" 
                      class="text-red-500"
                    />
                  </div>
                  <div class="text-[10px] text-slate-500 dark:text-slate-500 italic leading-tight mt-0.5">
                    {{ row.discrimination_label }}
                  </div>
                </td>
                <td class="px-3 py-3">
                  <div v-if="row.option_stats?.length" class="space-y-1">
                    <div v-for="opt in row.option_stats" :key="opt.option_id" class="text-[10px] flex gap-1 items-center">
                      <span class="font-bold text-slate-400 dark:text-slate-500 w-3">{{ opt.label }}</span>
                      <div class="h-1.5 flex-1 bg-slate-100 dark:bg-slate-800 rounded-full overflow-hidden max-w-[40px]">
                        <div class="h-full bg-slate-300 dark:bg-slate-600" :style="{ width: `${opt.selected_percent}%` }"></div>
                      </div>
                      <span v-if="opt.is_correct" class="rounded px-1 text-[9px] font-bold border border-emerald-500 text-emerald-600 dark:text-emerald-400">KEY</span>
                      <span class="text-slate-500 dark:text-slate-400">{{ opt.selected_percent }}%</span>
                    </div>
                  </div>
                  <span v-else class="text-xs text-slate-500">-</span>
                </td>
              </tr>
              <tr v-if="!itemAnalysis.length && !isLoadingItems">
                <td colspan="9" class="px-3 py-10 text-center text-slate-400 dark:text-slate-500 italic">Belum ada data item analysis untuk ujian ini.</td>
              </tr>
            </tbody>
          </table>
        </div>
      </CardBox>
      <div v-else class="mt-4 flex justify-end">
        <button @click="showItemAnalysis = true" class="text-xs font-bold text-blue-600 hover:underline">
          + Tampilkan Analisis Butir
        </button>
      </div>
    </SectionMain>

    <CardBoxModal
      v-model="isEssayModalActive"
      :title="`Koreksi Essay: ${selectedSessionForEssay?.student_name || ''}`"
      button-label="Tutup & Update Nilai"
      @confirm="handleEssayModalDone"
      has-custom-layout
    >
      <div class="p-6">
        <div v-if="isLoadingEssays" class="text-center py-10 italic text-slate-500">
          Memuat jawaban essay...
        </div>
        <div v-else-if="!essayAttempts.length" class="text-center py-10 italic text-slate-400">
          Tidak ada soal essay untuk peserta ini.
        </div>
        <div v-else class="space-y-8">
          <div v-for="(att, idx) in essayAttempts" :key="att.question_id" class="rounded-xl border border-slate-200 dark:border-slate-800 p-4 bg-slate-50/50 dark:bg-slate-900/30">
            <div class="mb-3 flex items-start justify-between gap-4">
              <div class="flex-1">
                <div class="text-[10px] font-bold uppercase tracking-widest text-slate-400 mb-1">Soal #{{ att.question_order }}</div>
                <div class="text-sm dark:text-slate-200 italic" v-html="att.question_stem"></div>
              </div>
              <div class="shrink-0 text-right">
                <div class="text-[10px] font-bold uppercase text-slate-400">Score Max</div>
                <div class="text-lg font-black text-slate-700 dark:text-slate-300">{{ att.max_score }}</div>
              </div>
            </div>

            <div class="mb-4 rounded-lg bg-white dark:bg-slate-950 p-3 border border-slate-100 dark:border-slate-800 shadow-inner">
              <div class="text-[10px] font-bold uppercase text-slate-400 mb-2">Jawaban Siswa:</div>
              <div class="text-sm whitespace-pre-wrap dark:text-slate-200">{{ att.answer_text || '(Tidak menjawab)' }}</div>
            </div>

            <div v-if="att.rubric_text" class="mb-4 rounded-lg bg-amber-50/50 dark:bg-amber-900/10 p-3 border border-amber-100 dark:border-amber-900/20">
               <div class="text-[10px] font-bold uppercase text-amber-600 dark:text-amber-500 mb-1">Rubrik/Kunci:</div>
               <div class="text-xs text-slate-600 dark:text-slate-400 italic" v-html="att.rubric_text"></div>
            </div>

            <div class="grid gap-4 md:grid-cols-3 items-end">
              <div class="md:col-span-1">
                <FormField label="Nilai">
                  <FormControl v-model="att.manual_score" type="number" :min="0" :max="att.max_score" placeholder="0" />
                </FormField>
              </div>
              <div class="md:col-span-1">
                <FormField label="Feedback (Opsional)">
                  <FormControl v-model="att.manual_feedback" placeholder="Komentar guru..." />
                </FormField>
              </div>
              <div class="md:col-span-1">
                <BaseButton
                  color="info"
                  label="Simpan Nilai Soal Ini"
                  class="w-full"
                  :disabled="isSavingEssay"
                  @click="saveEssayScore(att)"
                />
              </div>
            </div>
          </div>
        </div>
      </div>
    </CardBoxModal>
  </LayoutAuthenticated>
</template>
