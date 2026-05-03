<script setup>
import { onMounted, ref } from 'vue'
import { mdiBullhornOutline, mdiInformationOutline, mdiMessageOutline, mdiRefresh } from '@mdi/js'
import LayoutAuthenticated from '@/layouts/LayoutAuthenticated.vue'
import SectionMain from '@/components/SectionMain.vue'
import SectionTitleLineWithButton from '@/components/SectionTitleLineWithButton.vue'
import CardBox from '@/components/CardBox.vue'
import BaseButton from '@/components/BaseButton.vue'
import BaseIcon from '@/components/BaseIcon.vue'
import BaseSkeleton from '@/components/BaseSkeleton.vue'
import { api } from '@/services/api.js'

const announcements = ref([])
const isLoading = ref(false)
const errorMessage = ref('')

const renderAnnouncementHtml = (value) => {
  const html = String(value || '').trim()
  if (!html) return ''
  if (typeof window === 'undefined' || typeof DOMParser === 'undefined') return html

  const doc = new DOMParser().parseFromString(html, 'text/html')
  doc.querySelectorAll('script, style, iframe, object, embed').forEach((node) => node.remove())
  for (const el of doc.body.querySelectorAll('*')) {
    for (const attr of [...el.attributes]) {
      const name = attr.name.toLowerCase()
      const value = String(attr.value || '')
      if (name.startsWith('on')) {
        el.removeAttribute(attr.name)
        continue
      }
      if ((name === 'href' || name === 'src') && value.trim().toLowerCase().startsWith('javascript:')) {
        el.removeAttribute(attr.name)
      }
    }
  }
  return doc.body.innerHTML
}

const loadAnnouncements = async () => {
  isLoading.value = true
  errorMessage.value = ''
  try {
    const { data } = await api.get('/api/v1/student/announcements', {
      params: { limit: 50, offset: 0 },
    })
    announcements.value = data?.data || []
    const ids = announcements.value.map((item) => item?.id).filter(Boolean)
    if (ids.length) {
      try {
        await api.post('/api/v1/student/announcements/read', {
          announcement_ids: ids,
        })
      } catch (err) {
        console.error('Failed to mark announcements as read:', err)
      }
    }
  } catch (err) {
    errorMessage.value = err?.response?.data?.error?.message || 'Gagal memuat pengumuman'
  } finally {
    isLoading.value = false
  }
}

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

onMounted(() => {
  loadAnnouncements()
})
</script>

<template>
  <LayoutAuthenticated>
    <SectionMain>
      <SectionTitleLineWithButton :icon="mdiBullhornOutline" title="Pusat Informasi & Pengumuman" main>
        <BaseButton :icon="mdiRefresh" color="info" outline label="Segarkan" @click="loadAnnouncements" />
      </SectionTitleLineWithButton>

      <div class="grid gap-6">
        <template v-if="isLoading">
        <div v-for="i in 3" :key="i" class="rounded-3xl border border-blue-400/60 bg-white p-6 shadow-sm dark:border-blue-800/80 dark:bg-slate-950">
             <div class="mb-6 flex items-center gap-3">
               <BaseSkeleton width="w-10" height="h-10" rounded="rounded-2xl" />
               <div class="space-y-2">
                 <BaseSkeleton width="w-20" height="h-2" />
                 <BaseSkeleton width="w-32" height="h-3" />
               </div>
             </div>
             <BaseSkeleton width="w-3/4" height="h-8" class="mb-4" />
             <div class="space-y-3">
               <BaseSkeleton width="w-full" height="h-3" />
               <BaseSkeleton width="w-full" height="h-3" />
               <BaseSkeleton width="w-2/3" height="h-3" />
             </div>
          </div>
        </template>

        <CardBox v-else-if="errorMessage" class="border-red-100 bg-red-50/50 dark:border-red-900/20 dark:bg-red-900/10">
          <p class="text-sm text-red-700 dark:text-red-400 font-bold text-center py-4">{{ errorMessage }}</p>
        </CardBox>

        <div v-else-if="!announcements.length" class="flex flex-col items-center justify-center py-20 border border-emerald-400/60 dark:border-emerald-800/80 rounded-3xl bg-emerald-50/5 dark:bg-emerald-900/5">
           <BaseIcon :path="mdiBullhornOutline" size="64" class="text-emerald-100 dark:text-emerald-900 mb-4" />
           <p class="text-emerald-600 dark:text-emerald-400 font-medium text-center px-6">Saat ini belum ada pengumuman aktif untuk Anda.<br><span class="text-xs opacity-70">Cek kembali secara berkala untuk info terbaru.</span></p>
        </div>

        <div v-for="(item, idx) in announcements" :key="item.id" 
          class="group relative overflow-hidden rounded-3xl border bg-white p-6 shadow-sm transition-all hover:shadow-md dark:bg-slate-950"
          :class="[
            idx % 3 === 0 ? 'border-blue-400/60 dark:border-blue-800/80' : 
            idx % 3 === 1 ? 'border-purple-400/60 dark:border-purple-800/80' : 
            'border-emerald-400/60 dark:border-emerald-800/80'
          ]"
        >
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
            
            <div
              class="max-w-4xl border-l-2 border-slate-100 pl-4 text-sm leading-relaxed text-slate-600 dark:border-slate-800 dark:text-slate-400 [&_p]:mb-3 [&_p:last-child]:mb-0 [&_ul]:mb-3 [&_ul]:list-disc [&_ul]:pl-5 [&_ol]:mb-3 [&_ol]:list-decimal [&_ol]:pl-5 [&_a]:text-blue-600 [&_a]:underline dark:[&_a]:text-blue-400"
              v-html="renderAnnouncementHtml(item.body)"
            >
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
