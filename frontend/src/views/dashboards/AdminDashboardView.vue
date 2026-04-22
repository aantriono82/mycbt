<script setup>
import { computed, onMounted, ref } from 'vue'
import {
  mdiAccountCheckOutline,
  mdiAccountTie,
  mdiBookOpenVariant,
  mdiHomeOutline,
  mdiAccountGroup,
  mdiViewGridOutline,
  mdiBookOpenPageVariantOutline,
  mdiSchoolOutline,
  mdiAccountCircleOutline,
  mdiRocketLaunchOutline,
  mdiSend,
  mdiWeb,
  mdiEmailOutline,
  mdiAccountMultipleOutline,
  mdiFileDocumentCheckOutline,
  mdiChartBoxOutline,
} from '@mdi/js'
import LayoutAuthenticated from '@/layouts/LayoutAuthenticated.vue'
import SectionMain from '@/components/SectionMain.vue'
import DashboardCard from '@/components/DashboardCard.vue'
import QuickMenuCard from '@/components/QuickMenuCard.vue'
import BaseIcon from '@/components/BaseIcon.vue'
import { api } from '@/services/api.js'
import { useAuthStore } from '@/stores/auth.js'

const authStore = useAuthStore()

const isLoading = ref(false)
const errorMessage = ref('')
const stats = ref({
  totalSiswa: 0,
  siswaAktif: 0,
  bankSoal: 0,
  ujianAktif: 0,
  totalNilai: 0,
})

const canLoad = computed(() => authStore.isAuthenticated)

const loadStats = async () => {
  if (!canLoad.value) return
  isLoading.value = true
  errorMessage.value = ''
  try {
    const { data } = await api.get('/api/v1/admin/dashboard/stats')
    const res = data?.data
    
    stats.value = {
      totalSiswa: res.total_students || 0,
      siswaAktif: res.total_students - res.pending_registrations,
      bankSoal: res.total_question_sets || 0,
      ujianAktif: res.total_exams || 0,
      totalNilai: Math.round(res.average_score) || 0,
    }
  } catch (error) {
    errorMessage.value = error?.response?.data?.error?.message || 'Gagal memuat ringkasan dashboard'
  } finally {
    isLoading.value = false
  }
}

onMounted(loadStats)
</script>

<template>
  <LayoutAuthenticated>
    <SectionMain>
      <!-- Header Section -->
      <div class="flex flex-col md:flex-row md:items-center justify-between mb-8">
        <div>
          <h1 class="text-3xl font-bold text-gray-800 dark:text-slate-100">Selamat Datang, Admin Utama!</h1>
          <p class="text-gray-500 mt-1 dark:text-slate-400">Kelola sistem ujian berbasis komputer dengan mudah</p>
        </div>
        
        <div class="flex items-center space-x-6 mt-4 md:mt-0 text-sm font-medium">
          <span class="text-gray-400 dark:text-slate-500">Support:</span>
          <a href="https://t.me/aantriono" target="_blank" class="flex items-center text-blue-600 hover:text-blue-700">
            <BaseIcon :path="mdiSend" size="18" class="mr-1 rotate-[-30deg]" />
            Telegram
          </a>
          <a href="http://www.aantriono.com" target="_blank" class="flex items-center text-emerald-600 hover:text-emerald-700">
            <BaseIcon :path="mdiWeb" size="18" class="mr-1" />
            Website
          </a>
          <a href="mailto:aantriono82@gmail.com" class="flex items-center text-purple-600 hover:text-purple-700">
            <BaseIcon :path="mdiEmailOutline" size="18" class="mr-1" />
            Email
          </a>
        </div>
      </div>


      <div v-if="!authStore.isAuthenticated" class="mb-6 rounded-xl bg-amber-50 px-4 py-3 text-sm text-amber-700 border border-amber-100">
        Login terlebih dulu agar dashboard dapat menampilkan statistik backend.
      </div>
      <div v-else-if="errorMessage" class="mb-6 rounded-xl bg-red-50 px-4 py-3 text-sm text-red-700 border border-red-100">
        {{ errorMessage }}
      </div>

      <!-- Stats Cards Row -->
      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-5 gap-4 mb-8">
        <DashboardCard 
          label="Total Siswa" 
          :number="stats.totalSiswa" 
          :icon="mdiAccountMultipleOutline" 
          color="blue" 
        />
        <DashboardCard 
          label="Siswa Aktif" 
          :number="stats.siswaAktif" 
          :icon="mdiAccountCheckOutline" 
          color="emerald" 
        />
        <DashboardCard 
          label="Bank Soal" 
          :number="stats.bankSoal" 
          :icon="mdiBookOpenVariant" 
          color="purple" 
        />
        <DashboardCard 
          label="Ujian Aktif" 
          :number="stats.ujianAktif" 
          :icon="mdiFileDocumentCheckOutline" 
          color="orange" 
        />
        <DashboardCard 
          label="Total Nilai" 
          :number="stats.totalNilai" 
          :icon="mdiChartBoxOutline" 
          color="indigo" 
        />
      </div>

      <div class="grid grid-cols-1 lg:grid-cols-3 gap-8">
        <!-- Main Menu Section -->
        <div class="lg:col-span-2">
          <div class="flex items-center mb-6">
            <div class="p-1.5 bg-blue-100 dark:bg-blue-900/30 rounded-lg mr-3">
              <BaseIcon :path="mdiViewGridOutline" size="20" class="text-blue-600 dark:text-blue-400" />
            </div>
            <h2 class="text-2xl font-bold text-gray-800 dark:text-slate-100">Menu Utama</h2>
          </div>
          
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <QuickMenuCard 
              label="Data Siswa" 
              description="Kelola data siswa" 
              :icon="mdiAccountGroup" 
              to="/admin/master-data/siswa"
              color="blue"
            />
            <QuickMenuCard 
              label="Level & Group" 
              description="Kelola tingkat kelas" 
              :icon="mdiViewGridOutline" 
              to="/admin/master-data"
              color="indigo"
            />
            <QuickMenuCard 
              label="Mata Pelajaran" 
              description="Kelola mapel" 
              :icon="mdiBookOpenPageVariantOutline" 
              to="/admin/master-data"
              color="blue"
            />
            <QuickMenuCard 
              label="Program" 
              description="IPA, IPS, Bahasa" 
              :icon="mdiSchoolOutline" 
              to="/admin/master-data"
              color="indigo"
            />
            <QuickMenuCard 
              label="Users" 
              description="Admin & Guru" 
              :icon="mdiAccountCircleOutline" 
              to="/admin/master-data/guru"
              color="blue"
            />
            <QuickMenuCard 
              label="Bank Soal" 
              description="Kelola soal ujian" 
              :icon="mdiBookOpenVariant" 
              to="/admin/bank-soal"
              color="emerald"
            />
          </div>
        </div>

        <!-- Quick Start Section -->
        <div>
          <div class="bg-white dark:bg-slate-900 rounded-2xl border border-gray-100 dark:border-slate-800 shadow-sm p-6">
            <div class="flex items-center mb-6">
              <BaseIcon :path="mdiRocketLaunchOutline" size="24" class="text-orange-500 mr-2" />
              <h3 class="text-xl font-bold text-gray-800 dark:text-slate-100">Mulai Cepat AtigaCBT</h3>
            </div>
            
            <p class="text-sm text-gray-600 dark:text-slate-400 mb-6">
              Ikuti langkah berikut untuk mulai menggunakan AtigaCBT:
            </p>
            
            <div class="space-y-4">
              <div class="flex items-start">
                <div class="text-blue-600 dark:text-blue-400 font-bold mr-3 mt-0.5">1.</div>
                <div>
                  <p class="text-sm font-semibold text-gray-800 dark:text-slate-200">Lengkapi Master Data</p>
                  <p class="text-xs text-gray-500 dark:text-slate-500">(Group, Level, Program, Mata Pelajaran)</p>
                </div>
              </div>
              
              <div class="flex items-start">
                <div class="text-blue-600 dark:text-blue-400 font-bold mr-3 mt-0.5">2.</div>
                <div>
                  <p class="text-sm font-semibold text-gray-800 dark:text-slate-200">Input data Siswa</p>
                </div>
              </div>
              
              <div class="flex items-start">
                <div class="text-blue-600 dark:text-blue-400 font-bold mr-3 mt-0.5">3.</div>
                <div>
                  <p class="text-sm font-semibold text-gray-800 dark:text-slate-200">Buat Bank Soal</p>
                </div>
              </div>
              
              <div class="flex items-start">
                <div class="text-blue-600 dark:text-blue-400 font-bold mr-3 mt-0.5">4.</div>
                <div>
                  <p class="text-sm font-semibold text-gray-800 dark:text-slate-200">Buat Jadwal Ujian</p>
                </div>
              </div>
            </div>

            <div v-if="isLoading" class="mt-8 flex items-center text-sm text-blue-500 dark:text-blue-400 italic">
              <div class="animate-spin mr-2 h-4 w-4 border-2 border-blue-500 dark:border-blue-400 border-t-transparent rounded-full"></div>
              Sinkronisasi statistik...
            </div>
          </div>
        </div>
      </div>
    </SectionMain>
  </LayoutAuthenticated>
</template>

