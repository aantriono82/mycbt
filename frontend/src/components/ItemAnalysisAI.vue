<script setup>
import { computed, watch, ref } from 'vue'
import { mdiAutoFix, mdiAlertDecagram, mdiInformation, mdiFlag } from '@mdi/js'
import CardBox from '@/components/CardBox.vue'
import BaseIcon from '@/components/BaseIcon.vue'
import { api } from '@/services/api.js'

const props = defineProps({
  examId: {
    type: String,
    required: true
  }
})

const suggestions = ref([])
const isLoading = ref(false)

const loadSuggestions = async () => {
  if (!props.examId) return
  isLoading.value = true
  try {
    const { data } = await api.get(`/api/v1/exams/${props.examId}/item-analysis/suggestions`)
    suggestions.value = data?.data || []
  } catch (error) {
    suggestions.value = []
  } finally {
    isLoading.value = false
  }
}

watch(() => props.examId, loadSuggestions, { immediate: true })

const highPriorityCount = computed(() => suggestions.value.filter(s => s.priority === 'high').length)
</script>

<template>
  <div v-if="suggestions.length" class="mb-6 space-y-4">
    <div class="flex items-center gap-2 mb-2">
      <BaseIcon :path="mdiAutoFix" class="text-info w-6 h-6" />
      <h3 class="text-lg font-bold dark:text-slate-100">AI-Powered Item Analysis Suggestions</h3>
      <span v-if="highPriorityCount" class="bg-red-500 text-white text-[10px] font-bold px-1.5 py-0.5 rounded-full animate-bounce">
        {{ highPriorityCount }} Critical Fixes
      </span>
    </div>

    <div class="flex flex-wrap gap-3">
      <div 
        v-for="(s, idx) in suggestions" 
        :key="s.question_id"
        class="group relative flex items-center gap-2 p-2 px-3 rounded-2xl border transition-all hover:scale-105 cursor-help"
        :class="
          s.priority === 'high' 
            ? 'bg-red-50 dark:bg-red-900/20 border-red-200 dark:border-red-800 text-red-700 dark:text-red-400 shadow-sm shadow-red-100 dark:shadow-none' 
            : 'bg-slate-50 dark:bg-slate-800/50 border-slate-200 dark:border-slate-700 text-slate-600 dark:text-slate-400'
        "
      >
        <div 
          class="flex items-center justify-center w-6 h-6 rounded-lg font-black text-xs"
          :class="s.priority === 'high' ? 'bg-red-500 text-white' : 'bg-slate-200 dark:bg-slate-700 text-slate-600 dark:text-slate-400'"
        >
          {{ idx + 1 }}
        </div>
        <div class="text-[10px] font-bold uppercase tracking-tight truncate max-w-[80px]">
          Butir {{ s.question_id.split('-')[0] }}
        </div>
        
        <!-- Tooltip Hint (Optional but nice for UX) -->
        <div class="absolute bottom-full mb-2 left-1/2 -translate-x-1/2 w-48 p-2 bg-slate-900 text-white text-[10px] rounded-lg shadow-xl opacity-0 group-hover:opacity-100 pointer-events-none transition-opacity z-50">
          <div class="font-black mb-1 text-sky-400">SARAN AI:</div>
          {{ s.tips.join('. ') }}
          <div class="absolute top-full left-1/2 -translate-x-1/2 border-8 border-transparent border-t-slate-900"></div>
        </div>
      </div>
    </div>
  </div>
</template>
