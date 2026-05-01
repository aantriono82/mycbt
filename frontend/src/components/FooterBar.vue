<script setup>
import { containerMaxW } from '@/config.js'

const year = new Date().getFullYear()
const appVersion = __APP_VERSION__
</script>

<template>
  <footer class="px-6 py-2" :class="containerMaxW">
    <div class="flex w-full min-h-[48px] items-center justify-center">
      <b class="inline-flex items-center gap-1.5 font-medium text-slate-600 dark:text-slate-400 text-center">
        <span>&copy;{{ year }}</span>
        <span class="text-slate-300 dark:text-slate-600">|</span>
        <span class="font-semibold text-slate-700 dark:text-slate-200">Atiga CBT</span>
        <span class="text-slate-300 dark:text-slate-600">|</span>
        <span>v{{ appVersion }}</span>
      </b>
    </div>
  </footer>
</template>
