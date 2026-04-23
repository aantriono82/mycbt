<script setup>
import { RouterView } from 'vue-router'
import GlobalToast from '@/components/GlobalToast.vue'
import BackendHealthBanner from '@/components/BackendHealthBanner.vue'
</script>

<template>
  <BackendHealthBanner />
  <RouterView v-slot="{ Component }">
    <transition
      name="page"
      mode="out-in"
    >
      <component :is="Component" />
    </transition>
  </RouterView>
  <GlobalToast />
</template>
