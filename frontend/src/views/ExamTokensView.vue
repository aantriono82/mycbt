<script setup>
import { computed, onMounted, reactive, ref, watch } from 'vue'
import {
  mdiDeleteOutline,
  mdiKeyVariant,
  mdiPlus,
  mdiRefresh,
  mdiToggleSwitch,
  mdiToggleSwitchOffOutline,
} from '@mdi/js'
import LayoutAuthenticated from '@/layouts/LayoutAuthenticated.vue'
import SectionMain from '@/components/SectionMain.vue'
import SectionTitleLineWithButton from '@/components/SectionTitleLineWithButton.vue'
import CardBox from '@/components/CardBox.vue'
import BaseButton from '@/components/BaseButton.vue'
import FormField from '@/components/FormField.vue'
import FormControl from '@/components/FormControl.vue'
import { api } from '@/services/api.js'
import { useAuthStore } from '@/stores/auth.js'

const authStore = useAuthStore()

const exams = ref([])
const tokens = ref([])
const selectedExamId = ref('')
const errorMessage = ref('')
const successMessage = ref('')
const isLoading = ref(false)
const isMutating = ref(false)

const form = reactive({
  valid_from_date: new Date().toISOString().split('T')[0],
  valid_from_hour: new Date().getHours(),
  valid_from_minute: Math.floor(new Date().getMinutes() / 5) * 5,
  valid_to_date: new Date(Date.now() + 2 * 60 * 60 * 1000).toISOString().split('T')[0],
  valid_to_hour: new Date(Date.now() + 2 * 60 * 60 * 1000).getHours(),
  valid_to_minute: Math.floor(new Date(Date.now() + 2 * 60 * 60 * 1000).getMinutes() / 5) * 5,
  timezone: 'WIB',
  length: 6,
})

const hourOptions = Array.from({ length: 24 }, (_, i) => ({ id: i, label: String(i).padStart(2, '0') }))
const minuteOptions = Array.from({ length: 12 }, (_, i) => ({ id: i * 5, label: String(i * 5).padStart(2, '0') }))
const tzOptions = [
  { id: 'WIB', label: 'WIB (UTC+7)' },
  { id: 'WITA', label: 'WITA (UTC+8)' },
  { id: 'WIT', label: 'WIT (UTC+9)' }
]

const formatDisplayDate = (dateStr) => {
  if (!dateStr) return '-'
  try {
    const d = new Date(dateStr)
    if (isNaN(d.getTime())) return dateStr
    
    const formatted = d.toLocaleString('id-ID', {
      day: '2-digit',
      month: '2-digit',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
      hour12: false
    }).replace(/\//g, '-')

    const offset = -d.getTimezoneOffset() / 60
    let tz = ''
    if (offset === 7) tz = 'WIB'
    else if (offset === 8) tz = 'WITA'
    else if (offset === 9) tz = 'WIT'
    else tz = offset >= 0 ? `GMT+${Math.floor(offset)}` : `GMT${Math.floor(offset)}`

    return `${formatted} ${tz}`
  } catch {
    return dateStr
  }
}

const canLoad = computed(() => authStore.isAuthenticated)

const loadExams = async () => {
  if (!canLoad.value) return
  try {
    const { data } = await api.get('/api/v1/exams', {
      params: { limit: 100, offset: 0 },
    })
    exams.value = data?.data || []
    if (!selectedExamId.value && exams.value.length) {
      selectedExamId.value = exams.value[0].id
    }
  } catch (error) {
    errorMessage.value = error?.response?.data?.error?.message || 'Gagal memuat ujian'
  }
}

const loadTokens = async () => {
  if (!canLoad.value || !selectedExamId.value) {
    tokens.value = []
    return
  }
  isLoading.value = true
  errorMessage.value = ''
  try {
    const { data } = await api.get(`/api/v1/exams/${selectedExamId.value}/tokens`, {
      params: { limit: 100, offset: 0 },
    })
    tokens.value = data?.data || []
  } catch (error) {
    errorMessage.value = error?.response?.data?.error?.message || 'Gagal memuat token'
  } finally {
    isLoading.value = false
  }
}

const createToken = async () => {
  if (!selectedExamId.value) return
  successMessage.value = ''
  errorMessage.value = ''
  
  try {
    const tzOffsetMap = { 'WIB': '+07:00', 'WITA': '+08:00', 'WIT': '+09:00' }
    const offset = tzOffsetMap[form.timezone]
    
    const vfStr = `${form.valid_from_date}T${String(form.valid_from_hour).padStart(2, '0')}:${String(form.valid_from_minute).padStart(2, '0')}:00${offset}`
    const vtStr = `${form.valid_to_date}T${String(form.valid_to_hour).padStart(2, '0')}:${String(form.valid_to_minute).padStart(2, '0')}:00${offset}`

    await api.post(`/api/v1/exams/${selectedExamId.value}/tokens`, {
      valid_from: new Date(vfStr).toISOString(),
      valid_to: new Date(vtStr).toISOString(),
      length: Number(form.length) || 6,
    })
    successMessage.value = 'Token berhasil dibuat'
    await loadTokens()
  } catch (error) {
    errorMessage.value = error?.response?.data?.error?.message || 'Gagal membuat token'
  }
}

const rotateToken = async () => {
  if (!selectedExamId.value) return
  successMessage.value = ''
  errorMessage.value = ''

  const ok = confirm(
    'Rotate token untuk ujian ini?\n\nToken baru akan dibuat, dan token lain akan dinonaktifkan (default).',
  )
  if (!ok) return

  isMutating.value = true
  try {
    const tzOffsetMap = { WIB: '+07:00', WITA: '+08:00', WIT: '+09:00' }
    const offset = tzOffsetMap[form.timezone] || '+07:00'

    const vfStr = `${form.valid_from_date}T${String(form.valid_from_hour).padStart(2, '0')}:${String(form.valid_from_minute).padStart(2, '0')}:00${offset}`
    const vtStr = `${form.valid_to_date}T${String(form.valid_to_hour).padStart(2, '0')}:${String(form.valid_to_minute).padStart(2, '0')}:00${offset}`

    await api.post(`/api/v1/exams/${selectedExamId.value}/tokens/rotate`, {
      valid_from: new Date(vfStr).toISOString(),
      valid_to: new Date(vtStr).toISOString(),
      length: Number(form.length) || 6,
      deactivate_others: true,
    })
    successMessage.value = 'Token berhasil di-rotate'
    await loadTokens()
  } catch (error) {
    errorMessage.value = error?.response?.data?.error?.message || 'Gagal rotate token'
  } finally {
    isMutating.value = false
  }
}

const deactivateAllTokens = async () => {
  if (!selectedExamId.value) return
  successMessage.value = ''
  errorMessage.value = ''

  const ok = confirm('Nonaktifkan semua token aktif untuk ujian ini?')
  if (!ok) return

  isMutating.value = true
  try {
    const { data } = await api.post(`/api/v1/exams/${selectedExamId.value}/tokens/deactivate-all`, {})
    const n = data?.data?.deactivated ?? 0
    successMessage.value = `Token aktif dinonaktifkan: ${n}`
    await loadTokens()
  } catch (error) {
    errorMessage.value = error?.response?.data?.error?.message || 'Gagal menonaktifkan semua token'
  } finally {
    isMutating.value = false
  }
}

const toggleToken = async (token) => {
  errorMessage.value = ''
  successMessage.value = ''
  try {
    await api.patch(`/api/v1/tokens/${token.id}`, {
      is_active: !token.is_active,
    })
    successMessage.value = `Token ${!token.is_active ? 'diaktifkan' : 'dinonaktifkan'}`
    await loadTokens()
  } catch (error) {
    errorMessage.value = error?.response?.data?.error?.message || 'Gagal mengubah status token'
  }
}

const deleteToken = async (token) => {
  if (!token?.id) return
  errorMessage.value = ''
  successMessage.value = ''

  const ok = confirm(`Hapus token ${token.token}?\n\nTindakan ini tidak bisa dibatalkan.`)
  if (!ok) return

  isMutating.value = true
  try {
    await api.delete(`/api/v1/tokens/${token.id}`)
    successMessage.value = 'Token berhasil dihapus'
    await loadTokens()
  } catch (error) {
    errorMessage.value = error?.response?.data?.error?.message || 'Gagal menghapus token'
  } finally {
    isMutating.value = false
  }
}

watch(selectedExamId, loadTokens)

onMounted(async () => {
  await loadExams()
  await loadTokens()
})
</script>

<template>
  <LayoutAuthenticated>
    <SectionMain>
      <SectionTitleLineWithButton :icon="mdiKeyVariant" title="Token Ujian" main>
        <BaseButton :icon="mdiRefresh" color="info" label="Muat Ulang" @click="loadExams(); loadTokens()" />
      </SectionTitleLineWithButton>

      <div class="mb-6 grid gap-6 xl:grid-cols-5">
        <CardBox class="xl:col-span-2" color="blue">
          <h3 class="mb-4 text-lg font-semibold dark:text-slate-100 uppercase tracking-tight">Generate Token Baru</h3>
          <div class="grid gap-4">
            <FormField label="Pilih Ujian">
              <FormControl
                v-model="selectedExamId"
                :options="exams.map((item) => ({ id: item.id, label: item.title }))"
              />
            </FormField>

            <FormField label="Berlaku Mulai">
              <div class="grid grid-cols-3 gap-2">
                <div class="col-span-1">
                   <FormControl v-model="form.valid_from_date" type="date" />
                </div>
                <div>
                   <FormControl v-model="form.valid_from_hour" :options="hourOptions" />
                </div>
                <div>
                   <FormControl v-model="form.valid_from_minute" :options="minuteOptions" />
                </div>
              </div>
            </FormField>

            <FormField label="Berlaku Hingga">
              <div class="grid grid-cols-3 gap-2">
                <div class="col-span-1">
                   <FormControl v-model="form.valid_to_date" type="date" />
                </div>
                <div>
                   <FormControl v-model="form.valid_to_hour" :options="hourOptions" />
                </div>
                <div>
                   <FormControl v-model="form.valid_to_minute" :options="minuteOptions" />
                </div>
              </div>
            </FormField>

            <div class="grid grid-cols-2 gap-4">
               <FormField label="Zona Waktu">
                <FormControl v-model="form.timezone" :options="tzOptions" />
              </FormField>
              <FormField label="Panjang Token">
                <FormControl v-model="form.length" inputmode="numeric" />
              </FormField>
            </div>

            <div class="mt-2 flex flex-wrap gap-2">
              <BaseButton :icon="mdiPlus" color="info" label="Generate Token" :disabled="isMutating" @click="createToken" />
              <BaseButton :icon="mdiRefresh" color="purple" label="Rotate Token" :disabled="isMutating" @click="rotateToken" />
              <BaseButton
                :icon="mdiToggleSwitchOffOutline"
                color="danger"
                label="Off Semua"
                :disabled="isMutating || !tokens.length"
                @click="deactivateAllTokens"
              />
            </div>
          </div>
        </CardBox>

        <CardBox class="xl:col-span-3" color="purple">
          <h3 class="mb-4 text-lg font-semibold dark:text-slate-100">Daftar Token</h3>

          <div v-if="!authStore.isAuthenticated" class="rounded-lg bg-amber-50 dark:bg-amber-900/20 px-4 py-3 text-sm text-amber-700 dark:text-amber-400 border border-amber-100 dark:border-amber-900/40">
            Login terlebih dulu agar token dapat dimuat dari backend.
          </div>
          <div v-else-if="errorMessage" class="rounded-lg bg-red-50 dark:bg-red-900/20 px-4 py-3 text-sm text-red-700 dark:text-red-400 border border-red-100 dark:border-red-900/40">
            {{ errorMessage }}
          </div>
          <div v-if="successMessage" class="mb-4 rounded-lg bg-emerald-50 dark:bg-emerald-900/20 px-4 py-3 text-sm text-emerald-700 dark:text-emerald-400 border border-emerald-100 dark:border-emerald-900/40">
            {{ successMessage }}
          </div>

          <div v-if="isLoading" class="text-sm text-slate-500 dark:text-slate-400 italic">Memuat token ujian...</div>
          <div v-else class="overflow-x-auto">
            <table class="w-full text-left text-sm">
              <thead class="border-b dark:border-slate-800 bg-slate-50 dark:bg-slate-800/50 text-slate-600 dark:text-slate-300 uppercase text-xs tracking-wider font-bold">
                <tr>
                  <th class="px-3 py-3">Token</th>
                  <th class="px-3 py-3 text-center">Aktif</th>
                  <th class="px-3 py-3">Valid From</th>
                  <th class="px-3 py-3">Valid To</th>
                  <th class="px-3 py-3 text-center">Aksi</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="token in tokens" :key="token.id" class="border-b dark:border-slate-800 last:border-b-0 hover:bg-slate-50/50 dark:hover:bg-slate-800/50 transition-colors">
                  <td class="px-3 py-3 font-mono font-bold text-info dark:text-sky-400 tracking-widest">{{ token.token }}</td>
                  <td class="px-3 py-3 text-center">
                    <span 
                      class="rounded-full px-2 py-0.5 text-[10px] font-bold uppercase"
                      :class="token.is_active ? 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400' : 'bg-slate-100 text-slate-600 dark:bg-slate-700 dark:text-slate-400'"
                    >
                      {{ token.is_active ? 'Ya' : 'Tidak' }}
                    </span>
                  </td>
                  <td class="px-3 py-3 text-xs text-slate-500 dark:text-slate-400 font-medium">{{ formatDisplayDate(token.valid_from) }}</td>
                  <td class="px-3 py-3 text-xs text-slate-500 dark:text-slate-400 font-medium">{{ formatDisplayDate(token.valid_to) }}</td>
                  <td class="px-3 py-3 text-center">
                    <div class="flex items-center justify-center gap-2">
                      <BaseButton
                        :icon="token.is_active ? mdiToggleSwitchOffOutline : mdiToggleSwitch"
                        color="info"
                        small
                        :disabled="isMutating"
                        :label="token.is_active ? 'Off' : 'On'"
                        @click="toggleToken(token)"
                      />
                      <BaseButton
                        :icon="mdiDeleteOutline"
                        color="danger"
                        small
                        :disabled="isMutating"
                        label="Hapus"
                        @click="deleteToken(token)"
                      />
                    </div>
                  </td>
                </tr>
                <tr v-if="!tokens.length && !isLoading">
                  <td colspan="5" class="px-3 py-10 text-center text-slate-400 dark:text-slate-500 italic">Belum ada token untuk ujian ini.</td>
                </tr>
              </tbody>
            </table>
          </div>
        </CardBox>
      </div>
    </SectionMain>
  </LayoutAuthenticated>
</template>
