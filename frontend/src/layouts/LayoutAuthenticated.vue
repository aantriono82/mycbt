<script setup>
import { mdiForwardburger, mdiBackburger, mdiMenu, mdiAccountCircleOutline, mdiBellOutline, mdiClipboardTextClockOutline, mdiSchoolOutline } from '@mdi/js'
import { computed, onMounted, ref, watch, onUnmounted } from 'vue'
import { useQuery, useQueryClient } from '@tanstack/vue-query'
import { useRouter } from 'vue-router'
import { getMenuAsideMain, menuAsideBottom } from '@/menuAside.js'
import { getMenuNavBar } from '@/menuNavBar.js'
import { useDarkModeStore } from '@/stores/darkMode.js'
import { homeRouteForRole, useAuthStore } from '@/stores/auth.js'
import BaseIcon from '@/components/BaseIcon.vue'
import NavBar from '@/components/NavBar.vue'
import NavBarItemPlain from '@/components/NavBarItemPlain.vue'
import AsideMenu from '@/components/AsideMenu.vue'
import FooterBar from '@/components/FooterBar.vue'
import BackendHealthBanner from '@/components/BackendHealthBanner.vue'
import { api } from '@/services/api.js'

const layoutAsidePadding = 'xl:pl-60'

const darkModeStore = useDarkModeStore()
const authStore = useAuthStore()

const router = useRouter()

const isAsideMobileExpanded = ref(false)
const isAsideLgActive = ref(false)

router.beforeEach(() => {
  isAsideMobileExpanded.value = false
  isAsideLgActive.value = false
})

const menuAsideMain = computed(() => getMenuAsideMain(authStore.role))
const menuNavBar = computed(() => getMenuNavBar(authStore.role, authStore.userDisplayName))



const queryClient = useQueryClient()
const announcements = ref([])
const exams = ref([])
const isNotificationsExpanded = ref(false)
const showBadge = ref(false)

// Real-time listener
onMounted(() => {
  if (authStore.role === 'student') {
    const baseUrl = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'
    const token = localStorage.getItem('mycbt_token')
    const es = new EventSource(`${baseUrl}/api/v1/student/notifications/stream?token=${token}`)
    
    es.addEventListener('update', () => {
      queryClient.invalidateQueries({ queryKey: ['student', 'notifications'] })
    })

    onUnmounted(() => {
      es.close()
    })
  }
})

const { isLoading: isLoadingAnnouncements } = useQuery({
  queryKey: ['student', 'notifications'],
  queryFn: async () => {
    if (authStore.role !== 'student') return { announcements: [], exams: [] }
    const [annResp, examsResp] = await Promise.all([
      api.get('/api/v1/student/announcements', { params: { limit: 5, offset: 0 } }),
      api.get('/api/v1/student/exams', { params: { limit: 5, offset: 0 } })
    ])
    
    const newAnns = annResp?.data?.data || []
    const upcoming = (examsResp?.data?.data || []).filter(item => {
      const endsAt = item?.ends_at ? new Date(item.ends_at) : null
      return !endsAt || Date.now() < endsAt.getTime()
    })

    announcements.value = newAnns
    exams.value = upcoming.slice(0, 3)

    // Badge logic
    const lastSeenAnnId = localStorage.getItem('last_seen_announcement_id')
    const lastSeenExamId = localStorage.getItem('last_seen_exam_id')
    const hasNewAnn = newAnns[0]?.id?.toString() && newAnns[0]?.id?.toString() !== lastSeenAnnId
    const hasNewExam = upcoming[0]?.id?.toString() && upcoming[0]?.id?.toString() !== lastSeenExamId

    if (hasNewAnn || hasNewExam) {
      showBadge.value = true
    }

    return { announcements: newAnns, exams: upcoming }
  },
  refetchInterval: 1000 * 60, // Polling automatic every 1 min
  enabled: computed(() => authStore.role === 'student'),
})

const announcementCount = computed(() => announcements.value.length)
const examCount = computed(() => exams.value.length)
const notificationsCount = computed(() => announcementCount.value + examCount.value)

const markNotificationsAsRead = () => {
  isNotificationsExpanded.value = !isNotificationsExpanded.value
  if (isNotificationsExpanded.value) {
    showBadge.value = false
    const latestAnnId = announcements.value[0]?.id
    const latestExamId = exams.value[0]?.id
    if (latestAnnId) localStorage.setItem('last_seen_announcement_id', latestAnnId.toString())
    if (latestExamId) localStorage.setItem('last_seen_exam_id', latestExamId.toString())
  }
}

const menuClick = (event, item) => {
  if (item.isToggleLightDark) {
    darkModeStore.set(null, true)
  }

  if (item.isLogout) {
    authStore.logout()
    router.push('/login')
  }
}
</script>

<template>
  <div
    :class="{
      'overflow-hidden lg:overflow-visible': isAsideMobileExpanded,
    }"
  >
    <div
      :class="[layoutAsidePadding, { 'ml-60 lg:ml-0': isAsideMobileExpanded }]"
      class="min-h-screen w-screen bg-slate-50 pt-14 transition-(--transition-position) lg:w-auto dark:bg-slate-950 dark:text-slate-100"
    >
      <NavBar
        :menu="menuNavBar"
        :class="[layoutAsidePadding, { 'ml-60 lg:ml-0': isAsideMobileExpanded }]"
        @menu-click="menuClick"
      >
        <NavBarItemPlain
          display="flex lg:hidden"
          @click.prevent="isAsideMobileExpanded = !isAsideMobileExpanded"
        >
          <BaseIcon :path="isAsideMobileExpanded ? mdiBackburger : mdiForwardburger" size="24" />
        </NavBarItemPlain>
        <NavBarItemPlain display="hidden lg:flex xl:hidden" @click.prevent="isAsideLgActive = true">
          <BaseIcon :path="mdiMenu" size="24" />
        </NavBarItemPlain>

        <template #right>
          <div v-if="authStore.role === 'student'" class="relative flex items-center h-14">
            <NavBarItemPlain 
              @click.prevent="markNotificationsAsRead"
              class="relative"
            >
              <div :class="{ 'animate-pulse drop-shadow-[0_0_10px_rgba(59,130,246,0.6)]': showBadge && notificationsCount > 0 }">
                <BaseIcon :path="mdiBellOutline" size="24" :class="showBadge && notificationsCount > 0 ? 'text-blue-600 dark:text-blue-400' : ''" />
              </div>
              <div v-if="showBadge && notificationsCount > 0" class="absolute top-3 right-3 flex h-4 w-4 items-center justify-center rounded-full bg-red-500 text-[10px] font-bold text-white shadow-sm ring-2 ring-white dark:ring-slate-950">
                {{ notificationsCount > 9 ? '9+' : notificationsCount }}
              </div>
              <!-- Pulse ring -->
              <div v-if="showBadge && notificationsCount > 0" class="absolute inset-0 rounded-full bg-blue-400/20 animate-ping pointer-events-none"></div>
            </NavBarItemPlain>

            <!-- Notifications Dropdown -->
            <div 
              v-if="isNotificationsExpanded"
              class="absolute top-full right-0 mt-2 w-80 bg-white dark:bg-slate-900 rounded-2xl shadow-2xl border border-slate-100 dark:border-slate-800 z-50 overflow-hidden"
            >
              <div class="p-4 border-b border-slate-100 dark:border-slate-800 flex justify-between items-center bg-slate-50 dark:bg-slate-800/50">
                <h3 class="text-xs font-black uppercase tracking-widest text-slate-500">Pusat Informasi</h3>
                <span class="text-[10px] font-bold text-blue-600 dark:text-blue-400">Terbaru</span>
              </div>

              <div class="max-h-[400px] overflow-y-auto divide-y divide-slate-100 dark:divide-slate-800">
                <!-- Announcements Section -->
                <div v-if="announcements.length">
                  <div class="px-4 py-2 bg-slate-50/50 dark:bg-slate-800/30 text-[9px] font-black uppercase tracking-widest text-slate-400">Pengumuman</div>
                  <div 
                    v-for="a in announcements" 
                    :key="a.id"
                    class="p-4 hover:bg-slate-50 dark:hover:bg-slate-800/50 transition-colors cursor-pointer"
                    @click="router.push('/student/announcements'); isNotificationsExpanded = false"
                  >
                    <div class="text-[10px] font-bold text-blue-600 dark:text-blue-400 mb-1 flex justify-between uppercase">
                      <span>{{ a.category || 'INFO' }}</span>
                      <span class="opacity-60">{{ a.published_at ? new Date(a.published_at).toLocaleDateString('id-ID') : '-' }}</span>
                    </div>
                    <div class="text-xs font-black text-slate-900 dark:text-white mb-1 line-clamp-1">{{ a.title }}</div>
                    <div class="text-[10px] text-slate-500 dark:text-slate-400 line-clamp-2 leading-relaxed">{{ a.body }}</div>
                  </div>
                </div>

                <!-- Exams Section -->
                <div v-if="exams.length">
                  <div class="px-4 py-2 bg-slate-50/50 dark:bg-slate-800/30 text-[9px] font-black uppercase tracking-widest text-emerald-500">Ujian Mendatang</div>
                  <div 
                    v-for="e in exams" 
                    :key="e.id"
                    class="p-4 hover:bg-slate-50 dark:hover:bg-slate-800/50 transition-colors cursor-pointer flex gap-3"
                    @click="router.push('/student/ujian'); isNotificationsExpanded = false"
                  >
                    <div class="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-emerald-50 text-emerald-600 dark:bg-emerald-900/20 dark:text-emerald-400 shadow-sm border border-emerald-100 dark:border-emerald-800">
                      <BaseIcon :path="mdiClipboardTextClockOutline" size="20" />
                    </div>
                    <div class="min-w-0">
                      <div class="text-[10px] font-bold text-emerald-600 dark:text-emerald-400 mb-0.5 uppercase tracking-tighter">UJIAN BARU</div>
                      <div class="text-xs font-black text-slate-900 dark:text-white truncate">{{ e.title }}</div>
                      <div class="text-[10px] text-slate-500 dark:text-slate-400 italic">{{ e.subject_name || 'Mata Pelajaran' }}</div>
                    </div>
                  </div>
                </div>

                <div v-if="isLoadingAnnouncements" class="p-8 text-center text-xs text-slate-400 italic">
                  Memuat data...
                </div>
                <div v-else-if="!announcements.length && !exams.length" class="p-8 text-center text-xs text-slate-400 italic">
                  Tidak ada informasi baru.
                </div>
              </div>

              <div class="p-3 border-t border-slate-100 dark:border-slate-800 text-center bg-slate-50 dark:bg-slate-800/50">
                <button 
                  @click="router.push('/student/announcements'); isNotificationsExpanded = false"
                  class="text-[10px] font-black uppercase tracking-widest text-blue-600 hover:text-blue-700 dark:text-blue-400"
                >
                  Buka Pusat Informasi
                </button>
              </div>
            </div>
          </div>
        </template>
      </NavBar>
      <AsideMenu
        :is-aside-mobile-expanded="isAsideMobileExpanded"
        :is-aside-lg-active="isAsideLgActive"
        :menu="menuAsideMain"
        :menu-bottom="menuAsideBottom"
        @menu-click="menuClick"
        @aside-lg-close-click="isAsideLgActive = false"
      />
      <div
        v-if="authStore.isAuthenticated"
        class="border-b border-slate-200 bg-white px-6 py-2 text-xs text-slate-600 shadow-sm dark:border-slate-800 dark:bg-slate-900 dark:text-slate-300 flex items-center"
      >
        <BaseIcon :path="mdiAccountCircleOutline" size="14" class="mr-2 text-blue-600 dark:text-blue-400" />
        <span>
          Role: <span class="font-bold text-slate-900 dark:text-white">{{ authStore.roleLabel }}</span>
          <span v-if="authStore.user?.username" class="opacity-60"> (@{{ authStore.user.username }})</span>
        </span>
        <span class="mx-3 text-slate-300 dark:text-slate-700">|</span>
        <button class="font-extrabold text-blue-600 hover:text-blue-700 dark:text-blue-400 hover:underline transition-all uppercase" @click="router.push(homeRouteForRole(authStore.role))">
          Dashboard Utama
        </button>
        <span class="mx-3 text-slate-300 dark:text-slate-700">|</span>
        <BackendHealthBanner />
      </div>
      <slot />
      <FooterBar>
        <div />
      </FooterBar>
    </div>
  </div>
</template>
