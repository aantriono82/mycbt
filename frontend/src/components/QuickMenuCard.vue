<script setup>
import BaseIcon from '@/components/BaseIcon.vue'

defineProps({
  label: String,
  description: String,
  icon: String,
  to: String,
  color: {
    type: String,
    default: 'blue'
  }
})

const colorMap = {
  blue: 'text-blue-500 dark:text-blue-400 bg-blue-50 dark:bg-blue-900/30',
  indigo: 'text-indigo-500 dark:text-indigo-400 bg-indigo-50 dark:bg-indigo-900/30',
  emerald: 'text-emerald-500 dark:text-emerald-400 bg-emerald-50 dark:bg-emerald-900/30',
  amber: 'text-amber-500 dark:text-amber-400 bg-amber-50 dark:bg-amber-900/30',
  purple: 'text-purple-500 dark:text-purple-400 bg-purple-50 dark:bg-purple-900/30',
  rose: 'text-rose-500 dark:text-rose-400 bg-rose-50 dark:bg-rose-900/30'
}
</script>

<template>
  <router-link
    :to="to"
    class="flex items-center p-5 bg-white dark:bg-slate-900 border border-slate-100 dark:border-slate-800 rounded-2xl shadow-sm hover:shadow-md hover:border-blue-200 dark:hover:border-blue-800 transition-all group overflow-hidden relative"
  >
    <!-- Background highlight on hover -->
    <div class="absolute -right-4 -top-4 w-12 h-12 rounded-full bg-blue-500/0 group-hover:bg-blue-500/5 transition-all duration-500 scale-0 group-hover:scale-[4]"></div>

    <div :class="['p-3 rounded-2xl mr-4 transition-all duration-500 group-hover:scale-110 shadow-sm', colorMap[color] || colorMap.blue]">
      <BaseIcon :path="icon" size="24" />
    </div>
    <div class="flex-1 relative z-10">
      <h4 class="font-bold text-slate-800 dark:text-slate-100 group-hover:text-blue-600 dark:group-hover:text-blue-400 transition-colors leading-tight mb-0.5">{{ label }}</h4>
      <p v-if="description" class="text-xs text-slate-500 dark:text-slate-400">{{ description }}</p>
    </div>
  </router-link>
</template>
