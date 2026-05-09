<script setup>
import { ref, onMounted } from 'vue'
import { 
  mdiHomeOutline, 
  mdiBookOpenVariant, 
  mdiCalendarClockOutline, 
  mdiKeyVariant, 
  mdiMonitorEye,
  mdiClose 
} from '@mdi/js'
import LayoutAuthenticated from '@/layouts/LayoutAuthenticated.vue'
import SectionMain from '@/components/SectionMain.vue'
import SectionTitleLineWithButton from '@/components/SectionTitleLineWithButton.vue'
import CardBox from '@/components/CardBox.vue'
import DashboardCard from '@/components/DashboardCard.vue'
import QuickMenuCard from '@/components/QuickMenuCard.vue'
import { useAuthStore } from '@/stores/auth.js'
import { api } from '@/services/api.js'
import BaseChart from '@/components/Charts/BaseChart.vue'
import BaseIcon from '@/components/BaseIcon.vue'

const authStore = useAuthStore()
const isLoading = ref(false)
const stats = ref({
  bankSoal: 0,
  sesiUjian: 0,
  tokens: 0,
  onlineStudents: 0
})
const showGradeChart = ref(true)

const avgScoreData = ref({
  labels: ['Kelas 10', 'Kelas 11', 'Kelas 12'],
  datasets: [
    {
      label: 'Rata-rata Nilai',
      data: [78, 85, 82],
      backgroundColor: [
        'rgba(59, 130, 246, 0.6)',
        'rgba(16, 185, 129, 0.6)',
        'rgba(99, 102, 241, 0.6)'
      ],
      borderRadius: 8,
      borderWidth: 0
    }
  ]
})

const chartOptions = {
  indexAxis: 'y',
  scales: {
    x: {
      beginAtZero: true,
      max: 100,
      grid: {
        display: false
      }
    },
    y: {
      grid: {
        display: false
      }
    }
  },
  plugins: {
    legend: {
      display: false
    }
  }
}
const assignments = ref({
  levels: [],
  groups: [],
  subjects: []
})

const fetchStats = async () => {
  isLoading.value = true
  try {
    const [qbRes, examsRes, assignRes] = await Promise.all([
      api.get('/api/v1/question-sets', { params: { limit: 1, offset: 0 } }),
      api.get('/api/v1/exams', { params: { limit: 1, offset: 0 } }),
      api.get('/api/v1/lookups/my-assignments')
    ])

    stats.value.bankSoal = qbRes.data?.meta?.total || 0
    stats.value.sesiUjian = examsRes.data?.meta?.total || 0
    assignments.value = assignRes.data?.data || assignments.value
  } catch (err) {
    console.error('Failed to fetch teacher stats:', err)
  } finally {
    isLoading.value = false
  }
}

onMounted(() => {
  fetchStats()
})
</script>

<template>
  <LayoutAuthenticated>
    <SectionMain>
      <SectionTitleLineWithButton :icon="mdiHomeOutline" title="Dashboard Guru" main />

      <div class="mb-6 rounded-[2rem] bg-white dark:bg-slate-900/50 border border-emerald-400/60 dark:border-emerald-800/80 px-6 py-8 sm:px-8 sm:py-10 shadow-sm relative overflow-hidden transition-all hover:shadow-md group">
        <!-- Decoration -->
        <div class="absolute top-0 right-0 -mt-10 -mr-10 h-64 w-64 rounded-full bg-emerald-500/5 blur-3xl transition-all group-hover:scale-110"></div>
        <div class="absolute bottom-0 left-0 -mb-10 -ml-10 h-48 w-48 rounded-full bg-blue-500/5 blur-2xl"></div>
        
        <div class="relative z-10">
          <div class="mb-3 flex items-center gap-2">
            <span class="px-2 py-0.5 rounded-md bg-emerald-100 dark:bg-emerald-900/40 text-emerald-600 dark:text-emerald-400 text-[10px] font-black uppercase tracking-widest">Panel Guru</span>
            <span class="text-[10px] font-black uppercase tracking-[0.3em] text-slate-400">AtigaCBT Professional</span>
          </div>
          <h2 class="mb-4 text-2xl sm:text-4xl font-black text-slate-800 dark:text-white tracking-tight leading-tight">
            Selamat mengajar, <span class="text-emerald-600 dark:text-emerald-400">{{ authStore.user?.name?.split(' ')[0] || 'Rekan Guru' }}</span>!
          </h2>
          <p class="text-sm sm:text-base font-medium leading-relaxed text-slate-500 dark:text-slate-400 max-w-2xl">
            Kelola bank soal, pantau aktivitas ujian siswa secara real-time, dan evaluasi hasil belajar melalui panel kendali Anda.
          </p>

          <!-- Assigned Info (Mobile Optimized) -->
          <div class="mt-6 flex flex-wrap gap-4 pt-6 border-t border-slate-100 dark:border-slate-800">
            <div v-if="assignments.levels.length" class="flex flex-col gap-1.5">
              <span class="text-[9px] text-slate-400 font-black uppercase tracking-widest">Jenjang Ampuan</span>
              <div class="flex flex-wrap gap-1.5">
                <span v-for="l in assignments.levels" :key="l.id" class="px-2.5 py-1 bg-emerald-50 dark:bg-emerald-900/20 text-emerald-700 dark:text-emerald-400 text-[10px] font-black rounded-lg border border-emerald-100 dark:border-emerald-800/50 uppercase">
                  {{ l.name }}
                </span>
              </div>
            </div>
            <div v-if="assignments.groups.length" class="flex flex-col gap-1.5">
              <span class="text-[9px] text-slate-400 font-black uppercase tracking-widest">Kelas / Group</span>
              <div class="flex flex-wrap gap-1.5">
                <span v-for="g in assignments.groups" :key="g.id" class="px-2.5 py-1 bg-blue-50 dark:bg-blue-900/20 text-blue-700 dark:text-blue-400 text-[10px] font-black rounded-lg border border-blue-100 dark:border-blue-800/50 uppercase">
                  {{ g.name }}
                </span>
              </div>
            </div>
            <div v-if="assignments.subjects.length" class="flex flex-col gap-1.5">
              <span class="text-[9px] text-slate-400 font-black uppercase tracking-widest">Mata Pelajaran</span>
              <div class="flex flex-wrap gap-1.5">
                <span v-for="s in assignments.subjects" :key="s.id" class="px-2.5 py-1 bg-indigo-50 dark:bg-indigo-900/20 text-indigo-700 dark:text-indigo-400 text-[10px] font-black rounded-lg border border-indigo-100 dark:border-indigo-800/50 uppercase">
                  {{ s.name }}
                </span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div class="mb-6 grid gap-4 sm:gap-6 grid-cols-2 lg:grid-cols-4">
        <DashboardCard
          color="emerald"
          :icon="mdiBookOpenVariant"
          label="Bank Soal"
          :number="isLoading ? '...' : stats.bankSoal"
          small
        />
        <DashboardCard
          color="sky"
          :icon="mdiCalendarClockOutline"
          label="Sesi Ujian"
          :number="isLoading ? '...' : stats.sesiUjian"
          small
        />
        <DashboardCard
          color="amber"
          :icon="mdiKeyVariant"
          label="Token Aktif"
          :number="isLoading ? '...' : stats.tokens"
          small
        />
        <DashboardCard
          color="rose"
          :icon="mdiMonitorEye"
          label="Siswa Online"
          :number="isLoading ? '...' : stats.onlineStudents"
          small
        />
      </div>

      <div class="grid gap-4 sm:gap-6 grid-cols-2 lg:grid-cols-3">
        <QuickMenuCard
          label="Bank Soal"
          description="Kelola & Import Soal"
          :icon="mdiBookOpenVariant"
          to="/teacher/bank-soal"
          color="emerald"
        />
        <QuickMenuCard
          label="Jadwal Ujian"
          description="Atur Sesi & Waktu"
          :icon="mdiCalendarClockOutline"
          to="/teacher/ujian/jadwal"
          color="blue"
        />
        <div class="col-span-2 lg:col-span-1">
          <QuickMenuCard
            label="Monitor Live"
            description="Pantau Progress Siswa"
            :icon="mdiMonitorEye"
            to="/teacher/ujian/monitor-ujian"
            color="amber"
          />
        </div>
      </div>

      <CardBox v-if="showGradeChart" class="mt-6 p-6 animate-fade-in relative">
        <button 
          @click="showGradeChart = false"
          class="absolute top-6 right-6 p-2 hover:bg-slate-100 dark:hover:bg-slate-800 rounded-lg text-slate-400 transition-colors"
          title="Sembunyikan Panel"
        >
          <BaseIcon :path="mdiClose" size="18" />
        </button>
        <div class="flex items-center mb-6">
          <div class="p-2 bg-indigo-100 dark:bg-indigo-900/30 rounded-xl mr-3">
            <BaseIcon :path="mdiBookOpenVariant" size="24" class="text-indigo-600 dark:text-indigo-400" />
          </div>
          <div>
            <h3 class="text-lg font-bold text-slate-800 dark:text-slate-100">Ringkasan Nilai per Jenjang</h3>
            <p class="text-xs text-slate-500">Performa rata-rata siswa di kelas Anda</p>
          </div>
        </div>
        <div class="h-64 w-full">
          <BaseChart type="bar" :data="avgScoreData" :options="chartOptions" />
        </div>
      </CardBox>
      <div v-else class="mt-4 flex justify-end">
         <button 
            @click="showGradeChart = true"
            class="text-[10px] font-black uppercase tracking-widest text-indigo-600 hover:underline"
         >
            + Tampilkan Statistik Nilai
         </button>
      </div>
    </SectionMain>
  </LayoutAuthenticated>
</template>
