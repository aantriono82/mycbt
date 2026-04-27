<script setup>
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { mdiLockOutline, mdiArrowLeft, mdiPlayCircleOutline } from '@mdi/js'
import LayoutAuthenticated from '@/layouts/LayoutAuthenticated.vue'
import SectionMain from '@/components/SectionMain.vue'
import SectionTitleLineWithButton from '@/components/SectionTitleLineWithButton.vue'
import BaseButton from '@/components/BaseButton.vue'
import { api } from '@/services/api.js'

const route = useRoute()
const router = useRouter()

const examId = computed(() => String(route.params.examId || '').trim())

const loading = ref(false)
const examLoading = ref(false)
const token = ref('')
const errorMessage = ref('')
const exam = ref(null)
const activeSessionId = ref('')

const examTitle = computed(() => exam.value?.title || 'Ruang Ujian')
const subjectName = computed(() => exam.value?.subject_name || '-')

const loadExamContext = async () => {
  if (!examId.value) return
  examLoading.value = true
  errorMessage.value = ''
  try {
    const listResp = await api.get('/api/v1/student/exams', { params: { limit: 100, offset: 0 } })
    const items = Array.isArray(listResp?.data?.data) ? listResp.data.data : []
    exam.value = items.find((it) => String(it?.id || '') === examId.value) || null
    if (!exam.value) {
      errorMessage.value = 'Ujian tidak ditemukan atau tidak tersedia untuk Anda.'
      return
    }
    try {
      const sessionResp = await api.get(`/api/v1/student/exams/${examId.value}/session`)
      activeSessionId.value = String(sessionResp?.data?.data?.id || '').trim()
    } catch {
      activeSessionId.value = ''
    }
  } catch (err) {
    errorMessage.value = err?.response?.data?.error?.message || 'Gagal memuat data ujian.'
  } finally {
    examLoading.value = false
  }
}

const storeSessionTokenGate = (sessionId) => {
  const sid = String(sessionId || '').trim()
  if (!sid) return
  try {
    localStorage.setItem(`mycbt_session_token_ok_${sid}`, '1')
  } catch {
    // ignore
  }
}

const submitToken = async () => {
  if (!examId.value) return
  const v = String(token.value || '').trim()
  if (!v) {
    errorMessage.value = 'Token ujian wajib diisi.'
    return
  }
  loading.value = true
  errorMessage.value = ''
  try {
    const resp = await api.post(`/api/v1/student/exams/${examId.value}/join`, { token: v })
    const sid = String(resp?.data?.data?.session_id || resp?.data?.data?.session?.id || '').trim()
    if (!sid) throw new Error('Session tidak ditemukan.')
    storeSessionTokenGate(sid)
    await router.replace(`/student/workspace/${sid}`)
  } catch (err) {
    errorMessage.value = err?.response?.data?.error?.message || err?.message || 'Gagal memverifikasi token.'
  } finally {
    loading.value = false
  }
}

const continueExistingSession = async () => {
  if (!activeSessionId.value) return
  storeSessionTokenGate(activeSessionId.value)
  await router.replace(`/student/workspace/${activeSessionId.value}`)
}

onMounted(() => {
  loadExamContext()
})
</script>

<template>
  <LayoutAuthenticated>
    <SectionMain>
      <SectionTitleLineWithButton :icon="mdiLockOutline" title="Verifikasi Token Ujian" main>
        <BaseButton :icon="mdiArrowLeft" color="lightDark" label="Kembali" @click="router.push('/student/ujian')" />
      </SectionTitleLineWithButton>

      <div class="mx-auto w-full max-w-2xl rounded-3xl border border-slate-100 bg-white p-8 shadow-sm dark:border-slate-800 dark:bg-slate-900">
        <div v-if="examLoading" class="py-12 text-center text-sm text-slate-500">Memuat data ujian...</div>

        <template v-else>
          <div class="mb-6">
            <p class="text-xs font-bold uppercase tracking-widest text-slate-400">Ujian</p>
            <h1 class="mt-2 text-xl font-black text-slate-800 dark:text-slate-100">{{ examTitle }}</h1>
            <p class="mt-1 text-sm text-slate-500 dark:text-slate-400">{{ subjectName }}</p>
          </div>

          <div class="rounded-2xl border border-slate-100 bg-slate-50 p-5 dark:border-slate-800 dark:bg-slate-800/40">
            <label class="mb-2 block text-xs font-bold uppercase tracking-widest text-slate-500">Token Ujian</label>
            <input
              v-model="token"
              type="text"
              placeholder="Masukkan token dari pengawas"
              class="w-full rounded-xl border border-slate-200 bg-white px-4 py-3 text-sm text-slate-800 outline-none transition focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 dark:border-slate-700 dark:bg-slate-900 dark:text-slate-100"
              autocomplete="off"
              spellcheck="false"
              @keydown.enter.prevent="submitToken"
            />
          </div>

          <div v-if="errorMessage" class="mt-4 rounded-xl border border-rose-100 bg-rose-50 px-4 py-3 text-sm text-rose-700 dark:border-rose-900/40 dark:bg-rose-900/20 dark:text-rose-300">
            {{ errorMessage }}
          </div>

          <div class="mt-6 flex flex-wrap items-center gap-3">
            <BaseButton
              :icon="mdiPlayCircleOutline"
              color="info"
              :label="loading ? 'Memeriksa Token...' : 'Masuk Ruang Ujian'"
              :disabled="loading || examLoading"
              @click="submitToken"
            />
            <BaseButton
              v-if="activeSessionId"
              color="success"
              label="Lanjutkan Session Aktif"
              :disabled="loading || examLoading"
              @click="continueExistingSession"
            />
          </div>
        </template>
      </div>
    </SectionMain>
  </LayoutAuthenticated>
</template>
