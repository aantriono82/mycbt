<script setup>
import { onMounted, ref, computed } from 'vue'
import { api } from '@/services/api.js'

const health = ref({
  state: 'idle', // idle | loading | ok | error
  time: null,
  error: null,
})

const formattedTime = computed(() => {
  if (!health.value.time) return 'N/A'
  try {
    const date = new Date(health.value.time)
    return new Intl.DateTimeFormat('id-ID', {
      dateStyle: 'medium',
      timeStyle: 'medium',
      timeZone: 'Asia/Jakarta'
    }).format(date) + ' WIB'
  } catch (e) {
    return health.value.time
  }
})

const fetchHealth = async () => {
  health.value = { state: 'loading', time: null, error: null }

  try {
    const { data } = await api.get('/healthz')
    health.value = { state: 'ok', time: data?.time ?? null, error: null }
  } catch (e) {
    health.value = {
      state: 'error',
      time: null,
      error: e?.message ?? 'Failed to reach backend',
    }
  }
}

onMounted(() => {
  fetchHealth()
  // Refresh every 30 seconds
  setInterval(fetchHealth, 30000)
})
</script>

<template>
  <div class="group relative flex items-center">
    <!-- The glowing radio button -->
    <div 
      class="radio-indicator" 
      :class="{ 
        'status-ok': health.state === 'ok', 
        'status-error': health.state === 'error',
        'status-loading': health.state === 'loading' || health.state === 'idle'
      }"
    >
      <div class="radio-outer">
        <div class="radio-inner"></div>
      </div>
    </div>

    <!-- Tooltip on hover -->
    <div class="pointer-events-none absolute left-full ml-2 hidden w-max group-hover:block">
      <div class="rounded bg-slate-800 px-2 py-1 text-[10px] text-white shadow-lg dark:bg-slate-700">
        <div class="font-bold flex items-center">
          <div class="h-1.5 w-1.5 rounded-full mr-1.5" :class="health.state === 'ok' ? 'bg-purple-400' : 'bg-red-400'"></div>
          {{ health.state === 'ok' ? 'BACKEND ONLINE' : health.state === 'error' ? 'BACKEND OFFLINE' : 'CHECKING...' }}
        </div>
        <div v-if="health.state === 'ok'" class="opacity-70 mt-0.5">
          Server Time: {{ formattedTime }}
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.radio-indicator {
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: help;
  padding: 4px;
}

.radio-outer {
  width: 14px;
  height: 14px;
  border-radius: 50%;
  border: 1.5px solid currentColor;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.3s ease;
  position: relative;
}

.radio-inner {
  width: 6px;
  height: 6px;
  background-color: currentColor;
  border-radius: 50%;
  transition: all 0.3s ease;
}

/* OK Status (Purple Glow) */
.status-ok {
  color: #9333ea;
}

:global(.dark) .status-ok {
  color: #a855f7;
}

.status-ok .radio-outer {
  box-shadow: 0 0 8px rgba(168, 85, 247, 0.3);
  animation: glow-pulse-purple 2s infinite alternate ease-in-out;
}

.status-ok .radio-inner {
  box-shadow: 0 0 10px #a855f7;
  animation: radio-pulse 1.5s infinite ease-in-out;
}

/* Error Status (Red Glow) */
.status-error {
  color: #dc2626;
}

:global(.dark) .status-error {
  color: #f87171;
}

.status-error .radio-outer {
  box-shadow: 0 0 8px rgba(239, 68, 68, 0.3);
  animation: glow-pulse-red 1s infinite alternate ease-in-out;
}

.status-error .radio-inner {
  box-shadow: 0 0 10px #ef4444;
  animation: radio-pulse-red 0.5s infinite ease-in-out;
}

/* Loading Status */
.status-loading {
  color: #94a3b8;
}

.status-loading .radio-outer {
  opacity: 0.5;
}

@keyframes glow-pulse-purple {
  from { box-shadow: 0 0 4px rgba(168, 85, 247, 0.2); }
  to { box-shadow: 0 0 12px rgba(168, 85, 247, 0.6); }
}

@keyframes glow-pulse-red {
  from { box-shadow: 0 0 4px rgba(239, 68, 68, 0.2); }
  to { box-shadow: 0 0 12px rgba(239, 68, 68, 0.6); }
}

@keyframes radio-pulse {
  0%, 100% { transform: scale(1); opacity: 1; }
  50% { transform: scale(1.4); opacity: 0.8; }
}

@keyframes radio-pulse-red {
  0%, 100% { transform: scale(1); }
  50% { transform: scale(1.6); }
}
</style>
