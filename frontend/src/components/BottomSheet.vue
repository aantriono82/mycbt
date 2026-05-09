<script setup>
import { ref, onMounted, onUnmounted, watch } from 'vue'
import { mdiClose } from '@mdi/js'
import BaseIcon from '@/components/BaseIcon.vue'

const props = defineProps({
  modelValue: {
    type: Boolean,
    default: false
  },
  title: {
    type: String,
    default: 'Menu'
  }
})

const emit = defineEmits(['update:modelValue'])

const isVisible = ref(false)
const isAnimating = ref(false)

const close = () => {
  isVisible.value = false
  setTimeout(() => {
    emit('update:modelValue', false)
  }, 300)
}

watch(() => props.modelValue, (val) => {
  if (val) {
    isAnimating.value = true
    setTimeout(() => {
      isVisible.value = true
    }, 10)
  } else {
    isVisible.value = false
  }
})

const handleKeyDown = (e) => {
  if (e.key === 'Escape' && props.modelValue) {
    close()
  }
}

onMounted(() => {
  window.addEventListener('keydown', handleKeyDown)
  if (props.modelValue) {
    isVisible.value = true
  }
})

onUnmounted(() => {
  window.removeEventListener('keydown', handleKeyDown)
})
</script>

<template>
  <Teleport to="body">
    <div v-if="modelValue" class="fixed inset-0 z-[200] flex items-end justify-center">
      <!-- Backdrop -->
      <div 
        class="absolute inset-0 bg-slate-900/60 backdrop-blur-sm transition-opacity duration-300"
        :class="isVisible ? 'opacity-100' : 'opacity-0'"
        @click="close"
      ></div>

      <!-- Sheet -->
      <div 
        class="relative w-full max-w-2xl bg-white dark:bg-slate-900 rounded-t-[2.5rem] shadow-2xl transition-transform duration-300 ease-out flex flex-col max-h-[85vh]"
        :class="isVisible ? 'translate-y-0' : 'translate-y-full'"
      >
        <!-- Handle -->
        <div class="flex justify-center pt-3 pb-2 cursor-pointer" @click="close">
          <div class="w-12 h-1.5 rounded-full bg-slate-200 dark:bg-slate-700"></div>
        </div>

        <!-- Header -->
        <div class="px-6 py-4 border-b border-slate-100 dark:border-slate-800 flex items-center justify-between shrink-0">
          <h3 class="text-xs font-black uppercase tracking-[0.2em] text-slate-500 dark:text-slate-400">
            {{ title }}
          </h3>
          <button 
            type="button" 
            class="p-2 rounded-xl bg-slate-50 dark:bg-slate-800 text-slate-400 hover:text-slate-600 transition-colors"
            @click="close"
          >
            <BaseIcon :path="mdiClose" size="20" />
          </button>
        </div>

        <!-- Content -->
        <div class="flex-1 overflow-y-auto p-6 scrollbar-hide">
          <slot />
        </div>

        <!-- Safe Area Padding -->
        <div class="h-safe-bottom"></div>
      </div>
    </div>
  </Teleport>
</template>

<style scoped>
.h-safe-bottom {
  height: env(safe-area-inset-bottom, 1rem);
}
.scrollbar-hide::-webkit-scrollbar {
  display: none;
}
.scrollbar-hide {
  -ms-overflow-style: none;
  scrollbar-width: none;
}
</style>
