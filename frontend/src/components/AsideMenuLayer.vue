<script setup>
import { computed } from 'vue'
import { mdiClose } from '@mdi/js'
import { useRouter } from 'vue-router'
import AsideMenuList from '@/components/AsideMenuList.vue'
import BaseIcon from '@/components/BaseIcon.vue'
import { useAuthStore } from '@/stores/auth.js'

const authStore = useAuthStore()
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
</script>

<template>
  <aside
    id="aside"
    class="fixed top-0 z-40 flex h-screen w-60 overflow-hidden transition-(--transition-position) lg:py-2 lg:pl-2"
  >
    <div class="aside flex flex-1 flex-col overflow-hidden lg:rounded-2xl dark:bg-slate-900 border-r lg:border-none border-slate-100 dark:border-slate-800 shadow-sm">
      <div class="aside-brand flex h-16 flex-row items-center justify-between dark:bg-slate-900 px-4">
        <div class="flex items-center cursor-pointer" @click="router.push('/dashboard')">
          <img src="/logo_a3_blue.png" class="mr-3 h-10 w-10 rounded-xl shadow-lg hover:scale-110 transition-transform duration-300" alt="Logo" />
          <div class="flex flex-col leading-tight">
            <b class="font-black uppercase tracking-tighter text-xl bg-clip-text text-transparent bg-gradient-to-r from-blue-600 to-indigo-600 dark:from-blue-400 dark:to-indigo-400">AtigaCBT</b>
            <span class="text-[9px] font-bold text-slate-400 dark:text-slate-500 tracking-widest uppercase">Professional LMS</span>
          </div>
        </div>
        <button class="hidden p-3 lg:inline-block xl:hidden" @click.prevent="asideLgCloseClick">
          <BaseIcon :path="mdiClose" />
        </button>
      </div>
      <div
        class="aside-scrollbar flex-1 overflow-x-hidden overflow-y-auto dark:scrollbar-styled-dark"
      >
        <AsideMenuList :menu="menu" @menu-click="menuClick" />
      </div>

      <AsideMenuList v-if="menuBottom" :menu="menuBottom" @menu-click="menuClick" />
    </div>
  </aside>
</template>
