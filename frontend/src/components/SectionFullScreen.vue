<script setup>
import { computed } from 'vue'
import { useDarkModeStore } from '@/stores/darkMode.js'
import { gradientBgPurplePink, gradientBgDark, gradientBgPinkRed } from '@/colors.js'

const props = defineProps({
  bg: {
    type: String,
    default: 'white',
    validator: (value) => ['purplePink', 'pinkRed', 'white', 'cream'].includes(value),
  },
})

const colorClass = computed(() => {
  if (useDarkModeStore().isEnabled) {
    return gradientBgDark
  }

  switch (props.bg) {
    case 'purplePink':
      return gradientBgPurplePink
    case 'pinkRed':
      return gradientBgPinkRed
    case 'white':
      return 'bg-white dark:bg-slate-900'
    case 'cream':
      return 'bg-amber-50 dark:bg-slate-900'
  }

  return ''
})
</script>

<template>
  <div class="flex min-h-screen items-center justify-center relative overflow-hidden bg-slate-50 dark:bg-[#0b1120]">
    <!-- Mesh Gradient elements -->
    <div class="absolute top-[-10%] left-[-10%] w-[40%] h-[40%] rounded-full bg-blue-400/20 dark:bg-blue-600/10 blur-[120px] animate-pulse"></div>
    <div class="absolute bottom-[-10%] right-[-10%] w-[40%] h-[40%] rounded-full bg-indigo-400/20 dark:bg-indigo-600/10 blur-[120px] animate-pulse" style="animation-delay: 2s;"></div>
    <div class="absolute top-[20%] right-[10%] w-[30%] h-[30%] rounded-full bg-purple-400/10 dark:bg-purple-600/5 blur-[100px] animate-pulse" style="animation-delay: 4s;"></div>
    <div class="absolute bottom-[20%] left-[10%] w-[30%] h-[30%] rounded-full bg-teal-400/10 dark:bg-emerald-600/5 blur-[100px] animate-pulse" style="animation-delay: 6s;"></div>

    <div class="relative z-10 w-full flex items-center justify-center">
      <slot card-class="w-11/12 md:w-7/12 lg:w-6/12 xl:w-4/12 shadow-2xl" />
    </div>
  </div>
</template>
