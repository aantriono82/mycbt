<script setup>
import { computed, ref, watch } from 'vue'
import { mdiAccountCheckOutline, mdiRefresh, mdiCheck, mdiClose } from '@mdi/js'
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

const registrations = ref([])
const meta = ref({ total: 0 })
const isLoading = ref(false)
const errorMessage = ref('')
const successMessage = ref('')
const query = ref('')
const status = ref('pending')
const role = ref('')
const actionLoadingId = ref('')
const isBulkApproving = ref(false)

const canLoad = computed(() => authStore.isAuthenticated)
const canBulkApprove = computed(() => status.value === 'pending' && Number(meta.value?.total || registrations.value.length || 0) > 0)

const loadRegistrations = async () => {
  if (!canLoad.value) return
  isLoading.value = true
  errorMessage.value = ''
  try {
    const { data } = await api.get('/api/v1/admin/registrations', {
      params: {
        q: query.value,
        status: status.value,
        role: role.value,
        limit: 50,
        offset: 0,
      },
    })
    registrations.value = data?.data || []
    meta.value = data?.meta || { total: registrations.value.length }
  } catch (error) {
    errorMessage.value = error?.response?.data?.error?.message || 'Gagal memuat pendaftaran'
  } finally {
    isLoading.value = false
  }
}

const decide = async (id, nextStatus) => {
  actionLoadingId.value = id
  successMessage.value = ''
  errorMessage.value = ''
  try {
    const note =
      nextStatus === 'approve'
        ? 'Disetujui dari dashboard admin'
        : nextStatus === 'reject'
          ? 'Ditolak dari dashboard admin'
          : ''
    await api.post(`/api/v1/admin/registrations/${id}/${nextStatus}`, {
      note,
    })
    successMessage.value =
      nextStatus === 'approve'
        ? 'Pendaftaran berhasil disetujui'
        : nextStatus === 'reject'
          ? 'Pendaftaran berhasil ditolak'
          : 'Pendaftaran berhasil diset pending'
    await loadRegistrations()
  } catch (error) {
    errorMessage.value = error?.response?.data?.error?.message || 'Gagal memproses pendaftaran'
  } finally {
    actionLoadingId.value = ''
  }
}

const bulkApproveByFilter = async () => {
  if (!canBulkApprove.value) return

  const matchedTotal = Number(meta.value?.total || registrations.value.length || 0)
  const processLimit = Math.min(matchedTotal, 500)
  const roleLabel =
    role.value === 'student' ? ' siswa' : role.value === 'teacher' ? ' guru' : ''

  const ok = confirm(
    `Approve ${processLimit} pendaftaran pending${roleLabel} sesuai filter aktif?\n\nJika total melebihi 500, sistem akan memproses 500 data per klik.`,
  )
  if (!ok) return

  isBulkApproving.value = true
  successMessage.value = ''
  errorMessage.value = ''
  try {
    const { data } = await api.post('/api/v1/admin/registrations/approve-bulk', {
      role: role.value,
      q: query.value,
      limit: processLimit,
      note: 'Disetujui massal dari dashboard admin sesuai filter',
    })

    const result = data?.data || {}
    const approved = Number(result.approved || 0)
    const failed = Number(result.failed || 0)
    const remaining = Number(result.remaining || 0)
    let failureMessage = ''
    successMessage.value = `Bulk approve selesai. Approved: ${approved}, gagal: ${failed}, sisa pending: ${remaining}.`

    const failureDetails = Array.isArray(result.failure_details) ? result.failure_details : []
    if (failed > 0 && failureDetails.length) {
      failureMessage = `Sebagian data gagal diproses: ${failureDetails
        .slice(0, 5)
        .map((item) => `${item.username || item.name}: ${item.message}`)
        .join(' | ')}`
    }

    await loadRegistrations()
    errorMessage.value = failureMessage
  } catch (error) {
    errorMessage.value = error?.response?.data?.error?.message || 'Gagal memproses bulk approve'
  } finally {
    isBulkApproving.value = false
  }
}

watch(
  () => authStore.isAuthenticated,
  (isAuthenticated) => {
    if (isAuthenticated) {
      loadRegistrations()
    }
  },
  { immediate: true },
)
</script>

<template>
  <LayoutAuthenticated>
    <SectionMain>
      <SectionTitleLineWithButton
        :icon="mdiAccountCheckOutline"
        title="Verifikasi Pendaftaran"
        main
      >
        <BaseButton :icon="mdiRefresh" color="purple" label="Muat Ulang" @click="loadRegistrations" />
      </SectionTitleLineWithButton>

      <CardBox>
        <div class="mb-6 flex flex-wrap items-end gap-4 overflow-x-auto pb-2">
          <!-- 1: Cari -->
          <div class="w-full md:w-64 flex flex-col gap-1.5">
            <label class="text-sm font-bold text-slate-700 dark:text-slate-300 px-1">Cari</label>
            <FormControl v-model="query" placeholder="Nama, username, email, nis, nip" class="h-12" />
          </div>
          
          <!-- 2: Status -->
          <div class="w-full md:w-40 flex flex-col gap-1.5">
            <label class="text-sm font-bold text-slate-700 dark:text-slate-300 px-1">Status</label>
            <FormControl
              v-model="status"
              :options="[
                { id: 'pending', label: 'Pending' },
                { id: 'approved', label: 'Approved' },
                { id: 'rejected', label: 'Rejected' },
                { id: '', label: 'Semua Status' },
              ]"
              class="h-12"
            />
          </div>

          <!-- 3: Role -->
          <div class="w-full md:w-40 flex flex-col gap-1.5">
            <label class="text-sm font-bold text-slate-700 dark:text-slate-300 px-1">Role</label>
            <FormControl
              v-model="role"
              :options="[
                { id: '', label: 'Semua Role' },
                { id: 'student', label: 'Student' },
                { id: 'teacher', label: 'Teacher' },
              ]"
              class="h-12"
            />
          </div>

          <!-- 4: Terapkan -->
          <div class="flex flex-col gap-1.5">
            <label class="text-sm font-bold text-slate-700 dark:text-slate-300 px-1">Aksi</label>
            <BaseButton
              color="purple"
              label="Terapkan Filter"
              class="h-12 px-6"
              @click="loadRegistrations"
            />
          </div>

          <!-- 5: Bulk Approve -->
          <div v-if="status === 'pending'" class="flex flex-col gap-1.5">
            <label class="text-sm font-bold text-slate-700 dark:text-slate-300 px-1">Aksi</label>
            <BaseButton
              :icon="mdiCheck"
              color="success"
              label="Approve Semua"
              class="h-12 px-6"
              :disabled="isBulkApproving || !canBulkApprove"
              @click="bulkApproveByFilter"
            />
          </div>
        </div>

        <div
          v-if="!authStore.isAuthenticated"
          class="rounded-lg border border-amber-100 bg-amber-50 px-4 py-3 text-sm text-amber-700 dark:border-amber-900/40 dark:bg-amber-900/20 dark:text-amber-400"
        >
          Login terlebih dulu agar daftar verifikasi dapat dimuat dari backend.
        </div>
        <div
          v-else-if="errorMessage"
          class="rounded-lg border border-red-100 bg-red-50 px-4 py-3 text-sm text-red-700 dark:border-red-900/40 dark:bg-red-900/20 dark:text-red-400"
        >
          {{ errorMessage }}
        </div>
        <div
          v-if="successMessage"
          class="mb-4 rounded-lg border border-emerald-100 bg-emerald-50 px-4 py-3 text-sm text-emerald-700 dark:border-emerald-900/40 dark:bg-emerald-900/20 dark:text-emerald-400"
        >
          {{ successMessage }}
        </div>

        <div class="mb-4 text-sm text-slate-500 dark:text-slate-400">
          Total data: {{ meta.total || registrations.length }}
        </div>
        <div
          v-if="status === 'pending'"
          class="mb-4 rounded-lg border border-blue-100 bg-blue-50 px-4 py-3 text-xs text-blue-700 dark:border-blue-900/40 dark:bg-blue-900/20 dark:text-blue-400"
        >
          Bulk approve mengikuti filter aktif saat ini. Gunakan filter `Role` untuk memproses hanya siswa, hanya guru, atau keduanya sekaligus.
        </div>

        <div v-if="isLoading" class="text-sm text-slate-500 italic dark:text-slate-400">
          Memuat data verifikasi...
        </div>
        <div v-else class="overflow-x-auto">
          <table class="w-full text-left text-sm">
            <thead
              class="border-b bg-slate-50 text-xs tracking-wider text-slate-600 uppercase dark:border-slate-800 dark:bg-slate-800/50 dark:text-slate-300"
            >
              <tr>
                <th class="px-3 py-3">Akun</th>
                <th class="px-3 py-3">Role</th>
                <th class="px-3 py-3">Identitas</th>
                <th class="px-3 py-3 text-center">Status</th>
                <th class="px-3 py-3">Catatan</th>
                <th class="px-3 py-3">Aksi</th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="item in registrations"
                :key="item.id"
                class="border-b align-top transition-colors last:border-b-0 hover:bg-slate-50 dark:border-slate-800 dark:hover:bg-slate-800/20"
              >
                <td class="px-3 py-3">
                  <div class="font-medium dark:text-slate-200">{{ item.name }}</div>
                  <div class="flex items-center gap-1 text-xs text-slate-500 dark:text-slate-400">
                    {{ item.username }}
                    <span
                      v-if="item.google_id"
                      class="inline-flex items-center rounded-full bg-blue-50 px-1.5 py-0.5 text-[10px] font-medium text-blue-700 ring-1 ring-blue-700/10 ring-inset dark:bg-blue-900/20 dark:text-blue-400"
                    >
                      G
                    </span>
                  </div>
                  <div class="text-xs text-slate-500 dark:text-slate-400">
                    {{ item.email || '-' }}
                  </div>
                </td>
                <td class="px-3 py-3 text-xs font-bold uppercase dark:text-slate-300">
                  {{ item.role }}
                </td>
                <td class="px-3 py-3">
                  <div v-if="item.role === 'student'" class="dark:text-slate-200">
                    <div class="text-xs">NISN: {{ item.nisn || '-' }}</div>
                    <div class="text-xs font-bold">NIS: {{ item.nis || '-' }}</div>
                    <div class="text-[10px] text-slate-500">
                      {{ item.level_name }} - {{ item.group_name }} / {{ item.program_code }}
                    </div>
                  </div>
                  <div v-if="item.role === 'teacher'" class="dark:text-slate-200">
                    <div class="text-xs font-bold">NIP: {{ item.nip || '-' }}</div>
                    <div class="text-[10px] text-slate-500 italic">
                      Mapel: {{ item.mapel_codes }}
                    </div>
                  </div>
                  <div class="mt-1 text-[10px] font-medium text-blue-600 dark:text-blue-400">
                    {{ item.phone || '-' }}
                  </div>
                </td>
                <td class="px-3 py-3 text-center">
                  <span
                    class="rounded-full px-2 py-1 text-[10px] font-bold tracking-tight uppercase"
                    :class="
                      item.status === 'approved'
                        ? 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400'
                        : item.status === 'rejected'
                          ? 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400'
                          : 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400'
                    "
                  >
                    {{ item.status }}
                  </span>
                </td>
                <td class="px-3 py-3 text-slate-500 dark:text-slate-400">{{ item.note || '-' }}</td>
                <td class="px-3 py-3 w-48">
                  <div class="flex flex-col gap-2 p-2 rounded-xl border border-purple-100 bg-purple-50/30 dark:border-purple-900/50 dark:bg-purple-900/20 shadow-sm">
                    <BaseButton
                      v-if="item.status === 'pending'"
                      :icon="mdiCheck"
                      color="success"
                      label="Approve"
                      class="w-full"
                      :disabled="actionLoadingId === item.id"
                      @click="decide(item.id, 'approve')"
                    />
                    <BaseButton
                      v-if="item.status === 'pending'"
                      :icon="mdiClose"
                      color="danger"
                      label="Reject"
                      class="w-full"
                      :disabled="actionLoadingId === item.id"
                      @click="decide(item.id, 'reject')"
                    />
                    <BaseButton
                      v-if="item.status !== 'pending'"
                      color="purple"
                      label="Set Pending"
                      class="w-full"
                      :disabled="actionLoadingId === item.id"
                      @click="decide(item.id, 'pending')"
                    />
                    <div
                      v-if="item.status === 'approved'"
                      class="mt-1 text-center text-[10px] text-purple-600 font-medium dark:text-purple-400 italic"
                    >
                      Akun sudah dibuat
                    </div>
                  </div>
                </td>
              </tr>
              <tr v-if="!registrations.length && !isLoading">
                <td
                  colspan="6"
                  class="px-3 py-10 text-center text-slate-400 italic dark:text-slate-500"
                >
                  Belum ada data pendaftaran.
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </CardBox>
    </SectionMain>
  </LayoutAuthenticated>
</template>
