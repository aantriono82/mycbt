<script setup>
import { computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { 
  mdiHome, 
  mdiClipboardText, 
  mdiChartBox, 
  mdiBullhorn, 
  mdiBookOpenVariant, 
  mdiAccountCircle,
  mdiCog,
  mdiFolder
} from '@mdi/js'
import BaseIcon from '@/components/BaseIcon.vue'
import { useAuthStore } from '@/stores/auth.js'

const authStore = useAuthStore()
const router = useRouter()
const route = useRoute()

const menuItems = computed(() => {
  if (authStore.role === 'student') {
    return [
      { to: '/student/dashboard', icon: mdiHome, label: 'Home' },
      { to: '/student/ujian', icon: mdiClipboardText, label: 'Ujian' },
      { to: '/student/hasil', icon: mdiChartBox, label: 'Hasil' },
      { to: '/student/pengumuman', icon: mdiBullhorn, label: 'Info' },
    ]
  }
  if (authStore.role === 'teacher') {
    return [
      { to: '/teacher/dashboard', icon: mdiHome, label: 'Home' },
      { to: '/teacher/bank-soal', icon: mdiBookOpenVariant, label: 'Soal' },
      { to: '/teacher/ujian/jadwal', icon: mdiClipboardText, label: 'Jadwal' },
      { to: '/teacher/evaluasi', icon: mdiChartBox, label: 'Hasil' },
    ]
  }
  // Admin
  return [
    { to: '/admin/dashboard', icon: mdiHome, label: 'Home' },
    { to: '/admin/master-data/siswa', icon: mdiFolder, label: 'Master' },
    { to: '/admin/ujian/jadwal', icon: mdiClipboardText, label: 'Ujian' },
    { to: '/admin/settings', icon: mdiCog, label: 'Config' },
  ]
})

const isActive = (to) => route.path.startsWith(to)
</script>

<template>
  <nav class="lg:hidden fixed bottom-0 left-0 right-0 z-50 bg-white/60 backdrop-blur-xl border-t border-slate-200 pb-safe">
    <div class="flex items-center justify-around h-16 px-2">
      <router-link
        v-for="item in menuItems"
        :key="item.to"
        :to="item.to"
        class="flex flex-col items-center justify-center flex-1 transition-colors"
        :class="isActive(item.to) ? 'text-[#4f46e5]' : 'text-slate-500'"
      >
        <BaseIcon :path="item.icon" size="24" />
        <span class="text-[10px] font-bold mt-1 uppercase tracking-tighter">{{ item.label }}</span>
      </router-link>
      
      <!-- Profile / More -->
      <router-link
        :to="authStore.role === 'student' ? '/student/profil' : (authStore.role === 'teacher' ? '/teacher/profil' : '/profile')"
        class="flex flex-col items-center justify-center flex-1 text-slate-500"
        :class="isActive('/profil') || isActive('/profile') ? 'text-[#4f46e5]' : 'text-slate-500'"
      >
        <BaseIcon :path="mdiAccountCircle" size="24" />
        <span class="text-[10px] font-bold mt-1 uppercase tracking-tighter">Profil</span>
      </router-link>
    </div>
  </nav>
</template>

<style scoped>
.pb-safe {
  padding-bottom: env(safe-area-inset-bottom);
}
</style>
