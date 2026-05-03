<script setup>
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { mdiRefresh, mdiPlus, mdiPencil, mdiDelete, mdiContentCopy } from '@mdi/js'
import LayoutAuthenticated from '@/layouts/LayoutAuthenticated.vue'
import SectionMain from '@/components/SectionMain.vue'
import SectionTitleLineWithButton from '@/components/SectionTitleLineWithButton.vue'
import CardBox from '@/components/CardBox.vue'
import BaseButton from '@/components/BaseButton.vue'
import BaseButtons from '@/components/BaseButtons.vue'
import FormField from '@/components/FormField.vue'
import FormControl from '@/components/FormControl.vue'
import { api } from '@/services/api.js'
import { useAuthStore } from '@/stores/auth.js'
import { shortCode2 } from '@/utils/shortCode.js'

const route = useRoute()
const authStore = useAuthStore()

const config = computed(() => route.meta?.resourceConfig || {})
const title = computed(() => route.meta?.title || 'Master Data')
const icon = computed(() => route.meta?.icon || null)
const endpoint = computed(() => config.value.endpoint || '')
const fields = computed(() => config.value.fields || [])
const itemLabel = computed(() => config.value.itemLabel || title.value)

const items = ref([])
const isLoading = ref(false)
const errorMessage = ref('')
const successMessage = ref('')
const editingId = ref('')

const form = reactive({})

const resetForm = () => {
  editingId.value = ''
  for (const field of fields.value) {
    form[field.key] = field.type === 'select' ? null : ''
  }
}

const loadItems = async () => {
  if (!authStore.isAuthenticated || !endpoint.value) return
  isLoading.value = true
  errorMessage.value = ''
  try {
    const { data } = await api.get(endpoint.value)
    items.value = data?.data || []
  } catch (error) {
    errorMessage.value = error?.response?.data?.error?.message || `Gagal memuat ${itemLabel.value}`
  } finally {
    isLoading.value = false
  }
}

const submitForm = async () => {
  successMessage.value = ''
  errorMessage.value = ''
  const payload = {}
  for (const field of fields.value) {
    if (field.type === 'select') {
      payload[field.key] = form[field.key] !== '' && form[field.key] != null ? form[field.key] : null
    } else {
      payload[field.key] = form[field.key]
    }
  }

  try {
    if (editingId.value) {
      await api.patch(`${endpoint.value}/${editingId.value}`, payload)
      successMessage.value = `${itemLabel.value} berhasil diperbarui`
    } else {
      await api.post(endpoint.value, payload)
      successMessage.value = `${itemLabel.value} berhasil ditambahkan`
    }
    resetForm()
    await loadItems()
  } catch (error) {
    errorMessage.value =
      error?.response?.data?.error?.message || `Gagal menyimpan ${itemLabel.value}`
  }
}

const editItem = (item) => {
  editingId.value = item.id
  for (const field of fields.value) {
    const val = item[field.key]
    if (field.type === 'select') {
      form[field.key] = val != null ? val : null
    } else {
      form[field.key] = val || ''
    }
  }
}

const deleteItem = async (item) => {
  errorMessage.value = ''
  successMessage.value = ''
  try {
    await api.delete(`${endpoint.value}/${item.id}`)
    successMessage.value = `${itemLabel.value} berhasil dihapus`
    if (editingId.value === item.id) {
      resetForm()
    }
    await loadItems()
  } catch (error) {
    errorMessage.value =
      error?.response?.data?.error?.message || `Gagal menghapus ${itemLabel.value}`
  }
}

const copyId = (id) => {
  navigator.clipboard.writeText(id)
  alert('ID disalin ke clipboard')
}

const shortId = (id) => shortCode2(id)

onMounted(() => {
  resetForm()
  loadItems()
})

watch(() => route.path, () => {
  resetForm()
  loadItems()
})
</script>

<template>
  <LayoutAuthenticated>
    <SectionMain>
      <SectionTitleLineWithButton :icon="icon" :title="title" main>
        <BaseButton :icon="mdiRefresh" color="info" label="Muat Ulang" @click="loadItems" />
      </SectionTitleLineWithButton>

      <div class="grid gap-6 xl:grid-cols-5">
        <CardBox class="xl:col-span-2" color="blue">
          <h3 class="mb-4 text-lg font-semibold dark:text-slate-100">
            {{ editingId ? `Edit ${itemLabel}` : `Tambah ${itemLabel}` }}
          </h3>
          <div class="grid gap-4">
            <template v-for="field in fields" :key="field.key">
              <FormField :label="field.label">
                <FormControl
                  v-if="field.type === 'select'"
                  v-model="form[field.key]"
                  :options="[{ value: null, label: field.placeholder || `Pilih ${field.label}` }, ...field.options]"
                />
                <FormControl
                  v-else
                  v-model="form[field.key]"
                  :placeholder="field.placeholder || field.label"
                />
              </FormField>
            </template>
            <BaseButtons>
              <BaseButton
                :icon="editingId ? mdiPencil : mdiPlus"
                color="info"
                :label="editingId ? 'Simpan Perubahan' : 'Tambah Data'"
                @click="submitForm"
              />
              <BaseButton color="whiteDark" outline label="Reset" @click="resetForm" />
            </BaseButtons>
          </div>
        </CardBox>

        <CardBox class="xl:col-span-3" color="emerald">
          <h3 class="mb-4 text-lg font-semibold dark:text-slate-100">Daftar {{ title }}</h3>

          <div v-if="!authStore.isAuthenticated" class="rounded-lg bg-amber-50 dark:bg-amber-900/20 px-4 py-3 text-sm text-amber-700 dark:text-amber-400 border border-amber-100 dark:border-amber-900/40">
            Login terlebih dulu agar data dapat dimuat dari backend.
          </div>
          <div v-else-if="errorMessage" class="rounded-lg bg-red-50 dark:bg-red-900/20 px-4 py-3 text-sm text-red-700 dark:text-red-400 border border-red-100 dark:border-red-900/40">
            {{ errorMessage }}
          </div>
          <div v-if="successMessage" class="mb-4 rounded-lg bg-emerald-50 dark:bg-emerald-900/20 px-4 py-3 text-sm text-emerald-700 dark:text-emerald-400 border border-emerald-100 dark:border-emerald-900/40">
            {{ successMessage }}
          </div>

          <div v-if="isLoading" class="text-sm text-slate-500 dark:text-slate-400 italic">Memuat data...</div>
          <div v-else class="overflow-x-auto">
            <table class="w-full text-left text-sm">
              <thead class="border-b dark:border-slate-800 bg-slate-50 dark:bg-slate-800/50 text-slate-600 dark:text-slate-300 uppercase text-xs tracking-wider">
                <tr>
                  <th
                    v-for="field in fields"
                    :key="field.key"
                    class="px-3 py-3 font-bold"
                  >
                    {{ field.label }}
                  </th>
                  <th class="px-3 py-3 font-bold">Kode ID</th>
                  <th class="px-3 py-3 font-bold text-center">Aksi</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="item in items" :key="item.id" class="border-b dark:border-slate-800 last:border-b-0 hover:bg-slate-50 dark:hover:bg-slate-800/20 transition-colors">
                  <td v-for="field in fields" :key="field.key" class="px-3 py-3 dark:text-slate-300">
                    <template v-if="field.type === 'select'">
                      {{ field.options?.find(o => o.value == item[field.key])?.label || (item[field.key] != null ? item[field.key] : '-') }}
                    </template>
                    <template v-else>
                      {{ item[field.key] || '-' }}
                    </template>
                  </td>
                  <td class="px-3 py-3 font-mono text-[10px] text-slate-400">
                    <div class="flex items-center gap-2">
                      <span class="truncate w-20 font-black">{{ shortId(item.id) }}</span>
                      <BaseButton :icon="mdiContentCopy" small color="whiteDark" outline @click="copyId(item.id)" title="Salin ID" />
                    </div>
                  </td>
                  <td class="px-3 py-3 text-center">
                    <BaseButtons>
                      <BaseButton :icon="mdiPencil" color="info" small label="Edit" @click="editItem(item)" />
                      <BaseButton :icon="mdiDelete" color="danger" small label="Hapus" @click="deleteItem(item)" />
                    </BaseButtons>
                  </td>
                </tr>
                <tr v-if="!items.length">
                  <td :colspan="fields.length + 2" class="px-3 py-10 text-center text-slate-400 dark:text-slate-500 italic">
                    Belum ada data.
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </CardBox>
      </div>
    </SectionMain>
  </LayoutAuthenticated>
</template>
