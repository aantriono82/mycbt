<script setup>
import { onMounted, ref, watch, onUnmounted } from 'vue'
import {
  BarController,
  BarElement,
  CategoryScale,
  Chart,
  Filler,
  Legend,
  LineController,
  LineElement,
  LinearScale,
  PointElement,
  Tooltip,
} from 'chart.js'

Chart.register(
  BarController,
  BarElement,
  CategoryScale,
  Filler,
  Legend,
  LineController,
  LineElement,
  LinearScale,
  PointElement,
  Tooltip,
)

const props = defineProps({
  data: {
    type: Object,
    required: true
  },
  options: {
    type: Object,
    default: () => ({})
  },
  type: {
    type: String,
    default: 'line'
  }
})

const root = ref(null)
let chart = null

const initChart = () => {
  if (chart) {
    chart.destroy()
  }

  if (!root.value) return

  chart = new Chart(root.value, {
    type: props.type,
    data: props.data,
    options: {
      responsive: true,
      maintainAspectRatio: false,
      plugins: {
        legend: {
          display: false
        }
      },
      ...props.options
    }
  })
}

onMounted(initChart)

watch(() => props.data, initChart, { deep: true })

onUnmounted(() => {
  if (chart) {
    chart.destroy()
  }
})
</script>

<template>
  <div class="relative h-full w-full">
    <canvas ref="root" />
  </div>
</template>
