<script setup>
import { computed, ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import {
  mdiHome,
  mdiClipboardText,
  mdiChartBox,
  mdiBullhorn,
  mdiBookOpenVariant,
  mdiAccountCircle,
  mdiCog,
  mdiFolder,
  mdiDotsGrid,
  mdiClose,
  mdiLogout,
  mdiCalendarClock,
  mdiKeyVariant,
  mdiMonitorEye,
  mdiAccountSearch,
  mdiAccountSwitch,
  mdiPrinter,
  mdiHistory,
  mdiClipboardTextSearch,
  mdiDatabaseExport,
  mdiAccountTie,
  mdiAccountSchool,
  mdiAccountCheck,
  mdiPlus,
  mdiFileDocument,
  mdiChartLine,
  mdiBullhornOutline,
  mdiWeatherSunny,
  mdiMoonWaningCrescent,
} from '@mdi/js'
import BaseIcon from '@/components/BaseIcon.vue'
import { useAuthStore } from '@/stores/auth.js'
import { useDarkModeStore } from '@/stores/darkMode.js'

const authStore = useAuthStore()
const darkModeStore = useDarkModeStore()
const router = useRouter()
const route = useRoute()

const isMoreOpen = ref(false)

// --- Main bottom nav items (max 4 + "More") ---
const mainItems = computed(() => {
  if (authStore.role === 'student') {
    return [
      { to: '/student/dashboard', icon: mdiHome, label: 'Home' },
      { to: '/student/ujian', icon: mdiClipboardText, label: 'Ujian' },
      { to: '/student/hasil', icon: mdiChartBox, label: 'Hasil' },
      { to: '/student/pengumuman', icon: mdiBullhorn, label: 'Info' },
    ]
  }
  if (authStore.role === 'teacher') {
    return [
      { to: '/teacher/dashboard', icon: mdiHome, label: 'Home' },
      { to: '/teacher/bank-soal', icon: mdiBookOpenVariant, label: 'Soal' },
      { to: '/teacher/ujian/jadwal', icon: mdiClipboardText, label: 'Jadwal' },
      { to: '/teacher/evaluasi', icon: mdiChartBox, label: 'Hasil' },
    ]
  }
  // Admin
  return [
    { to: '/admin/dashboard', icon: mdiHome, label: 'Home' },
    { to: '/admin/master-data/siswa', icon: mdiFolder, label: 'Master' },
    { to: '/admin/ujian/jadwal', icon: mdiClipboardText, label: 'Ujian' },
    { to: '/admin/evaluasi', icon: mdiChartBox, label: 'Evaluasi' },
  ]
})

// --- Extra items shown in "More" drawer ---
const moreItems = computed(() => {
  if (authStore.role === 'student') {
    return [{ label: 'Profil', icon: mdiAccountCircle, to: '/student/profil' }]
  }
  if (authStore.role === 'teacher') {
    return [
      { label: 'Profil', icon: mdiAccountCircle, to: '/teacher/profil' },
      {
        label: 'Monitor Ujian',
        icon: mdiMonitorEye,
        to: '/teacher/ujian/monitor-ujian',
      },
      {
        label: 'Monitor Peserta',
        icon: mdiAccountSearch,
        to: '/teacher/ujian/monitor-peserta',
      },
      {
        label: 'Token Ujian',
        icon: mdiKeyVariant,
        to: '/teacher/ujian/token',
      },
      {
        label: 'Reset Login',
        icon: mdiAccountSwitch,
        to: '/teacher/ujian/reset-login',
      },
    ]
  }
  // Admin
  return [
    { label: 'Pengumuman', icon: mdiBullhornOutline, to: '/admin/pengumuman' },
    { label: 'Guru', icon: mdiAccountTie, to: '/admin/master-data/guru' },
    { label: 'Siswa', icon: mdiAccountSchool, to: '/admin/master-data/siswa' },
    { label: 'Registrasi', icon: mdiAccountCheck, to: '/admin/master-data/verifikasi-pendaftaran' },
    { label: 'Sesi', icon: mdiCalendarClock, to: '/admin/master-data/sesi' },
    { label: 'Bank Soal', icon: mdiBookOpenVariant, to: '/admin/bank-soal' },
    { label: 'Buat Soal', icon: mdiPlus, to: '/admin/bank-soal/new' },
    { label: 'Impor Soal', icon: mdiFileDocument, to: '/admin/bank-soal/import' },
    { label: 'Token Ujian', icon: mdiKeyVariant, to: '/admin/ujian/token' },
    { label: 'Monitor Ujian', icon: mdiMonitorEye, to: '/admin/ujian/monitor-ujian' },
    { label: 'Mon. Peserta', icon: mdiAccountSearch, to: '/admin/ujian/monitor-peserta' },
    { label: 'Reset Login', icon: mdiAccountSwitch, to: '/admin/ujian/reset-login' },
    { label: 'Analitik', icon: mdiChartLine, to: '/admin/analytics' },
    { label: 'Cetak', icon: mdiPrinter, to: '/admin/cetak' },
    { label: 'Settings', icon: mdiCog, to: '/admin/settings' },
    { label: 'Log Aktivitas', icon: mdiHistory, to: '/admin/settings/activity-log' },
    { label: 'Audit Log', icon: mdiClipboardTextSearch, to: '/admin/settings/audit-log' },
    { label: 'Integrasi LMS', icon: mdiDatabaseExport, to: '/admin/lms' },
  ]
})

const hasMore = computed(() => moreItems.value.length > 0)

const isActive = (to) => route.path.startsWith(to)

const goTo = (to) => {
  isMoreOpen.value = false
  if (router.currentRoute.value.path !== to) {
    router.push(to)
  }
}

const handleLogout = () => {
  isMoreOpen.value = false
  authStore.logout()
  router.push('/login')
}

const toggleDarkMode = () => {
  darkModeStore.set(null, true)
}
</script>

<template>
  <!-- Bottom Navigation Bar -->
  <nav class="lg:hidden fixed bottom-0 left-0 right-0 z-50 bg-white/70 dark:bg-slate-950/70 backdrop-blur-xl border-t border-slate-200 dark:border-slate-800 pb-safe shadow-[0_-1px_24px_rgba(79,70,229,0.07)]">
    <div class="flex items-center justify-around h-16 px-1">
      <router-link
        v-for="item in mainItems"
        :key="item.to"
        :to="item.to"
        class="flex flex-col items-center justify-center flex-1 gap-0.5 transition-all duration-200 py-1 rounded-xl mx-0.5"
        :class="isActive(item.to)
          ? 'text-indigo-600 dark:text-indigo-400'
          : 'text-slate-400 dark:text-slate-500 hover:text-slate-600'"
      >
        <div
          class="flex items-center justify-center w-9 h-6 rounded-full transition-all duration-200"
          :class="isActive(item.to) ? 'bg-indigo-100 dark:bg-indigo-900/40' : ''"
        >
          <BaseIcon :path="item.icon" size="20" />
        </div>
        <span class="text-[9px] font-bold uppercase tracking-tighter leading-none">{{ item.label }}</span>
      </router-link>

      <!-- More button (only if there are extra items) -->
      <button
        v-if="hasMore"
        type="button"
        class="flex flex-col items-center justify-center flex-1 gap-0.5 transition-all duration-200 py-1 rounded-xl mx-0.5"
        :class="isMoreOpen ? 'text-indigo-600 dark:text-indigo-400' : 'text-slate-400 dark:text-slate-500'"
        @click="isMoreOpen = !isMoreOpen"
      >
        <div
          class="flex items-center justify-center w-9 h-6 rounded-full transition-all duration-200"
          :class="isMoreOpen ? 'bg-indigo-100 dark:bg-indigo-900/40' : ''"
        >
          <BaseIcon :path="isMoreOpen ? mdiClose : mdiDotsGrid" size="20" />
        </div>
        <span class="text-[9px] font-bold uppercase tracking-tighter leading-none">{{ isMoreOpen ? 'Tutup' : 'Lainnya' }}</span>
      </button>
    </div>
  </nav>

  <!-- More Drawer Overlay -->
  <Transition name="overlay-fade">
    <div
      v-if="isMoreOpen"
      class="lg:hidden fixed inset-0 z-40 bg-black/30 backdrop-blur-sm"
      @click="isMoreOpen = false"
    />
  </Transition>

  <!-- More Drawer Panel -->
  <Transition name="drawer-slide">
    <div
      v-if="isMoreOpen"
      class="lg:hidden fixed bottom-16 left-0 right-0 z-40 bg-white/95 dark:bg-slate-900/95 backdrop-blur-xl border-t border-slate-200 dark:border-slate-700 rounded-t-3xl shadow-2xl pb-safe-extra"
    >
      <!-- Drawer Handle -->
      <div class="flex justify-center pt-3 pb-1">
        <div class="w-10 h-1 rounded-full bg-slate-300 dark:bg-slate-600" />
      </div>

      <!-- Drawer Header -->
      <div class="px-5 pb-3 pt-1 flex items-center justify-between border-b border-slate-100 dark:border-slate-800">
        <span class="text-xs font-black uppercase tracking-widest text-slate-400 dark:text-slate-500">Menu Lainnya</span>
        <button
          type="button"
          class="text-slate-400 hover:text-slate-600 dark:hover:text-slate-300 transition-colors"
          @click="isMoreOpen = false"
        >
          <BaseIcon :path="mdiClose" size="18" />
        </button>
      </div>

      <!-- Menu Grid -->
      <div class="px-4 pt-3 pb-4 max-h-[60vh] overflow-y-auto">
        <div class="grid grid-cols-4 gap-2">
          <button
            v-for="item in moreItems"
            :key="item.to"
            type="button"
            class="flex flex-col items-center gap-1.5 p-2 rounded-2xl transition-all duration-150 active:scale-95"
            :class="isActive(item.to)
              ? 'bg-indigo-50 dark:bg-indigo-900/30 text-indigo-600 dark:text-indigo-400'
              : 'text-slate-500 dark:text-slate-400 hover:bg-slate-50 dark:hover:bg-slate-800'"
            @click="goTo(item.to)"
          >
            <div
              class="flex items-center justify-center w-12 h-12 rounded-2xl shadow-sm border transition-all duration-150"
              :class="isActive(item.to)
                ? 'bg-indigo-100 dark:bg-indigo-900/50 border-indigo-200 dark:border-indigo-700'
                : 'bg-white dark:bg-slate-800 border-slate-200 dark:border-slate-700'"
            >
              <BaseIcon :path="item.icon" size="22" />
            </div>
            <span class="text-[9px] font-bold text-center leading-tight uppercase tracking-tighter line-clamp-2 w-full">{{ item.label }}</span>
          </button>
        </div>

        <!-- Divider + Logout -->
        <div class="mt-4 pt-3 border-t border-slate-100 dark:border-slate-800">
          <button
            type="button"
            class="mb-2 w-full flex items-center gap-3 px-4 py-3 rounded-2xl transition-colors active:scale-[0.98]"
            :class="darkModeStore.isEnabled
              ? 'text-indigo-300 hover:bg-indigo-900/20'
              : 'text-purple-700 hover:bg-purple-50'"
            @click="toggleDarkMode"
          >
            <div
              class="flex items-center justify-center w-10 h-10 rounded-xl border transition-colors"
              :class="darkModeStore.isEnabled
                ? 'bg-indigo-900/40 border-indigo-700/60'
                : 'bg-purple-100 border-purple-200'"
            >
              <BaseIcon
                :path="darkModeStore.isEnabled ? mdiWeatherSunny : mdiMoonWaningCrescent"
                size="20"
                :class="darkModeStore.isEnabled ? 'text-indigo-300' : 'text-purple-600'"
              />
            </div>
            <div class="flex flex-col items-start">
              <span class="text-sm font-black uppercase tracking-tight">Mode Tampilan</span>
              <span
                class="text-[10px]"
                :class="darkModeStore.isEnabled ? 'text-indigo-300/80' : 'text-purple-700/70'"
              >{{ darkModeStore.isEnabled ? 'Dark Mode Aktif' : 'Light Mode Aktif' }}</span>
            </div>
          </button>

          <button
            type="button"
            class="w-full flex items-center gap-3 px-4 py-3 rounded-2xl text-red-500 dark:text-red-400 hover:bg-red-50 dark:hover:bg-red-900/20 transition-colors active:scale-[0.98]"
            @click="handleLogout"
          >
            <div class="flex items-center justify-center w-10 h-10 rounded-xl bg-red-50 dark:bg-red-900/20 border border-red-100 dark:border-red-800">
              <BaseIcon :path="mdiLogout" size="20" />
            </div>
            <div class="flex flex-col items-start">
              <span class="text-sm font-black uppercase tracking-tight">Keluar</span>
              <span class="text-[10px] opacity-60">{{ authStore.userDisplayName || authStore.user?.username }}</span>
            </div>
          </button>
        </div>
      </div>
    </div>
  </Transition>
</template>

<style scoped>
.pb-safe {
  padding-bottom: env(safe-area-inset-bottom);
}
.pb-safe-extra {
  padding-bottom: calc(env(safe-area-inset-bottom) + 0.5rem);
}

/* Overlay fade */
.overlay-fade-enter-active,
.overlay-fade-leave-active {
  transition: opacity 0.2s ease;
}
.overlay-fade-enter-from,
.overlay-fade-leave-to {
  opacity: 0;
}

/* Drawer slide up */
.drawer-slide-enter-active,
.drawer-slide-leave-active {
  transition: transform 0.28s cubic-bezier(0.34, 1.56, 0.64, 1), opacity 0.2s ease;
}
.drawer-slide-enter-from,
.drawer-slide-leave-to {
  transform: translateY(100%);
  opacity: 0;
}
</style>
