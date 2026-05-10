<script setup>
import { onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { mdiArrowLeft, mdiBookOpenPageVariant } from '@mdi/js'
import LayoutAuthenticated from '@/layouts/LayoutAuthenticated.vue'
import SectionMain from '@/components/SectionMain.vue'
import SectionTitleLineWithButton from '@/components/SectionTitleLineWithButton.vue'
import BaseButton from '@/components/BaseButton.vue'
import { api } from '@/services/api.js'

const route = useRoute()
const router = useRouter()
const isLoading = ref(false)
const errorMessage = ref('')
const discussion = ref(null)

const loadDiscussion = async () => {
  const examId = String(route.params.examId || '').trim()
  if (!examId) return
  isLoading.value = true
  errorMessage.value = ''
  try {
    const { data } = await api.get(`/api/v1/student/results/${examId}/discussion`)
    discussion.value = data?.data || null
  } catch (error) {
    discussion.value = null
    errorMessage.value = error?.response?.data?.error?.message || 'Gagal memuat pembahasan'
  } finally {
    isLoading.value = false
  }
}

const boolLabel = (value) => {
  if (value === true) return 'Benar'
  if (value === false) return 'Salah'
  return 'Dinilai Manual'
}

onMounted(loadDiscussion)
</script>

<template>
  <LayoutAuthenticated>
    <SectionMain>
      <SectionTitleLineWithButton :icon="mdiBookOpenPageVariant" title="Pembahasan Soal" main>
        <BaseButton :icon="mdiArrowLeft" color="info" label="Kembali ke Hasil" @click="router.push('/student/hasil')" />
      </SectionTitleLineWithButton>

      <div v-if="isLoading" class="rounded-xl border border-slate-200 bg-white px-4 py-3 text-sm text-slate-500 dark:border-slate-800 dark:bg-slate-900">
        Memuat pembahasan...
      </div>
      <div v-else-if="errorMessage" class="rounded-xl border border-red-100 bg-red-50 px-4 py-3 text-sm text-red-700 dark:border-red-900/40 dark:bg-red-900/20 dark:text-red-300">
        {{ errorMessage }}
      </div>
      <div v-else-if="discussion" class="space-y-4">
        <div class="rounded-xl border border-slate-200 bg-white p-4 dark:border-slate-800 dark:bg-slate-900">
          <div class="text-lg font-bold text-slate-800 dark:text-slate-100">{{ discussion.exam_title }}</div>
          <div class="mt-1 text-xs text-slate-500 dark:text-slate-400">
            Attempt {{ discussion.attempt_no }} dari {{ discussion.max_attempts }}
          </div>
        </div>
        <div
          v-for="item in discussion.items || []"
          :key="item.question_id"
          class="rounded-xl border border-slate-200 bg-white p-4 dark:border-slate-800 dark:bg-slate-900"
        >
          <div class="mb-2 flex items-center justify-between gap-3">
            <div class="text-sm font-bold text-slate-700 dark:text-slate-200">Soal {{ item.order_no }}</div>
            <div class="text-xs font-bold" :class="item.is_correct === true ? 'text-emerald-600' : item.is_correct === false ? 'text-rose-600' : 'text-amber-600'">
              {{ boolLabel(item.is_correct) }}
            </div>
          </div>
          <div class="prose max-w-none text-sm dark:prose-invert" v-html="item.stem"></div>
          <div class="mt-3 grid gap-3 md:grid-cols-2">
            <div class="rounded-lg border border-slate-200 bg-slate-50 p-3 dark:border-slate-700 dark:bg-slate-800/40">
              <div class="mb-1 text-[10px] font-black uppercase tracking-widest text-slate-500">Jawaban Siswa</div>
              <pre class="whitespace-pre-wrap break-words text-xs text-slate-700 dark:text-slate-300">{{ JSON.stringify(item.student_answer ?? null, null, 2) }}</pre>
            </div>
            <div class="rounded-lg border border-emerald-200 bg-emerald-50 p-3 dark:border-emerald-900/40 dark:bg-emerald-900/20">
              <div class="mb-1 text-[10px] font-black uppercase tracking-widest text-emerald-700 dark:text-emerald-300">Kunci Jawaban</div>
              <pre class="whitespace-pre-wrap break-words text-xs text-emerald-800 dark:text-emerald-200">{{ JSON.stringify(item.correct_answer ?? null, null, 2) }}</pre>
            </div>
          </div>
          <div v-if="item.explanation" class="mt-3 rounded-lg border border-indigo-200 bg-indigo-50 p-3 dark:border-indigo-900/40 dark:bg-indigo-900/20">
            <div class="mb-1 text-[10px] font-black uppercase tracking-widest text-indigo-700 dark:text-indigo-300">Pembahasan</div>
            <div class="prose max-w-none text-sm dark:prose-invert" v-html="item.explanation"></div>
          </div>
        </div>
      </div>
    </SectionMain>
  </LayoutAuthenticated>
</template>
