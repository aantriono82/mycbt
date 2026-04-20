<script setup>
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import LayoutAuthenticated from '@/layouts/LayoutAuthenticated.vue'
import SectionMain from '@/components/SectionMain.vue'
import SectionTitleLineWithButton from '@/components/SectionTitleLineWithButton.vue'
import CardBox from '@/components/CardBox.vue'
import { useAuthStore } from '@/stores/auth.js'

const route = useRoute()
const authStore = useAuthStore()

const title = computed(() => route.meta?.title ?? 'Halaman')
const icon = computed(() => route.meta?.icon ?? null)

const moduleNotes = computed(() => {
  if (String(route.path).includes('/monitor-ujian')) {
    return 'Modul ini menunggu backend realtime (monitor session/SSE). Struktur Ujian dan Token sudah siap sebagai fondasi.'
  }
  if (String(route.path).includes('/monitor-peserta')) {
    return 'Monitor peserta akan memanfaatkan session ujian, status login, dan heartbeat siswa. Backend belum masuk tahap ini.'
  }
  if (String(route.path).includes('/reset-login')) {
    return 'Reset login akan dihubungkan ke session ujian siswa setelah modul pengerjaan ujian selesai.'
  }
  if (String(route.path).includes('/evaluasi')) {
    return 'Evaluasi akan membaca hasil ujian, analisis butir, dan rekap nilai setelah modul pengerjaan ujian aktif.'
  }
  if (String(route.path).includes('/cetak')) {
    return 'Cetak akan mengandalkan data ujian, kartu, absensi, dan hasil. Belum diimplementasikan di backend.'
  }
  if (String(route.path).includes('/settings')) {
    return 'Settings akan berisi identitas sekolah dan pengaturan sistem. Belum ada endpoint backend khusus untuk modul ini.'
  }
  return 'Halaman ini belum penuh, tetapi roadmap backend dan frontend-nya sudah dipetakan di blueprint.'
})
</script>

<template>
  <LayoutAuthenticated>
    <SectionMain>
      <SectionTitleLineWithButton :icon="icon" :title="title" main />
      <CardBox>
        <div class="space-y-4">
          <div class="rounded-xl bg-slate-50 dark:bg-slate-800/50 p-6 border border-slate-100 dark:border-slate-800">
            <h3 class="mb-2 text-lg font-bold dark:text-slate-100 uppercase tracking-tight">Status Modul</h3>
            <p class="text-[13px] leading-relaxed text-slate-600 dark:text-slate-400 italic">
              {{ moduleNotes }}
            </p>
          </div>

          <div class="grid gap-4 md:grid-cols-3">
            <div class="rounded-xl border border-slate-200 dark:border-slate-800 p-4 bg-white dark:bg-slate-900/50">
              <div class="mb-2 text-[10px] font-bold uppercase tracking-widest text-slate-400 dark:text-slate-500">Frontend</div>
              <p class="text-xs text-slate-600 dark:text-slate-400 leading-normal">Layout dan navigasi siap. View ini menunggu integrasi data spesifik.</p>
            </div>
            <div class="rounded-xl border border-slate-200 dark:border-slate-800 p-4 bg-white dark:bg-slate-900/50">
              <div class="mb-2 text-[10px] font-bold uppercase tracking-widest text-slate-400 dark:text-slate-500">Backend</div>
              <p class="text-xs text-slate-600 dark:text-slate-400 leading-normal">Master Data, Bank Soal, dan API Gateway Core sudah tersedia.</p>
            </div>
            <div class="rounded-xl border border-slate-200 dark:border-slate-800 p-4 bg-white dark:bg-slate-900/50">
              <div class="mb-2 text-[10px] font-bold uppercase tracking-widest text-slate-400 dark:text-slate-500">Sesi</div>
              <p class="text-xs text-slate-600 dark:text-slate-400 leading-normal">
                {{ authStore.isAuthenticated ? 'Auth aktif & valid.' : 'Login diperlukan untuk akses data.' }}
              </p>
            </div>
          </div>
        </div>
      </CardBox>
    </SectionMain>
  </LayoutAuthenticated>
</template>
