<script setup>
import { onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { mdiLinkVariant, mdiCheck, mdiAlert } from '@mdi/js'
import SectionMain from '@/components/SectionMain.vue'
import CardBox from '@/components/CardBox.vue'
import BaseButton from '@/components/BaseButton.vue'
import BaseIcon from '@/components/BaseIcon.vue'
import { api } from '@/services/api.js'
import { useAuthStore } from '@/stores/auth.js'

const route = useRoute()
const authStore = useAuthStore()

const exams = ref([])
const isLoading = ref(true)
const errorMessage = ref('')
const sessionId = ref(route.query.session_id || '')
const token = ref(route.query.token || '')

const loadExams = async () => {
  isLoading.value = true
  try {
    const { data } = await api.get('/api/v1/exams', {
      params: { limit: 100, offset: 0 }
    })
    exams.value = data?.data || []
  } catch (err) {
    errorMessage.value = 'Gagal memuat daftar ujian'
  } finally {
    isLoading.value = false
  }
}

const selectExam = async (exam) => {
  try {
    const { data } = await api.post('/api/v1/lti/deep-link', {
      session_id: sessionId.value,
      exam_id: exam.id,
      title: exam.title
    })

    // Automatic FORM POST back to LMS
    const form = document.createElement('form')
    form.method = 'POST'
    form.action = data.data.return_url
    
    const jwtInput = document.createElement('input')
    jwtInput.type = 'hidden'
    jwtInput.name = 'JWT'
    jwtInput.value = data.data.jwt
    form.appendChild(jwtInput)
    
    document.body.appendChild(form)
    form.submit()
  } catch (err) {
    alert('Gagal mengirim pilihan ke LMS: ' + (err.response?.data?.error?.message || err.message))
  }
}

onMounted(async () => {
  // If we have a token in query, login with it first
  if (token.value) {
    const ok = await authStore.loginWithToken(token.value)
    if (!ok) {
      errorMessage.value = 'Sesi LTI tidak valid atau telah berakhir'
      isLoading.value = false
      return
    }
  }
  
  if (!sessionId.value) {
    errorMessage.value = 'Session ID LTI diperlukan'
    isLoading.value = false
    return
  }
  
  await loadExams()
})
</script>

<template>
  <div class="min-h-screen bg-slate-50 dark:bg-slate-900 p-6">
    <div class="max-w-4xl mx-auto">
      <div class="flex items-center justify-between mb-8">
        <div>
          <h1 class="text-3xl font-black text-slate-800 dark:text-white flex items-center gap-3 uppercase tracking-tight">
            <BaseIcon :path="mdiLinkVariant" size="32" class="text-blue-600" />
            LTI Resource Picker
          </h1>
          <p class="text-slate-500 mt-2">Pilih jadwal ujian yang ingin dibagikan ke LMS (Moodle/Canvas).</p>
        </div>
      </div>

      <div v-if="isLoading" class="flex flex-col items-center justify-center p-20 animate-pulse">
        <div class="h-12 w-12 border-4 border-blue-600 border-t-transparent rounded-full animate-spin mb-4"></div>
        <div class="text-slate-400 font-bold uppercase tracking-widest text-xs">Memuat daftar ujian...</div>
      </div>

      <div v-else-if="errorMessage" class="bg-red-50 border border-red-200 rounded-2xl p-8 text-center text-red-700">
        <BaseIcon :path="mdiAlert" size="48" class="mx-auto mb-4" />
        <h3 class="text-xl font-bold mb-2">Terjadi Kesalahan</h3>
        <p>{{ errorMessage }}</p>
      </div>

      <div v-else class="grid gap-4">
        <CardBox 
          v-for="exam in exams" 
          :key="exam.id"
          class="hover:shadow-xl transition-all cursor-pointer group"
          @click="selectExam(exam)"
        >
          <div class="flex items-center justify-between">
            <div class="flex-1">
              <div class="flex items-center gap-2 mb-1">
                <span class="text-[10px] bg-blue-100 text-blue-700 px-2 py-0.5 rounded-full font-black uppercase tracking-tighter">
                  {{ exam.status }}
                </span>
                <span v-if="exam.session_name" class="text-[10px] bg-slate-100 text-slate-600 px-2 py-0.5 rounded-full font-bold">
                  {{ exam.session_name }}
                </span>
              </div>
              <h3 class="text-xl font-black text-slate-800 dark:text-slate-100 group-hover:text-blue-600 transition-colors uppercase">
                {{ exam.title }}
              </h3>
              <div class="text-sm text-slate-500 mt-1 flex items-center gap-4">
                <span>Mapel: <strong>{{ exam.subject_name || exam.subject_id }}</strong></span>
                <span>•</span>
                <span>Mulai: {{ new Date(exam.starts_at).toLocaleString() }}</span>
              </div>
            </div>
            <div>
              <BaseButton :icon="mdiCheck" color="info" label="Pilih Ujian" rounded-full />
            </div>
          </div>
        </CardBox>

        <div v-if="!exams.length" class="bg-white border-2 border-dashed border-slate-200 rounded-3xl p-16 text-center text-slate-400">
          Belum ada jadwal ujian yang tersedia. Silakan buat jadwal ujian terlebih dahulu di dashboard.
        </div>
      </div>
    </div>
  </div>
</template>
