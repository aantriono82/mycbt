<script setup>
import { containerMaxW } from '@/config.js'
</script>

<template>
  <section class="px-4 sm:px-6 md:px-10 pt-2 sm:pt-3 md:pt-4 pb-4 sm:pb-6 md:pb-10" :class="containerMaxW">
    <slot />
  </section>
</template>
