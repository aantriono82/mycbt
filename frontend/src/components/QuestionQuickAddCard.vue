<script setup>
import { mdiPlus } from '@mdi/js'
import BaseIcon from '@/components/BaseIcon.vue'

defineProps({
  title: {
    type: String,
    default: 'Tambah Soal',
  },
  subtitle: {
    type: String,
    default: 'Buat banyak soal sekaligus (template)',
  },
})

const emit = defineEmits(['click'])
</script>

<template>
  <button
    type="button"
    class="group relative flex min-h-[140px] w-full flex-col justify-between overflow-hidden rounded-2xl border-2 border-dashed border-slate-200 bg-linear-to-br from-white to-slate-50 p-5 text-left transition-all hover:-translate-y-0.5 hover:border-blue-300 hover:shadow-lg hover:shadow-blue-100 dark:border-slate-800 dark:from-slate-950 dark:to-slate-900/30 dark:hover:border-blue-700/60 dark:hover:shadow-blue-900/20"
    @click="emit('click')"
  >
    <div class="pointer-events-none absolute -right-10 -top-10 h-28 w-28 rounded-full bg-blue-600/8 blur-2xl transition-opacity group-hover:opacity-100 dark:bg-blue-400/10" />

    <div class="flex items-start justify-between gap-4">
      <div>
        <div class="text-sm font-black tracking-tight text-slate-800 dark:text-slate-100">
          {{ title }}
        </div>
        <div class="mt-1 text-xs text-slate-500 dark:text-slate-400">
          {{ subtitle }}
        </div>
      </div>

      <span
        class="flex h-12 w-12 items-center justify-center rounded-2xl bg-blue-600 text-white shadow-lg shadow-blue-200 transition-transform group-hover:scale-105 dark:shadow-blue-900/30"
      >
        <BaseIcon :path="mdiPlus" size="22" />
      </span>
    </div>

    <div class="mt-4 flex items-center gap-2 text-[10px] font-black uppercase tracking-widest text-slate-400">
      Klik untuk quick-add
    </div>
  </button>
</template>

