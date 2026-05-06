<script setup>
import { computed } from 'vue'
import BaseIcon from '@/components/BaseIcon.vue'
import BaseSkeleton from '@/components/BaseSkeleton.vue'

const props = defineProps({
  label: String,
  number: [Number, String],
  icon: String,
  color: {
    type: String,
    default: 'blue'
  },
  small: {
    type: Boolean,
    default: false
  },
  loading: {
    type: Boolean,
    default: false
  }
})

const colorClasses = computed(() => {
  const map = {
    blue: {
      text: 'text-slate-900 dark:text-white',
      label: 'text-blue-700 dark:text-blue-400',
      bg: 'bg-blue-100 dark:bg-blue-900/40',
      cardBg: 'bg-blue-50 dark:bg-blue-950/40',
      cardBorder: 'border-blue-400/60 dark:border-blue-800/80',
      icon: 'text-blue-600 dark:text-blue-400'
    },
    emerald: {
      text: 'text-slate-900 dark:text-white',
      label: 'text-emerald-700 dark:text-emerald-400',
      bg: 'bg-emerald-100 dark:bg-emerald-900/40',
      cardBg: 'bg-emerald-50 dark:bg-emerald-950/40',
      cardBorder: 'border-emerald-400/60 dark:border-emerald-800/80',
      icon: 'text-emerald-600 dark:text-emerald-400'
    },
    cyan: {
      text: 'text-slate-900 dark:text-white',
      label: 'text-cyan-700 dark:text-cyan-400',
      bg: 'bg-cyan-100 dark:bg-cyan-900/40',
      cardBg: 'bg-cyan-50 dark:bg-cyan-950/40',
      cardBorder: 'border-cyan-400/60 dark:border-cyan-800/80',
      icon: 'text-cyan-600 dark:text-cyan-400'
    },
    orange: {
      text: 'text-slate-900 dark:text-white',
      label: 'text-orange-700 dark:text-orange-400',
      bg: 'bg-orange-100 dark:bg-orange-900/40',
      cardBg: 'bg-orange-50 dark:bg-orange-950/40',
      cardBorder: 'border-orange-400/60 dark:border-orange-800/80',
      icon: 'text-orange-600 dark:text-orange-400'
    },
    indigo: {
      text: 'text-slate-900 dark:text-white',
      label: 'text-indigo-700 dark:text-indigo-400',
      bg: 'bg-indigo-100 dark:bg-indigo-900/40',
      cardBg: 'bg-indigo-50 dark:bg-indigo-950/40',
      cardBorder: 'border-indigo-400/60 dark:border-indigo-800/80',
      icon: 'text-indigo-600 dark:text-indigo-400'
    },
    red: {
      text: 'text-slate-900 dark:text-white',
      label: 'text-red-700 dark:text-red-400',
      bg: 'bg-red-100 dark:bg-red-900/40',
      cardBg: 'bg-red-50 dark:bg-red-950/40',
      cardBorder: 'border-red-400/60 dark:border-red-800/80',
      icon: 'text-red-600 dark:text-red-400'
    },
    amber: {
      text: 'text-slate-900 dark:text-white',
      label: 'text-amber-700 dark:text-amber-400',
      bg: 'bg-amber-100 dark:bg-amber-900/40',
      cardBg: 'bg-amber-50 dark:bg-amber-950/40',
      cardBorder: 'border-amber-400/60 dark:border-amber-800/80',
      icon: 'text-amber-600 dark:text-amber-400'
    }
  }
  return map[props.color] || map.blue
})
</script>

<template>
  <div :class="[
    'rounded-[1.75rem] shadow-sm border flex flex-col justify-between transition-all hover:shadow-xl hover:-translate-y-1 active:scale-95 relative overflow-hidden group cursor-pointer',
    small ? 'min-h-[7rem] p-4 sm:min-h-32 sm:p-5' : 'min-h-[8.5rem] p-4 sm:min-h-[10rem] sm:p-6',
    colorClasses.cardBg,
    colorClasses.cardBorder
  ]">
    <!-- Decorative Circle -->
    <div :class="['absolute -top-5 -right-5 h-20 w-20 rounded-full opacity-20 sm:-top-6 sm:-right-6 sm:h-24 sm:w-24', colorClasses.bg]"></div>
    
    <div :class="[small ? 'text-[9px] sm:text-[10px]' : 'text-[10px] sm:text-xs', 'font-black uppercase tracking-[0.18em] sm:tracking-[0.2em]', colorClasses.label]">
      {{ label }}
    </div>
    <div class="flex items-end justify-between relative z-10">
      <div v-if="loading">
        <BaseSkeleton :width="small ? 'w-16' : 'w-24'" :height="small ? 'h-8' : 'h-12'" />
      </div>
      <div v-else :class="[small ? 'text-2xl sm:text-3xl' : 'text-2xl sm:text-5xl', 'font-black leading-none tracking-tight', colorClasses.text]">
        {{ number }}
      </div>
      <div :class="['rounded-xl sm:rounded-2xl shadow-sm transition-transform group-hover:scale-110 shrink-0', small ? 'p-2' : 'p-2.5 sm:p-3', colorClasses.bg]">
        <BaseSkeleton v-if="loading" :width="small ? 'w-6' : 'w-8'" :height="small ? 'w-6' : 'w-8'" rounded="rounded-lg" />
        <BaseIcon v-else :path="icon" :size="small ? 22 : 26" :class="colorClasses.icon" />
      </div>
    </div>
  </div>
</template>
