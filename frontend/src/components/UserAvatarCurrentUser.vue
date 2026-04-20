<script setup>
import { computed } from 'vue'
import { useAuthStore } from '@/stores/auth.js'
import UserAvatar from '@/components/UserAvatar.vue'

const authStore = useAuthStore()

const username = computed(() => authStore.userDisplayName || 'Pengguna')

const avatarUrl = computed(() => {
  const user = authStore.user
  if (!user) return null
  
  const photo = user.photo_url || user.photo || user.avatar
  if (photo) {
    if (photo.startsWith('/uploads')) {
      const baseUrl = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'
      return `${baseUrl}${photo}`
    }
    return photo
  }
  return null
})
</script>

<template>
  <UserAvatar :username="username" :avatar="avatarUrl">
    <slot />
  </UserAvatar>
</template>
