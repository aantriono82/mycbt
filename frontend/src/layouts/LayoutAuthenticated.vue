<script setup>
import { mdiForwardburger, mdiBackburger, mdiMenu, mdiAccountCircleOutline, mdiBellOutline } from '@mdi/js'
import { computed, onMounted, ref } from 'vue'
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

const announcementCount = ref(0)
const loadAnnouncementCount = async () => {
  if (authStore.role !== 'student') return
  try {
    const { data } = await api.get('/api/v1/student/announcements', { params: { limit: 1, offset: 0 } })
    announcementCount.value = data?.meta?.total || 0
  } catch {
    announcementCount.value = 0
  }
}

onMounted(loadAnnouncementCount)

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

        <NavBarItemPlain 
          v-if="authStore.role === 'student'"
          @click.prevent="router.push('/student/announcements')"
          class="relative"
        >
          <BaseIcon :path="mdiBellOutline" size="24" />
          <div v-if="announcementCount > 0" class="absolute top-3 right-3 flex h-4 w-4 items-center justify-center rounded-full bg-red-500 text-[10px] font-bold text-white shadow-sm ring-2 ring-white dark:ring-slate-950">
            {{ announcementCount > 9 ? '9+' : announcementCount }}
          </div>
        </NavBarItemPlain>
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
