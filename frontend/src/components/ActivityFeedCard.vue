<script setup>
import { computed, onMounted, ref } from 'vue'
import { api } from '@/services/api.js'
import {
  mdiHistory,
  mdiAlertCircle,
  mdiCheckCircle,
  mdiPencil,
  mdiDelete,
  mdiPlus,
  mdiAccount,
} from '@mdi/js'
import BaseIcon from '@/components/BaseIcon.vue'
import { useAuthStore } from '@/stores/auth.js'
import BaseSkeleton from '@/components/BaseSkeleton.vue'

const authStore = useAuthStore()

const isLoading = ref(true)
const logs = ref([])
const error = ref('')

const fetchLogs = async () => {
  if (!authStore.isAuthenticated) return
  isLoading.value = true
  error.value = ''
  try {
    const { data } = await api.get('/api/v1/admin/audit-logs', {
      params: { limit: 5 }
    })
    logs.value = data?.data || []
  } catch (err) {
    error.value = 'Gagal memuat log aktivitas'
  } finally {
    isLoading.value = false
  }
}

onMounted(() => {
  fetchLogs()
})

const getIconForAction = (action) => {
  if (!action) return mdiHistory
  action = action.toLowerCase()
  if (action.includes('create') || action.includes('add')) return mdiPlus
  if (action.includes('update') || action.includes('edit')) return mdiPencil
  if (action.includes('delete') || action.includes('remove')) return mdiDelete
  if (action.includes('login')) return mdiAccount
  return mdiHistory
}

const getColorForAction = (action) => {
  if (!action) return 'text-slate-500 bg-slate-100 dark:bg-slate-800'
  action = action.toLowerCase()
  if (action.includes('create') || action.includes('add')) return 'text-emerald-500 bg-emerald-100 dark:bg-emerald-900/30'
  if (action.includes('update') || action.includes('edit')) return 'text-blue-500 bg-blue-100 dark:bg-blue-900/30'
  if (action.includes('delete') || action.includes('remove')) return 'text-red-500 bg-red-100 dark:bg-red-900/30'
  if (action.includes('login')) return 'text-indigo-500 bg-indigo-100 dark:bg-indigo-900/30'
  return 'text-slate-500 bg-slate-100 dark:bg-slate-800'
}

const formatTimeAgo = (dateStr) => {
  if (!dateStr) return ''
  const date = new Date(dateStr)
  const now = new Date()
  const diffInSeconds = Math.floor((now - date) / 1000)

  if (diffInSeconds < 60) return 'Baru saja'
  if (diffInSeconds < 3600) return `${Math.floor(diffInSeconds / 60)} mnt lalu`
  if (diffInSeconds < 86400) return `${Math.floor(diffInSeconds / 3600)} jam lalu`
  return `${Math.floor(diffInSeconds / 86400)} hr lalu`
}
</script>

<template>
  <div class="bg-white dark:bg-slate-900 rounded-2xl border border-slate-100 dark:border-slate-800 shadow-sm p-6 flex flex-col h-full">
    <div class="flex items-center justify-between mb-6">
      <div class="flex items-center">
        <BaseIcon :path="mdiHistory" size="24" class="text-blue-500 mr-2" />
        <h3 class="text-xl font-bold text-gray-800 dark:text-slate-100">Aktivitas Live</h3>
      </div>
      <button @click="fetchLogs" class="text-xs text-blue-500 hover:underline flex items-center">
        Refresh
      </button>
    </div>

    <div v-if="isLoading" class="flex-1 space-y-4">
      <div v-for="i in 5" :key="i" class="flex gap-3">
        <BaseSkeleton class="w-8 h-8 rounded-full shrink-0" />
        <div class="flex-1 space-y-2 py-1">
          <BaseSkeleton class="w-3/4 h-3" />
          <BaseSkeleton class="w-1/4 h-2" />
        </div>
      </div>
    </div>

    <div v-else-if="error" class="flex-1 flex flex-col items-center justify-center text-slate-400 py-6">
      <BaseIcon :path="mdiAlertCircle" size="32" class="mb-2 opacity-50" />
      <p class="text-sm">{{ error }}</p>
    </div>

    <div v-else-if="logs.length === 0" class="flex-1 flex flex-col items-center justify-center text-slate-400 py-6">
      <BaseIcon :path="mdiCheckCircle" size="32" class="mb-2 opacity-50 text-emerald-500" />
      <p class="text-sm">Belum ada aktivitas</p>
    </div>

    <div v-else class="flex-1 overflow-y-auto pr-2 custom-scrollbar space-y-4 relative">
      <!-- Continuous vertical line -->
      <div class="absolute left-4 top-4 bottom-4 w-0.5 bg-slate-100 dark:bg-slate-800 z-0"></div>
      
      <div v-for="(log, idx) in logs" :key="log.id || idx" class="flex gap-4 relative z-10 animate-fade-in group" :style="{ animationDelay: `${idx * 0.1}s` }">
        <div class="shrink-0 flex items-center justify-center w-8 h-8 rounded-full ring-4 ring-white dark:ring-slate-900 transition-transform group-hover:scale-110" :class="getColorForAction(log.action)">
          <BaseIcon :path="getIconForAction(log.action)" size="14" />
        </div>
        <div class="flex-1 min-w-0 bg-slate-50/50 dark:bg-slate-800/20 p-3 rounded-xl border border-slate-100 dark:border-slate-800 transition-colors group-hover:bg-slate-50 dark:group-hover:bg-slate-800/40">
          <p class="text-sm font-medium text-slate-800 dark:text-slate-200 truncate">
            <span class="font-bold text-blue-600 dark:text-blue-400">{{ log.username || 'System' }}</span>
            {{ log.action }}
          </p>
          <p class="text-xs text-slate-500 dark:text-slate-400 truncate mt-0.5">
            {{ log.target_type }} {{ log.target_id ? `(#${log.target_id})` : '' }}
          </p>
          <p class="text-[10px] text-slate-400 mt-2 font-mono flex items-center gap-1">
            <BaseIcon :path="mdiHistory" size="10" />
            {{ formatTimeAgo(log.created_at) }}
          </p>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.animate-fade-in {
  animation: fadeIn 0.4s ease-out forwards;
  opacity: 0;
}
@keyframes fadeIn {
  from { opacity: 0; transform: translateX(-10px); }
  to { opacity: 1; transform: translateX(0); }
}
.custom-scrollbar::-webkit-scrollbar {
  width: 4px;
}
.custom-scrollbar::-webkit-scrollbar-track {
  background: transparent;
}
.custom-scrollbar::-webkit-scrollbar-thumb {
  background-color: #cbd5e1;
  border-radius: 20px;
}
.dark .custom-scrollbar::-webkit-scrollbar-thumb {
  background-color: #334155;
}
</style>
