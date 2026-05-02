<script setup>
defineProps({
  noPadding: Boolean,
})
</script>

<template>
  <div class="min-w-0 max-w-full flex-1" :class="{ 'p-6 md:p-8': !noPadding }">
    <slot />
  </div>
</template>
