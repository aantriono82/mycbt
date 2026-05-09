<script setup>
import { 
  mdiCheckCircle, 
  mdiAlertCircle, 
  mdiAlert, 
  mdiInformation,
  mdiClose
} from '@mdi/js'
import { onMounted, ref } from 'vue'
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
  danger: 'bg-red-600/90 text-white shadow-red-500/30 border-red-400/30',
  success: 'bg-emerald-600/90 text-white shadow-emerald-500/30 border-emerald-400/30',
  warning: 'bg-amber-500/90 text-white shadow-amber-500/30 border-amber-400/30',
  info: 'bg-blue-600/90 text-white shadow-blue-500/30 border-blue-400/30'
}

const progressClasses = {
  danger: 'bg-red-300/40',
  success: 'bg-emerald-300/40',
  warning: 'bg-amber-200/40',
  info: 'bg-blue-200/40'
}
</script>

<template>
  <div class="fixed top-6 right-6 z-[9999] flex flex-col gap-3 w-full max-w-sm pointer-events-none">
    <transition-group name="toast">
      <div
        v-for="item in store.items"
        :key="item.id"
        :class="colorClasses[item.color] || colorClasses.info"
        class="group relative flex items-start gap-4 px-5 py-4 rounded-2xl shadow-2xl border backdrop-blur-xl transition-all duration-500 pointer-events-auto overflow-hidden active:scale-95"
      >
        <!-- Progress Bar Background -->
        <div 
          class="absolute bottom-0 left-0 h-1 transition-all duration-[linear]"
          :class="progressClasses[item.color] || progressClasses.info"
          :style="{ 
            width: '100%',
            animation: `shrink ${item.timeout}ms linear forwards`
          }"
        ></div>

        <div class="shrink-0 mt-0.5">
          <div class="p-1.5 bg-white/20 rounded-xl shadow-inner">
            <BaseIcon :path="getIcon(item.color)" size="20" />
          </div>
        </div>
        
        <div class="flex-1">
          <div class="text-sm font-black leading-tight mb-0.5 uppercase tracking-tight">
            {{ item.color === 'danger' ? 'Error' : item.color === 'success' ? 'Success' : item.color === 'warning' ? 'Warning' : 'Info' }}
          </div>
          <div class="text-xs font-bold opacity-90 leading-relaxed">
            {{ item.message }}
          </div>
        </div>

        <button
          @click="store.remove(item.id)"
          class="shrink-0 -mt-1 -mr-2 p-1.5 rounded-full hover:bg-white/20 transition-all active:scale-90"
        >
          <BaseIcon :path="mdiClose" size="18" />
        </button>
      </div>
    </transition-group>
  </div>
</template>

<style scoped>
@keyframes shrink {
  from { width: 100%; }
  to { width: 0%; }
}

.toast-enter-active {
  transition: all 0.6s cubic-bezier(0.34, 1.56, 0.64, 1);
}
.toast-leave-active {
  transition: all 0.4s cubic-bezier(0.4, 0, 0.2, 1);
}
.toast-enter-from {
  opacity: 0;
  transform: translateX(100px) scale(0.8) rotate(5deg);
}
.toast-leave-to {
  opacity: 0;
  transform: translateX(50px) scale(0.9);
  filter: blur(8px);
}

.toast-move {
  transition: transform 0.4s ease;
}
</style>
