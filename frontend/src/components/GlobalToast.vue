<script setup>
import { 
  mdiCheckCircle, 
  mdiAlertCircle, 
  mdiAlert, 
  mdiInformation,
  mdiClose
} from '@mdi/js'
import BaseIcon from '@/components/BaseIcon.vue'
import { useNotificationStore } from '@/stores/notification.js'

const store = useNotificationStore()

const getIcon = (color) => {
  if (color === 'danger') return mdiAlertCircle
  if (color === 'success') return mdiCheckCircle
  if (color === 'warning') return mdiAlert
  return mdiInformation
}

const colorClasses = {
  danger: 'bg-red-600 text-white shadow-red-500/20',
  success: 'bg-emerald-600 text-white shadow-emerald-500/20',
  warning: 'bg-amber-500 text-white shadow-amber-500/20',
  info: 'bg-sky-600 text-white shadow-sky-500/20'
}
</script>

<template>
  <div class="fixed top-6 right-6 z-[9999] flex flex-col gap-3 w-full max-w-sm pointer-events-none">
    <transition-group name="toast">
      <div
        v-for="item in store.items"
        :key="item.id"
        :class="colorClasses[item.color] || colorClasses.info"
        class="flex items-start gap-3 px-5 py-4 rounded-2xl shadow-2xl border border-white/20 backdrop-blur-md transition-all duration-300 pointer-events-auto"
      >
        <div class="shrink-0 mt-0.5">
          <BaseIcon :path="getIcon(item.color)" size="20" />
        </div>
        <div class="flex-1 text-sm font-bold leading-snug">
          {{ item.message }}
        </div>
        <button
          @click="store.remove(item.id)"
          class="shrink-0 -mt-1 -mr-2 p-1.5 rounded-full hover:bg-black/10 transition-colors"
        >
          <BaseIcon :path="mdiClose" size="18" />
        </button>
      </div>
    </transition-group>
  </div>
</template>

<style scoped>
.toast-enter-active {
  transition: all 0.5s cubic-bezier(0.175, 0.885, 0.32, 1.275);
}
.toast-leave-active {
  transition: all 0.3s ease-in;
}
.toast-enter-from {
  opacity: 0;
  transform: translateX(50px) scale(0.9);
}
.toast-leave-to {
  opacity: 0;
  transform: translateX(20px) scale(0.9);
}
</style>
