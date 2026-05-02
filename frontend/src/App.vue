<script setup>
import { RouterView } from 'vue-router'
import GlobalToast from '@/components/GlobalToast.vue'
</script>

<template>
  <RouterView v-slot="{ Component, route }">
    <Suspense>
      <transition name="page">
        <component :is="Component" :key="route.fullPath" />
      </transition>
      <template #fallback>
        <div class="min-h-screen flex items-center justify-center bg-slate-50 dark:bg-slate-950">
          <div class="text-sm font-semibold text-slate-500 dark:text-slate-400">Memuat halaman...</div>
        </div>
      </template>
    </Suspense>
  </RouterView>
  <GlobalToast />
</template>
