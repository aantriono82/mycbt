<script setup>
import { mdiChevronUp, mdiChevronDown, mdiWeatherSunny, mdiMoonWaningCrescent } from '@mdi/js'
import { RouterLink } from 'vue-router'
import { computed, ref, onMounted, onBeforeUnmount } from 'vue'
import { useMainStore } from '@/stores/main.js'
import BaseIcon from '@/components/BaseIcon.vue'
import UserAvatarCurrentUser from '@/components/UserAvatarCurrentUser.vue'
import NavBarMenuList from '@/components/NavBarMenuList.vue'
import BaseDivider from '@/components/BaseDivider.vue'
import { useDarkModeStore } from '@/stores/darkMode'

const props = defineProps({
  item: {
    type: Object,
    required: true,
  },
})

const emit = defineEmits(['menu-click'])

const is = computed(() => {
  if (props.item.href) {
    return 'a'
  }

  if (props.item.to) {
    return RouterLink
  }

  return 'div'
})

const componentClass = computed(() => {
  const base = [
    isDropdownActive.value
      ? `navbar-item-label-active dark:text-slate-400`
      : `navbar-item-label dark:text-white dark:hover:text-slate-400`,
    props.item.menu ? 'lg:py-2 lg:px-3' : 'py-2 px-3',
  ]

  if (props.item.isDesktopNoLabel) {
    base.push('lg:w-16', 'lg:justify-center')
  }

  return base
})

const itemLabel = computed(() =>
  props.item.isCurrentUser ? props.item.label || useMainStore().userName : props.item.label,
)

const isDropdownActive = ref(false)

const menuClick = (event) => {
  emit('menu-click', event, props.item)

  if (props.item.menu) {
    isDropdownActive.value = !isDropdownActive.value
  }
}

const menuClickDropdown = (event, item) => {
  emit('menu-click', event, item)
}

const root = ref(null)

const forceClose = (event) => {
  if (root.value && !root.value.contains(event.target)) {
    isDropdownActive.value = false
  }
}

onMounted(() => {
  if (props.item.menu) {
    window.addEventListener('click', forceClose)
  }
})

onBeforeUnmount(() => {
  if (props.item.menu) {
    window.removeEventListener('click', forceClose)
  }
})

const darkModeStore = useDarkModeStore()
</script>

<template>
  <BaseDivider v-if="item.isDivider" nav-bar />
  <component
    :is="is"
    v-else
    ref="root"
    class="relative block cursor-pointer items-center lg:flex"
    :class="componentClass"
    :to="item.to ?? null"
    :href="item.href ?? null"
    :target="item.target ?? null"
    @click="menuClick"
  >
    <div
      class="flex items-center"
      :class="{
        'bg-gray-100 p-3 lg:bg-transparent lg:p-0 dark:bg-slate-800 lg:dark:bg-transparent':
          item.menu,
      }"
    >
      <UserAvatarCurrentUser v-if="item.isCurrentUser" class="mr-3 inline-flex h-6 w-6" />
      <!-- Toggle Switch Premium untuk Dark/Light Mode -->
      <div
        v-if="item.isToggleLightDark"
        class="relative flex h-7 w-16 items-center rounded-full p-1 transition-all duration-500 shadow-inner group"
        :class="darkModeStore.isEnabled 
          ? 'bg-gradient-to-r from-indigo-600 to-blue-700 shadow-indigo-900/20' 
          : 'bg-gradient-to-r from-purple-600 to-fuchsia-500 shadow-purple-900/20'"
      >
        <!-- Circle sliding part -->
        <div
          class="flex h-5 w-5 transform items-center justify-center rounded-full bg-white shadow-md transition-all duration-500 ease-in-out"
          :class="darkModeStore.isEnabled ? 'translate-x-9 rotate-0' : 'translate-x-0 rotate-180'"
        >
          <BaseIcon
            :path="darkModeStore.isEnabled ? mdiWeatherSunny : mdiMoonWaningCrescent"
            size="14"
            :class="darkModeStore.isEnabled ? 'text-purple-500' : 'text-indigo-600'"
          />
        </div>
        
        <!-- Text indicator (hidden/visible based on state) -->
        <span 
          class="absolute left-0 right-0 top-0 bottom-0 flex items-center justify-center pointer-events-none"
        >
          <span 
            class="text-[8px] font-bold uppercase tracking-tighter transition-opacity duration-300"
            :class="darkModeStore.isEnabled 
              ? 'opacity-100 -translate-x-2 text-white' 
              : 'opacity-0 translate-x-2'"
          >DARK</span>
          <span 
            class="text-[8px] font-bold uppercase tracking-tighter transition-opacity duration-300"
            :class="!darkModeStore.isEnabled 
              ? 'opacity-100 translate-x-2 text-white' 
              : 'opacity-0 -translate-x-2'"
          >LIGHT</span>
        </span>
      </div>

      <!-- Ikon biasa untuk item lainnya -->
      <BaseIcon
        v-else-if="item.icon"
        :path="item.icon"
        :class="{ 'transition-colors': !darkModeStore.isInProgress }"
      />
      <span
        v-if="!item.isToggleLightDark"
        class="px-2"
        :class="{
          'lg:hidden': item.isDesktopNoLabel && item.icon,
          'transition-colors': !darkModeStore.isInProgress,
        }"
        >{{ itemLabel }}</span
      >
      <BaseIcon
        v-if="item.menu"
        :path="isDropdownActive ? mdiChevronUp : mdiChevronDown"
        :class="{ 'transition-colors': !darkModeStore.isInProgress }"
        class="hidden lg:inline-flex"
      />
    </div>
    <div
      v-if="item.menu"
      class="border-b border-gray-100 text-sm lg:absolute lg:top-full lg:left-0 lg:z-20 lg:min-w-full lg:rounded-lg lg:border lg:bg-white lg:shadow-lg dark:border-slate-700 lg:dark:bg-slate-800"
      :class="{ 'lg:hidden': !isDropdownActive }"
    >
      <NavBarMenuList :menu="item.menu" @menu-click="menuClickDropdown" />
    </div>
  </component>
</template>
