<script setup>
import { ref } from 'vue'
import { mdiClose, mdiDotsVertical } from '@mdi/js'
import { containerMaxW } from '@/config.js'
import BaseIcon from '@/components/BaseIcon.vue'
import NavBarMenuList from '@/components/NavBarMenuList.vue'
import NavBarItemPlain from '@/components/NavBarItemPlain.vue'

defineProps({
  menu: {
    type: Array,
    required: true,
  },
  showMobileMenuToggle: {
    type: Boolean,
    default: true,
  },
})

const emit = defineEmits(['menu-click'])

const menuClick = (event, item) => {
  emit('menu-click', event, item)
}

const isMenuNavBarActive = ref(false)
</script>

<template>
  <nav
    class="fixed inset-x-0 top-0 z-30 h-14 w-screen bg-white/60 dark:bg-slate-950/60 backdrop-blur-xl transition-(--transition-position) lg:w-auto border-b border-slate-100 dark:border-slate-800"
  >
    <div class="flex items-stretch px-0 md:px-0" :class="containerMaxW">
      <div class="flex h-14 flex-1 items-stretch overflow-hidden">
        <slot />
      </div>
      <div class="flex h-14 flex-none items-stretch">
        <slot name="right-actions" />
        <div v-if="showMobileMenuToggle" class="lg:hidden flex items-stretch">
          <NavBarItemPlain @click.prevent="isMenuNavBarActive = !isMenuNavBarActive">
            <BaseIcon :path="isMenuNavBarActive ? mdiClose : mdiDotsVertical" size="24" />
          </NavBarItemPlain>
        </div>
      </div>
      <div
        class="absolute top-14 left-0 max-h-[calc(100dvh-(--spacing(14)))] w-screen overflow-y-auto bg-gray-50 shadow-lg lg:static lg:flex lg:w-auto lg:overflow-visible lg:shadow-none dark:bg-slate-800"
        :class="[
          showMobileMenuToggle ? (isMenuNavBarActive ? 'block' : 'hidden') : 'hidden',
          'lg:flex',
        ]"
      >
        <slot name="right" />
        <div class="lg:mr-12 lg:flex lg:items-stretch">
          <NavBarMenuList :menu="menu" @menu-click="menuClick" />
        </div>
      </div>
    </div>
  </nav>
</template>
