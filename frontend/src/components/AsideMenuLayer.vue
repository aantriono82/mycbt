<script setup>
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { mdiClose } from '@mdi/js'
import { useRouter } from 'vue-router'
import AsideMenuList from '@/components/AsideMenuList.vue'
import BaseIcon from '@/components/BaseIcon.vue'
import { useAuthStore } from '@/stores/auth.js'
import { useSchoolIdentityStore } from '@/stores/schoolIdentity.js'

const authStore = useAuthStore()
const schoolStore = useSchoolIdentityStore()
const router = useRouter()

defineProps({
  menu: {
    type: Array,
    required: true,
  },
  menuBottom: Array,
})

const emit = defineEmits(['menu-click', 'aside-lg-close-click'])

const menuClick = (event, item) => {
  emit('menu-click', event, item)
}

const asideLgCloseClick = (event) => {
  emit('aside-lg-close-click', event)
}

const handleSchoolIdentityUpdated = (event) => {
  const detail = event?.detail
  if (detail && typeof detail === 'object') {
    schoolStore.updateIdentity(detail)
    return
  }
  schoolStore.loadSchoolIdentity(authStore.role === 'admin')
}

onMounted(() => {
  schoolStore.loadSchoolIdentity(authStore.role === 'admin')
  if (typeof window !== 'undefined') {
    window.addEventListener('school-identity-updated', handleSchoolIdentityUpdated)
  }
})

onBeforeUnmount(() => {
  if (typeof window !== 'undefined') {
    window.removeEventListener('school-identity-updated', handleSchoolIdentityUpdated)
  }
})

watch(
  () => authStore.isAuthenticated,
  (isAuthenticated) => {
    if (isAuthenticated) schoolStore.loadSchoolIdentity(authStore.role === 'admin')
  },
)
</script>

<template>
  <aside
    id="aside"
    class="fixed top-0 z-40 flex h-screen w-60 overflow-hidden transition-(--transition-position) lg:pl-2"
  >
    <div class="aside flex flex-1 flex-col overflow-hidden lg:rounded-[2rem] bg-white/70 dark:bg-[#13082a]/70 backdrop-blur-xl border-r lg:border border-slate-200 dark:border-blue-900/30 shadow-2xl shadow-blue-300/40 dark:shadow-blue-900/30">
      <div class="aside-brand flex h-[102.5px] flex-row items-center justify-between px-6 pt-6 border-b border-slate-200/80 dark:border-slate-800/80">
        <div class="flex items-center cursor-pointer" @click="router.push('/dashboard')">
          <img :src="schoolStore.schoolBrandLogo" class="mr-3 h-10 w-10 rounded-lg shadow-lg hover:scale-110 transition-transform duration-300 brightness-110 object-contain bg-white p-0.5" alt="School Logo" />
          <div class="flex flex-col leading-tight">
            <b class="font-black uppercase tracking-tighter text-[10px] bg-clip-text text-transparent bg-gradient-to-br from-blue-600 via-indigo-500 to-violet-600 dark:from-blue-400 dark:via-indigo-400 dark:to-violet-400 truncate max-w-[145px]">{{ schoolStore.schoolBrandName }}</b>
            <span class="text-[8px] font-bold text-slate-400 dark:text-slate-500 tracking-wider uppercase truncate max-w-[145px]">{{ schoolStore.schoolBrandTagline }}</span>
          </div>
        </div>
        <button class="hidden p-3 lg:inline-block xl:hidden text-slate-400 hover:text-slate-700 dark:text-white/70 dark:hover:text-white" @click.prevent="asideLgCloseClick">
          <BaseIcon :path="mdiClose" />
        </button>
      </div>
      <div
        class="aside-scrollbar flex-1 overflow-x-hidden overflow-y-auto scrollbar-styled-light dark:scrollbar-styled-dark px-3 mt-2"
      >
        <AsideMenuList :menu="menu" @menu-click="menuClick" />
      </div>

      <div class="px-3 pb-6">
        <AsideMenuList v-if="menuBottom" :menu="menuBottom" @menu-click="menuClick" />
      </div>
    </div>
  </aside>
</template>
