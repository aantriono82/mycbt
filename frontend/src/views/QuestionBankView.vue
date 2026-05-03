<script setup>
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import {
  mdiDatabaseOutline,
  mdiPlus,
  mdiRefresh,
  mdiContentSave,
  mdiDelete,
  mdiPencil,
  mdiEye,
  mdiContentCopy,
  mdiAccountArrowRightOutline,
  mdiFileDocumentOutline,
} from '@mdi/js'
import LayoutAuthenticated from '@/layouts/LayoutAuthenticated.vue'
import SectionMain from '@/components/SectionMain.vue'
import SectionTitleLineWithButton from '@/components/SectionTitleLineWithButton.vue'
import CardBox from '@/components/CardBox.vue'
import CardBoxModal from '@/components/CardBoxModal.vue'
import BaseButton from '@/components/BaseButton.vue'
import BaseButtons from '@/components/BaseButtons.vue'
import BaseIcon from '@/components/BaseIcon.vue'
import FormField from '@/components/FormField.vue'
import FormControl from '@/components/FormControl.vue'
import { api } from '@/services/api.js'
import { useAuthStore } from '@/stores/auth.js'

const authStore = useAuthStore()
const router = useRouter()
const route = useRoute()

const subjects = ref([])
const levels = ref([])
const questionSets = ref([])
const isLoadingSets = ref(false)
const errorMessage = ref('')
const successMessage = ref('')

const teachers = ref([])
const isCloneModalActive = ref(false)
const cloningItemID = ref(null)
const cloneTargetTeacherID = ref('')
const isCloning = ref(false)

const isAdmin = computed(() => authStore.role === 'admin')

const formErrors = reactive({
  subject_id: '',
  title: '',
})

const setForm = reactive({
  subject_id: '',
  title: '',
})

const resetSetForm = () => {
  setForm.subject_id = ''
  setForm.title = ''
}

const loadSubjects = async () => {
  try {
    const { data } = await api.get('/api/v1/lookups/subjects')
    subjects.value = data?.data || []
  } catch {
    subjects.value = []
  }
}

const loadLevels = async () => {
  try {
    const { data } = await api.get('/api/v1/lookups/levels')
    levels.value = data?.data || []
  } catch {
    levels.value = []
  }
}

const loadTeachers = async () => {
  if (!isAdmin.value) return
  try {
    const { data } = await api.get('/api/v1/lookups/teachers')
    teachers.value = data?.data || []
  } catch {
    teachers.value = []
  }
}

const loadQuestionSets = async () => {
  isLoadingSets.value = true
  errorMessage.value = ''
  try {
    const { data } = await api.get('/api/v1/question-sets', {
      params: { limit: 100, offset: 0 },
    })
    questionSets.value = data?.data || []
  } catch (error) {
    errorMessage.value = error?.response?.data?.error?.message || 'Gagal memuat bank soal'
  } finally {
    isLoadingSets.value = false
  }
}

const saveQuestionSet = async () => {
  successMessage.value = ''
  errorMessage.value = ''
  
  if (!setForm.title) {
    formErrors.title = 'Judul wajib diisi'
    return
  }

  try {
    const { data } = await api.post('/api/v1/question-sets', {
      subject_id: setForm.subject_id,
      title: setForm.title,
    })
    successMessage.value = 'Set soal ditambahkan. Mengalihkan ke editor...'
    setTimeout(() => {
      goToEditor(data.data.id)
    }, 1000)
    resetSetForm()
    await loadQuestionSets()
  } catch (error) {
    errorMessage.value = error?.response?.data?.error?.message || 'Gagal menyimpan'
  }
}

const deleteQuestionSet = async (id) => {
  if (!confirm('Hapus set soal ini?')) return
  try {
    await api.delete(`/api/v1/question-sets/${id}`)
    await loadQuestionSets()
  } catch (error) {
    errorMessage.value = 'Gagal menghapus'
  }
}

const openCloneModal = (id) => {
  cloningItemID.value = id
  isCloneModalActive.value = true
}

const submitClone = async () => {
  if (!cloneTargetTeacherID.value) {
    alert('Silakan pilih guru tujuan')
    return
  }
  
  isCloning.value = true
  try {
    await api.post(`/api/v1/question-sets/${cloningItemID.value}/clone`, {
      teacher_id: cloneTargetTeacherID.value
    })
    successMessage.value = 'Bank soal berhasil disalin ke guru tersebut.'
    isCloneModalActive.value = false
    await loadQuestionSets()
  } catch (error) {
    alert(error?.response?.data?.error?.message || 'Gagal menyalin bank soal')
  } finally {
    isCloning.value = false
  }
}

const goToEditor = (id) => {
  const role = route.path.startsWith('/admin') ? 'admin' : 'teacher'
  router.push(`/${role}/bank-soal/new?id=${id}`)
}

const goToPreview = (id) => {
  const role = route.path.startsWith('/admin') ? 'admin' : 'teacher'
  router.push(`/${role}/bank-soal/preview/${id}`)
}

const goToImport = (id = '') => {
  const role = route.path.startsWith('/admin') ? 'admin' : 'teacher'
  const suffix = id ? `?set_id=${id}` : ''
  router.push(`/${role}/bank-soal/import${suffix}`)
}

onMounted(async () => {
  await loadSubjects()
  await loadLevels()
  await loadQuestionSets()
  if (isAdmin.value) {
    await loadTeachers()
  }
})
</script>

<template>
  <LayoutAuthenticated>
    <SectionMain>
      <SectionTitleLineWithButton :icon="mdiDatabaseOutline" title="Daftar Bank Soal" main>
        <div class="flex items-center gap-2">
          <BaseButton :icon="mdiFileDocumentOutline" color="purple" label="Impor Soal" @click="goToImport()" />
          <BaseButton :icon="mdiRefresh" color="info" label="Refresh" @click="loadQuestionSets" />
        </div>
      </SectionTitleLineWithButton>

      <CardBoxModal
        v-model="isCloneModalActive"
        title="Salin Bank Soal ke Guru"
        button-label="Salin Sekarang"
        has-cancel
        :is-loading="isCloning"
        @confirm="submitClone"
      >
        <p class="mb-4 text-sm text-slate-500">Pilih guru yang akan menerima salinan bank soal ini. Seluruh butir soal akan diduplikasi ke panel guru terpilih.</p>
        <FormField label="Pilih Guru Tujuan">
          <FormControl
            v-model="cloneTargetTeacherID"
            :options="[
              { value: '', label: 'Pilih guru' },
              ...teachers.map(t => ({ value: t.id, label: t.name }))
            ]"
          />
        </FormField>
      </CardBoxModal>

      <div class="mb-6 grid gap-6 xl:grid-cols-5 animate-fade-in">
        <!-- Create/Edit Form -->
        <CardBox class="xl:col-span-2 shadow-md" color="blue">
          <h3 class="mb-4 text-lg font-bold dark:text-slate-100 flex items-center gap-2">
            <BaseIcon :path="mdiPlus" size="20" />
            Tambah Bank Soal
          </h3>
          <div class="grid gap-4">
            <FormField label="Mata Pelajaran" :error="formErrors.subject_id">
              <FormControl
                v-model="setForm.subject_id"
                :options="[
                  { value: '', label: 'Pilih mata pelajaran' },
                  ...subjects.map((item) => ({
                    value: item.id,
                    label: `${item.code || '-'} - ${item.name}`,
                  })),
                ]"
              />
            </FormField>
            <FormField label="Judul Set" :error="formErrors.title">
              <FormControl v-model="setForm.title" placeholder="Contoh: Matematika Kelas X PAS" />
            </FormField>
            <BaseButtons>
              <BaseButton
                :icon="mdiContentSave"
                color="info"
                label="Buat & Lanjut"
                @click="saveQuestionSet"
              />
            </BaseButtons>
          </div>
        </CardBox>

        <!-- List Table -->
        <CardBox class="xl:col-span-3 shadow-md" color="emerald">
          <div v-if="errorMessage" class="mb-4 rounded-lg bg-red-50 p-3 text-xs text-red-700 border border-red-100">{{ errorMessage }}</div>
          <div v-if="successMessage" class="mb-4 rounded-lg bg-emerald-50 p-3 text-xs text-emerald-700 border border-emerald-100">{{ successMessage }}</div>

          <div class="overflow-x-auto">
            <table class="w-full text-left text-sm">
              <thead class="border-b dark:border-slate-800 bg-slate-50 dark:bg-slate-800/50 text-slate-600 dark:text-slate-300 uppercase text-[10px] tracking-widest font-black">
                <tr>
                  <th class="px-3 py-3">Nama Bank Soal</th>
                  <th class="px-3 py-3 text-center">Soal</th>
                  <th class="px-3 py-3 text-center">Status</th>
                  <th class="px-3 py-3 text-right">Aksi</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="item in questionSets" :key="item.id" class="border-b dark:border-slate-800 last:border-b-0 hover:bg-slate-50 dark:hover:bg-slate-800/50 transition-colors group cursor-pointer" @click="goToEditor(item.id)">
                  <td class="px-3 py-4">
                    <div class="font-bold text-slate-800 dark:text-slate-200 group-hover:text-blue-600">{{ item.title }}</div>
                    <div class="text-[10px] text-slate-400 mt-1 uppercase flex gap-2">
                       <span>{{ subjects.find(s => s.id === item.subject_id)?.name || item.subject_id }}</span>
                       <span v-if="item.jenjang" class="text-blue-500 font-black">• {{ item.jenjang }}</span>
                       <span v-if="item.level_id" class="text-blue-500 font-black">• {{ levels.find(l => l.id === item.level_id)?.name || 'Class' }}</span>
                    </div>
                  </td>
                  <td class="px-3 py-4 text-center font-mono text-xs text-slate-500">-</td>
                  <td class="px-3 py-4 text-center text-[10px] font-black uppercase tracking-tighter" :class="item.status === 'published' ? 'text-emerald-600' : 'text-slate-400'">
                    {{ item.status }}
                  </td>
                  <td class="px-3 py-4" @click.stop>
                    <div class="flex items-center justify-end gap-4">
                      <BaseIcon v-if="isAdmin" :path="mdiAccountArrowRightOutline" size="18" class="text-emerald-500 hover:scale-125 transition-transform cursor-pointer" title="Salin ke Guru" @click="openCloneModal(item.id)" />
                      <BaseIcon :path="mdiFileDocumentOutline" size="18" class="text-purple-500 hover:scale-125 transition-transform cursor-pointer" title="Impor soal ke bank soal ini" @click="goToImport(item.id)" />
                      <BaseIcon :path="mdiEye" size="18" class="text-indigo-500 hover:scale-125 transition-transform cursor-pointer" title="Pratinjau Soal" @click="goToPreview(item.id)" />
                      <BaseIcon :path="mdiPencil" size="18" class="text-blue-500 hover:scale-125 transition-transform cursor-pointer" title="Kelola Bank Soal (Edit Judul & Isi)" @click="goToEditor(item.id)" />
                      <BaseIcon :path="mdiDelete" size="18" class="text-red-500 hover:scale-125 transition-transform cursor-pointer" @click="deleteQuestionSet(item.id)" />
                    </div>
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

<style scoped>
.animate-fade-in { animation: fadeIn 0.4s ease-out; }
@keyframes fadeIn { from { opacity: 0; } to { opacity: 1; } }
</style>
