<script setup>
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useRoute } from 'vue-router'

const formatDateIndonesian = (dateStr) => {
  if (!dateStr) return '-'
  const d = new Date(dateStr)
  if (isNaN(d.getTime())) return dateStr
  const formatted = d.toLocaleString('id-ID', {
    dateStyle: 'medium',
    timeStyle: 'short',
    hour12: false
  }).replace(/\./g, ':')

  const offset = -d.getTimezoneOffset() / 60
  let tz = ''
  if (offset === 7) tz = 'WIB'
  else if (offset === 8) tz = 'WITA'
  else if (offset === 9) tz = 'WIT'
  else tz = offset >= 0 ? `GMT+${offset}` : `GMT${offset}`

  return `${formatted} ${tz}`
}
  import { mdiClipboardTextOutline, mdiContentSave, mdiDelete, mdiPlus, mdiRefresh } from '@mdi/js'
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

const route = useRoute()
const authStore = useAuthStore()

const exams = ref([])
const teachers = ref([])
const questionSets = ref([])
const subjects = ref([])
const levels = ref([])
const groups = ref([])
const students = ref([])
const attachedQuestionSets = ref([])
const targets = ref([])
const sessions = ref([])

const errorMessage = ref('')
const successMessage = ref('')
const isLoading = ref(false)
const isLoadingDetails = ref(false)
const formErrors = reactive({
  teacher_id: '',
  subject_id: '',
  title: '',
  starts_at: '',
  ends_at: '',
  duration_minutes: '',
  question_set_id: '',
  num_questions: '',
  target_ids: '',
})

const selectedExamId = ref('')
const selectedExamSessionId = ref('')
const selectedExamScoringMode = ref('partial')
const selectedExam = computed(() => exams.value.find((x) => x.id === selectedExamId.value) || null)

const form = reactive({
  teacher_id: '',
  subject_id: '',
  session_id: '',
  title: '',
  starts_at: '',
  ends_at: '',
  duration_minutes: 60,
  shuffle_questions: true,
  shuffle_options: true,
  scoring_mode: 'partial',
})

const scheduleForm = reactive({
  starts_date: '',
  starts_hour: '',
  starts_minute: '',
  ends_date: '',
  ends_hour: '',
  ends_minute: '',
})

const hourOptions = Array.from({ length: 24 }, (_, i) => {
  const value = String(i).padStart(2, '0')
  return { value, label: value }
})

const minuteOptions = Array.from({ length: 60 }, (_, i) => {
  const value = String(i).padStart(2, '0')
  return { value, label: value }
})

const buildDateTimeLocalValue = (dateValue, hourValue, minuteValue) => {
  if (!dateValue || hourValue === '' || minuteValue === '') return ''
  return `${dateValue}T${String(hourValue).padStart(2, '0')}:${String(minuteValue).padStart(2, '0')}`
}

const syncScheduleDateTimes = () => {
  form.starts_at = buildDateTimeLocalValue(
    scheduleForm.starts_date,
    scheduleForm.starts_hour,
    scheduleForm.starts_minute,
  )
  form.ends_at = buildDateTimeLocalValue(
    scheduleForm.ends_date,
    scheduleForm.ends_hour,
    scheduleForm.ends_minute,
  )
}

const attachForm = reactive({
  question_set_id: '',
  num_questions: '',
})

const targetForm = reactive({
  level_ids_text: '',
  group_ids_text: '',
  student_ids_text: '',
})

const helperLevel = ref('')
const helperGroup = ref('')
const helperStudent = ref('')

const appendToTarget = (key, value) => {
  if (!value) return
  const current = targetForm[key] || ''
  const items = current.split(/[\n,]/).map((s) => s.trim()).filter(Boolean)
  if (!items.includes(value)) {
    items.push(value)
    targetForm[key] = items.join('\n')
  }
}

watch(helperLevel, (val) => {
  if (val) {
    appendToTarget('level_ids_text', val)
    setTimeout(() => { helperLevel.value = '' }, 100)
  }
})
watch(helperGroup, (val) => {
  if (val) {
    appendToTarget('group_ids_text', val)
    setTimeout(() => { helperGroup.value = '' }, 100)
  }
})
watch(helperStudent, (val) => {
  if (val) {
    appendToTarget('student_ids_text', val)
    setTimeout(() => { helperStudent.value = '' }, 100)
  }
})

const selectedLevelNames = computed(() => {
  const ids = parseCsvLines(targetForm.level_ids_text)
  return ids.map((id) => levels.value.find((l) => l.id === id)?.name || id)
})

const selectedGroupNames = computed(() => {
  const ids = parseCsvLines(targetForm.group_ids_text)
  return ids.map((id) => groups.value.find((g) => g.id === id)?.name || id)
})

const selectedStudentNames = computed(() => {
  const ids = parseCsvLines(targetForm.student_ids_text)
  return ids.map((id) => {
    const s = students.value.find((x) => x.id === id)
    return s ? `${s.name} (${s.nis})` : id
  })
})

const isTeacherArea = computed(() => String(route.path || '').startsWith('/teacher/'))
const isAdminArea = computed(() => !isTeacherArea.value)

const parseCsvLines = (value) =>
  String(value || '')
    .split(/[\n,]/)
    .map((item) => item.trim())
    .filter(Boolean)

const uuidPattern =
  /^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i

const clearExamFormErrors = () => {
  formErrors.teacher_id = ''
  formErrors.subject_id = ''
  formErrors.title = ''
  formErrors.starts_at = ''
  formErrors.ends_at = ''
  formErrors.duration_minutes = ''
}

const clearAttachErrors = () => {
  formErrors.question_set_id = ''
  formErrors.num_questions = ''
}

const clearTargetErrors = () => {
  formErrors.target_ids = ''
}

const validateCreateExamForm = () => {
  clearExamFormErrors()
  syncScheduleDateTimes()
  if (!isTeacherArea.value && !String(form.teacher_id || '').trim()) {
    formErrors.teacher_id = 'Guru wajib dipilih'
  }
  if (!String(form.subject_id || '').trim()) formErrors.subject_id = 'Mata pelajaran wajib dipilih'
  const title = String(form.title || '').trim()
  if (title.length < 3) formErrors.title = 'Judul ujian minimal 3 karakter'
  if (title.length > 120) formErrors.title = 'Judul ujian maksimal 120 karakter'
  const startsAt = new Date(form.starts_at)
  const endsAt = new Date(form.ends_at)
  if (!form.starts_at || Number.isNaN(startsAt.getTime())) formErrors.starts_at = 'Waktu mulai wajib diisi'
  if (!form.ends_at || Number.isNaN(endsAt.getTime())) formErrors.ends_at = 'Waktu selesai wajib diisi'
  if (!formErrors.starts_at && !formErrors.ends_at && endsAt <= startsAt) {
    formErrors.ends_at = 'Waktu selesai harus setelah waktu mulai'
  }
  const duration = Number(form.duration_minutes)
  if (!Number.isInteger(duration) || duration < 1 || duration > 600) {
    formErrors.duration_minutes = 'Durasi harus bilangan bulat 1-600'
  }
  return !Object.values({
    teacher_id: formErrors.teacher_id,
    subject_id: formErrors.subject_id,
    title: formErrors.title,
    starts_at: formErrors.starts_at,
    ends_at: formErrors.ends_at,
    duration_minutes: formErrors.duration_minutes,
  }).some(Boolean)
}

const validateAttachForm = () => {
  clearAttachErrors()
  if (!String(attachForm.question_set_id || '').trim()) {
    formErrors.question_set_id = 'Question set wajib dipilih'
  }
  if (attachForm.num_questions === '' || attachForm.num_questions === null) return ''
  const count = Number(attachForm.num_questions)
  if (!Number.isInteger(count) || count < 1 || count > 500) {
    formErrors.num_questions = 'Jumlah soal harus bilangan bulat 1-500'
  }
  return !formErrors.question_set_id && !formErrors.num_questions
}

const validateTargetsForm = () => {
  clearTargetErrors()
  const levelIDs = parseCsvLines(targetForm.level_ids_text)
  const groupIDs = parseCsvLines(targetForm.group_ids_text)
  const studentIDs = parseCsvLines(targetForm.student_ids_text)
  if (!levelIDs.length && !groupIDs.length && !studentIDs.length) {
    formErrors.target_ids = 'Minimal satu target (level/group/siswa) harus diisi'
  }
  const allIDs = [...levelIDs, ...groupIDs, ...studentIDs]
  const invalid = allIDs.find((id) => !uuidPattern.test(id))
  if (invalid) formErrors.target_ids = `ID target tidak valid: ${invalid}`
  return !formErrors.target_ids
}

const loadTeachers = async () => {
  if (!authStore.isAuthenticated || isTeacherArea.value) return
  try {
    const { data } = await api.get('/api/v1/admin/teachers', {
      params: { limit: 100, offset: 0 },
    })
    teachers.value = data?.data || []
  } catch {
    teachers.value = []
  }
}

const loadTargetLookups = async () => {
  if (!authStore.isAuthenticated) return
  try {
    const [levelsRes, groupsRes, studentsRes] = await Promise.all([
      api.get('/api/v1/lookups/levels'),
      api.get('/api/v1/lookups/groups'),
      api.get('/api/v1/lookups/students', { params: { limit: 100, offset: 0 } }),
    ])
    levels.value = levelsRes?.data?.data || []
    groups.value = groupsRes?.data?.data || []
    students.value = studentsRes?.data?.data || []
  } catch {
    levels.value = []
    groups.value = []
    students.value = []
  }
}

const loadQuestionSets = async () => {
  if (!authStore.isAuthenticated) return
  try {
    const { data } = await api.get('/api/v1/question-sets', {
      params: { limit: 100, offset: 0 },
    })
    questionSets.value = data?.data || []
  } catch {
    questionSets.value = []
  }
}

const loadSubjects = async () => {
  if (!authStore.isAuthenticated) return
  try {
    const { data } = await api.get('/api/v1/lookups/subjects')
    subjects.value = data?.data || []
  } catch {
    subjects.value = []
  }
}

const loadSessions = async () => {
  if (!authStore.isAuthenticated) return
  try {
    const { data } = await api.get('/api/v1/lookups/sessions')
    sessions.value = data?.data || []
  } catch {
    sessions.value = []
  }
}

const loadExams = async () => {
  if (!authStore.isAuthenticated) return
  isLoading.value = true
  errorMessage.value = ''
  try {
    const { data } = await api.get('/api/v1/exams', {
      params: { limit: 100, offset: 0 },
    })
    exams.value = data?.data || []
    if (!selectedExamId.value && exams.value.length) {
      selectedExamId.value = exams.value[0].id
    }
  } catch (error) {
    errorMessage.value = error?.response?.data?.error?.message || 'Gagal memuat jadwal ujian'
  } finally {
    isLoading.value = false
  }
}

const loadSelectedExamDetails = async () => {
  if (!authStore.isAuthenticated || !selectedExamId.value) {
    attachedQuestionSets.value = []
    targets.value = []
    return
  }
  isLoadingDetails.value = true
  try {
    const [questionSetsRes, targetsRes] = await Promise.all([
      api.get(`/api/v1/exams/${selectedExamId.value}/question-sets`),
      api.get(`/api/v1/exams/${selectedExamId.value}/targets`),
    ])
    attachedQuestionSets.value = questionSetsRes?.data?.data || []
    targets.value = targetsRes?.data?.data || []

    targetForm.level_ids_text = targets.value
      .map((item) => item.level_id)
      .filter(Boolean)
      .join('\n')
    targetForm.group_ids_text = targets.value
      .map((item) => item.group_id)
      .filter(Boolean)
      .join('\n')
    targetForm.student_ids_text = targets.value
      .map((item) => item.student_id)
      .filter(Boolean)
      .join('\n')
  } catch (error) {
    errorMessage.value = error?.response?.data?.error?.message || 'Gagal memuat detail ujian'
  } finally {
    isLoadingDetails.value = false
  }
}

const setExamStatus = async (nextStatus) => {
  if (!selectedExamId.value) return
  errorMessage.value = ''
  successMessage.value = ''
  try {
    await api.patch(`/api/v1/exams/${selectedExamId.value}`, { status: nextStatus })
    successMessage.value = `Status ujian diubah menjadi ${nextStatus}`
    await loadExams()
    await loadSelectedExamDetails()
  } catch (error) {
    errorMessage.value = error?.response?.data?.error?.message || 'Gagal mengubah status ujian'
  }
}

const deleteExam = async () => {
  if (!selectedExamId.value) return
  if (!confirm(`Hapus jadwal ujian "${selectedExam.value?.title}"? Tindakan ini tidak dapat dibatalkan.`)) return

  errorMessage.value = ''
  successMessage.value = ''
  try {
    await api.delete(`/api/v1/exams/${selectedExamId.value}`)
    successMessage.value = 'Jadwal ujian berhasil dihapus'
    selectedExamId.value = ''
    await loadExams()
  } catch (error) {
    errorMessage.value = error?.response?.data?.error?.message || 'Gagal menghapus jadwal ujian'
  }
}

const updateExamSession = async () => {
  if (!selectedExamId.value) return
  errorMessage.value = ''
  successMessage.value = ''
  try {
    await api.patch(`/api/v1/exams/${selectedExamId.value}`, { session_id: selectedExamSessionId.value })
    successMessage.value = 'Sesi ujian berhasil diperbarui'
    await loadExams()
  } catch (error) {
    errorMessage.value = error?.response?.data?.error?.message || 'Gagal memperbarui sesi ujian'
  }
}

const updateExamScoringMode = async () => {
  if (!selectedExamId.value) return
  errorMessage.value = ''
  successMessage.value = ''
  try {
    await api.patch(`/api/v1/exams/${selectedExamId.value}`, {
      scoring_mode: selectedExamScoringMode.value,
    })
    successMessage.value = 'Mode penilaian berhasil diperbarui'
    await loadExams()
  } catch (error) {
    errorMessage.value = error?.response?.data?.error?.message || 'Gagal memperbarui mode penilaian'
  }
}

const createExam = async () => {
  successMessage.value = ''
  errorMessage.value = ''
  syncScheduleDateTimes()
  const isValid = validateCreateExamForm()
  if (!isValid) {
    errorMessage.value = 'Periksa kembali form jadwal ujian'
    return
  }
  const payload = {
    subject_id: String(form.subject_id || '').trim(),
    session_id: String(form.session_id || '').trim(),
    title: String(form.title || '').trim(),
    starts_at: form.starts_at ? new Date(form.starts_at).toISOString() : '',
    ends_at: form.ends_at ? new Date(form.ends_at).toISOString() : '',
    duration_minutes: Number(form.duration_minutes),
    shuffle_questions: form.shuffle_questions,
    shuffle_options: form.shuffle_options,
    scoring_mode: String(form.scoring_mode || 'partial').trim() || 'partial',
  }
  if (!isTeacherArea.value) {
    payload.teacher_id = form.teacher_id
  }

  try {
    const { data } = await api.post('/api/v1/exams', payload)
    successMessage.value = 'Jadwal ujian berhasil ditambahkan'
    form.subject_id = ''
    form.session_id = ''
    form.title = ''
    form.starts_at = ''
    form.ends_at = ''
    scheduleForm.starts_date = ''
    scheduleForm.starts_hour = ''
    scheduleForm.starts_minute = ''
    scheduleForm.ends_date = ''
    scheduleForm.ends_hour = ''
    scheduleForm.ends_minute = ''
    form.duration_minutes = 60
    form.scoring_mode = 'partial'
    await loadExams()
    selectedExamId.value = data?.data?.id || selectedExamId.value
  } catch (error) {
    errorMessage.value = error?.response?.data?.error?.message || 'Gagal menambah jadwal ujian'
  }
}

const addAttachedQuestionSet = () => {
  const isValid = validateAttachForm()
  if (!isValid) {
    errorMessage.value = 'Periksa kembali form attach bank soal'
    return
  }
  if (attachedQuestionSets.value.some((item) => item.question_set_id === attachForm.question_set_id)) {
    errorMessage.value = 'Question set sudah ada pada ujian ini'
    return
  }
  attachedQuestionSets.value = [
    ...attachedQuestionSets.value,
    {
      question_set_id: attachForm.question_set_id,
      num_questions: attachForm.num_questions ? Number(attachForm.num_questions) : null,
    },
  ]
  attachForm.question_set_id = ''
  attachForm.num_questions = ''
}

const removeAttachedQuestionSet = (questionSetId) => {
  attachedQuestionSets.value = attachedQuestionSets.value.filter(
    (item) => item.question_set_id !== questionSetId,
  )
}

const saveAttachedQuestionSets = async () => {
  if (!selectedExamId.value) return
  errorMessage.value = ''
  successMessage.value = ''
  if (!attachedQuestionSets.value.length) {
    clearAttachErrors()
    errorMessage.value = 'Minimal satu bank soal harus di-attach ke ujian'
    return
  }
  try {
    await api.put(`/api/v1/exams/${selectedExamId.value}/question-sets`, {
      items: attachedQuestionSets.value,
    })
    successMessage.value = 'Bank soal untuk ujian berhasil diperbarui'
    await loadSelectedExamDetails()
  } catch (error) {
    errorMessage.value = error?.response?.data?.error?.message || 'Gagal menyimpan bank soal ujian'
  }
}

const saveTargets = async () => {
  if (!selectedExamId.value) return
  errorMessage.value = ''
  successMessage.value = ''
  const isValid = validateTargetsForm()
  if (!isValid) {
    errorMessage.value = 'Periksa kembali target ujian'
    return
  }
  try {
    await api.put(`/api/v1/exams/${selectedExamId.value}/targets`, {
      level_ids: parseCsvLines(targetForm.level_ids_text),
      group_ids: parseCsvLines(targetForm.group_ids_text),
      student_ids: parseCsvLines(targetForm.student_ids_text),
    })
    successMessage.value = 'Target ujian berhasil diperbarui'
    await loadSelectedExamDetails()
  } catch (error) {
    errorMessage.value = error?.response?.data?.error?.message || 'Gagal menyimpan target ujian'
  }
}

const questionSetTitle = (id) =>
  questionSets.value.find((item) => item.id === id)?.title || id

watch(selectedExamId, (newId) => {
  loadSelectedExamDetails()
  if (selectedExam.value) {
    selectedExamSessionId.value = selectedExam.value.session_id || ''
    selectedExamScoringMode.value = selectedExam.value.scoring_mode || 'partial'
  } else {
    selectedExamSessionId.value = ''
    selectedExamScoringMode.value = 'partial'
  }
})

onMounted(async () => {
  await Promise.all([loadTeachers(), loadTargetLookups(), loadSubjects(), loadSessions(), loadQuestionSets(), loadExams()])
  await loadSelectedExamDetails()
})
</script>

<template>
  <LayoutAuthenticated>
    <SectionMain>
      <SectionTitleLineWithButton :icon="mdiClipboardTextOutline" title="Jadwal Ujian" main>
        <BaseButton :icon="mdiRefresh" color="info" label="Muat Ulang" @click="loadExams(); loadQuestionSets(); loadSelectedExamDetails()" />
      </SectionTitleLineWithButton>

      <div class="mb-6 grid gap-6 xl:grid-cols-5">
        <CardBox class="xl:col-span-2">
          <h3 class="mb-4 text-lg font-semibold dark:text-slate-100">Tambah Jadwal Ujian</h3>
          <div class="grid gap-4">
            <FormField v-if="!isTeacherArea" label="Guru" :error="formErrors.teacher_id">
              <FormControl
                v-model="form.teacher_id"
                :options="[{ id: '', label: 'Pilih guru' }, ...teachers.map((item) => ({ id: item.id, label: item.name }))]"
              />
            </FormField>
            <FormField label="Mata Pelajaran" :error="formErrors.subject_id">
              <FormControl
                v-model="form.subject_id"
                :options="[{ id: '', label: 'Pilih mapel' }, ...subjects.map((item) => ({ id: item.id, label: item.code ? `${item.code} - ${item.name}` : item.name }))]"
              />
            </FormField>
            <FormField label="Sesi (Opsional)">
              <FormControl
                v-model="form.session_id"
                :options="[{ id: '', label: 'Pilih sesi (pilihan)' }, ...sessions.map((item) => ({ id: item.id, label: item.start_time ? `${item.name} (${item.start_time} - ${item.end_time})` : item.name }))]"
              />
            </FormField>
            <FormField label="Judul Ujian" :error="formErrors.title">
              <FormControl v-model="form.title" placeholder="Ujian Harian 1" />
            </FormField>
            <FormField label="Mulai" :error="formErrors.starts_at" help="Gunakan format 24 jam. Waktu akan ditampilkan dalam WIB/WITA/WIT sesuai zona waktu Anda">
              <div class="grid grid-cols-3 gap-2">
                <FormControl v-model="scheduleForm.starts_date" type="date" />
                <FormControl
                  v-model="scheduleForm.starts_hour"
                  :options="[{ value: '', label: 'Jam' }, ...hourOptions]"
                />
                <FormControl
                  v-model="scheduleForm.starts_minute"
                  :options="[{ value: '', label: 'Menit' }, ...minuteOptions]"
                />
              </div>
            </FormField>
            <FormField label="Selesai" :error="formErrors.ends_at" help="Gunakan format 24 jam. Waktu akan ditampilkan dalam WIB/WITA/WIT sesuai zona waktu Anda">
              <div class="grid grid-cols-3 gap-2">
                <FormControl v-model="scheduleForm.ends_date" type="date" />
                <FormControl
                  v-model="scheduleForm.ends_hour"
                  :options="[{ value: '', label: 'Jam' }, ...hourOptions]"
                />
                <FormControl
                  v-model="scheduleForm.ends_minute"
                  :options="[{ value: '', label: 'Menit' }, ...minuteOptions]"
                />
              </div>
            </FormField>
            <FormField label="Durasi (menit)" :error="formErrors.duration_minutes">
              <FormControl v-model="form.duration_minutes" inputmode="numeric" />
            </FormField>
            <FormField label="Acak Soal">
              <FormControl
                v-model="form.shuffle_questions"
                :options="[
                  { id: true, label: 'Ya' },
                  { id: false, label: 'Tidak' },
                ]"
              />
            </FormField>
            <FormField label="Acak Opsi">
              <FormControl
                v-model="form.shuffle_options"
                :options="[
                  { id: true, label: 'Ya' },
                  { id: false, label: 'Tidak' },
                ]"
              />
            </FormField>
            <FormField label="Mode Penilaian">
              <FormControl
                v-model="form.scoring_mode"
                :options="[
                  { id: 'partial', label: 'Parsial (nilai proporsional)' },
                  { id: 'absolute', label: 'Absolut (harus tepat penuh)' },
                ]"
              />
            </FormField>
            <BaseButton :icon="mdiPlus" color="info" label="Tambah Jadwal" @click="createExam" />
          </div>
        </CardBox>

        <CardBox class="xl:col-span-3">
	          <div class="mb-4 flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
	            <div>
	              <h3 class="text-lg font-semibold dark:text-slate-100">Daftar Jadwal Ujian</h3>
	              <p class="text-sm text-slate-500 dark:text-slate-400">Pilih satu ujian untuk mengatur bank soal dan target peserta.</p>
	            </div>
	            <div class="w-full lg:max-w-sm">
	              <FormField label="Ujian Aktif">
	                <FormControl
	                  v-model="selectedExamId"
	                  :options="exams.map((item) => ({ id: item.id, label: item.title }))"
	                />
	              </FormField>
	            </div>
	          </div>

	          <div v-if="selectedExamId" class="mb-4 rounded-xl border border-slate-200 dark:border-slate-800 px-4 py-3 text-sm bg-slate-50/50 dark:bg-slate-800/30">
	            <div class="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
	              <div>
	                <div class="text-slate-500 dark:text-slate-400 font-medium">Status Ujian</div>
	                <div class="mt-1">
	                  <span
	                    class="rounded-full px-2 py-1 text-xs font-semibold"
	                    :class="
	                      selectedExam?.status === 'published'
	                        ? 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400'
	                        : selectedExam?.status === 'archived'
	                          ? 'bg-slate-200 text-slate-700 dark:bg-slate-800 dark:text-slate-400'
	                          : 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400'
	                    "
	                  >
	                    {{ selectedExam?.status || '-' }}
	                  </span>
	                </div>
	                <div class="mt-2 text-[10px] text-slate-500 dark:text-slate-500 uppercase font-bold tracking-tight">
	                  Catatan: ujian hanya muncul di area siswa ketika status = <span class="text-emerald-600 dark:text-emerald-400">published</span>.
	                </div>
                  <div class="mt-4 border-t dark:border-slate-800 pt-3">
                    <div class="text-slate-500 dark:text-slate-400 font-medium text-xs mb-2">PENGATURAN SESI</div>
                    <div class="flex items-center gap-3">
                      <div class="flex-1">
                        <FormControl
                          v-model="selectedExamSessionId"
                          :options="[{ id: '', label: 'Tanpa Sesi' }, ...sessions.map((item) => ({ id: item.id, label: item.start_time ? `${item.name} (${item.start_time} - ${item.end_time})` : item.name }))]"
                          small
                        />
                      </div>
                      <BaseButton color="info" label="Update Sesi" small @click="updateExamSession" />
                    </div>
                  </div>
                  <div class="mt-4 border-t dark:border-slate-800 pt-3">
                    <div class="text-slate-500 dark:text-slate-400 font-medium text-xs mb-2">PENGATURAN PENILAIAN</div>
                    <div class="flex items-center gap-3">
                      <div class="flex-1">
                        <FormControl
                          v-model="selectedExamScoringMode"
                          :options="[
                            { id: 'partial', label: 'Partial' },
                            { id: 'absolute', label: 'Absolute' },
                          ]"
                          small
                        />
                      </div>
                      <BaseButton color="info" label="Update Mode" small @click="updateExamScoringMode" />
                    </div>
                  </div>
	              </div>
	              <div class="flex flex-wrap gap-2">
	                <BaseButton
	                  :icon="mdiContentSave"
	                  color="info"
	                  small
	                  label="Publish"
	                  :disabled="selectedExam?.status === 'published'"
	                  @click="setExamStatus('published')"
	                />
	                <BaseButton
	                  color="purple"
	                  small
	                  label="Draft"
	                  :disabled="selectedExam?.status === 'draft'"
	                  @click="setExamStatus('draft')"
	                />
	                <BaseButton
	                  color="contrast"
	                  small
	                  label="Archive"
	                  :disabled="selectedExam?.status === 'archived'"
	                  @click="setExamStatus('archived')"
	                />
	                <BaseButton
	                  :icon="mdiDelete"
	                  color="danger"
	                  small
	                  label="Hapus"
	                  @click="deleteExam"
	                />
	              </div>
	            </div>
	          </div>

          <div v-if="!authStore.isAuthenticated" class="rounded-lg bg-amber-50 dark:bg-amber-900/20 px-4 py-3 text-sm text-amber-700 dark:text-amber-400 border border-amber-100 dark:border-amber-900/40">
            Login terlebih dulu agar jadwal ujian dapat dimuat dari backend.
          </div>
          <div v-else-if="errorMessage" class="rounded-lg bg-red-50 dark:bg-red-900/20 px-4 py-3 text-sm text-red-700 dark:text-red-400 border border-red-100 dark:border-red-900/40">
            {{ errorMessage }}
          </div>
          <div v-if="successMessage" class="mb-4 rounded-lg bg-emerald-50 dark:bg-emerald-900/20 px-4 py-3 text-sm text-emerald-700 dark:text-emerald-400 border border-emerald-100 dark:border-emerald-900/40">
            {{ successMessage }}
          </div>

          <div v-if="isLoading" class="text-sm text-slate-500 dark:text-slate-400 italic">Memuat jadwal ujian...</div>
          <div v-else class="mb-6 overflow-x-auto">
            <table class="w-full text-left text-sm">
              <thead class="border-b dark:border-slate-800 bg-slate-50 dark:bg-slate-800/50 text-slate-600 dark:text-slate-300 uppercase text-xs tracking-wider font-bold">
                <tr>
                  <th class="px-3 py-3">Judul</th>
                  <th class="px-3 py-3">Mata Pelajaran</th>
                  <th class="px-3 py-3">Sesi</th>
                  <th class="px-3 py-3 text-center">Mode</th>
                  <th class="px-3 py-3">Mulai (WIB/WITA/WIT)</th>
                  <th class="px-3 py-3">Selesai (WIB/WITA/WIT)</th>
                  <th class="px-3 py-3 text-center">Status</th>
                </tr>
              </thead>
              <tbody>
                <tr
                  v-for="exam in exams"
                  :key="exam.id"
                  class="border-b dark:border-slate-800 last:border-b-0 transition-colors"
                  :class="selectedExamId === exam.id ? 'bg-sky-50 dark:bg-sky-900/20' : 'hover:bg-slate-50 dark:hover:bg-slate-800/20'"
                >
                  <td class="px-3 py-3 font-medium dark:text-slate-200">{{ exam.title }}</td>
                  <td class="px-3 py-3 text-slate-500 dark:text-slate-400">{{ subjects.find(s => s.id === exam.subject_id)?.name || exam.subject_id }}</td>
                  <td class="px-3 py-3 text-slate-500 dark:text-slate-400">
                    <div v-if="exam.session_id" class="flex flex-col gap-1 items-start">
                      <span class="rounded-lg bg-indigo-50 dark:bg-indigo-900/20 px-2 py-0.5 text-indigo-600 dark:text-indigo-400 border border-indigo-100 dark:border-indigo-900/30 font-bold">
                        {{ exam.session_name || 'Sesi' }}
                      </span>
                      <span v-if="exam.session_start_time" class="text-[10px] text-slate-400 font-mono">
                        {{ exam.session_start_time }} - {{ exam.session_end_time }}
                      </span>
                    </div>
                    <span v-else class="text-slate-400 italic">N/A</span>
                  </td>
                  <td class="px-3 py-3 text-center">
                    <span
                      class="rounded-full px-2 py-1 text-[10px] font-bold uppercase tracking-tight"
                      :class="exam.scoring_mode === 'absolute' ? 'bg-rose-100 text-rose-700 dark:bg-rose-900/30 dark:text-rose-400' : 'bg-indigo-100 text-indigo-700 dark:bg-indigo-900/30 dark:text-indigo-400'"
                    >
                      {{ exam.scoring_mode || 'partial' }}
                    </span>
                  </td>
                  <td class="px-3 py-3 text-xs text-slate-500 dark:text-slate-400">{{ formatDateIndonesian(exam.starts_at) }}</td>
                  <td class="px-3 py-3 text-xs text-slate-500 dark:text-slate-400">{{ formatDateIndonesian(exam.ends_at) }}</td>
                  <td class="px-3 py-3 text-center">
                    <span 
                      class="rounded-full px-2 py-1 text-[10px] font-bold uppercase tracking-tight"
                      :class="exam.status === 'published' ? 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400' : 'bg-slate-100 text-slate-600 dark:bg-slate-700 dark:text-slate-400'"
                    >
                      {{ exam.status }}
                    </span>
                  </td>
                </tr>
                <tr v-if="!exams.length && !isLoading">
                  <td colspan="7" class="px-3 py-10 text-center text-slate-400 dark:text-slate-500 italic">Belum ada jadwal ujian.</td>
                </tr>
              </tbody>
            </table>
          </div>

          <div v-if="selectedExamId" class="grid gap-6 xl:grid-cols-2">
            <div class="rounded-2xl border border-emerald-100 dark:border-emerald-900/30 p-4 bg-emerald-50/30 dark:bg-emerald-900/10">
              <h4 class="mb-4 text-base font-semibold dark:text-slate-200">Attach Bank Soal</h4>
              <div class="grid gap-4">
                <FormField label="Question Set" :error="formErrors.question_set_id">
                  <FormControl
                    v-model="attachForm.question_set_id"
                    :options="questionSets.map((item) => ({ id: item.id, label: item.title }))"
                  />
                </FormField>
                <FormField label="Jumlah Soal" :error="formErrors.num_questions">
                  <FormControl v-model="attachForm.num_questions" inputmode="numeric" placeholder="Opsional" />
                </FormField>
                <BaseButtons>
                  <BaseButton :icon="mdiPlus" color="info" label="Tambah Set" @click="addAttachedQuestionSet" />
                  <BaseButton
                    :icon="mdiContentSave"
                    color="success"
                    label="Simpan Attach"
                    @click="saveAttachedQuestionSets"
                  />
                </BaseButtons>
              </div>

              <div class="mt-4 space-y-3">
                <div
                  v-for="item in attachedQuestionSets"
                  :key="item.question_set_id"
                  class="flex items-center justify-between rounded-xl border border-slate-200 dark:border-slate-700 px-4 py-3 text-sm bg-white dark:bg-slate-800/40"
                >
                  <div>
                    <div class="font-medium dark:text-slate-200">{{ questionSetTitle(item.question_set_id) }}</div>
                    <div class="text-slate-500 dark:text-slate-400 text-xs">Jumlah soal: {{ item.num_questions ?? 'Semua' }}</div>
                  </div>
                  <BaseButton color="danger" small label="Hapus" @click="removeAttachedQuestionSet(item.question_set_id)" />
                </div>
                 <div v-if="!attachedQuestionSets.length" class="text-sm text-slate-500 dark:text-slate-500 italic">
                  Belum ada bank soal yang di-attach ke ujian ini.
                </div>
              </div>
            </div>

            <div class="rounded-2xl border border-slate-200 dark:border-slate-800 p-4 bg-slate-50/20 dark:bg-slate-800/10">
              <h4 class="mb-4 text-base font-semibold dark:text-slate-200">Target Ujian</h4>
              <div class="grid gap-4">
                <div class="grid gap-6 md:grid-cols-3">
                  <div>
                    <FormField label="Target Level" :error="formErrors.target_ids">
                      <div class="flex flex-col gap-2">
                        <FormControl
                          v-model="helperLevel"
                          :options="[{ id: '', label: 'Pilih level...' }, ...levels.map(l => ({ id: l.id, label: l.name }))]"
                        />
                        <FormControl
                          v-model="targetForm.level_ids_text"
                          type="textarea"
                          placeholder="List UUID Level"
                        />
                      </div>
                    </FormField>
                    <div class="mt-2 flex flex-wrap gap-1">
                      <span v-for="name in selectedLevelNames" :key="name" class="rounded bg-sky-100 dark:bg-sky-900/30 px-2 py-0.5 text-[10px] text-sky-700 dark:text-sky-400 border border-sky-200 dark:border-sky-800">
                        {{ name }}
                      </span>
                    </div>
                  </div>

                  <div>
                    <FormField label="Target Group" :error="formErrors.target_ids">
                      <div class="flex flex-col gap-2">
                        <FormControl
                          v-model="helperGroup"
                          :options="[{ id: '', label: 'Pilih group...' }, ...groups.map(g => ({ id: g.id, label: g.name }))]"
                        />
                        <FormControl
                          v-model="targetForm.group_ids_text"
                          type="textarea"
                          placeholder="List UUID Group"
                        />
                      </div>
                    </FormField>
                    <div class="mt-2 flex flex-wrap gap-1">
                      <span v-for="name in selectedGroupNames" :key="name" class="rounded bg-indigo-100 dark:bg-indigo-900/30 px-2 py-0.5 text-[10px] text-indigo-700 dark:text-indigo-400 border border-indigo-200 dark:border-indigo-800">
                        {{ name }}
                      </span>
                    </div>
                  </div>

                  <div>
                    <FormField label="Target Siswa" :error="formErrors.target_ids">
                      <div class="flex flex-col gap-2">
                        <FormControl
                          v-model="helperStudent"
                          :options="[{ id: '', label: 'Pilih siswa...' }, ...students.map(s => ({ id: s.id, label: `${s.name} (${s.nis})` }))]"
                        />
                        <FormControl
                          v-model="targetForm.student_ids_text"
                          type="textarea"
                          placeholder="List UUID Siswa"
                        />
                      </div>
                    </FormField>
                    <div class="mt-2 flex flex-wrap gap-1">
                      <span v-for="name in selectedStudentNames" :key="name" class="rounded bg-emerald-100 dark:bg-emerald-900/30 px-2 py-0.5 text-[10px] text-emerald-700 dark:text-emerald-400 border border-emerald-200 dark:border-emerald-800">
                        {{ name }}
                      </span>
                    </div>
                  </div>
                </div>
                <BaseButton
                  :icon="mdiContentSave"
                  color="info"
                  label="Simpan Target"
                  @click="saveTargets"
                />
              </div>

              <div class="mt-4">
                <div class="mb-2 text-sm font-medium text-slate-700 dark:text-slate-300 border-t dark:border-slate-800 pt-3">Target Tersimpan</div>
                <div v-if="isLoadingDetails" class="text-sm text-slate-500 dark:text-slate-400 italic">Memuat target...</div>
                <div v-else class="space-y-2 text-sm">
                  <div
                    v-for="item in targets"
                    :key="item.id"
                    class="rounded-lg bg-slate-50 dark:bg-slate-800 px-3 py-2 text-slate-700 dark:text-slate-300 border border-slate-100 dark:border-slate-700"
                  >
                    <span class="font-mono text-[10px] break-all">
                      Level: {{ item.level_id || '-' }} · Group: {{ item.group_id || '-' }} · Student: {{ item.student_id || '-' }}
                    </span>
                  </div>
                  <div v-if="!targets.length" class="text-slate-500 dark:text-slate-500 italic">
                    Belum ada target tersimpan untuk ujian ini.
                  </div>
                </div>
              </div>
            </div>
          </div>
        </CardBox>
      </div>
    </SectionMain>
  </LayoutAuthenticated>
</template>
