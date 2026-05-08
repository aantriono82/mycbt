<script setup>
import { onMounted, onUnmounted, ref, computed } from 'vue'
import { api } from '@/services/api.js'

const props = defineProps({
  subtle: {
    type: Boolean,
    default: false,
  },
  tapTooltip: {
    type: Boolean,
    default: false,
  },
})

const health = ref({
  state: 'idle', // idle | loading | ok | error
  time: null,
  error: null,
})
let healthTimer = null
const isTooltipOpen = ref(false)
let tooltipTimer = null

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
  healthTimer = setInterval(fetchHealth, 60000)
})

onUnmounted(() => {
  if (healthTimer) {
    clearInterval(healthTimer)
    healthTimer = null
  }
  if (tooltipTimer) {
    clearTimeout(tooltipTimer)
    tooltipTimer = null
  }
})

const toggleTooltipOnTap = () => {
  if (!props.tapTooltip) return
  isTooltipOpen.value = !isTooltipOpen.value
  if (tooltipTimer) {
    clearTimeout(tooltipTimer)
    tooltipTimer = null
  }
  if (isTooltipOpen.value) {
    tooltipTimer = setTimeout(() => {
      isTooltipOpen.value = false
      tooltipTimer = null
    }, 2000)
  }
}
</script>

<template>
  <div class="group relative flex items-center" @click.stop="toggleTooltipOnTap">
    <!-- The glowing radio button -->
    <div 
      class="radio-indicator" 
      :class="{ 
        [props.subtle ? 'status-ok-subtle' : 'status-ok']: health.state === 'ok',
        [props.subtle ? 'status-error-subtle' : 'status-error']: health.state === 'error',
        'status-loading': health.state === 'loading' || health.state === 'idle'
      }"
    >
      <!-- Ping ring overlay (only when online) -->
      <div
        v-if="health.state === 'ok'"
        class="absolute inset-0 rounded-full bg-emerald-400/30 animate-ping pointer-events-none"
      ></div>
      <!-- Error ping overlay -->
      <div
        v-else-if="health.state === 'error'"
        class="absolute inset-0 rounded-full bg-red-400/30 animate-ping pointer-events-none"
      ></div>

      <div class="radio-outer">
        <div class="radio-inner"></div>
      </div>
    </div>

    <!-- Tooltip on hover -->
    <div
      class="pointer-events-none absolute left-full ml-2 w-max"
      :class="props.tapTooltip ? (isTooltipOpen ? 'block' : 'hidden') : 'hidden group-hover:block'"
    >
      <div class="rounded bg-slate-800 px-2 py-1 text-[10px] text-white shadow-lg dark:bg-slate-700">
        <div class="font-bold flex items-center">
          <div class="h-1.5 w-1.5 rounded-full mr-1.5" :class="health.state === 'ok' ? 'bg-emerald-400' : 'bg-red-400'"></div>
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
  position: relative;
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

/* OK Status (Emerald Green Glow) */
.status-ok {
  color: #10b981;
  filter: drop-shadow(0 0 6px rgba(16, 185, 129, 0.6));
  animation: drop-glow-green 1.8s infinite alternate ease-in-out;
}

:global(.dark) .status-ok {
  color: #34d399;
}

.status-ok .radio-outer {
  box-shadow: 0 0 10px rgba(16, 185, 129, 0.5), 0 0 20px rgba(16, 185, 129, 0.25);
  animation: glow-pulse-green 1.8s infinite alternate ease-in-out;
}

.status-ok .radio-inner {
  box-shadow: 0 0 12px #10b981, 0 0 24px rgba(16, 185, 129, 0.5);
  animation: radio-pulse 1.4s infinite ease-in-out;
}

.status-ok-subtle {
  color: #10b981;
  filter: drop-shadow(0 0 4px rgba(16, 185, 129, 0.4));
}

:global(.dark) .status-ok-subtle {
  color: #34d399;
}

.status-ok-subtle .radio-outer {
  box-shadow: 0 0 6px rgba(16, 185, 129, 0.35);
}

.status-ok-subtle .radio-inner {
  box-shadow: 0 0 6px rgba(16, 185, 129, 0.6);
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

.status-error-subtle {
  color: #dc2626;
}

:global(.dark) .status-error-subtle {
  color: #f87171;
}

.status-error-subtle .radio-outer {
  box-shadow: 0 0 5px rgba(239, 68, 68, 0.25);
}

.status-error-subtle .radio-inner {
  box-shadow: 0 0 6px rgba(239, 68, 68, 0.5);
}

/* Loading Status */
.status-loading {
  color: #94a3b8;
}

.status-loading .radio-outer {
  opacity: 0.5;
}

@keyframes glow-pulse-green {
  from { box-shadow: 0 0 6px rgba(16, 185, 129, 0.3), 0 0 12px rgba(16, 185, 129, 0.15); }
  to   { box-shadow: 0 0 16px rgba(16, 185, 129, 0.7), 0 0 30px rgba(16, 185, 129, 0.35); }
}

@keyframes drop-glow-green {
  from { filter: drop-shadow(0 0 3px rgba(16, 185, 129, 0.4)); }
  to   { filter: drop-shadow(0 0 10px rgba(16, 185, 129, 0.8)); }
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
