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
      <!-- Header Section with Mesh Accent -->
      <div class="relative overflow-hidden rounded-[2rem] bg-white dark:bg-slate-900 shadow-sm border border-blue-400/60 dark:border-blue-800/80 p-8 mb-8">
        <div class="absolute top-[-20%] right-[-10%] w-64 h-64 rounded-full bg-blue-400/10 blur-3xl"></div>
        <div class="absolute bottom-[-20%] left-[-10%] w-64 h-64 rounded-full bg-indigo-400/10 blur-3xl"></div>
        
        <div class="relative z-10 flex flex-col md:flex-row md:items-center justify-between">
          <div>
            <h1 class="text-3xl font-extrabold text-slate-800 dark:text-slate-100 tracking-tight">Selamat Datang, Admin Utama!</h1>
            <p class="text-slate-500 mt-1 dark:text-slate-400 font-medium">Kelola sistem ujian berbasis komputer dengan lebih elegan</p>
          </div>
          
          <div class="flex items-center space-x-6 mt-6 md:mt-0 text-sm font-bold">
            <span class="text-slate-400 dark:text-slate-500 uppercase tracking-widest text-[10px]">Support:</span>
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
    </div>



      <div v-if="!authStore.isAuthenticated" class="mb-6 rounded-xl bg-amber-50 px-4 py-3 text-sm text-amber-700 border border-amber-100">
        Login terlebih dulu agar dashboard dapat menampilkan statistik backend.
      </div>
      <div v-else-if="errorMessage" class="mb-6 rounded-xl bg-red-50 px-4 py-3 text-sm text-red-700 border border-red-100">
        {{ errorMessage }}
      </div>

      <!-- Stats Cards Row -->
      <div class="mb-8 grid grid-cols-2 gap-3 sm:gap-4 lg:grid-cols-3 xl:grid-cols-5">
        <DashboardCard 
          label="Total Siswa" 
          :number="stats.totalSiswa" 
          :icon="mdiAccountMultipleOutline" 
          color="blue" 
          :loading="isLoading"
        />
        <DashboardCard 
          label="Siswa Aktif" 
          :number="stats.siswaAktif" 
          :icon="mdiAccountCheckOutline" 
          color="emerald" 
          :loading="isLoading"
        />
        <DashboardCard 
          label="Bank Soal" 
          :number="stats.bankSoal" 
          :icon="mdiBookOpenVariant" 
          color="cyan" 
          :loading="isLoading"
        />
        <DashboardCard 
          label="Ujian Aktif" 
          :number="stats.ujianAktif" 
          :icon="mdiFileDocumentCheckOutline" 
          color="orange" 
          :loading="isLoading"
        />
        <DashboardCard 
          label="Total Nilai" 
          :number="stats.totalNilai" 
          :icon="mdiChartBoxOutline" 
          color="indigo" 
          :loading="isLoading"
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
          <!-- Spacer to align with "Menu Utama" title on the left -->
          <div class="hidden lg:flex items-center mb-6 invisible">
            <div class="p-1.5 rounded-lg mr-3">
              <BaseIcon :path="mdiViewGridOutline" size="20" />
            </div>
            <h2 class="text-2xl font-bold">Spacer</h2>
          </div>

          <div class="bg-white dark:bg-slate-900 rounded-2xl border border-orange-400/60 dark:border-orange-800/80 shadow-sm p-6 pb-8">
            <div class="flex items-center mb-4">
              <BaseIcon :path="mdiRocketLaunchOutline" size="24" class="text-orange-500 mr-2" />
              <h3 class="text-xl font-bold text-gray-800 dark:text-slate-100">Mulai Cepat AtigaCBT</h3>
            </div>
            
            <p class="text-sm text-gray-600 dark:text-slate-400 mb-4">
              Ikuti langkah berikut untuk mulai menggunakan AtigaCBT:
            </p>
            
            <div class="space-y-3">
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

            <div v-if="isLoading" class="mt-4 flex items-center text-sm text-blue-500 dark:text-blue-400 italic">
              <div class="animate-spin mr-2 h-4 w-4 border-2 border-blue-500 dark:border-blue-400 border-t-transparent rounded-full"></div>
              Sinkronisasi statistik...
            </div>
          </div>
        </div>
      </div>
    </SectionMain>

  </LayoutAuthenticated>
</template>
