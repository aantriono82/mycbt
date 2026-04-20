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

    <div class="grid gap-3 md:grid-cols-2 lg:grid-cols-3">
      <div 
        v-for="s in suggestions" 
        :key="s.question_id"
        class="p-3 rounded-xl border bg-white dark:bg-slate-900 shadow-sm flex flex-col h-full"
        :class="
          s.priority === 'high' 
            ? 'border-red-200 dark:border-red-900/40 bg-red-50/10' 
            : 'border-slate-200 dark:border-slate-800'
        "
      >
        <div class="flex items-start gap-2 mb-2">
          <div 
            class="p-1.5 rounded-lg shrink-0"
            :class="s.priority === 'high' ? 'bg-red-100 text-red-600' : 'bg-slate-100 text-slate-500 dark:bg-slate-800'"
          >
            <BaseIcon :path="s.priority === 'high' ? mdiAlertDecagram : mdiFlag" size="18" />
          </div>
          <div>
            <div class="text-[10px] font-bold uppercase text-slate-400 tracking-tighter">Butir ID: {{ s.question_id.split('-')[0] }}...</div>
            <div class="text-xs font-bold dark:text-slate-200" v-for="(tip, idx) in s.tips" :key="idx">
              {{ tip }}
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
