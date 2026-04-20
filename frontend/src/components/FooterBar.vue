<script setup>
import { containerMaxW } from '@/config.js'
import BaseLevel from '@/components/BaseLevel.vue'

const year = new Date().getFullYear()
</script>

<template>
  <footer class="px-6 py-2" :class="containerMaxW">
    <div class="flex w-full min-h-[48px] items-center justify-center">
      <b class="inline-flex items-center gap-1.5 font-medium text-slate-600 dark:text-slate-400 text-center">
        <span>&copy;{{ year }}</span>
        <span class="text-slate-300 dark:text-slate-600">|</span>
        <span>Dibuat dengan</span>
        <svg class="h-4 w-4 shrink-0 text-red-500 fill-current" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg" aria-label="cinta">
          <path d="M12 21.593c-5.63-5.539-11-10.297-11-14.402 0-3.791 3.068-5.191 5.281-5.191 1.312 0 4.151.501 5.719 4.457 1.59-3.968 4.464-4.447 5.726-4.447 2.54 0 5.274 1.621 5.274 5.181 0 4.069-5.136 8.625-11 14.402z"/>
        </svg>
        <span>oleh <span class="font-semibold text-slate-700 dark:text-slate-200">Aan Triono</span></span>
      </b>
    </div>
  </footer>
</template>
