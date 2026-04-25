<script setup>
import { computed, onMounted, ref } from 'vue'
import { mdiChartBar, mdiTrophyOutline } from '@mdi/js'
import LayoutAuthenticated from '@/layouts/LayoutAuthenticated.vue'
import SectionMain from '@/components/SectionMain.vue'
import SectionTitleLineWithButton from '@/components/SectionTitleLineWithButton.vue'
import CardBox from '@/components/CardBox.vue'
import DashboardCard from '@/components/DashboardCard.vue'
import BaseIcon from '@/components/BaseIcon.vue'
import { api } from '@/services/api.js'

import { useResultStore } from '@/stores/result.js'
import { storeToRefs } from 'pinia'

const resultStore = useResultStore()
const { results, isLoading, errorMessage, averageScore, totalExams } = storeToRefs(resultStore)

onMounted(() => resultStore.loadResults())

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

      <div class="mb-8 grid gap-4 grid-cols-1 sm:grid-cols-2 lg:grid-cols-4">
        <DashboardCard
          :icon="mdiTrophyOutline"
          color="amber"
          label="Rata-rata Nilai"
          :number="averageScore"
          :loading="isLoading"
        />
        <DashboardCard
          :icon="mdiChartBar"
          color="blue"
          label="Total Ujian"
          :number="totalExams"
          :loading="isLoading"
        />
        <DashboardCard
          v-if="results.length"
          :icon="mdiChartBar"
          color="emerald"
          label="Nilai Tertinggi"
          :number="Math.max(...results.map(r => r.score || 0))"
        />
        <DashboardCard
          v-if="results.length"
          :icon="mdiChartBar"
          color="indigo"
          label="Ujian Terakhir"
          :number="results[0]?.score || 0"
        />
      </div>

      <div class="space-y-4">
        <div v-if="isLoading" class="flex flex-col items-center py-20">
           <div class="h-10 w-10 border-4 border-blue-600 border-t-transparent rounded-full animate-spin mb-4"></div>
           <p class="text-slate-400 font-bold uppercase tracking-widest text-xs">Memuat Hasil...</p>
        </div>
        <div v-else-if="errorMessage" class="mb-4 rounded-2xl bg-red-50 px-4 py-3 text-sm text-red-700 border border-red-100">{{ errorMessage }}</div>

        <div v-for="item in results" :key="item.session_id || item.id" class="bg-white dark:bg-slate-900 rounded-[2rem] border border-slate-100 dark:border-slate-800 shadow-sm hover:shadow-lg transition-all p-6 flex flex-col md:flex-row md:items-center gap-6 group">
           <!-- Score Circle -->
           <div class="flex-none h-20 w-20 rounded-full flex items-center justify-center relative shadow-inner overflow-hidden"
             :class="item.score >= 75 ? 'bg-emerald-50 text-emerald-600' : 'bg-blue-50 text-blue-600'"
           >
              <!-- Decorative background flare -->
              <div class="absolute inset-0 bg-white/20 blur-xl opacity-0 group-hover:opacity-100 transition-opacity"></div>
              <span class="text-2xl font-black relative z-10">{{ item.score }}</span>
           </div>

           <!-- Info -->
           <div class="grow">
              <div class="text-[10px] font-black uppercase tracking-[0.2em] text-slate-400 mb-1">{{ item.subject }}</div>
              <h3 class="text-xl font-bold text-slate-800 dark:text-white mb-2">{{ item.exam_title }}</h3>
              <div class="flex flex-wrap gap-4 text-xs font-bold text-slate-500">
                 <div class="flex items-center">
                    <span class="h-1.5 w-1.5 rounded-full bg-slate-300 mr-2"></span>
                    {{ formatDateTime(item.submitted_at) }}
                 </div>
                 <div class="flex items-center">
                    <span class="h-1.5 w-1.5 rounded-full bg-emerald-400 mr-2"></span>
                    Benar: {{ item.correct_count }}/{{ item.auto_scorable_questions ?? item.total_questions }}
                 </div>
              </div>
           </div>

           <!-- Status Badge & Action -->
           <div class="flex-none flex flex-col items-end gap-3">
              <span
                class="rounded-xl px-4 py-1.5 text-[10px] font-black uppercase tracking-widest"
                :class="
                  String(item.session_status || item.status) === 'submitted' || item.status === 'Tuntas'
                    ? 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400 border border-emerald-200/50'
                    : 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400 border border-amber-200/50'
                "
              >
                {{ statusLabel(item.session_status || item.status) }}
              </span>
              <div v-if="item.pending_grading_count > 0" class="text-[9px] font-black text-amber-500 dark:text-amber-400 uppercase tracking-widest animate-pulse">
                Menunggu Koreksi Guru
              </div>
           </div>
        </div>

        <div v-if="!results.length && !isLoading" class="bg-white rounded-[2rem] border border-dashed border-slate-200 p-20 text-center">
           <BaseIcon :path="mdiTrophyOutline" size="48" class="text-slate-200 mb-4 mx-auto" />
           <p class="text-slate-400 font-bold italic">Belum ada hasil ujian yang tersedia.</p>
        </div>
      </div>
    </SectionMain>
  </LayoutAuthenticated>
</template>
