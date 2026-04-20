<script setup>
import { onMounted, ref } from 'vue'
import { mdiBullhornOutline, mdiInformationOutline, mdiMessageOutline, mdiRefresh } from '@mdi/js'
import LayoutAuthenticated from '@/layouts/LayoutAuthenticated.vue'
import SectionMain from '@/components/SectionMain.vue'
import SectionTitleLineWithButton from '@/components/SectionTitleLineWithButton.vue'
import CardBox from '@/components/CardBox.vue'
import BaseButton from '@/components/BaseButton.vue'
import BaseIcon from '@/components/BaseIcon.vue'
import { api } from '@/services/api.js'

const isLoading = ref(false)
const errorMessage = ref('')
const announcements = ref([])

const formatDateTime = (value) => {
  if (!value) return '-'
  const parsed = new Date(value)
  if (Number.isNaN(parsed.getTime())) return value
  return parsed.toLocaleString('id-ID', {
    dateStyle: 'medium',
    timeStyle: 'short',
  })
}

const categoryConfig = (value) => {
  const v = String(value || '').toLowerCase()
  if (v === 'informasi') {
    return { label: 'Informasi', color: 'bg-emerald-100 text-emerald-700 dark:border-emerald-900/40 dark:bg-emerald-900/20 dark:text-emerald-400', icon: mdiInformationOutline }
  }
  if (v === 'pengumuman') {
    return { label: 'Pengumuman', color: 'bg-amber-100 text-amber-700 dark:border-amber-900/40 dark:bg-amber-900/20 dark:text-amber-400', icon: mdiBullhornOutline }
  }
  return { label: value || 'General', color: 'bg-sky-100 text-sky-700 dark:border-sky-900/40 dark:bg-sky-900/20 dark:text-sky-400', icon: mdiMessageOutline }
}

const loadAnnouncements = async () => {
  isLoading.value = true
  errorMessage.value = ''
  try {
    const { data } = await api.get('/api/v1/student/announcements', {
      params: { limit: 50, offset: 0 },
    })
    announcements.value = data?.data || []
  } catch (error) {
    announcements.value = []
    errorMessage.value = error?.response?.data?.error?.message || 'Gagal memuat pengumuman'
  } finally {
    isLoading.value = false
  }
}

onMounted(loadAnnouncements)
</script>

<template>
  <LayoutAuthenticated>
    <SectionMain>
      <SectionTitleLineWithButton :icon="mdiBullhornOutline" title="Pusat Informasi & Pengumuman" main>
        <BaseButton :icon="mdiRefresh" color="info" outline label="Segarkan" @click="loadAnnouncements" />
      </SectionTitleLineWithButton>

      <div class="grid gap-6">
        <div v-if="isLoading" class="flex flex-col items-center justify-center py-20 animate-pulse">
           <BaseIcon :path="mdiBullhornOutline" size="48" class="text-slate-200 dark:text-slate-800 mb-4" />
           <p class="text-sm text-slate-400 italic">Mencari informasi terbaru...</p>
        </div>

        <CardBox v-else-if="errorMessage" class="border-red-100 bg-red-50/50 dark:border-red-900/20 dark:bg-red-900/10">
          <p class="text-sm text-red-700 dark:text-red-400 font-bold text-center py-4">{{ errorMessage }}</p>
        </CardBox>

        <div v-else-if="!announcements.length" class="flex flex-col items-center justify-center py-20 border-2 border-dashed border-slate-100 dark:border-slate-800 rounded-3xl">
           <BaseIcon :path="mdiBullhornOutline" size="64" class="text-slate-100 dark:text-slate-900 mb-4" />
           <p class="text-slate-400 text-center px-6">Saat ini belum ada pengumuman aktif untuk Anda.<br><span class="text-xs">Cek kembali secara berkala untuk info terbaru.</span></p>
        </div>

        <div v-for="item in announcements" :key="item.id" class="group relative overflow-hidden rounded-3xl border border-slate-100 bg-white p-6 shadow-sm transition-all hover:shadow-md dark:border-slate-800 dark:bg-slate-950">
          <!-- Premium Background Decoration -->
          <div class="absolute -top-12 -right-12 h-32 w-32 rounded-full bg-slate-50 opacity-0 transition-opacity group-hover:opacity-100 dark:bg-slate-900" />
          
          <div class="relative z-10">
            <div class="mb-4 flex flex-wrap items-center justify-between gap-4">
              <div class="flex items-center gap-3">
                <div 
                  class="flex h-10 w-10 items-center justify-center rounded-2xl shadow-sm border border-transparent"
                  :class="categoryConfig(item.category).color"
                >
                  <BaseIcon :path="categoryConfig(item.category).icon" size="20" />
                </div>
                <div>
                  <div class="text-[10px] font-black uppercase tracking-[0.2em] text-slate-400 dark:text-slate-500">
                    {{ categoryConfig(item.category).label }}
                  </div>
                  <div class="text-[11px] font-bold text-slate-500 dark:text-slate-400">
                    {{ formatDateTime(item.published_at) }}
                  </div>
                </div>
              </div>
              
              <div v-if="item.expires_at" class="hidden sm:block rounded-full border border-slate-100 dark:border-slate-800 px-3 py-1 text-[10px] font-bold text-slate-400 dark:text-slate-500 uppercase tracking-tight">
                Berlaku s/d {{ formatDateTime(item.expires_at) }}
              </div>
            </div>

            <h3 class="mb-3 text-xl font-black italic tracking-tight text-slate-800 dark:text-slate-100">
              {{ item.title }}
            </h3>
            
            <div class="max-w-4xl border-l-2 border-slate-100 pl-4 dark:border-slate-800">
              <p class="text-sm leading-relaxed text-slate-600 dark:text-slate-400 whitespace-pre-wrap">
                {{ item.body }}
              </p>
            </div>
            
            <div v-if="item.expires_at" class="mt-4 sm:hidden text-[10px] font-bold text-slate-400 uppercase">
               Berlaku s/d {{ formatDateTime(item.expires_at) }}
            </div>
          </div>
        </div>
      </div>
    </SectionMain>
  </LayoutAuthenticated>
</template>
