<script setup>
import {
  mdiAccountCircleOutline,
  mdiBellOutline,
  mdiClipboardTextClockOutline,
  mdiCloseCircle,
  mdiMagnify,
  mdiMenu,
  mdiMenuOpen,
} from '@mdi/js'
import { computed, onMounted, ref, onUnmounted, watch } from 'vue'
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
import BottomNavigation from '@/components/BottomNavigation.vue'
import { api, getApiBaseUrl } from '@/services/api.js'

const darkModeStore = useDarkModeStore()
const authStore = useAuthStore()

const router = useRouter()

const isAsideLgActive = ref(false)
const isAsideDesktopHidden = ref(false)
const searchQuery = ref('')
const isSearchFocused = ref(false)
const isSearchMobileActive = ref(false)
const activeSearchIndex = ref(0)

const flattenSearchItems = (items = [], parents = []) => {
  const results = []
  for (const item of items) {
    const currentLabel = item?.label ? [...parents, item.label] : [...parents]
    if (item?.to && item?.label) {
      results.push({
        key: `${item.to}-${item.label}`,
        label: item.label,
        to: item.to,
        breadcrumbs: currentLabel.join(' > '),
      })
    }
    if (Array.isArray(item?.menu) && item.menu.length) {
      results.push(...flattenSearchItems(item.menu, currentLabel))
    }
  }
  return results
}

const searchItems = computed(() => {
  const base = flattenSearchItems(menuAsideMain.value)
  const dashboardRoute = homeRouteForRole(authStore.role)
  if (!base.some((item) => item.to === dashboardRoute)) {
    base.unshift({
      key: `${dashboardRoute}-dashboard`,
      label: 'Dashboard',
      to: dashboardRoute,
      breadcrumbs: 'Dashboard',
    })
  }
  return base
})

const filteredSearchItems = computed(() => {
  const query = searchQuery.value.trim().toLowerCase()
  if (!query) return searchItems.value.slice(0, 8)
  return searchItems.value
    .filter((item) =>
      [item.label, item.breadcrumbs, item.to].some((value) =>
        String(value || '').toLowerCase().includes(query),
      ),
    )
    .slice(0, 8)
})

const selectSearchItem = (item) => {
  if (!item?.to) return
  isSearchFocused.value = false
  searchQuery.value = item.label
  activeSearchIndex.value = 0
  if (router.currentRoute.value.path !== item.to) {
    router.push(item.to)
  }
}

const onSearchKeydown = (event) => {
  if (!filteredSearchItems.value.length) return
  if (event.key === 'ArrowDown') {
    event.preventDefault()
    activeSearchIndex.value = (activeSearchIndex.value + 1) % filteredSearchItems.value.length
    return
  }
  if (event.key === 'ArrowUp') {
    event.preventDefault()
    activeSearchIndex.value = (activeSearchIndex.value - 1 + filteredSearchItems.value.length) % filteredSearchItems.value.length
    return
  }
  if (event.key === 'Enter') {
    event.preventDefault()
    selectSearchItem(filteredSearchItems.value[activeSearchIndex.value] || filteredSearchItems.value[0])
  }
}

const clearSearch = () => {
  searchQuery.value = ''
  activeSearchIndex.value = 0
}

const toggleDesktopAside = () => {
  isAsideDesktopHidden.value = !isAsideDesktopHidden.value
}

watch(() => router.currentRoute.value.fullPath, () => {
  isAsideLgActive.value = false
  isSearchFocused.value = false
})

const layoutAsidePadding = computed(() => {
  if (isAsideDesktopHidden.value) {
    return ''
  }
  return 'lg:pl-60'
})

const menuAsideMain = computed(() => getMenuAsideMain(authStore.role))
const menuNavBar = computed(() => getMenuNavBar(authStore.role, authStore.userDisplayName))

const announcements = ref([])
const exams = ref([])
const announcementsTotal = ref(0)
const examsTotal = ref(0)
const isNotificationsExpanded = ref(false)
const showBadge = ref(false)
const isLoadingAnnouncements = ref(false)
const notificationsFingerprint = ref('')
const lastSeenNotificationsFingerprint = ref('')
const NOTIF_READ_KEY_PREFIX = 'atigacbt:last_seen_notifications:'
let notificationsPollTimer = null
let notificationsEventSource = null
let clockTimer = null
const digitalClock = ref('')

const updateDigitalClock = () => {
  const now = new Date()
  digitalClock.value = now.toLocaleTimeString('id-ID', {
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
    hour12: false,
  }).replace(/\./g, ':')
}

const currentUserId = computed(() => String(authStore.user?.id || authStore.user?.username || 'anon'))
const unreadAnnouncements = computed(() =>
  announcements.value.filter((item) => !item?.is_read),
)
const notificationsCount = computed(() => announcementsTotal.value + examsTotal.value)
const notificationsReadStorageKey = computed(() => {
  return `${NOTIF_READ_KEY_PREFIX}${currentUserId.value}`
})

const buildNotificationsFingerprint = (newAnns = [], newExams = [], annTotal = 0, examTotal = 0) => {
  const annKeys = (newAnns || []).map((item) =>
    `ann:${String(item?.id || '')}:${String(item?.updated_at || item?.published_at || '')}`,
  )
  const examKeys = (newExams || []).map((item) =>
    `exam:${String(item?.id || '')}:${String(item?.updated_at || item?.starts_at || '')}`,
  )
  return [`ann_total:${annTotal}`, ...annKeys, `exam_total:${examTotal}`, ...examKeys].join('|')
}

const markNotificationsAsRead = () => {
  lastSeenNotificationsFingerprint.value = notificationsFingerprint.value
  try {
    localStorage.setItem(notificationsReadStorageKey.value, lastSeenNotificationsFingerprint.value || '')
  } catch {
    // ignore storage errors
  }
  showBadge.value = false
}

const markAnnouncementItemsAsRead = async (items = []) => {
  const ids = items.map((item) => item?.id).filter(Boolean)
  if (!ids.length) return
  await api.post('/api/v1/student/announcements/read', {
    announcement_ids: ids,
  })
}

const openAnnouncementCenter = async () => {
  try {
    await markAnnouncementItemsAsRead(unreadAnnouncements.value)
  } catch (e) {
    console.error('Failed to mark announcements as read:', e)
  }
  markNotificationsAsRead()
  isNotificationsExpanded.value = false
  router.push('/student/announcements')
}

const openAnnouncementItem = async (item) => {
  if (item?.id) {
    try {
      await markAnnouncementItemsAsRead([item])
    } catch (e) {
      console.error('Failed to mark announcement as read:', e)
    }
  }
  markNotificationsAsRead()
  isNotificationsExpanded.value = false
  router.push('/student/announcements')
}

const loadNotifications = async ({ silent = false } = {}) => {
  if (authStore.role !== 'student') {
    announcements.value = []
    exams.value = []
    announcementsTotal.value = 0
    examsTotal.value = 0
    return
  }

  if (!silent) {
    isLoadingAnnouncements.value = true
  }
  try {
    const [annResp, examResp] = await Promise.all([
      api.get('/api/v1/student/announcements', { params: { limit: 5, offset: 0, unread_only: true } }),
      api.get('/api/v1/student/exams', { params: { limit: 5, offset: 0 } }),
    ])

    const newAnns = annResp.data?.data || []
    const newExams = examResp.data?.data || []
    announcements.value = newAnns
    exams.value = newExams
    announcementsTotal.value = Number(annResp.data?.meta?.total ?? newAnns.length)
    examsTotal.value = Number(examResp.data?.meta?.total ?? newExams.length)
    notificationsFingerprint.value = buildNotificationsFingerprint(newAnns, newExams, announcementsTotal.value, examsTotal.value)

    if (!notificationsFingerprint.value) {
      showBadge.value = false
      lastSeenNotificationsFingerprint.value = ''
    } else if (!lastSeenNotificationsFingerprint.value) {
      showBadge.value = true
    } else {
      showBadge.value = notificationsFingerprint.value !== lastSeenNotificationsFingerprint.value
    }

    if (isNotificationsExpanded.value && notificationsFingerprint.value) {
      markNotificationsAsRead()
    }
  } catch (e) {
    console.error('Failed to fetch notifications:', e)
    announcements.value = []
    exams.value = []
    announcementsTotal.value = 0
    examsTotal.value = 0
  } finally {
    isLoadingAnnouncements.value = false
  }
}

const stopNotificationsSync = () => {
  if (notificationsPollTimer) {
    clearInterval(notificationsPollTimer)
    notificationsPollTimer = null
  }
  if (notificationsEventSource) {
    notificationsEventSource.close()
    notificationsEventSource = null
  }
}

onMounted(() => {
  updateDigitalClock()
  clockTimer = setInterval(updateDigitalClock, 1000)

  if (authStore.role !== 'student') return
  try {
    lastSeenNotificationsFingerprint.value = localStorage.getItem(notificationsReadStorageKey.value) || ''
  } catch {
    lastSeenNotificationsFingerprint.value = ''
  }
  loadNotifications()

  notificationsPollTimer = setInterval(() => {
    loadNotifications({ silent: true })
  }, 60000)

  const baseUrl = getApiBaseUrl()
  const token = localStorage.getItem('atigacbt_token')
  notificationsEventSource = new EventSource(`${baseUrl}/api/v1/student/notifications/stream?access_token=${token}`)
  notificationsEventSource.addEventListener('update', () => {
    loadNotifications({ silent: true })
  })
})

onUnmounted(() => {
  if (clockTimer) {
    clearInterval(clockTimer)
    clockTimer = null
  }
  stopNotificationsSync()
})

watch(
  () => currentUserId.value,
  () => {
    if (authStore.role === 'student') {
      loadNotifications({ silent: true })
    }
  },
)

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
  <div>
      <div
        :class="[layoutAsidePadding]"
        class="min-h-screen w-screen bg-white pt-14 transition-(--transition-position) lg:w-auto dark:bg-slate-950 dark:text-slate-100 pb-24 lg:pb-0"
      >
      <NavBar
        :menu="menuNavBar"
        :class="[layoutAsidePadding]"
        @menu-click="menuClick"
      >

        <NavBarItemPlain 
          class="flex h-14 flex-none items-center px-4"
          @click="router.push(homeRouteForRole(authStore.role))"
        >
          <img src="/logo_atiga.png" alt="AtigaCBT Logo" class="w-8 h-8 md:w-9 md:h-9 object-contain rounded-full" />
          <div class="ml-2 flex flex-col justify-center leading-none">
            <span class="text-base md:text-lg font-black tracking-tighter text-slate-900 dark:text-white uppercase">AtigaCBT</span>
            <span class="text-[7px] md:text-[8px] font-bold text-blue-600 dark:text-blue-400 tracking-[0.08em] uppercase opacity-80">Professional CBT</span>
          </div>
        </NavBarItemPlain>

        <button
          type="button"
          class="hidden md:flex h-14 flex-none items-center px-3 text-slate-600 hover:text-blue-600 dark:text-slate-200 dark:hover:text-blue-400 transition"
          @click="toggleDesktopAside"
          aria-label="Toggle sidebar"
          title="Toggle Sidebar"
        >
          <BaseIcon :path="isAsideDesktopHidden ? mdiMenu : mdiMenuOpen" size="22" />
        </button>
        <!-- Desktop/Mobile Search Wrapper -->
        <div 
          class="flex flex-1 items-center px-4 transition-all duration-300"
          :class="[
            isSearchMobileActive ? 'absolute inset-0 z-50 bg-white dark:bg-slate-950 flex' : 'hidden md:flex'
          ]"
        >
          <!-- Mobile Close Search -->
          <button 
            v-if="isSearchMobileActive"
            type="button"
            class="mr-2 text-slate-500 md:hidden"
            @click="isSearchMobileActive = false"
          >
            <BaseIcon :path="mdiCloseCircle" size="24" />
          </button>

          <div class="relative w-full max-w-xl mx-auto">
            <label class="sr-only" for="global-search-input">Pencarian data</label>
            <BaseIcon
              :path="mdiMagnify"
              size="18"
              class="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-slate-400 dark:text-slate-500"
            />
            <input
              id="global-search-input"
              ref="searchInput"
              v-model="searchQuery"
              type="text"
              autocomplete="off"
              placeholder="Cari data/menu..."
              class="h-9 w-full rounded-xl border border-slate-200 bg-white pl-10 pr-10 text-sm font-semibold text-slate-700 shadow-sm outline-none transition focus:border-blue-400 focus:ring-2 focus:ring-blue-100 dark:border-slate-700 dark:bg-slate-900 dark:text-slate-100 dark:focus:border-blue-500 dark:focus:ring-blue-900/50"
              @focus="isSearchFocused = true"
              @blur="setTimeout(() => { isSearchFocused = false }, 120)"
              @keydown="onSearchKeydown"
            />
            <button
              v-if="searchQuery"
              type="button"
              class="absolute right-3 top-1/2 -translate-y-1/2 text-slate-400 transition hover:text-slate-600 dark:text-slate-500 dark:hover:text-slate-300"
              @click="clearSearch"
              aria-label="Hapus pencarian"
            >
              <BaseIcon :path="mdiCloseCircle" size="16" />
            </button>

            <div
              v-if="isSearchFocused"
              class="absolute z-50 mt-2 w-full overflow-hidden rounded-xl border border-slate-200 bg-white shadow-xl dark:border-slate-700 dark:bg-slate-900"
            >
              <div v-if="filteredSearchItems.length" class="max-h-80 overflow-y-auto py-1">
                <button
                  v-for="(item, idx) in filteredSearchItems"
                  :key="item.key"
                  type="button"
                  class="flex w-full flex-col px-3 py-2 text-left transition"
                  :class="idx === activeSearchIndex ? 'bg-blue-50 dark:bg-blue-900/30' : 'hover:bg-slate-50 dark:hover:bg-slate-800/70'"
                  @mouseenter="activeSearchIndex = idx"
                  @click="selectSearchItem(item); isSearchMobileActive = false"
                >
                  <span class="text-sm font-bold text-slate-800 dark:text-slate-100">{{ item.label }}</span>
                  <span class="text-[11px] text-slate-500 dark:text-slate-400">{{ item.breadcrumbs }}</span>
                </button>
              </div>
              <div v-else class="px-3 py-3 text-xs text-slate-500 dark:text-slate-400">
                Data tidak ditemukan
              </div>
            </div>
          </div>
        </div>

        <template #right-actions>
          <!-- Mobile Search Toggle -->
          <NavBarItemPlain 
            class="md:hidden flex items-center"
            @click="isSearchMobileActive = true"
          >
            <BaseIcon :path="mdiMagnify" size="22" />
          </NavBarItemPlain>

          <!-- Student Notifications (Mobile/Tablet) -->
          <div v-if="authStore.role === 'student'" class="flex items-center md:hidden">
            <NavBarItemPlain 
              @click.prevent="isNotificationsExpanded = !isNotificationsExpanded; markNotificationsAsRead()"
              class="relative"
            >
              <div :class="{ 'animate-pulse drop-shadow-[0_0_10px_rgba(59,130,246,0.6)]': showBadge && notificationsCount > 0 }">
                <BaseIcon :path="mdiBellOutline" size="22" :class="showBadge && notificationsCount > 0 ? 'text-blue-600 dark:text-blue-400' : ''" />
              </div>
              <div v-if="showBadge && notificationsCount > 0" class="absolute top-3 right-3 flex h-3.5 w-3.5 items-center justify-center rounded-full bg-red-500 text-[8px] font-bold text-white shadow-sm ring-1 ring-white dark:ring-slate-950">
                {{ notificationsCount > 9 ? '9+' : notificationsCount }}
              </div>
            </NavBarItemPlain>
          </div>

          <!-- Digital Clock (Mobile) -->
          <div class="md:hidden flex items-center px-2">
            <div class="px-2 py-1 bg-slate-100 dark:bg-slate-800 rounded-lg border border-slate-200 dark:border-slate-700">
              <span class="text-[10px] font-mono font-black text-slate-700 dark:text-slate-200 tracking-wider">
                {{ digitalClock }}
              </span>
            </div>
          </div>
        </template>

        <template #right>
          <div v-if="authStore.role === 'student'" class="relative hidden md:flex items-center h-14 mr-4">
            <NavBarItemPlain 
              @click.prevent="isNotificationsExpanded = !isNotificationsExpanded; markNotificationsAsRead()"
              class="relative"
            >
              <div :class="{ 'animate-pulse drop-shadow-[0_0_10px_rgba(59,130,246,0.6)]': showBadge && notificationsCount > 0 }">
                <BaseIcon :path="mdiBellOutline" size="24" :class="showBadge && notificationsCount > 0 ? 'text-blue-600 dark:text-blue-400' : ''" />
              </div>
              <div v-if="showBadge && notificationsCount > 0" class="absolute top-3 right-3 flex h-4 w-4 items-center justify-center rounded-full bg-red-500 text-[10px] font-bold text-white shadow-sm ring-2 ring-white dark:ring-slate-950">
                {{ notificationsCount > 9 ? '9+' : notificationsCount }}
              </div>
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
                <div v-if="unreadAnnouncements.length">
                  <div class="px-4 py-2 bg-slate-50/50 dark:bg-slate-800/30 text-[9px] font-black uppercase tracking-widest text-slate-400">Pengumuman</div>
                  <div 
                    v-for="a in unreadAnnouncements" 
                    :key="a.id"
                    class="p-4 hover:bg-slate-50 dark:hover:bg-slate-800/50 transition-colors cursor-pointer"
                    @click="openAnnouncementItem(a)"
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
                <div v-else-if="!unreadAnnouncements.length && !exams.length" class="p-8 text-center text-xs text-slate-400 italic">
                  Tidak ada informasi baru.
                </div>
              </div>

              <div class="p-3 border-t border-slate-100 dark:border-slate-800 text-center bg-slate-50 dark:bg-slate-800/50">
                <button 
                  @click="openAnnouncementCenter"
                  class="text-[10px] font-black uppercase tracking-widest text-blue-600 hover:text-blue-700 dark:text-blue-400"
                >
                  Buka Pusat Informasi
                </button>
              </div>
            </div>
          </div>

          <!-- Notifications Dropdown (Mobile Action Bar version) -->
          <div 
            v-if="isNotificationsExpanded && authStore.role === 'student'"
            class="md:hidden fixed top-14 right-4 w-[calc(100vw-32px)] max-w-sm bg-white dark:bg-slate-900 rounded-2xl shadow-2xl border border-slate-100 dark:border-slate-800 z-50 overflow-hidden"
          >
            <div class="p-4 border-b border-slate-100 dark:border-slate-800 flex justify-between items-center bg-slate-50 dark:bg-slate-800/50">
              <h3 class="text-xs font-black uppercase tracking-widest text-slate-500">Notifikasi</h3>
              <button @click="isNotificationsExpanded = false" class="text-slate-400"><BaseIcon :path="mdiCloseCircle" size="20" /></button>
            </div>
            <div class="max-h-[60vh] overflow-y-auto divide-y divide-slate-100 dark:divide-slate-800">
              <!-- Re-use the notification list content here or similar -->
              <div v-if="unreadAnnouncements.length">
                <div v-for="a in unreadAnnouncements" :key="a.id" class="p-4" @click="openAnnouncementItem(a)">
                  <div class="text-[10px] font-bold text-blue-600 mb-1">{{ a.title }}</div>
                  <div class="text-[10px] text-slate-500 line-clamp-1">{{ a.body }}</div>
                </div>
              </div>
              <div v-if="exams.length">
                <div v-for="e in exams" :key="e.id" class="p-4" @click="router.push('/student/ujian'); isNotificationsExpanded = false">
                  <div class="text-[10px] font-bold text-emerald-600 mb-1">{{ e.title }}</div>
                </div>
              </div>
              <div v-if="!unreadAnnouncements.length && !exams.length" class="p-6 text-center text-[10px] text-slate-400 italic">Tidak ada notifikasi baru</div>
            </div>
            <div class="p-3 border-t border-slate-100 dark:border-slate-800 text-center">
              <button @click="openAnnouncementCenter" class="text-[10px] font-black uppercase text-blue-600">Buka Semua</button>
            </div>
          </div>
        </template>
      </NavBar>
      <AsideMenu
        :is-aside-lg-active="isAsideLgActive"
        :is-aside-desktop-hidden="isAsideDesktopHidden"
        :menu="menuAsideMain"
        :menu-bottom="menuAsideBottom"
        @menu-click="menuClick"
        @aside-lg-close-click="isAsideLgActive = false"
      />
      <div
        v-if="authStore.isAuthenticated"
        class="border-b border-slate-200 bg-white/60 px-4 md:px-6 py-2 text-[10px] md:text-xs text-slate-600 shadow-sm dark:border-slate-800 dark:bg-slate-900/60 dark:text-slate-300 flex items-center backdrop-blur-md overflow-x-auto no-scrollbar"
      >
        <div class="flex items-center shrink-0">
          <BaseIcon :path="mdiAccountCircleOutline" size="14" class="mr-1.5 md:mr-2 text-blue-600 dark:text-blue-400" />
          <span class="whitespace-nowrap">
            <span class="md:inline hidden">Role: </span>
            <span class="font-bold text-slate-900 dark:text-white uppercase tracking-tighter">{{ authStore.roleLabel }}</span>
            <span v-if="authStore.user?.username" class="opacity-60 hidden sm:inline"> (@{{ authStore.user.username }})</span>
          </span>
        </div>
        
        <span class="mx-2 md:mx-3 text-slate-300 dark:text-slate-700 shrink-0">|</span>
        
        <button class="font-extrabold text-blue-600 hover:text-blue-700 dark:text-blue-400 hover:underline transition-all uppercase whitespace-nowrap tracking-tighter" @click="router.push(homeRouteForRole(authStore.role))">
          Dashboard
        </button>

        <span class="mx-2 md:mx-3 text-slate-300 dark:text-slate-700 shrink-0">|</span>
        
        <div class="flex items-center shrink-0">
          <BackendHealthBanner />
        </div>

        <div class="ml-auto hidden md:flex items-center shrink-0">
          <span class="mx-3 text-slate-300 dark:text-slate-700">|</span>
          <span class="font-mono font-black tracking-widest text-slate-700 dark:text-slate-100">
            {{ digitalClock }}
          </span>
        </div>
      </div>
      <slot />
      <FooterBar>
        <div />
      </FooterBar>
    </div>
    <BottomNavigation />
  </div>
</template>
