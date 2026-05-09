<script setup>
import BaseButton from '@/components/BaseButton.vue'

defineProps({
  title: {
    type: String,
    default: 'Data Kosong'
  },
  description: {
    type: String,
    default: 'Belum ada data yang bisa ditampilkan di sini.'
  },
  buttonLabel: {
    type: String,
    default: ''
  },
  buttonTo: {
    type: [String, Object],
    default: ''
  },
  buttonIcon: {
    type: String,
    default: ''
  }
})
</script>

<template>
  <div class="flex flex-col items-center justify-center py-16 px-6 text-center animate-fade-in">
    <!-- Premium SVG Illustration -->
    <div class="relative mb-8 group">
      <div class="absolute inset-0 bg-blue-500/10 blur-3xl rounded-full group-hover:bg-blue-500/20 transition-colors duration-500"></div>
      <svg
        class="relative w-48 h-48 sm:w-64 sm:h-64 text-slate-200 dark:text-slate-800 transition-transform duration-700 group-hover:scale-110"
        viewBox="0 0 200 200"
        fill="none"
        xmlns="http://www.w3.org/2000/svg"
      >
        <circle cx="100" cy="100" r="80" class="fill-slate-50 dark:fill-slate-900/50" />
        <path
          d="M140 70H60C54.4772 70 50 74.4772 50 80V120C50 125.523 54.4772 130 60 130H140C145.523 130 150 125.523 150 120V80C150 74.4772 145.523 70 140 70Z"
          stroke="currentColor"
          stroke-width="2"
          stroke-linecap="round"
          stroke-linejoin="round"
          class="text-blue-500/30"
        />
        <path
          d="M80 100H120"
          stroke="currentColor"
          stroke-width="3"
          stroke-linecap="round"
          class="text-blue-500/40"
        />
        <path
          d="M90 115H110"
          stroke="currentColor"
          stroke-width="3"
          stroke-linecap="round"
          class="text-blue-500/20"
        />
        <circle cx="100" cy="85" r="4" fill="currentColor" class="text-blue-500" />
        
        <!-- Decorative elements -->
        <circle cx="160" cy="60" r="6" class="fill-indigo-400/20 animate-pulse" />
        <circle cx="40" cy="140" r="4" class="fill-blue-400/20" />
        <rect x="170" y="130" width="8" height="8" rx="2" class="fill-purple-400/20 animate-bounce" style="animation-duration: 3s;" />
      </svg>
    </div>

    <h3 class="text-2xl font-black text-slate-800 dark:text-slate-100 mb-2 uppercase tracking-tight">
      {{ title }}
    </h3>
    <p class="text-slate-500 dark:text-slate-400 max-w-sm mx-auto mb-8 font-medium leading-relaxed">
      {{ description }}
    </p>

    <BaseButton
      v-if="buttonLabel"
      :label="buttonLabel"
      :icon="buttonIcon"
      :to="buttonTo"
      color="info"
      rounded-full
      class="px-8 shadow-lg shadow-blue-500/20 transition-all hover:scale-105 active:scale-95"
    />
  </div>
</template>

<style scoped>
.animate-fade-in {
  animation: fadeIn 0.8s cubic-bezier(0.16, 1, 0.3, 1);
}
@keyframes fadeIn {
  from { opacity: 0; transform: translateY(20px); }
  to { opacity: 1; transform: translateY(0); }
}
</style>
