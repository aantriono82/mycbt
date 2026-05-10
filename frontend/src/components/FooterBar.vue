<script setup>
import { containerMaxW } from '@/config.js'
import BaseIcon from '@/components/BaseIcon.vue'
import { mdiSend, mdiEmailOutline } from '@mdi/js'

const year = new Date().getFullYear()
const appVersion = __APP_VERSION__
</script>

<template>
  <footer class="px-6 pt-4 pb-8 lg:pb-12" :class="containerMaxW">
    <div class="flex flex-col w-full items-center justify-center gap-2">
      <b class="inline-flex items-center gap-1.5 font-medium text-slate-600 dark:text-slate-400 text-center text-sm">
        <span>&copy;{{ year }}</span>
        <span class="text-slate-300 dark:text-slate-600">|</span>
        <span class="font-semibold text-slate-700 dark:text-slate-200">Atiga CBT</span>
        <span class="text-slate-300 dark:text-slate-600">|</span>
        <span>v{{ appVersion }}</span>
      </b>
      
      <div class="flex flex-wrap justify-center items-center gap-x-4 gap-y-2 text-xs font-bold">
        <span class="text-slate-400 dark:text-slate-500 uppercase tracking-widest text-[10px]">Contact Us:</span>
        <a href="https://t.me/aantriono" target="_blank" class="flex items-center text-blue-600 hover:text-blue-700 dark:text-blue-400 dark:hover:text-blue-300 transition-colors">
          <BaseIcon :path="mdiSend" size="14" class="mr-1 rotate-[-30deg]" />
          Telegram
        </a>
        <a href="mailto:aantriono82@gmail.com" class="flex items-center text-purple-600 hover:text-purple-700 dark:text-purple-400 dark:hover:text-purple-300 transition-colors">
          <BaseIcon :path="mdiEmailOutline" size="14" class="mr-1" />
          Email
        </a>
      </div>
    </div>
  </footer>
</template>
