<script setup>
import { computed, onMounted, ref } from 'vue'
import {
  mdiBellOutline,
  mdiClipboardTextClockOutline,
  mdiHomeOutline,
  mdiSchoolOutline,
} from '@mdi/js'
import LayoutAuthenticated from '@/layouts/LayoutAuthenticated.vue'
import SectionMain from '@/components/SectionMain.vue'
import SectionTitleLineWithButton from '@/components/SectionTitleLineWithButton.vue'
import CardBox from '@/components/CardBox.vue'
import DashboardCard from '@/components/DashboardCard.vue'
import BaseButton from '@/components/BaseButton.vue'
import BaseIcon from '@/components/BaseIcon.vue'
import BaseSkeleton from '@/components/BaseSkeleton.vue'
import { mdiContentCopy, mdiLockOutline } from '@mdi/js'
import { api } from '@/services/api.js'
import { useAuthStore } from '@/stores/auth.js'

const authStore = useAuthStore()

const exams = ref([])
const results = ref([])
const isExamsLoading = ref(false)
const isResultsLoading = ref(false)
const errorMessage = ref('')

const isLoading = computed(() => isExamsLoading.value || isResultsLoading.value)

const loadDashboard = async () => {
  isExamsLoading.value = true
  isResultsLoading.value = true
  errorMessage.value = ''
  try {
    const [examRes, resultRes] = await Promise.all([
      api.get('/api/v1/student/exams', { params: { limit: 50, offset: 0 } }),
      api.get('/api/v1/student/results', { params: { limit: 50, offset: 0 } }),
    ])
    exams.value = examRes?.data?.data || []
    results.value = resultRes?.data?.data || []
  } catch (err) {
    errorMessage.value = err?.response?.data?.error?.message || 'Gagal memuat dashboard'
  } finally {
    isExamsLoading.value = false
    isResultsLoading.value = false
  }
}

const classifyStatus = (item) => {
  const endsAt = item?.ends_at ? new Date(item.ends_at) : null
  const isEnded = endsAt ? Date.now() > endsAt.getTime() : false
  const sessionStatus = String(item?.session_status || '')

  if (sessionStatus && sessionStatus !== 'in_progress') return 'completed'
  if (isEnded) return 'completed'
  return 'upcoming'
}

const formatDateTime = (value) => {
  if (!value) return '-'
  const d = new Date(value)
  if (Number.isNaN(d.getTime())) return value
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
}

const upcomingExams = computed(() =>
  exams.value
    .filter((item) => classifyStatus(item) === 'upcoming')
    .sort((a, b) => new Date(a.starts_at).getTime() - new Date(b.starts_at).getTime())
    .slice(0, 5),
)

const completedExamsCount = computed(() => exams.value.filter((item) => classifyStatus(item) === 'completed').length)
const averageScore = computed(() => {
  if (!results.value.length) return 0
  const total = results.value.reduce((sum, item) => sum + Number(item.score || 0), 0)
  return Math.round(total / results.value.length)
})

const copyToken = (token) => {
  navigator.clipboard.writeText(token)
  alert('Token berhasil disalin: ' + token)
}

onMounted(() => {
  loadDashboard()
})
</script>

<template>
  <LayoutAuthenticated>
    <SectionMain>
      <SectionTitleLineWithButton :icon="mdiHomeOutline" title="Dashboard Siswa" main />

      <div class="mb-6 rounded-2xl bg-white dark:bg-slate-900/50 border border-blue-400/60 dark:border-blue-800/80 px-6 py-8 shadow-sm relative overflow-hidden transition-all hover:shadow-md">
        <!-- Decoration -->
        <div class="absolute top-0 right-0 -mt-10 -mr-10 h-64 w-64 rounded-full bg-blue-500/5 blur-3xl"></div>
        
        <div class="relative z-10 max-w-3xl">
          <div class="mb-2 text-[10px] font-black uppercase tracking-[0.4em] text-blue-600 dark:text-sky-400">Portal Peserta AtigaCBT</div>
          <h2 class="mb-2 text-2xl font-bold text-slate-800 dark:text-white">
            {{ authStore.user?.name || 'Peserta' }}, cek jadwal ujian terbaru kamu.
          </h2>
          <p class="text-sm font-medium text-slate-500 dark:text-slate-400">
            Semua aktivitas ujian dan hasil belajar kamu dapat dipantau dari panel ini.
          </p>
        </div>
      </div>

      <div v-if="errorMessage" class="mb-6 rounded-xl bg-red-50 dark:bg-red-900/20 px-4 py-3 text-sm text-red-700 dark:text-red-400 border border-red-100 dark:border-red-900/40 font-bold uppercase tracking-tighter">
        {{ errorMessage }}
      </div>

      <div class="mb-6 grid gap-6 md:grid-cols-3">
        <template v-if="isLoading">
          <CardBox v-for="i in 3" :key="i" class="h-32 flex flex-col justify-center">
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
          <DashboardCard
            :icon="mdiClipboardTextClockOutline"
            color="sky"
            label="Ujian Mendatang"
            :number="upcomingExams.length"
          />
          <DashboardCard
            :icon="mdiSchoolOutline"
            color="emerald"
            label="Ujian Selesai"
            :number="completedExamsCount"
          />
          <DashboardCard
            :icon="mdiBellOutline"
            color="amber"
            label="Rata-rata Nilai"
            :number="averageScore"
          />
        </template>
      </div>

      <div class="grid gap-6 xl:grid-cols-2">
        <CardBox color="blue">
          <div class="mb-6 flex items-center justify-between">
            <h3 class="text-lg font-bold dark:text-slate-100 uppercase tracking-tight">Jadwal Terdekat</h3>
            <BaseButton to="/student/ujian" color="info" label="Lihat Semua" small />
          </div>
          <div v-if="isLoading" class="space-y-4">
            <div v-for="i in 2" :key="i" class="rounded-xl border border-slate-200 dark:border-slate-800 p-5">
              <div class="flex justify-between mb-4">
                <BaseSkeleton width="w-40" height="h-6" />
                <BaseSkeleton width="w-20" height="h-5" rounded="rounded-full" />
              </div>
              <div class="space-y-2">
                <BaseSkeleton width="w-full" height="h-3" />
                <BaseSkeleton width="w-3/4" height="h-3" />
                <BaseSkeleton width="w-1/2" height="h-3" />
              </div>
            </div>
          </div>
          <div v-else-if="!upcomingExams.length" class="text-sm text-slate-500 dark:text-slate-400 italic">Belum ada ujian mendatang.</div>
          <div v-else class="space-y-4">
            <div
              v-for="exam in upcomingExams"
              :key="exam.id"
              class="rounded-xl border border-indigo-400/60 dark:border-indigo-800/80 p-5 bg-slate-50/30 dark:bg-slate-800/20"
            >
              <div class="mb-3 flex items-start justify-between gap-3">
                <div class="font-black text-slate-900 dark:text-white uppercase tracking-tighter text-lg">{{ exam.title }}</div>
                <span class="rounded-full bg-blue-100 text-blue-700 dark:bg-blue-900/40 dark:text-blue-400 px-3 py-1 text-[10px] font-black uppercase tracking-widest whitespace-nowrap shadow-sm">
                  {{ exam.subject_name || exam.subject_id }}
                </span>
              </div>
              <div class="space-y-2 text-xs text-slate-600 dark:text-slate-400">
                <div class="flex items-center gap-2">Guru: <span class="font-bold dark:text-slate-200">{{ exam.teacher_name || exam.teacher_id }}</span></div>
                <div v-if="exam.master_session_name" class="flex items-center gap-2 font-mono">Sesi: <span class="px-1.5 py-0.5 rounded bg-indigo-50 dark:bg-indigo-900/30 text-indigo-700 dark:text-indigo-300 font-bold uppercase tracking-tight text-[10px]">{{ exam.master_session_name }}</span></div>
                <div class="flex items-center gap-2 font-mono">Mulai: <span class="italic text-slate-500">{{ formatDateTime(exam.starts_at) }}</span></div>
                <div class="flex items-center gap-2 font-mono">Selesai: <span class="italic text-slate-500">{{ formatDateTime(exam.ends_at) }}</span></div>
                
                <!-- Token Display -->
                <div v-if="exam.active_token" class="mt-4 flex items-center justify-between gap-2 rounded-lg border-2 border-dashed border-sky-100 bg-sky-50 px-3 py-2 dark:border-sky-900/30 dark:bg-sky-900/10 animate-fade-in">
                  <div class="flex items-center gap-2">
                    <BaseIcon :path="mdiLockOutline" size="14" class="text-sky-600 dark:text-sky-400" />
                    <span class="text-[10px] font-black uppercase tracking-widest text-sky-600 dark:text-sky-400">Token:</span>
                  </div>
                  <div class="flex items-center gap-2">
                    <span class="font-mono text-sm font-black text-sky-700 dark:text-sky-400">{{ exam.active_token }}</span>
                    <button 
                      @click.stop="copyToken(exam.active_token)"
                      class="text-sky-600 hover:text-sky-800 dark:text-sky-400 dark:hover:text-white transition-colors"
                      title="Salin Token"
                    >
                      <BaseIcon :path="mdiContentCopy" size="14" />
                    </button>
                  </div>
                </div>

                <div class="mt-4">
                  <BaseButton 
                    :to="`/student/kerjakan/${exam.id}`" 
                    color="info" 
                    label="Masuk Ujian" 
                    small 
                    class="w-full" 
                  />
                </div>
              </div>
            </div>
          </div>
        </CardBox>

        <CardBox color="purple">
          <div class="mb-6 flex items-center justify-between">
            <h3 class="text-lg font-bold dark:text-slate-100 uppercase tracking-tight">Info Peserta</h3>
            <BaseButton to="/student/hasil" color="purple" label="Lihat Hasil" small />
          </div>
          <div class="space-y-4 text-[13px] leading-relaxed text-slate-600 dark:text-slate-400 font-medium">
            <p class="flex items-start gap-2">
              <span class="mt-1 flex h-4 w-4 shrink-0 items-center justify-center rounded-full bg-blue-100 text-[10px] font-bold text-blue-600 dark:bg-blue-900/40 dark:text-blue-400">1</span>
              <span>Pilih ujian yang berstatus <b>Aktif</b> pada daftar jadwal terdekat atau menu Ruang Ujian.</span>
            </p>
            <p class="flex items-start gap-2">
              <span class="mt-1 flex h-4 w-4 shrink-0 items-center justify-center rounded-full bg-blue-100 text-[10px] font-bold text-blue-600 dark:bg-blue-900/40 dark:text-blue-400">2</span>
              <span>Gunakan kode <b>Token</b> yang tertera untuk masuk ke dalam ruang ujian.</span>
            </p>
            <p class="flex items-start gap-2">
              <span class="mt-1 flex h-4 w-4 shrink-0 items-center justify-center rounded-full bg-blue-100 text-[10px] font-bold text-blue-600 dark:bg-blue-900/40 dark:text-blue-400">3</span>
              <span>Pastikan mengklik tombol <b>Simpan & Selesai</b> setelah Anda selesai menjawab semua soal.</span>
            </p>
            <p class="flex items-start gap-2">
              <span class="mt-1 flex h-4 w-4 shrink-0 items-center justify-center rounded-full bg-amber-100 text-[10px] font-bold text-amber-600 dark:bg-amber-900/40 dark:text-amber-400">!</span>
              <span>Jika terjadi kendala teknis (terputus/logout), Anda dapat masuk kembali selama waktu ujian masih tersedia.</span>
            </p>
          </div>
        </CardBox>
      </div>
    </SectionMain>
  </LayoutAuthenticated>
</template>
