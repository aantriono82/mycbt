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
      label: 'text-blue-600 dark:text-blue-400',
      bg: 'bg-blue-100/50 dark:bg-blue-900/30',
      cardBg: 'bg-blue-400/10 dark:bg-blue-500/10 backdrop-blur-md',
      cardBorder: 'border-blue-200/50 dark:border-blue-700/50',
      icon: 'text-blue-600 dark:text-blue-400'
    },
    emerald: {
      text: 'text-slate-900 dark:text-white',
      label: 'text-emerald-600 dark:text-emerald-400',
      bg: 'bg-emerald-100/50 dark:bg-emerald-900/30',
      cardBg: 'bg-emerald-400/10 dark:bg-emerald-500/10 backdrop-blur-md',
      cardBorder: 'border-emerald-200/50 dark:border-emerald-700/50',
      icon: 'text-emerald-600 dark:text-emerald-400'
    },
    cyan: {
      text: 'text-slate-900 dark:text-white',
      label: 'text-cyan-600 dark:text-cyan-400',
      bg: 'bg-cyan-100/50 dark:bg-cyan-900/30',
      cardBg: 'bg-cyan-400/10 dark:bg-cyan-500/10 backdrop-blur-md',
      cardBorder: 'border-cyan-200/50 dark:border-cyan-700/50',
      icon: 'text-cyan-600 dark:text-cyan-400'
    },
    orange: {
      text: 'text-slate-900 dark:text-white',
      label: 'text-orange-600 dark:text-orange-400',
      bg: 'bg-orange-100/50 dark:bg-orange-900/30',
      cardBg: 'bg-orange-400/10 dark:bg-orange-500/10 backdrop-blur-md',
      cardBorder: 'border-orange-200/50 dark:border-orange-700/50',
      icon: 'text-orange-600 dark:text-orange-400'
    },
    indigo: {
      text: 'text-slate-900 dark:text-white',
      label: 'text-indigo-600 dark:text-indigo-400',
      bg: 'bg-indigo-100/50 dark:bg-indigo-900/30',
      cardBg: 'bg-indigo-400/10 dark:bg-indigo-500/10 backdrop-blur-md',
      cardBorder: 'border-indigo-200/50 dark:border-indigo-700/50',
      icon: 'text-indigo-600 dark:text-indigo-400'
    },
    red: {
      text: 'text-slate-900 dark:text-white',
      label: 'text-red-600 dark:text-red-400',
      bg: 'bg-red-100/50 dark:bg-red-900/30',
      cardBg: 'bg-red-400/10 dark:bg-red-500/10 backdrop-blur-md',
      cardBorder: 'border-red-200/50 dark:border-red-700/50',
      icon: 'text-red-600 dark:text-red-400'
    },
    amber: {
      text: 'text-slate-900 dark:text-white',
      label: 'text-amber-600 dark:text-amber-400',
      bg: 'bg-amber-100/50 dark:bg-amber-900/30',
      cardBg: 'bg-amber-400/10 dark:bg-amber-500/10 backdrop-blur-md',
      cardBorder: 'border-amber-200/50 dark:border-amber-700/50',
      icon: 'text-amber-600 dark:text-amber-400'
    }
  }
  return map[props.color] || map.blue
})
</script>

<template>
  <div :class="[
    'rounded-[2rem] shadow-sm border flex flex-col justify-between transition-all hover:shadow-xl hover:-translate-y-1 active:scale-95 relative overflow-hidden group cursor-pointer',
    small ? 'h-32 p-5' : 'h-40 p-6',
    colorClasses.cardBg,
    colorClasses.cardBorder
  ]">
    <!-- Decorative Circle -->
    <div :class="['absolute -top-6 -right-6 h-24 w-24 rounded-full opacity-20', colorClasses.bg]"></div>
    
    <div :class="[small ? 'text-[10px]' : 'text-xs', 'font-black uppercase tracking-[0.2em]', colorClasses.label]">
      {{ label }}
    </div>
    <div class="flex items-end justify-between relative z-10">
      <div v-if="loading">
        <BaseSkeleton :width="small ? 'w-16' : 'w-24'" :height="small ? 'h-8' : 'h-12'" />
      </div>
      <div v-else :class="[small ? 'text-3xl' : 'text-5xl', 'font-black tracking-tighter', colorClasses.text]">
        {{ number }}
      </div>
      <div :class="['rounded-2xl shadow-sm transition-transform group-hover:scale-110', small ? 'p-2' : 'p-3', colorClasses.bg]">
        <BaseSkeleton v-if="loading" :width="small ? 'w-6' : 'w-8'" :height="small ? 'w-6' : 'w-8'" rounded="rounded-lg" />
        <BaseIcon v-else :path="icon" :size="small ? 24 : 32" :class="colorClasses.icon" />
      </div>
    </div>
  </div>
</template>
