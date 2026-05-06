<script setup>
import { computed, onMounted, onBeforeUnmount, ref } from 'vue'
import { useRouter } from 'vue-router'
import {
  mdiClipboardTextOutline,
  mdiRefresh,
  mdiCheckCircleOutline,
  mdiDeleteOutline,
} from '@mdi/js'
import BaseIcon from '@/components/BaseIcon.vue'
import LayoutAuthenticated from '@/layouts/LayoutAuthenticated.vue'
import SectionMain from '@/components/SectionMain.vue'
import SectionTitleLineWithButton from '@/components/SectionTitleLineWithButton.vue'
import BaseButton from '@/components/BaseButton.vue'
import { api } from '@/services/api.js'
const router = useRouter()

const filter = ref('all')
const search = ref('')

const isLoading = ref(false)
const errorMessage = ref('')
const backendExams = ref([])
const deletingExamId = ref('')

let nowInterval = null
const nowMs = ref(Date.now())

const classifyStatus = (item) => {
  const endsAt = item?.ends_at ? new Date(item.ends_at) : null
  const startsAt = item?.starts_at ? new Date(item.starts_at) : null
  const isEnded = endsAt ? nowMs.value > endsAt.getTime() : false
  const isStarted = startsAt ? nowMs.value >= startsAt.getTime() : true
  const sessionStatus = String(item?.session_status || '')

  if (sessionStatus && sessionStatus !== 'in_progress') return 'completed'
  if (isEnded) return 'completed'
  if (!isStarted) return 'upcoming'

  // Refine active status based on backend can_join (which handles session windows)
  if (!item.can_join) {
    // If we're between startsAt and endsAt but can't join, it's a session issue
    // We'll treat it as upcoming (not yet) or completed (passed) based on session window
    if (item.session_end) {
       // Best effort check on frontend
       // But for now, returning 'upcoming' is safer to show "Belum Waktunya"
       return 'upcoming'
    }
  }

  return 'active'
}

const formatDateTime = (value) => {
  if (!value) return '-'
  const d = new Date(value)
  if (Number.isNaN(d.getTime())) return value
  const formatted = d
    .toLocaleString('id-ID', {
      dateStyle: 'medium',
      timeStyle: 'short',
      hour12: false,
    })
    .replace(/\./g, ':')

  const offset = -d.getTimezoneOffset() / 60
  let tz = ''
  if (offset === 7) tz = 'WIB'
  else if (offset === 8) tz = 'WITA'
  else if (offset === 9) tz = 'WIT'
  else tz = offset >= 0 ? `GMT+${offset}` : `GMT${offset}`

  return `${formatted} ${tz}`
}

const formatCountdown = (ms) => {
  if (ms <= 0) return '00:00:00'
  const totalSec = Math.floor(ms / 1000)
  const h = String(Math.floor(totalSec / 3600)).padStart(2, '0')
  const m = String(Math.floor((totalSec % 3600) / 60)).padStart(2, '0')
  const s = String(totalSec % 60).padStart(2, '0')
  return `${h}:${m}:${s}`
}

const exams = computed(() => {
  const q = search.value.trim().toLowerCase()
  const normalized = backendExams.value.map((item) => ({
    ...item,
    _ui_status: classifyStatus(item),
  }))
  let filtered = filter.value === 'all' ? normalized : normalized.filter((item) => item._ui_status === filter.value)
  if (q) {
    filtered = filtered.filter(
      (item) =>
        (item.title || '').toLowerCase().includes(q) ||
        (item.subject_name || '').toLowerCase().includes(q) ||
        (item.teacher_name || '').toLowerCase().includes(q),
    )
  }
  return filtered
})

const upcomingCount = computed(() => backendExams.value.filter((e) => classifyStatus(e) === 'upcoming').length)
const activeCount = computed(() => backendExams.value.filter((e) => classifyStatus(e) === 'active').length)
const completedCount = computed(() => backendExams.value.filter((e) => classifyStatus(e) === 'completed').length)

const loadExams = async () => {
  isLoading.value = true
  errorMessage.value = ''
  try {
    const { data } = await api.get('/api/v1/student/exams', { params: { limit: 50, offset: 0 } })
    backendExams.value = data?.data || []
  } catch (error) {
    backendExams.value = []
    errorMessage.value = error?.response?.data?.error?.message || 'Gagal memuat daftar ujian dari backend'
  } finally {
    isLoading.value = false
  }
}

const joinExam = async (exam) => {
  if (!exam?.id) return
  router.push(`/student/ujian/${exam.id}/token`)
}

const dismissCompletedExam = async (exam) => {
  if (!exam?.id) return
  if (exam._ui_status !== 'completed') return
  if (!confirm(`Hapus kartu ujian "${exam.title}" dari daftar?`)) return

  deletingExamId.value = exam.id
  errorMessage.value = ''
  try {
    await api.delete(`/api/v1/student/exams/${exam.id}/dismiss`)
    backendExams.value = backendExams.value.filter((item) => item.id !== exam.id)
  } catch (error) {
    errorMessage.value = error?.response?.data?.error?.message || 'Gagal menghapus kartu ujian'
  } finally {
    deletingExamId.value = ''
  }
}

onMounted(() => {
  loadExams()
  nowInterval = window.setInterval(() => {
    nowMs.value = Date.now()
  }, 1000)
})

onBeforeUnmount(() => {
  if (nowInterval) window.clearInterval(nowInterval)
})
</script>

<template>
  <LayoutAuthenticated>
    <SectionMain>
      <SectionTitleLineWithButton :icon="mdiClipboardTextOutline" title="Ruang Ujian" main>
        <BaseButton :icon="mdiRefresh" color="info" label="Refresh" @click="loadExams" />
      </SectionTitleLineWithButton>

      <!-- Stats bar -->
      <div class="mb-6 grid gap-4 grid-cols-2 sm:grid-cols-3">
        <div
          class="flex items-center gap-4 rounded-2xl border border-sky-400/60 bg-white px-5 py-4 shadow-sm dark:border-sky-800/80 dark:bg-slate-900/50"
        >
          <div class="flex h-11 w-11 items-center justify-center rounded-xl bg-sky-100 dark:bg-sky-900/30">
            <svg class="h-5 w-5 text-sky-600 dark:text-sky-400" viewBox="0 0 24 24" fill="currentColor">
              <path d="M12 2a10 10 0 1 0 10 10A10 10 0 0 0 12 2m1 11h-5V7h2v4h3Z" />
            </svg>
          </div>
          <div>
            <div class="text-2xl font-black text-slate-800 dark:text-white">{{ upcomingCount }}</div>
            <div class="text-[11px] font-bold uppercase tracking-widest text-slate-400 dark:text-slate-500">Mendatang</div>
          </div>
        </div>

        <div
          class="flex items-center gap-4 rounded-2xl border border-emerald-400/60 bg-emerald-50/40 px-5 py-4 shadow-sm dark:border-emerald-800/80 dark:bg-emerald-900/10"
        >
          <div class="flex h-11 w-11 items-center justify-center rounded-xl bg-emerald-100 dark:bg-emerald-900/40">
            <svg class="h-5 w-5 text-emerald-600 dark:text-emerald-400" viewBox="0 0 24 24" fill="currentColor">
              <path d="m10 16.4-4-4L7.4 11l2.6 2.6L16.6 7 18 8.4Z" />
            </svg>
          </div>
          <div>
            <div class="text-2xl font-black text-slate-800 dark:text-white">{{ activeCount }}</div>
            <div class="text-[11px] font-bold uppercase tracking-widest text-emerald-600 dark:text-emerald-400">Aktif Sekarang</div>
          </div>
        </div>

        <div
          class="flex items-center gap-4 rounded-2xl border border-purple-400/60 bg-white px-5 py-4 shadow-sm dark:border-purple-800/80 dark:bg-slate-900/50"
        >
          <div class="flex h-11 w-11 items-center justify-center rounded-xl bg-slate-100 dark:bg-slate-800">
            <svg class="h-5 w-5 text-slate-500 dark:text-slate-400" viewBox="0 0 24 24" fill="currentColor">
              <path d="M9 16.17 4.83 12l-1.42 1.41L9 19 21 7l-1.41-1.41Z" />
            </svg>
          </div>
          <div>
            <div class="text-2xl font-black text-slate-800 dark:text-white">{{ completedCount }}</div>
            <div class="text-[11px] font-bold uppercase tracking-widest text-slate-400 dark:text-slate-500">Selesai</div>
          </div>
        </div>
      </div>

      <!-- Filter & Search -->
      <div class="mb-6 flex flex-col gap-3 sm:flex-row sm:items-center border border-blue-400/60 dark:border-blue-800/80 p-4 rounded-2xl bg-blue-50/10 dark:bg-blue-900/10">
        <div class="relative flex-1">
          <svg
            class="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-400"
            viewBox="0 0 24 24"
            fill="currentColor"
          >
            <path
              d="M9.5 3A6.5 6.5 0 0 1 16 9.5c0 1.61-.59 3.09-1.56 4.23l.27.27H15l5 5-1.5 1.5-5-5v-.79l-.27-.27A6.516 6.516 0 0 1 9.5 16 6.5 6.5 0 0 1 3 9.5 6.5 6.5 0 0 1 9.5 3m0 2C7 5 5 7 5 9.5S7 14 9.5 14 14 12 14 9.5 12 5 9.5 5Z"
            />
          </svg>
          <input
            v-model="search"
            type="text"
            placeholder="Cari ujian, mapel, pengajar..."
            class="w-full rounded-xl border border-slate-200 bg-white py-2.5 pl-9 pr-4 text-sm text-slate-700 placeholder-slate-400 shadow-sm outline-none transition focus:border-sky-500 focus:ring-2 focus:ring-sky-500/20 dark:border-slate-700 dark:bg-slate-900 dark:text-slate-200 dark:placeholder-slate-500 dark:focus:border-sky-500"
          />
        </div>
        <div class="flex gap-2 overflow-x-auto pb-1 scrollbar-hide">
          <button
            v-for="tab in [
              { value: 'all', label: 'Semua' },
              { value: 'active', label: 'Aktif' },
              { value: 'upcoming', label: 'Mendatang' },
              { value: 'completed', label: 'Selesai' },
            ]"
            :key="tab.value"
            type="button"
            class="whitespace-nowrap rounded-lg px-4 py-2 text-xs font-bold uppercase tracking-widest transition-all"
            :class="
              filter === tab.value
                ? 'bg-sky-600 text-white shadow-sm shadow-sky-500/30'
                : 'bg-white text-slate-600 border border-slate-200 hover:bg-slate-50 dark:bg-slate-900 dark:text-slate-400 dark:border-slate-700 dark:hover:bg-slate-800'
            "
            @click="filter = tab.value"
          >
            {{ tab.label }}
          </button>
        </div>
      </div>

      <!-- Error state -->
      <div
        v-if="errorMessage"
        class="mb-6 rounded-xl border border-red-100 bg-red-50 px-4 py-3 text-sm text-red-700 dark:border-red-900/30 dark:bg-red-900/20 dark:text-red-400"
      >
        {{ errorMessage }}
      </div>

      <!-- Loading skeleton -->
      <div v-if="isLoading" class="grid gap-5 md:grid-cols-2 xl:grid-cols-3">
        <div
          v-for="i in 3"
          :key="i"
          class="animate-pulse rounded-2xl border border-slate-100 bg-white p-6 dark:border-slate-800 dark:bg-slate-900/50"
        >
          <div class="mb-4 h-4 rounded bg-slate-100 dark:bg-slate-800 w-3/4"></div>
          <div class="mb-2 h-3 rounded bg-slate-100 dark:bg-slate-800 w-1/2"></div>
          <div class="mt-6 h-9 rounded-xl bg-slate-100 dark:bg-slate-800"></div>
        </div>
      </div>

      <!-- Empty state -->
      <div
        v-else-if="!exams.length"
        class="flex flex-col items-center justify-center rounded-2xl border border-dashed border-slate-200 bg-white py-20 dark:border-slate-700 dark:bg-slate-900/30"
      >
        <div class="mb-4 flex h-16 w-16 items-center justify-center rounded-2xl bg-slate-100 dark:bg-slate-800">
          <svg class="h-8 w-8 text-slate-400 dark:text-slate-500" viewBox="0 0 24 24" fill="currentColor">
            <path
              d="M9 3v1H4v2h1v13a2 2 0 0 0 2 2h10a2 2 0 0 0 2-2V6h1V4h-5V3H9m0 5h2v9H9V8m4 0h2v9h-2V8Z"
            />
          </svg>
        </div>
        <p class="text-sm font-bold text-slate-400 dark:text-slate-500">Belum ada ujian yang tersedia</p>
        <p class="mt-1 text-xs text-slate-400 dark:text-slate-600">Ujian akan muncul di sini setelah ditargetkan oleh guru/admin.</p>
      </div>

      <!-- Exam cards grid -->
      <div v-else class="grid gap-5 md:grid-cols-2 xl:grid-cols-3">
        <div
          v-for="exam in exams"
          :key="exam.id"
          class="group relative flex flex-col overflow-hidden rounded-2xl border bg-white shadow-sm transition-all duration-300 hover:-translate-y-0.5 hover:shadow-lg dark:bg-slate-900/60"
          :class="{
            'border-emerald-400/60 dark:border-emerald-800/80 shadow-emerald-500/10': exam._ui_status === 'active',
            'border-sky-400/60 dark:border-sky-800/80': exam._ui_status === 'upcoming',
            'border-purple-400/60 dark:border-purple-800/80': exam._ui_status === 'completed',
          }"
        >
          <!-- Top accent bar -->
          <div
            class="h-1 w-full"
            :class="{
              'bg-gradient-to-r from-emerald-400 to-teal-500': exam._ui_status === 'active',
              'bg-gradient-to-r from-sky-400 to-blue-500': exam._ui_status === 'upcoming',
              'bg-gradient-to-r from-slate-300 to-slate-400 dark:from-slate-700 dark:to-slate-600': exam._ui_status === 'completed',
            }"
          ></div>

          <div class="flex flex-1 flex-col p-5">
            <!-- Header -->
            <div class="mb-4 flex items-start justify-between gap-3">
              <div class="flex-1 min-w-0">
                <h3 class="truncate text-base font-bold uppercase tracking-tight text-slate-800 dark:text-slate-100">
                  {{ exam.title }}
                </h3>
                <p class="mt-1 text-xs text-slate-500 dark:text-slate-400">
                  {{ exam.subject_name || exam.subject_id }}
                  <span v-if="exam.teacher_name">· {{ exam.teacher_name }}</span>
                  <span
                    v-if="exam.master_session_name"
                    class="ml-1 rounded border border-indigo-100 bg-indigo-50 px-1.5 py-0.5 text-[10px] font-bold uppercase text-indigo-600 dark:border-indigo-900/30 dark:bg-indigo-900/40 dark:text-indigo-400"
                  >
                    {{ exam.master_session_name }}
                  </span>
                </p>
              </div>

              <!-- Status badge -->
              <span
                class="shrink-0 rounded-full px-2.5 py-1 text-[10px] font-black uppercase tracking-wider"
                :class="{
                  'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/40 dark:text-emerald-400': exam._ui_status === 'active',
                  'bg-sky-100 text-sky-700 dark:bg-sky-900/30 dark:text-sky-400': exam._ui_status === 'upcoming',
                  'bg-slate-100 text-slate-500 dark:bg-slate-800 dark:text-slate-400': exam._ui_status === 'completed',
                }"
              >
                <span v-if="exam._ui_status === 'active'">⚡ Aktif</span>
                <span v-else-if="exam._ui_status === 'upcoming'">⏳ Mendatang</span>
                <span v-else>✓ Selesai</span>
              </span>
            </div>

            <!-- Info rows -->
            <div class="mb-4 space-y-2 text-xs text-slate-500 dark:text-slate-400">
              <div class="flex items-center gap-2">
                <svg class="h-3.5 w-3.5 shrink-0 text-slate-400 dark:text-slate-500" viewBox="0 0 24 24" fill="currentColor">
                  <path d="M12 2a10 10 0 1 0 10 10A10 10 0 0 0 12 2m1 11h-5V7h2v4h3Z" />
                </svg>
                <span class="font-medium">Mulai:</span>
                <span class="font-mono">{{ formatDateTime(exam.starts_at) }}</span>
              </div>
              <div v-if="exam.session_start && exam.session_end" class="flex items-center gap-2">
                <svg class="h-3.5 w-3.5 shrink-0 text-indigo-400 dark:text-indigo-500" viewBox="0 0 24 24" fill="currentColor">
                  <path d="M12 2a10 10 0 1 0 10 10A10 10 0 0 0 12 2m0 18a8 8 0 1 1 8-8 8 8 0 0 1-8 8Z" />
                  <path d="M12.5 7H11v6l5.2 3.2.8-1.3-4.5-2.7V7Z" />
                </svg>
                <span class="font-medium text-indigo-600 dark:text-indigo-400">Jadwal Sesi:</span>
                <span class="font-mono font-bold text-indigo-600 dark:text-indigo-400">{{ exam.session_start }} - {{ exam.session_end }}</span>
              </div>
              <div class="flex items-center gap-2">
                <svg class="h-3.5 w-3.5 shrink-0 text-slate-400 dark:text-slate-500" viewBox="0 0 24 24" fill="currentColor">
                  <path d="M12 20a8 8 0 0 1-8-8 8 8 0 0 1 8-8 8 8 0 0 1 8 8 8 8 0 0 1-8 8m0-18A10 10 0 0 0 2 12a10 10 0 0 0 10 10 10 10 0 0 0 10-10A10 10 0 0 0 12 2m-1 5v6l5.25 3.15.75-1.23-4.5-2.67V7H11Z" />
                </svg>
                <span class="font-medium">Durasi:</span>
                <span class="font-bold text-slate-700 dark:text-slate-200">{{ exam.duration_minutes ?? '-' }} menit</span>
              </div>

              <!-- Countdown timer for active exams -->
              <div
                v-if="exam._ui_status === 'active' && exam.ends_at && nowMs < new Date(exam.ends_at).getTime()"
                class="flex items-center gap-2 rounded-lg bg-emerald-50/80 px-3 py-1.5 dark:bg-emerald-900/20"
              >
                <span class="inline-block h-2 w-2 animate-pulse rounded-full bg-emerald-500"></span>
                <span class="font-bold text-emerald-700 dark:text-emerald-400">Sisa:</span>
                <span class="font-mono font-black text-emerald-700 dark:text-emerald-400">
                  {{ formatCountdown(new Date(exam.ends_at).getTime() - nowMs) }}
                </span>
              </div>

              <!-- Session status -->
              <div
                v-if="exam.session_status"
                class="flex items-center gap-2 text-[10px] font-bold uppercase tracking-widest"
              >
                <span
                  class="inline-block h-1.5 w-1.5 rounded-full"
                  :class="exam.session_status === 'in_progress' ? 'bg-amber-400 animate-pulse' : 'bg-slate-400'"
                ></span>
                <span>Sesi: {{ exam.session_status }}</span>
                <span v-if="exam.session_finished_at">
                  ({{ formatDateTime(exam.session_finished_at) }})
                </span>
              </div>
            </div>

            <!-- CTA Button -->
            <div class="mt-auto pt-3 border-t border-slate-100 dark:border-slate-800">
              <button
                v-if="exam._ui_status === 'active' || (exam._ui_status === 'upcoming' && exam.can_join)"
                @click="joinExam(exam)"
                class="flex w-full items-center justify-center gap-2 rounded-xl bg-gradient-to-r from-sky-500 to-blue-600 px-4 py-2.5 text-xs font-black uppercase tracking-widest text-white shadow-md shadow-sky-500/20 transition-all hover:shadow-sky-500/40 hover:opacity-90"
              >
                <svg class="h-4 w-4" viewBox="0 0 24 24" fill="currentColor">
                  <path d="M8 5.14v14l11-7-11-7Z" />
                </svg>
                Masuk Ruang Ujian
              </button>

              <div
                v-else-if="exam._ui_status === 'upcoming'"
                class="flex w-full items-center justify-center gap-2 rounded-xl border border-slate-200 bg-slate-50 px-4 py-2.5 text-xs font-bold uppercase tracking-widest text-slate-400 dark:border-slate-700 dark:bg-slate-800/50 dark:text-slate-500"
              >
                <svg class="h-4 w-4" viewBox="0 0 24 24" fill="currentColor">
                  <path d="M12 2a10 10 0 1 0 10 10A10 10 0 0 0 12 2m1 11h-5V7h2v4h3Z" />
                </svg>
                Belum Waktunya
              </div>

              <RouterLink
                v-else
                to="/student/hasil"
                class="flex w-full items-center justify-center gap-2 rounded-xl border border-slate-200 bg-white px-4 py-2.5 text-xs font-bold uppercase tracking-widest text-slate-600 shadow-sm transition-all hover:bg-slate-50 dark:border-slate-700 dark:bg-slate-800 dark:text-slate-400 dark:hover:bg-slate-700"
              >
                <svg class="h-4 w-4" viewBox="0 0 24 24" fill="currentColor">
                  <path d="M9 16.17 4.83 12l-1.42 1.41L9 19 21 7l-1.41-1.41Z" />
                </svg>
                Lihat Hasil
              </RouterLink>

              <button
                v-if="exam._ui_status === 'completed'"
                type="button"
                class="mt-2 flex w-full items-center justify-center gap-2 rounded-xl border border-red-200 bg-red-50 px-4 py-2.5 text-xs font-bold uppercase tracking-widest text-red-600 transition-all hover:bg-red-100 disabled:cursor-not-allowed disabled:opacity-60 dark:border-red-900/40 dark:bg-red-900/20 dark:text-red-300"
                :disabled="deletingExamId === exam.id"
                @click="dismissCompletedExam(exam)"
              >
                <BaseIcon :path="mdiDeleteOutline" size="16" />
                {{ deletingExamId === exam.id ? 'Menghapus...' : 'Hapus Kartu' }}
              </button>
            </div>
          </div>
        </div>
      </div>
    </SectionMain>
  </LayoutAuthenticated>
</template>
