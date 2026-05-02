<script setup>
import { computed } from 'vue'
import { useAuthStore } from '@/stores/auth.js'
import UserAvatar from '@/components/UserAvatar.vue'
import { resolveBackendAssetUrl } from '@/utils/assetUrl.js'

const authStore = useAuthStore()

const username = computed(() => authStore.userDisplayName || 'Pengguna')

const avatarUrl = computed(() => {
  const user = authStore.user
  if (!user) return null

  const photo = user.photo_url || user.photo || user.avatar
  return photo ? resolveBackendAssetUrl(photo) : null
})
</script>

<template>
  <UserAvatar :username="username" :avatar="avatarUrl">
    <slot />
  </UserAvatar>
</template>
