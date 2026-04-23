import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useNotificationStore = defineStore('notification', () => {
    const items = ref([])

    const add = (message, color = 'info', icon = null, timeout = 5000) => {
        const id = Date.now()
        items.value.push({ id, message, color, icon, timeout })

        if (timeout > 0) {
            setTimeout(() => {
                remove(id)
            }, timeout)
        }
    }

    const remove = (id) => {
        const idx = items.value.findIndex(i => i.id === id)
        if (idx >= 0) items.value.splice(idx, 1)
    }

    const pushError = (message) => add(message, 'danger', null, 7000)
    const pushSuccess = (message) => add(message, 'success', null, 5000)
    const pushWarning = (message) => add(message, 'warning', null, 6000)

    return { items, add, remove, pushError, pushSuccess, pushWarning }
})
