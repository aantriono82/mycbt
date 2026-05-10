<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { mdiDomain, mdiRefresh, mdiPlus, mdiDelete, mdiPencil, mdiContentSave, mdiUpload } from '@mdi/js'
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

const schools = ref([])
const isLoading = ref(false)
const isSaving = ref(false)
const errorMessage = ref('')
const successMessage = ref('')
const editingId = ref('')

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
const isEditing = computed(() => !!editingId.value)

const resetForm = () => {
  editingId.value = ''
  form.name = ''
  form.logo_url = ''
  form.address = ''
  form.phone = ''
  form.email = ''
  form.website = ''
  form.principal_name = ''
}

const loadSchools = async () => {
  if (!canLoad.value) return
  isLoading.value = true
  errorMessage.value = ''
  try {
    const { data } = await api.get('/api/v1/admin/schools')
    schools.value = data?.data || []
  } catch (error) {
    errorMessage.value = error?.response?.data?.error?.message || 'Gagal memuat data sekolah'
  } finally {
    isLoading.value = false
  }
}

const startEdit = (school) => {
  editingId.value = school.id
  form.name = school.name || ''
  form.logo_url = school.logo_url || ''
  form.address = school.address || ''
  form.phone = school.phone || ''
  form.email = school.email || ''
  form.website = school.website || ''
  form.principal_name = school.principal_name || ''
}

const saveSchool = async () => {
  isSaving.value = true
  successMessage.value = ''
  errorMessage.value = ''
  try {
    if (isEditing.value) {
      await api.put(`/api/v1/admin/schools/${editingId.value}`, { ...form })
      successMessage.value = 'Data sekolah berhasil diperbarui'
    } else {
      await api.post('/api/v1/admin/schools', { ...form })
      successMessage.value = 'Data sekolah berhasil ditambahkan'
    }
    resetForm()
    await loadSchools()
  } catch (error) {
    errorMessage.value = error?.response?.data?.error?.message || 'Gagal menyimpan data sekolah'
  } finally {
    isSaving.value = false
  }
}

const deleteSchool = async (id) => {
  if (!confirm('Hapus sekolah ini? Pengguna yang terhubung mungkin akan kehilangan akses branding.')) return
  successMessage.value = ''
  errorMessage.value = ''
  try {
    await api.delete(`/api/v1/admin/schools/${id}`)
    successMessage.value = 'Data sekolah berhasil dihapus'
    if (editingId.value === id) resetForm()
    await loadSchools()
  } catch (error) {
    errorMessage.value = error?.response?.data?.error?.message || 'Gagal menghapus sekolah'
  }
}

const uploadLogo = async (event, schoolId) => {
  const file = event.target.files[0]
  if (!file) return

  const formData = new FormData()
  formData.append('file', file)

  try {
    const { data } = await api.post(`/api/v1/admin/schools/${schoolId}/logo`, formData, {
      headers: { 'Content-Type': 'multipart/form-data' }
    })
    if (editingId.value === schoolId) {
      form.logo_url = data.data.logo_url
    }
    await loadSchools()
    successMessage.value = 'Logo berhasil diunggah'
  } catch (e) {
    errorMessage.value = 'Gagal mengunggah logo'
  } finally {
    event.target.value = ''
  }
}

onMounted(async () => {
  await loadSchools()
})
</script>

<template>
  <LayoutAuthenticated>
    <SectionMain>
      <SectionTitleLineWithButton :icon="mdiDomain" title="Manajemen Sekolah" main>
        <BaseButton :icon="mdiRefresh" color="info" label="Muat Ulang" @click="loadSchools" small />
      </SectionTitleLineWithButton>

      <div class="grid gap-6 xl:grid-cols-5">
        <CardBox class="xl:col-span-2" color="blue">
          <h3 class="mb-4 text-lg font-semibold dark:text-slate-100">
            {{ isEditing ? 'Edit Sekolah' : 'Tambah Sekolah Baru' }}
          </h3>
          <div class="grid gap-4">
            <FormField label="Nama Sekolah">
              <FormControl v-model="form.name" placeholder="Misal: SMA Negeri 1 Jakarta" />
            </FormField>
            
            <FormField label="Nama Kepala Sekolah">
              <FormControl v-model="form.principal_name" placeholder="Nama Lengkap & Gelar" />
            </FormField>

            <FormField label="Alamat">
              <FormControl v-model="form.address" type="textarea" placeholder="Alamat lengkap sekolah" />
            </FormField>

            <div class="grid grid-cols-2 gap-4">
              <FormField label="Telepon">
                <FormControl v-model="form.phone" placeholder="021-xxxx" />
              </FormField>
              <FormField label="Email">
                <FormControl v-model="form.email" placeholder="info@sekolah.sch.id" />
              </FormField>
            </div>

            <FormField label="Website">
              <FormControl v-model="form.website" placeholder="https://www.sekolah.sch.id" />
            </FormField>

            <BaseButtons>
              <BaseButton
                :icon="isEditing ? mdiContentSave : mdiPlus"
                color="info"
                :label="isEditing ? 'Simpan Perubahan' : 'Tambah Sekolah'"
                :disabled="isSaving"
                @click="saveSchool"
              />
              <BaseButton
                v-if="isEditing"
                color="whiteDark"
                outline
                label="Batal"
                @click="resetForm"
              />
            </BaseButtons>
          </div>
        </CardBox>

        <CardBox class="xl:col-span-3" color="indigo">
          <div v-if="errorMessage" class="mb-4 rounded-lg bg-red-50 dark:bg-red-900/20 px-4 py-3 text-sm text-red-700 dark:text-red-400 border border-red-100 dark:border-red-900/40">
            {{ errorMessage }}
          </div>
          <div v-if="successMessage" class="mb-4 rounded-lg bg-emerald-50 dark:bg-emerald-900/20 px-4 py-3 text-sm text-emerald-700 dark:text-emerald-400 border border-emerald-100 dark:border-emerald-900/40">
            {{ successMessage }}
          </div>

          <div v-if="isLoading" class="py-12 text-center text-slate-500">Memuat data...</div>
          <div v-else class="overflow-x-auto">
            <table class="w-full text-left text-sm">
              <thead class="border-b dark:border-slate-800 bg-slate-50 dark:bg-slate-800/50 text-slate-600 dark:text-slate-300 uppercase text-xs tracking-wider">
                <tr>
                  <th class="px-3 py-3">Logo</th>
                  <th class="px-3 py-3">Sekolah</th>
                  <th class="px-3 py-3">Kontak</th>
                  <th class="px-3 py-3">Aksi</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="school in schools" :key="school.id" class="border-b dark:border-slate-800 last:border-b-0 hover:bg-slate-50 dark:hover:bg-slate-800/30 transition-colors">
                  <td class="px-3 py-3">
                    <div class="relative h-12 w-12 rounded-lg bg-slate-100 dark:bg-slate-800 border border-slate-200 dark:border-slate-700 overflow-hidden group">
                      <img v-if="school.logo_url" :src="school.logo_url" class="h-full w-full object-contain p-1" />
                      <div v-else class="flex h-full w-full items-center justify-center text-[10px] text-slate-400 text-center leading-tight">No Logo</div>
                      
                      <label class="absolute inset-0 flex items-center justify-center bg-black/40 opacity-0 group-hover:opacity-100 cursor-pointer transition-opacity">
                        <BaseIcon :path="mdiUpload" size="18" class="text-white" />
                        <input type="file" class="hidden" accept="image/*" @change="e => uploadLogo(e, school.id)" />
                      </label>
                    </div>
                  </td>
                  <td class="px-3 py-3">
                    <div class="font-bold dark:text-slate-200">{{ school.name }}</div>
                    <div class="text-xs text-slate-500 line-clamp-1">{{ school.address || '-' }}</div>
                  </td>
                  <td class="px-3 py-3">
                    <div class="text-xs text-slate-500">{{ school.phone || '-' }}</div>
                    <div class="text-xs text-slate-500">{{ school.email || '-' }}</div>
                  </td>
                  <td class="px-3 py-3">
                    <div class="flex items-center gap-2">
                      <BaseButton :icon="mdiPencil" color="info" @click="startEdit(school)" small outline />
                      <BaseButton :icon="mdiDelete" color="danger" @click="deleteSchool(school.id)" small outline />
                    </div>
                  </td>
                </tr>
                <tr v-if="!schools.length">
                  <td colspan="4" class="py-12 text-center text-slate-400 italic">Belum ada data sekolah.</td>
                </tr>
              </tbody>
            </table>
          </div>
        </CardBox>
      </div>
    </SectionMain>
  </LayoutAuthenticated>
</template>
