<script setup>
import { computed } from 'vue'

const props = defineProps({
  value: {
    type: Number,
    default: 0
  },
  max: {
    type: Number,
    default: 100
  },
  size: {
    type: Number,
    default: 100
  },
  strokeWidth: {
    type: Number,
    default: 8
  },
  color: {
    type: String,
    default: 'blue'
  },
  label: String
})

const radius = computed(() => (props.size - props.strokeWidth) / 2)
const circumference = computed(() => 2 * Math.PI * radius.value)
const offset = computed(() => {
  const percentage = Math.min(Math.max(props.value / props.max, 0), 1)
  return circumference.value * (1 - percentage)
})

const colorClasses = computed(() => {
  const map = {
    blue: 'text-blue-500 dark:text-blue-400',
    emerald: 'text-emerald-500 dark:text-emerald-400',
    amber: 'text-amber-500 dark:text-amber-400',
    rose: 'text-rose-500 dark:text-rose-400',
    indigo: 'text-indigo-500 dark:text-indigo-400'
  }
  return map[props.color] || map.blue
})
</script>

<template>
  <div class="relative inline-flex items-center justify-center" :style="{ width: size + 'px', height: size + 'px' }">
    <svg :width="size" :height="size" class="transform -rotate-90">
      <!-- Background Circle -->
      <circle
        class="text-slate-100 dark:text-slate-800"
        stroke="currentColor"
        :stroke-width="strokeWidth"
        fill="transparent"
        :r="radius"
        :cx="size / 2"
        :cy="size / 2"
      />
      <!-- Progress Circle -->
      <circle
        :class="['transition-all duration-1000 ease-out', colorClasses]"
        stroke="currentColor"
        :stroke-width="strokeWidth"
        stroke-linecap="round"
        fill="transparent"
        :r="radius"
        :cx="size / 2"
        :cy="size / 2"
        :style="{
          strokeDasharray: circumference,
          strokeDashoffset: offset
        }"
      />
    </svg>
    <div class="absolute inset-0 flex flex-col items-center justify-center text-center">
      <span class="text-xl font-black text-slate-800 dark:text-white leading-none">
        {{ value }}
      </span>
      <span v-if="label" class="text-[9px] font-bold uppercase tracking-tighter text-slate-400 dark:text-slate-500 mt-0.5">
        {{ label }}
      </span>
    </div>
  </div>
</template>
