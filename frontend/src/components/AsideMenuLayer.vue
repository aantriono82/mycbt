<script setup>
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { mdiClose } from '@mdi/js'
import { useRouter } from 'vue-router'
import AsideMenuList from '@/components/AsideMenuList.vue'
import BaseIcon from '@/components/BaseIcon.vue'
import { useAuthStore } from '@/stores/auth.js'
import { api } from '@/services/api.js'

const authStore = useAuthStore()
const router = useRouter()
const schoolIdentity = ref({
  school_name: '',
  logo_url: '',
})
const logoVersion = ref(Date.now())

defineProps({
  menu: {
    type: Array,
    required: true,
  },
  menuBottom: Array,
})

const emit = defineEmits(['menu-click', 'aside-lg-close-click'])

const menuClick = (event, item) => {
  emit('menu-click', event, item)
}

const asideLgCloseClick = (event) => {
  emit('aside-lg-close-click', event)
}

const schoolBrandName = computed(() => schoolIdentity.value.school_name?.trim() || 'ATIGACBT')
const extractLogoUrl = (payload = {}) => {
  const candidates = [
    payload.logo_url,
    payload.logoUrl,
    payload.url,
    payload.file_url,
    payload.path,
    payload.location,
  ]
  const found = candidates.find((item) => String(item || '').trim())
  return found ? String(found).trim() : ''
}
const pickIdentityPayload = (raw = {}) => {
  if (!raw || typeof raw !== 'object') return {}
  if (raw.data && typeof raw.data === 'object') return raw.data
  if (raw.school_identity && typeof raw.school_identity === 'object') return raw.school_identity
  if (raw.schoolIdentity && typeof raw.schoolIdentity === 'object') return raw.schoolIdentity
  return raw
}
const normalizeSchoolIdentity = (raw = {}) => {
  const identity = pickIdentityPayload(raw)
  return {
    ...identity,
    school_name: String(identity.school_name || identity.name || '').trim(),
    logo_url: extractLogoUrl(identity),
  }
}
const apiOrigin = computed(() => {
  const baseURL = String(api.defaults.baseURL || '').trim()
  if (!baseURL) return ''
  try {
    return new URL(baseURL, window.location.origin).origin
  } catch {
    return ''
  }
})
const schoolBrandLogo = computed(() => {
  const logo = extractLogoUrl(normalizeSchoolIdentity(schoolIdentity.value))
  if (!logo) return '/logo_atiga.png'
  if (/^(data:|blob:|https?:\/\/|\/\/)/i.test(logo)) return logo
  const withVersion = (value) => `${value}${String(value).includes('?') ? '&' : '?'}v=${logoVersion.value}`
  if (apiOrigin.value) {
    if (logo.startsWith('/')) return withVersion(`${apiOrigin.value}${logo}`)
    return withVersion(`${apiOrigin.value}/${logo.replace(/^\/+/, '')}`)
  }
  if (logo.startsWith('/')) return withVersion(logo)
  return withVersion(`/${logo.replace(/^\/+/, '')}`)
})
const schoolBrandTagline = computed(() => (schoolIdentity.value.school_name?.trim() ? 'School CBT' : 'Professional CBT'))

const loadSchoolIdentity = async () => {
  if (!authStore.isAuthenticated) return
  try {
    const { data } = await api.get('/api/v1/settings/school-identity')
    const normalized = normalizeSchoolIdentity(data || {})
    schoolIdentity.value = {
      ...schoolIdentity.value,
      ...normalized,
    }
    logoVersion.value = Date.now()
  } catch {
    // Fallback ke brand default jika identity tidak tersedia.
  }
}

const handleSchoolIdentityUpdated = (event) => {
  const detail = event?.detail
  if (detail && typeof detail === 'object') {
    const normalized = normalizeSchoolIdentity(detail)
    schoolIdentity.value = {
      ...schoolIdentity.value,
      ...normalized,
    }
    logoVersion.value = Date.now()
    return
  }
  loadSchoolIdentity()
}

onMounted(() => {
  loadSchoolIdentity()
  if (typeof window !== 'undefined') {
    window.addEventListener('school-identity-updated', handleSchoolIdentityUpdated)
  }
})

onBeforeUnmount(() => {
  if (typeof window !== 'undefined') {
    window.removeEventListener('school-identity-updated', handleSchoolIdentityUpdated)
  }
})

watch(
  () => authStore.isAuthenticated,
  (isAuthenticated) => {
    if (isAuthenticated) loadSchoolIdentity()
  },
)
</script>

<template>
  <aside
    id="aside"
    class="fixed top-0 z-40 flex h-screen w-60 overflow-hidden transition-(--transition-position) lg:py-2 lg:pl-2"
  >
    <div class="aside flex flex-1 flex-col overflow-hidden lg:rounded-[2rem] bg-white/70 dark:bg-[#13082a]/70 backdrop-blur-xl border-r lg:border border-slate-200 dark:border-blue-900/30 shadow-2xl shadow-blue-300/40 dark:shadow-blue-900/30 mb-2">
      <div class="aside-brand flex h-20 flex-row items-center justify-between px-6">
        <div class="flex items-center cursor-pointer" @click="router.push('/dashboard')">
          <img :src="schoolBrandLogo" class="mr-3 h-10 w-10 rounded-full shadow-lg hover:scale-110 transition-transform duration-300 brightness-110 object-contain bg-white" alt="School Logo" />
          <div class="flex flex-col leading-tight">
            <b class="font-black uppercase tracking-tight text-sm bg-clip-text text-transparent bg-gradient-to-br from-blue-600 via-indigo-500 to-violet-600 dark:from-blue-400 dark:via-indigo-400 dark:to-violet-400 truncate max-w-[145px]">{{ schoolBrandName }}</b>
            <span class="text-[9px] font-bold text-slate-400 dark:text-slate-500 tracking-widest uppercase truncate max-w-[145px]">{{ schoolBrandTagline }}</span>
          </div>
        </div>
        <button class="hidden p-3 lg:inline-block xl:hidden text-slate-400 hover:text-slate-700 dark:text-white/70 dark:hover:text-white" @click.prevent="asideLgCloseClick">
          <BaseIcon :path="mdiClose" />
        </button>
      </div>
      <div
        class="aside-scrollbar flex-1 overflow-x-hidden overflow-y-auto scrollbar-styled-light dark:scrollbar-styled-dark px-3 mt-2"
      >
        <AsideMenuList :menu="menu" @menu-click="menuClick" />
      </div>

      <div class="px-3 pb-6">
        <AsideMenuList v-if="menuBottom" :menu="menuBottom" @menu-click="menuClick" />
      </div>
    </div>
  </aside>
</template>
