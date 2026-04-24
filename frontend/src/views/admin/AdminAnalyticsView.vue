<script setup>
import { onMounted, ref, computed } from 'vue'
import {
  mdiChartLine,
  mdiChartBar,
  mdiRefresh,
  mdiChartBoxOutline,
  mdiAccountGroup,
  mdiBookOpenVariant,
  mdiTrendingUp,
} from '@mdi/js'
import LayoutAuthenticated from '@/layouts/LayoutAuthenticated.vue'
import SectionMain from '@/components/SectionMain.vue'
import SectionTitleLineWithButton from '@/components/SectionTitleLineWithButton.vue'
import CardBox from '@/components/CardBox.vue'
import BaseButton from '@/components/BaseButton.vue'
import BaseIcon from '@/components/BaseIcon.vue'
import DashboardCard from '@/components/DashboardCard.vue'
import { api } from '@/services/api.js'

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
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  BarElement,
  LineController,
  BarController,
  Tooltip,
  Legend,
  Filler,
)

const isLoading = ref(false)
const errorMessage = ref('')
const stats = ref({
  total_exams: 0,
  total_participants: 0,
  average_score: 0,
})

const trendData = ref([])
const subjectData = ref([])
const groupData = ref([])

const trendChartCanvas = ref(null)
const subjectChartCanvas = ref(null)
const groupChartCanvas = ref(null)

let trendChart = null
let subjectChart = null
let groupChart = null

const loadData = async () => {
  isLoading.value = true
  errorMessage.value = ''
  try {
    const { data } = await api.get('/api/v1/analytics/dashboard')
    const res = data?.data
    stats.value = res.summary || stats.value
    trendData.value = res.trends || []
    subjectData.value = res.subjects || []
    groupData.value = res.groups || []

    renderCharts()
  } catch (error) {
    errorMessage.value = error?.response?.data?.error?.message || 'Gagal memuat data analitik'
  } finally {
    isLoading.value = false
  }
}

const renderCharts = () => {
  // Destroy existing charts if they exist
  if (trendChart) trendChart.destroy()
  if (subjectChart) subjectChart.destroy()
  if (groupChart) groupChart.destroy()

  // 1. Trend Chart
  if (trendChartCanvas.value && trendData.value.length > 0) {
    const ctx = trendChartCanvas.value.getContext('2d')
    const gradient = ctx.createLinearGradient(0, 0, 0, 400)
    gradient.addColorStop(0, 'rgba(79, 70, 229, 0.4)')
    gradient.addColorStop(1, 'rgba(79, 70, 229, 0)')

    trendChart = new Chart(trendChartCanvas.value, {
      type: 'line',
      data: {
        labels: trendData.value.map((d) => d.label),
        datasets: [
          {
            label: 'Rata-rata Nilai',
            data: trendData.value.map((d) => d.score),
            borderColor: '#6366f1',
            backgroundColor: gradient,
            fill: true,
            tension: 0.4,
            pointRadius: 4,
            pointBackgroundColor: '#6366f1',
            borderWidth: 3,
          },
        ],
      },
      options: {
        responsive: true,
        maintainAspectRatio: false,
        plugins: { legend: { display: false } },
        scales: {
          y: { min: 0, max: 100, ticks: { stepSize: 20 } },
        },
      },
    })
  }

  // 2. Subject Chart
  if (subjectChartCanvas.value && subjectData.value.length > 0) {
    const ctx = subjectChartCanvas.value.getContext('2d')
    const gradient = ctx.createLinearGradient(0, 0, 400, 0)
    gradient.addColorStop(0, '#10b981')
    gradient.addColorStop(1, '#34d399')

    subjectChart = new Chart(subjectChartCanvas.value, {
      type: 'bar',
      data: {
        labels: subjectData.value.map((d) => d.subject_name),
        datasets: [
          {
            label: 'Rata-rata Nilai',
            data: subjectData.value.map((d) => d.avg_score),
            backgroundColor: gradient,
            borderRadius: 10,
          },
        ],
      },
      options: {
        indexAxis: 'y',
        responsive: true,
        maintainAspectRatio: false,
        plugins: { legend: { display: false } },
        scales: {
          x: { min: 0, max: 100 },
        },
      },
    })
  }

  // 3. Group Chart
  if (groupChartCanvas.value && groupData.value.length > 0) {
    const ctx = groupChartCanvas.value.getContext('2d')
    const gradient = ctx.createLinearGradient(0, 0, 0, 400)
    gradient.addColorStop(0, '#6366f1')
    gradient.addColorStop(1, '#a855f7')

    groupChart = new Chart(groupChartCanvas.value, {
      type: 'bar',
      data: {
        labels: groupData.value.map((d) => d.group_name),
        datasets: [
          {
            label: 'Rata-rata Nilai',
            data: groupData.value.map((d) => d.avg_score),
            backgroundColor: gradient,
            borderRadius: 10,
          },
        ],
      },
      options: {
        responsive: true,
        maintainAspectRatio: false,
        plugins: { legend: { display: false } },
        scales: {
          y: { min: 0, max: 100 },
        },
      },
    })
  }
}

onMounted(loadData)
</script>

<template>
  <LayoutAuthenticated>
    <SectionMain>
      <SectionTitleLineWithButton :icon="mdiChartLine" title="Advanced Analytics Dashboard" main>
        <BaseButton
          :icon="mdiRefresh"
          color="info"
          label="Refresh"
          :disabled="isLoading"
          @click="loadData"
        />
      </SectionTitleLineWithButton>

      <div v-if="errorMessage" class="mb-6 rounded-xl bg-red-50 dark:bg-red-900/20 px-4 py-3 text-sm text-red-700 dark:text-red-400 border border-red-100 dark:border-red-900/40">
        {{ errorMessage }}
      </div>

      <!-- Stats Row -->
      <div class="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
        <DashboardCard 
          label="Total Ujian Terlaksana" 
          :number="stats.total_exams" 
          :icon="mdiChartBoxOutline" 
          color="blue" 
        />
        <DashboardCard 
          label="Total Partisipan (Siswa)" 
          :number="stats.total_participants" 
          :icon="mdiAccountGroup" 
          color="emerald" 
        />
        <DashboardCard 
          label="Harapan Rata-rata (Global)" 
          :number="Math.round(stats.average_score)" 
          :icon="mdiTrendingUp" 
          color="indigo" 
        />
      </div>

      <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <!-- Trend Chart -->
        <CardBox class="lg:col-span-2">
          <div class="flex items-center justify-between mb-4">
            <h3 class="text-lg font-bold dark:text-slate-100">Tren Performa Ujian (10 Terakhir)</h3>
            <div class="text-xs text-slate-500 font-mono italic">Average score per exam</div>
          </div>
          <div class="h-80 relative">
            <canvas ref="trendChartCanvas"></canvas>
            <div v-if="trendData.length === 0 && !isLoading" class="absolute inset-0 flex items-center justify-center text-slate-400 italic">
              Belum ada data ujian yang selesai.
            </div>
          </div>
        </CardBox>

        <!-- Subject Performance -->
        <CardBox>
          <div class="flex items-center justify-between mb-4">
            <h3 class="text-lg font-bold dark:text-slate-100">Performa per Mata Pelajaran</h3>
            <div class="text-xs text-slate-500 font-mono italic">Lower score = difficult</div>
          </div>
          <div class="h-80 relative">
            <canvas ref="subjectChartCanvas"></canvas>
            <div v-if="subjectData.length === 0 && !isLoading" class="absolute inset-0 flex items-center justify-center text-slate-400 italic">
              Belum ada data performa mapel.
            </div>
          </div>
        </CardBox>

        <!-- Group Performance -->
        <CardBox>
          <div class="flex items-center justify-between mb-4">
            <h3 class="text-lg font-bold dark:text-slate-100">Performa per Rombel / Group</h3>
            <div class="text-xs text-slate-500 font-mono italic">Comparing class averages</div>
          </div>
          <div class="h-80 relative">
            <canvas ref="groupChartCanvas"></canvas>
            <div v-if="groupData.length === 0 && !isLoading" class="absolute inset-0 flex items-center justify-center text-slate-400 italic">
              Belum ada data performa group.
            </div>
          </div>
        </CardBox>
      </div>

      <!-- Recommendation Section -->
      <div v-if="stats.average_score > 0" class="mt-8 rounded-2xl bg-indigo-50 dark:bg-indigo-900/20 border border-indigo-100 dark:border-indigo-800 p-6">
        <h4 class="text-indigo-800 dark:text-indigo-300 font-bold flex items-center mb-2">
          <BaseIcon :path="mdiTrendingUp" class="mr-2" size="20" />
          Data-Driven Insights
        </h4>
        <div class="grid md:grid-cols-2 gap-4">
          <div class="text-sm text-indigo-700 dark:text-indigo-400">
            <p v-if="stats.average_score < 60">
              ⚠ **Peringatan**: Rata-rata global berada di bawah standar ({{ Math.round(stats.average_score) }}). Pertimbangkan untuk mengevaluasi kembali tingkat kesulitan bank soal atau memberikan bimbingan tambahan bagi kelompok dengan performa rendah.
            </p>
            <p v-else-if="stats.average_score < 75">
              💡 **Info**: Performa siswa stabil. Fokus pada peningkatan kualitas distraktor di bank soal untuk membedakan siswa dengan kemampuan tinggi.
            </p>
            <p v-else>
              ✅ **Sangat Baik**: Rata-rata global tinggi ({{ Math.round(stats.average_score) }}). Siswa memahami materi dengan sangat baik atau soal kurang menantang.
            </p>
          </div>
          <div class="text-sm text-indigo-700 dark:text-indigo-400 border-l border-indigo-200 dark:border-indigo-800 pl-4">
            <ul class="list-disc list-inside space-y-1">
              <li>Identifikasi Subject dengan skor terendah untuk fokus pendampingan.</li>
              <li>Bandingkan performa antar Group untuk pemerataan kualitas pembelajaran.</li>
              <li>Gunakan grafik tren untuk melihat efektivitas kurikulum dari waktu ke waktu.</li>
            </ul>
          </div>
        </div>
      </div>
    </SectionMain>
  </LayoutAuthenticated>
</template>
