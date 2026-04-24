<script setup>
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'

const props = defineProps({
  duration: { type: Number, required: true },
  submitExam: { type: Function, default: null },
})

const emit = defineEmits(['time-up'])

const remaining = ref(Math.max(0, Number(props.duration || 0)))
const intervalId = ref(null)
const fired = ref(false)

const formatTime = (seconds) => {
  const safe = Math.max(0, Number(seconds || 0))
  const h = Math.floor(safe / 3600)
  const m = Math.floor((safe % 3600) / 60)
  const s = safe % 60
  return [h, m, s].map((v) => String(v).padStart(2, '0')).join(':')
}

const display = computed(() => formatTime(remaining.value))

const clearTimer = () => {
  if (intervalId.value) {
    clearInterval(intervalId.value)
    intervalId.value = null
  }
}

const onTimeUp = () => {
  if (fired.value) return
  fired.value = true
  emit('time-up')
  if (typeof props.submitExam === 'function') {
    props.submitExam()
  }
}

const start = () => {
  clearTimer()
  fired.value = false
  remaining.value = Math.max(0, Number(props.duration || 0))
  if (remaining.value === 0) {
    onTimeUp()
    return
  }

  intervalId.value = setInterval(() => {
    if (remaining.value > 0) {
      remaining.value -= 1
    }
    if (remaining.value <= 0) {
      clearTimer()
      onTimeUp()
    }
  }, 1000)
}

watch(
  () => props.duration,
  () => start(),
)

onMounted(() => start())
onUnmounted(() => clearTimer())
</script>

<template>
  <span data-testid="exam-timer">{{ display }}</span>
</template>

