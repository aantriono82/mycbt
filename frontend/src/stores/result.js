import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { api } from '@/services/api.js'

export const useResultStore = defineStore('result', () => {
    const results = ref([])
    const isLoading = ref(false)
    const errorMessage = ref('')

    const loadResults = async (limit = 50, offset = 0) => {
        isLoading.value = true
        errorMessage.value = ''
        try {
            const { data } = await api.get('/api/v1/student/results', { params: { limit, offset } })
            results.value = data?.data || []
        } catch (error) {
            results.value = []
            errorMessage.value = error?.response?.data?.error?.message || 'Gagal memuat hasil ujian dari backend'
        } finally {
            isLoading.value = false
        }
    }

    const averageScore = computed(() => {
        if (!results.value.length) return 0
        const total = results.value.reduce((sum, item) => sum + Number(item.score || 0), 0)
        return Math.round(total / results.value.length)
    })

    const totalExams = computed(() => results.value.length)

    return {
        results,
        isLoading,
        errorMessage,
        loadResults,
        averageScore,
        totalExams
    }
})
