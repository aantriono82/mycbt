<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { mdiDomain, mdiRefresh, mdiContentSave, mdiUpload } from '@mdi/js'
import LayoutAuthenticated from '@/layouts/LayoutAuthenticated.vue'
import SectionMain from '@/components/SectionMain.vue'
import SectionTitleLineWithButton from '@/components/SectionTitleLineWithButton.vue'
import CardBox from '@/components/CardBox.vue'
import BaseButton from '@/components/BaseButton.vue'
import BaseButtons from '@/components/BaseButtons.vue'
import BaseIcon from '@/components/BaseIcon.vue'
import FormField from '@/components/FormField.vue'
import FormControl from '@/components/FormControl.vue'
import { api } from '@/services/api.js'
import { useAuthStore } from '@/stores/auth.js'

const authStore = useAuthStore()

const isLoading = ref(false)
const isSaving = ref(false)
const errorMessage = ref('')
const successMessage = ref('')

const form = reactive({
  name: '',
  logo_url: '',
  address: '',
  phone: '',
  email: '',
  website: '',
  principal_name: '',
})

const canLoad = computed(() => authStore.isAuthenticated)

const loadSchoolData = async () => {
  if (!canLoad.value) return
  isLoading.value = true
  errorMessage.value = ''
  try {
    const { data } = await api.get('/api/v1/teacher/school')
    const school = data?.data
    if (school) {
      form.name = school.name || ''
      form.logo_url = school.logo_url || ''
      form.address = school.address || ''
      form.phone = school.phone || ''
      form.email = school.email || ''
      form.website = school.website || ''
      form.principal_name = school.principal_name || ''
    }
  } catch (error) {
    errorMessage.value = error?.response?.data?.error?.message || 'Gagal memuat data sekolah. Pastikan akun Anda sudah terasosiasi dengan sekolah.'
  } finally {
    isLoading.value = false
  }
}

const saveSchool = async () => {
  isSaving.value = true
  successMessage.value = ''
  errorMessage.value = ''
  try {
    await api.put('/api/v1/teacher/school', { ...form })
    successMessage.value = 'Profil sekolah berhasil diperbarui'
    await loadSchoolData()
  } catch (error) {
    errorMessage.value = error?.response?.data?.error?.message || 'Gagal menyimpan data sekolah'
  } finally {
    isSaving.value = false
  }
}

const uploadLogo = async (event) => {
  const file = event.target.files[0]
  if (!file) return

  const formData = new FormData()
  formData.append('file', file)

  try {
    const { data } = await api.post('/api/v1/teacher/school/logo', formData, {
      headers: { 'Content-Type': 'multipart/form-data' }
    })
    form.logo_url = data.data.logo_url
    successMessage.value = 'Logo sekolah berhasil diperbarui'
  } catch (e) {
    errorMessage.value = 'Gagal mengunggah logo'
  } finally {
    event.target.value = ''
  }
}

onMounted(async () => {
  await loadSchoolData()
})
</script>

<template>
  <LayoutAuthenticated>
    <SectionMain>
      <SectionTitleLineWithButton :icon="mdiDomain" title="Profil Sekolah" main>
        <BaseButton :icon="mdiRefresh" color="info" label="Segarkan" @click="loadSchoolData" small />
      </SectionTitleLineWithButton>

      <div class="grid gap-6">
        <CardBox color="blue" is-form @submit.prevent="saveSchool">
          <div v-if="errorMessage" class="mb-6 rounded-lg bg-red-50 dark:bg-red-900/20 px-4 py-3 text-sm text-red-700 dark:text-red-400 border border-red-100 dark:border-red-900/40">
            {{ errorMessage }}
          </div>
          <div v-if="successMessage" class="mb-6 rounded-lg bg-emerald-50 dark:bg-emerald-900/20 px-4 py-3 text-sm text-emerald-700 dark:text-emerald-400 border border-emerald-100 dark:border-emerald-900/40">
            {{ successMessage }}
          </div>

          <div class="flex flex-col md:flex-row gap-8">
            <!-- Logo Section -->
            <div class="flex flex-col items-center gap-4 md:w-1/3">
              <div class="group relative h-48 w-48 rounded-2xl bg-slate-50 dark:bg-slate-800 border-2 border-dashed border-slate-200 dark:border-slate-700 overflow-hidden flex items-center justify-center p-4">
                <img v-if="form.logo_url" :src="form.logo_url" class="max-h-full max-w-full object-contain drop-shadow-sm" />
                <div v-else class="text-center">
                  <BaseIcon :path="mdiDomain" size="48" class="text-slate-300 mb-2" />
                  <p class="text-xs text-slate-400">Belum ada logo</p>
                </div>
                
                <label class="absolute inset-0 flex items-center justify-center bg-black/50 opacity-0 group-hover:opacity-100 cursor-pointer transition-opacity duration-300 backdrop-blur-[1px]">
                  <div class="flex flex-col items-center gap-2">
                    <BaseIcon :path="mdiUpload" size="32" class="text-white" />
                    <span class="text-xs font-bold text-white uppercase tracking-wider">Ganti Logo</span>
                  </div>
                  <input type="file" class="hidden" accept="image/*" @change="uploadLogo" />
                </label>
              </div>
              <p class="text-[10px] text-slate-400 text-center px-4 italic leading-relaxed">
                Gunakan file PNG transparan untuk hasil terbaik di semua tema.
              </p>
            </div>

            <!-- Info Section -->
            <div class="flex-1 space-y-4">
              <FormField label="Nama Sekolah">
                <FormControl v-model="form.name" placeholder="Misal: SMA Negeri 1 Jakarta" icon="mdiDomain" />
              </FormField>

              <FormField label="Nama Kepala Sekolah">
                <FormControl v-model="form.principal_name" placeholder="Nama Lengkap & Gelar" />
              </FormField>

              <FormField label="Alamat Lengkap">
                <FormControl v-model="form.address" type="textarea" placeholder="Alamat lengkap sekolah" />
              </FormField>

              <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                <FormField label="Nomor Telepon">
                  <FormControl v-model="form.phone" placeholder="Contoh: 021-xxxx" />
                </FormField>
                <FormField label="Email Sekolah">
                  <FormControl v-model="form.email" placeholder="info@sekolah.sch.id" />
                </FormField>
              </div>

              <FormField label="Website Resmi">
                <FormControl v-model="form.website" placeholder="https://www.sekolah.sch.id" />
              </FormField>

              <BaseButtons class="pt-4">
                <BaseButton
                  type="submit"
                  :icon="mdiContentSave"
                  color="info"
                  label="Simpan Profil Sekolah"
                  :disabled="isSaving || isLoading"
                  :loading="isSaving"
                />
              </BaseButtons>
            </div>
          </div>
        </CardBox>
      </div>
    </SectionMain>
  </LayoutAuthenticated>
</template>
