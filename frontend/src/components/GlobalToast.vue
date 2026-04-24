<script setup>
import { useNotificationStore } from '@/stores/notification.js'

const store = useNotificationStore()

const getIconText = (item) => {
  if (item.color === 'danger') return '!'
  if (item.color === 'success') return '✓'
  if (item.color === 'warning') return '!'
  return 'i'
}

const colorClasses = {
  danger: 'bg-red-500 text-white',
  success: 'bg-emerald-500 text-white',
  warning: 'bg-amber-500 text-white',
  info: 'bg-sky-500 text-white'
}
</script>

<template>
  <div class="fixed bottom-6 right-6 z-[9999] flex flex-col gap-3 w-full max-w-sm">
    <transition-group name="toast">
      <div
        v-for="item in store.items"
        :key="item.id"
        :class="colorClasses[item.color] || colorClasses.info"
        class="flex items-center gap-3 px-4 py-3 rounded-2xl shadow-2xl border border-white/10 backdrop-blur-sm transition-all duration-300 pointer-events-auto"
      >
        <span class="inline-flex h-5 w-5 items-center justify-center rounded-full bg-black/15 text-xs font-black">
          {{ getIconText(item) }}
        </span>
        <div class="flex-1 text-sm font-semibold leading-snug">
          {{ item.message }}
        </div>
        <button
          @click="store.remove(item.id)"
          class="inline-flex h-6 w-6 items-center justify-center rounded-full text-xs font-black hover:bg-black/10 transition-colors"
        >
          ×
        </button>
      </div>
    </transition-group>
  </div>
</template>

<style scoped>
.toast-enter-active,
.toast-leave-active {
  transition: all 0.4s cubic-bezier(0.175, 0.885, 0.32, 1.275);
}
.toast-enter-from {
  opacity: 0;
  transform: translateX(100%) scale(0.9);
}
.toast-leave-to {
  opacity: 0;
  transform: translateY(-20px) scale(0.8);
}
</style>
