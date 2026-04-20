<script setup>
import { ref, onMounted } from 'vue'
import { 
  mdiHomeOutline, 
  mdiBookOpenVariant, 
  mdiCalendarClockOutline, 
  mdiKeyVariant, 
  mdiMonitorEye 
} from '@mdi/js'
import LayoutAuthenticated from '@/layouts/LayoutAuthenticated.vue'
import SectionMain from '@/components/SectionMain.vue'
import SectionTitleLineWithButton from '@/components/SectionTitleLineWithButton.vue'
import CardBox from '@/components/CardBox.vue'
import DashboardCard from '@/components/DashboardCard.vue'
import BackendHealthBanner from '@/components/BackendHealthBanner.vue'
import QuickMenuCard from '@/components/QuickMenuCard.vue'
import { useAuthStore } from '@/stores/auth.js'
import { api } from '@/services/api.js'

const authStore = useAuthStore()
const isLoading = ref(false)
const stats = ref({
  bankSoal: 0,
  sesiUjian: 0,
  tokens: 0,
  onlineStudents: 0
})
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
    
    // For Token/Student counts, they are per-exam. 
    // Aggregate data not available yet in individual teacher endpoints.
    // Defaulting to 0/empty for now but could be expanded.
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
      <BackendHealthBanner />

      <div class="mb-6 rounded-2xl bg-white dark:bg-slate-900/50 border border-slate-100 dark:border-slate-800 px-6 py-8 shadow-sm relative overflow-hidden transition-all hover:shadow-md">
        <!-- Decoration -->
        <div class="absolute top-0 right-0 -mt-10 -mr-10 h-64 w-64 rounded-full bg-blue-500/5 blur-3xl"></div>
        
        <div class="relative z-10 max-w-3xl">
          <div class="mb-2 text-[10px] font-black uppercase tracking-[0.4em] text-blue-600 dark:text-sky-400">Panel Guru / Pengajar</div>
          <h2 class="mb-3 text-3xl font-bold text-slate-800 dark:text-white">
            Selamat datang, {{ authStore.user?.name || 'Rekan Guru' }}!
          </h2>
          <p class="text-sm font-medium leading-relaxed text-slate-500 dark:text-slate-400">
            Kelola bank soal, pantau aktivitas ujian siswa, dan evaluasi hasil belajar secara mandiri melalui panel dashboard Anda.
          </p>

          <!-- Assigned Info -->
          <div v-if="assignments.levels.length || assignments.groups.length" class="mt-4 flex flex-wrap gap-2 animate-fade-in">
            <div v-if="assignments.levels.length" class="flex items-center">
              <span class="text-[10px] text-slate-400 font-bold uppercase mr-2">Jenjang:</span>
              <div class="flex flex-wrap gap-1">
                <span v-for="l in assignments.levels" :key="l.id" class="px-2 py-0.5 bg-blue-50 dark:bg-blue-900/30 text-blue-600 dark:text-blue-400 text-[10px] font-bold rounded-md border border-blue-100 dark:border-blue-900/50">
                  {{ l.name }}
                </span>
              </div>
            </div>
            <div v-if="assignments.groups.length" class="flex items-center ml-2 border-l pl-3 dark:border-slate-800">
              <span class="text-[10px] text-slate-400 font-bold uppercase mr-2">Kelas/Group:</span>
              <div class="flex flex-wrap gap-1">
                <span v-for="g in assignments.groups" :key="g.id" class="px-2 py-0.5 bg-emerald-50 dark:bg-emerald-900/30 text-emerald-600 dark:text-emerald-400 text-[10px] font-bold rounded-md border border-emerald-100 dark:border-emerald-900/50">
                  {{ g.name }}
                </span>
              </div>
            </div>
          </div>
          <div v-if="assignments.subjects.length" class="mt-2 flex items-center animate-fade-in">
             <span class="text-[10px] text-slate-400 font-bold uppercase mr-2">Mata Pelajaran:</span>
             <div class="flex flex-wrap gap-1">
                <span v-for="s in assignments.subjects" :key="s.id" class="px-2 py-0.5 bg-indigo-50 dark:bg-indigo-900/30 text-indigo-600 dark:text-indigo-400 text-[10px] font-bold rounded-md border border-indigo-100 dark:border-indigo-900/50">
                  {{ s.name }} ({{ s.code }})
                </span>
              </div>
          </div>
        </div>
      </div>

      <div class="mb-6 grid gap-6 md:grid-cols-2 xl:grid-cols-4">
        <DashboardCard
          color="emerald"
          :icon="mdiBookOpenVariant"
          label="Bank Soal"
          :number="isLoading ? '...' : stats.bankSoal"
        />
        <DashboardCard
          color="sky"
          :icon="mdiCalendarClockOutline"
          label="Sesi Ujian"
          :number="isLoading ? '...' : stats.sesiUjian"
        />
        <DashboardCard
          color="amber"
          :icon="mdiKeyVariant"
          label="Token"
          :number="isLoading ? '...' : stats.tokens"
        />
        <DashboardCard
          color="rose"
          :icon="mdiMonitorEye"
          label="Siswa Online"
          :number="isLoading ? '...' : stats.onlineStudents"
        />
      </div>

      <div class="grid gap-6 grid-cols-1 md:grid-cols-2 lg:grid-cols-3">
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
        <QuickMenuCard
          label="Monitor Live"
          description="Pantau Progress Siswa"
          :icon="mdiMonitorEye"
          to="/teacher/ujian/monitor-ujian"
          color="amber"
        />
      </div>

      <CardBox class="mt-6">
        <div class="text-center py-10">
          <p class="text-sm font-semibold text-slate-400 dark:text-slate-500">
            Ringkasan aktivitas mengajar Anda akan muncul di sini.
          </p>
        </div>
      </CardBox>
    </SectionMain>
  </LayoutAuthenticated>
</template>

