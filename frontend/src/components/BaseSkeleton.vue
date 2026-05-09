<script setup>
defineProps({
  width: {
    type: String,
    default: 'w-full'
  },
  height: {
    type: String,
    default: 'h-4'
  },
  rounded: {
    type: String,
    default: 'rounded-lg'
  }
})
</script>

<template>
  <div 
    :class="[
      'animate-pulse relative overflow-hidden bg-slate-200 dark:bg-slate-700/50',
      width,
      height,
      rounded
    ]"
  >
    <div class="absolute inset-0 -translate-x-full bg-gradient-to-r from-transparent via-white/30 dark:via-white/10 to-transparent animate-shimmer"></div>
  </div>
</template>
