<script setup>
import { computed } from 'vue'
import { mdiCog } from '@mdi/js'
import CardBox from '@/components/CardBox.vue'
import NumberDynamic from '@/components/NumberDynamic.vue'
import BaseIcon from '@/components/BaseIcon.vue'
import BaseLevel from '@/components/BaseLevel.vue'
import PillTagTrend from '@/components/PillTagTrend.vue'
import BaseButton from '@/components/BaseButton.vue'

const props = defineProps({
  number: {
    type: [Number, String],
    default: 0,
  },
  icon: String,
  prefix: String,
  suffix: String,
  label: String,
  color: {
    type: String,
    default: 'blue'
  },
  trend: String,
  trendType: String,
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
    purple: {
      text: 'text-slate-900 dark:text-white',
      label: 'text-purple-600 dark:text-purple-400',
      bg: 'bg-purple-100/50 dark:bg-purple-900/30',
      cardBg: 'bg-purple-400/10 dark:bg-purple-500/10 backdrop-blur-md',
      cardBorder: 'border-purple-200/50 dark:border-purple-700/50',
      icon: 'text-purple-600 dark:text-purple-400'
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
    },
    sky: {
      text: 'text-slate-900 dark:text-white',
      label: 'text-sky-600 dark:text-sky-400',
      bg: 'bg-sky-100/50 dark:bg-sky-900/30',
      cardBg: 'bg-sky-400/10 dark:bg-sky-500/10 backdrop-blur-md',
      cardBorder: 'border-sky-200/50 dark:border-sky-700/50',
      icon: 'text-sky-600 dark:text-sky-400'
    },
    rose: {
      text: 'text-slate-900 dark:text-white',
      label: 'text-rose-600 dark:text-rose-400',
      bg: 'bg-rose-100/50 dark:bg-rose-900/30',
      cardBg: 'bg-rose-400/10 dark:bg-rose-500/10 backdrop-blur-md',
      cardBorder: 'border-rose-200/50 dark:border-rose-700/50',
      icon: 'text-rose-600 dark:text-rose-400'
    }
  }
  return map[props.color] || map.blue
})
</script>

<template>
  <CardBox :class="['relative overflow-hidden group border transition-all hover:shadow-md', colorClasses.cardBg, colorClasses.cardBorder]">
    <!-- Decorative Circle -->
    <div :class="['absolute -bottom-6 -right-6 h-20 w-20 rounded-full opacity-20', colorClasses.bg]"></div>
    
    <BaseLevel v-if="trend" class="mb-3" mobile>
      <PillTagTrend :trend="trend" :trend-type="trendType" small />
      <BaseButton :icon="mdiCog" icon-w="w-4" icon-h="h-4" color="lightDark" small />
    </BaseLevel>
    <BaseLevel mobile class="relative z-10">
      <div>
        <h3 :class="['text-xs font-black uppercase tracking-[0.2em] mb-1.5', colorClasses.label]">
          {{ label }}
        </h3>
        <h1 :class="['text-3xl leading-tight font-black tracking-tighter', colorClasses.text]">
          <NumberDynamic :value="number" :prefix="prefix" :suffix="suffix" />
        </h1>
      </div>
      <div v-if="icon" :class="['p-3 rounded-2xl shadow-sm transition-transform group-hover:scale-110', colorClasses.bg]">
        <BaseIcon :path="icon" size="32" w="w-10" h="h-10" :class="colorClasses.icon" />
      </div>
    </BaseLevel>
  </CardBox>
</template>
