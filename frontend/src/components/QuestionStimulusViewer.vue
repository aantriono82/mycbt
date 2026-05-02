<template>
  <section class="space-y-4">
    <div class="flex flex-wrap items-center justify-between gap-3">
      <div class="flex items-center gap-3">
        <div
          v-if="title"
          class="px-3 py-1 rounded-lg text-[10px] font-black uppercase tracking-[0.2em]"
          :class="titleClass"
        >
          {{ title }}
        </div>
      </div>

      <div class="flex items-center gap-2 rounded-xl border border-slate-200 bg-white/95 px-2 py-1.5 shadow-sm">
        <div class="min-w-[52px] text-center text-xs font-bold text-slate-500">
          {{ zoomPercent }}%
        </div>
        <button
          type="button"
          class="h-9 w-9 rounded-lg border border-slate-300 bg-white text-slate-700 shadow-sm transition-colors hover:border-slate-400 hover:bg-slate-50 disabled:cursor-not-allowed disabled:opacity-40"
          :disabled="!canZoomOut"
          title="Perkecil gambar"
          @click="zoomOut"
        >
          <BaseIcon :path="mdiMagnifyMinus" size="18" class="mx-auto" />
        </button>
        <button
          type="button"
          class="h-9 w-9 rounded-lg border border-slate-300 bg-white text-slate-700 shadow-sm transition-colors hover:border-slate-400 hover:bg-slate-50 disabled:cursor-not-allowed disabled:opacity-40"
          :disabled="!canReset"
          title="Pas ke ukuran awal"
          @click="resetZoom"
        >
          <BaseIcon :path="mdiFitToScreen" size="18" class="mx-auto" />
        </button>
        <button
          type="button"
          class="h-9 w-9 rounded-lg border border-slate-300 bg-white text-slate-700 shadow-sm transition-colors hover:border-slate-400 hover:bg-slate-50 disabled:cursor-not-allowed disabled:opacity-40"
          :disabled="!canZoomIn"
          title="Perbesar gambar"
          @click="zoomIn"
        >
          <BaseIcon :path="mdiMagnifyPlus" size="18" class="mx-auto" />
        </button>
      </div>
    </div>

    <div ref="contentEl" class="stimulus-content" :class="contentClass">
      <slot />
    </div>
  </section>
</template>

<script setup>
import { computed, nextTick, ref, watch } from 'vue'
import { mdiFitToScreen, mdiMagnifyMinus, mdiMagnifyPlus } from '@mdi/js'
import BaseIcon from '@/components/BaseIcon.vue'

const props = defineProps({
  zoom: { type: Number, default: 1 },
  minZoom: { type: Number, default: 0.25 },
  maxZoom: { type: Number, default: 3 },
  step: { type: Number, default: 0.25 },
  contentKey: { type: [String, Number], default: '' },
  title: { type: String, default: 'STIMULUS / SOAL UTAMA' },
  titleClass: {
    type: String,
    default: 'bg-slate-100 text-slate-500',
  },
  contentClass: { type: String, default: '' },
})

const emit = defineEmits(['update:zoom'])

const contentEl = ref(null)

const roundZoom = (value) => Math.round(Number(value || 1) * 100) / 100
const clampZoom = (value) => {
  const n = Number(value || 1)
  return roundZoom(Math.min(props.maxZoom, Math.max(props.minZoom, n)))
}

const zoomPercent = computed(() => Math.round(clampZoom(props.zoom) * 100))
const canZoomOut = computed(() => clampZoom(props.zoom) > props.minZoom)
const canZoomIn = computed(() => clampZoom(props.zoom) < props.maxZoom)
const canReset = computed(() => Math.abs(clampZoom(props.zoom) - 1) > 0.001)

const updateZoom = (value) => {
  emit('update:zoom', clampZoom(value))
}

const zoomOut = () => updateZoom(clampZoom(props.zoom) - props.step)
const zoomIn = () => updateZoom(clampZoom(props.zoom) + props.step)
const resetZoom = () => updateZoom(1)

const clearManagedStyle = (img) => {
  img.style.removeProperty('width')
  img.style.removeProperty('max-width')
  img.style.removeProperty('height')
}

const syncImages = async () => {
  await nextTick()
  const root = contentEl.value
  if (!root) return

  const images = Array.from(root.querySelectorAll('img'))
  for (const img of images) {
    img.setAttribute('loading', 'lazy')
    if (!img.complete) {
      img.addEventListener('load', syncImages, { once: true })
    }
    clearManagedStyle(img)
  }

  if (Math.abs(clampZoom(props.zoom) - 1) <= 0.001) return

  for (const img of images) {
    const baseWidth = img.getBoundingClientRect().width
    if (!baseWidth) continue
    img.style.width = `${Math.max(baseWidth * clampZoom(props.zoom), 48)}px`
    img.style.maxWidth = 'none'
    img.style.height = 'auto'
  }
}

watch(
  () => [props.zoom, props.contentKey],
  () => {
    syncImages()
  },
  { immediate: true, flush: 'post' },
)
</script>

<style scoped>
.stimulus-content {
  min-width: 0;
}

.stimulus-content :deep(img) {
  display: block;
  cursor: zoom-in;
}
</style>
