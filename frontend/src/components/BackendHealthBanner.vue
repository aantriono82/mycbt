<script setup>
import { onMounted, ref, computed } from 'vue'
import { mdiServer } from '@mdi/js'
import NotificationBar from '@/components/NotificationBar.vue'
import { api } from '@/services/api.js'

const health = ref({
  state: 'idle', // idle | loading | ok | error
  time: null,
  error: null,
})

const formattedTime = computed(() => {
  if (!health.value.time) return ''
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
})
</script>

<template>
  <NotificationBar
    :icon="mdiServer"
    :color="health.state === 'ok' ? 'success' : health.state === 'error' ? 'danger' : 'info'"
  >
    <template v-if="health.state === 'ok'">
      Backend reachable. Server Time: <b>{{ formattedTime }}</b>
    </template>
    <template v-else-if="health.state === 'error'">
      Backend unreachable. {{ health.error }}
    </template>
    <template v-else>
      Checking backend...
    </template>
  </NotificationBar>
</template>
