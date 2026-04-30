<script setup>
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'

const props = defineProps({
  question: {
    type: Object,
    required: true,
  },
  modelValue: {
    type: Object,
    default: () => ({}),
  },
})

const emit = defineEmits(['update:modelValue'])

const matchingBoardEl = ref(null)
const linesSvgEl = ref(null)
const activeMatchingLeftId = ref('')
const internalPairs = ref({})
let matchingRaf = 0

const stripHtml = (html) => {
  const raw = String(html || '')
  if (!raw.trim()) return ''
  try {
    const doc = new DOMParser().parseFromString(raw, 'text/html')
    return String(doc.body.textContent || '')
      .replace(/\u00a0/g, ' ')
      .replace(/&nbsp;/gi, ' ')
      .replace(/\s+/g, ' ')
      .trim()
  } catch {
    return raw.replace(/<[^>]*>/g, ' ').replace(/&nbsp;/gi, ' ').replace(/\s+/g, ' ').trim()
  }
}

const renderHtml = (html) => {
  const raw = String(html || '')
  if (!raw.trim()) return '<p><em>(Konten soal kosong)</em></p>'
  try {
    const doc = new DOMParser().parseFromString(raw, 'text/html')
    doc.querySelectorAll('script, iframe, object, embed').forEach((n) => n.remove())
    doc.querySelectorAll('*').forEach((el) => {
      for (const attr of Array.from(el.attributes || [])) {
        if (/^on/i.test(attr.name)) el.removeAttribute(attr.name)
      }
    })
    return doc.body.innerHTML
  } catch {
    return raw
  }
}

const escapeAttrSelector = (value) => {
  const raw = String(value ?? '')
  if (typeof CSS !== 'undefined' && typeof CSS.escape === 'function') return CSS.escape(raw)
  return raw.replace(/["\\]/g, '\\$&')
}

const matchingLeftItems = computed(() => (Array.isArray(props.question?.pairs) ? props.question.pairs : []))
const normalizedLeftItems = computed(() => {
  if (Array.isArray(props.question?.pairs)) {
    return props.question.pairs
      .map((pair, index) => ({
        id: String(pair?.id || ''),
        content: pair?.left_content || '',
        orderNo: pair?.order_no ?? index + 1,
      }))
      .filter((item) => stripHtml(item.content))
  }
  if (Array.isArray(props.question?.matching_left)) {
    return props.question.matching_left
      .map((item, index) => ({
        id: String(item?.id || ''),
        content: item?.content || '',
        orderNo: item?.order_no ?? index + 1,
      }))
      .filter((item) => stripHtml(item.content))
  }
  return []
})
const matchingRightItems = computed(() =>
  Array.isArray(props.question?.matching_right)
    ? props.question.matching_right.map((item) => ({
        id: String(item?.id || ''),
        text: stripHtml(item?.content || ''),
      }))
    : matchingLeftItems.value.map((pair) => ({
        id: String(pair?.id || ''),
        text: stripHtml(pair?.right_content || ''),
      })),
)

const matchingPairsMap = computed(() => {
  const raw = internalPairs.value
  if (!raw || typeof raw !== 'object' || Array.isArray(raw)) return {}
  return raw
})

const updatePairs = (nextMap) => {
  internalPairs.value = { ...nextMap }
  emit('update:modelValue', { ...nextMap })
}

const setMatching = (leftPairId, rightPickId) => {
  const pid = String(leftPairId || '')
  if (!pid) return
  const currentMap = { ...matchingPairsMap.value }
  const rightId = String(rightPickId || '')
  if (!rightId) {
    delete currentMap[pid]
    updatePairs(currentMap)
    return
  }
  for (const [leftId, selectedRight] of Object.entries(currentMap)) {
    if (leftId !== pid && String(selectedRight || '') === rightId) delete currentMap[leftId]
  }
  currentMap[pid] = rightId
  updatePairs(currentMap)
}

const setActiveMatchingLeft = (leftId) => {
  activeMatchingLeftId.value = String(leftId || '').trim()
}

const onMatchingRightClick = (rightId) => {
  let leftId = String(activeMatchingLeftId.value || '').trim()
  if (!leftId) {
    const firstUnpaired = normalizedLeftItems.value.find((item) => !matchingPairsMap.value?.[String(item?.id || '')])
    if (firstUnpaired?.id != null) leftId = String(firstUnpaired.id)
  }
  if (!leftId) return
  setMatching(leftId, rightId)
  activeMatchingLeftId.value = ''
}

const clearMatchingForLeft = (leftId) => {
  setMatching(leftId, '')
  if (activeMatchingLeftId.value === String(leftId || '')) activeMatchingLeftId.value = ''
}

const isMatchingLeftActive = (leftId) => String(activeMatchingLeftId.value) === String(leftId || '')
const isMatchingLeftPaired = (leftId) => !!matchingPairsMap.value?.[String(leftId || '')]
const isMatchingRightUsed = (rightId) => Object.values(matchingPairsMap.value || {}).some((v) => String(v) === String(rightId || ''))
const activeLeftLabel = computed(() => {
  const activeId = String(activeMatchingLeftId.value || '')
  if (!activeId) return ''
  const item = normalizedLeftItems.value.find((entry) => String(entry.id) === activeId)
  return item ? stripHtml(item.content) : ''
})

const clearRenderedLines = () => {
  const svg = linesSvgEl.value
  if (!svg) return
  while (svg.firstChild) svg.removeChild(svg.firstChild)
}

const appendSvgEl = (parent, tag, attrs) => {
  const el = document.createElementNS('http://www.w3.org/2000/svg', tag)
  for (const [key, value] of Object.entries(attrs)) el.setAttribute(key, String(value))
  parent.appendChild(el)
}

const scheduleMatchingLines = async () => {
  if (matchingRaf) cancelAnimationFrame(matchingRaf)
  await nextTick()
  matchingRaf = requestAnimationFrame(() => {
    try {
      const root = matchingBoardEl.value
      const svg = linesSvgEl.value
      if (!root || !svg) {
        return
      }
      clearRenderedLines()
      const rootRect = root.getBoundingClientRect()
      for (const [leftId, rightId] of Object.entries(matchingPairsMap.value || {})) {
        const leftEl = root.querySelector(`[data-left-id="${escapeAttrSelector(leftId)}"]`)
        const rightEl = root.querySelector(`[data-right-id="${escapeAttrSelector(rightId)}"]`)
        if (!leftEl || !rightEl) continue
        const lRect = leftEl.getBoundingClientRect()
        const rRect = rightEl.getBoundingClientRect()
        const x1 = lRect.right - rootRect.left
        const y1 = lRect.top + lRect.height / 2 - rootRect.top
        const x2 = rRect.left - rootRect.left
        const y2 = rRect.top + rRect.height / 2 - rootRect.top
        const bend = Math.max(36, Math.min(120, (x2 - x1) * 0.35))
        const d = `M ${x1} ${y1} C ${x1 + bend} ${y1}, ${x2 - bend} ${y2}, ${x2} ${y2}`
        const active = String(activeMatchingLeftId.value || '') === String(leftId)
        appendSvgEl(svg, 'path', {
          d,
          fill: 'none',
          stroke: active ? '#22c55e' : '#34d399',
          'stroke-width': active ? 7 : 6,
          'stroke-linecap': 'round',
          'stroke-opacity': '0.16',
        })
        appendSvgEl(svg, 'path', {
          d,
          fill: 'none',
          stroke: '#10b981',
          'stroke-width': active ? 3 : 2.5,
          'stroke-linecap': 'round',
        })
        appendSvgEl(svg, 'circle', { cx: x1, cy: y1, r: 5, fill: '#10b981' })
        appendSvgEl(svg, 'circle', { cx: x2, cy: y2, r: 5, fill: '#10b981' })
      }
    } catch {
      clearRenderedLines()
    }
  })
}

watch(
  () => props.question?.id,
  () => {
    internalPairs.value = props.modelValue && typeof props.modelValue === 'object' && !Array.isArray(props.modelValue) ? { ...props.modelValue } : {}
    activeMatchingLeftId.value = ''
    scheduleMatchingLines()
  },
  { immediate: true },
)

watch(
  () => props.modelValue,
  (value) => {
    internalPairs.value = value && typeof value === 'object' && !Array.isArray(value) ? { ...value } : {}
    scheduleMatchingLines()
  },
  { deep: true },
)

watch(matchingPairsMap, scheduleMatchingLines, { deep: true })

watch(activeMatchingLeftId, () => {
  scheduleMatchingLines()
})

onMounted(() => {
  window.addEventListener('resize', scheduleMatchingLines)
  window.addEventListener('scroll', scheduleMatchingLines, true)
  scheduleMatchingLines()
})

onUnmounted(() => {
  window.removeEventListener('resize', scheduleMatchingLines)
  window.removeEventListener('scroll', scheduleMatchingLines, true)
  if (matchingRaf) cancelAnimationFrame(matchingRaf)
  clearRenderedLines()
})
</script>

<template>
  <div ref="matchingBoardEl" class="relative bg-white rounded-2xl border-2 border-slate-100 shadow-inner overflow-hidden">
    <div class="px-6 md:px-8 py-5 border-b border-slate-100 bg-slate-50/70 flex items-center justify-between">
      <div class="text-[10px] font-black uppercase tracking-[0.3em] text-slate-500">MENJODOHKAN</div>
      <div class="hidden sm:block text-[10px] font-black uppercase tracking-widest text-slate-400">Pilih pasangan yang sesuai</div>
    </div>

    <div class="hidden md:block absolute inset-0 pointer-events-none" aria-hidden="true">
      <svg ref="linesSvgEl" class="w-full h-full overflow-visible"></svg>
    </div>

    <div class="p-5 md:p-8">
      <div class="hidden md:grid md:grid-cols-[1fr_1.05fr] gap-10 items-start">
        <div class="space-y-4">
          <div class="text-[10px] font-black uppercase tracking-widest text-slate-400">Pernyataan</div>
          <div
            v-for="left in normalizedLeftItems"
            :key="left.id"
            role="button"
            tabindex="0"
            :data-left-id="left.id"
            @click="setActiveMatchingLeft(left.id)"
            @keydown.enter.prevent="setActiveMatchingLeft(left.id)"
            @keydown.space.prevent="setActiveMatchingLeft(left.id)"
            class="group w-full min-h-[76px] text-left px-5 py-4 rounded-xl border-2 transition-all bg-white relative"
            :class="[
              isMatchingLeftActive(left.id) ? 'border-emerald-400 bg-emerald-50/70 ring-2 ring-emerald-100' : 'border-slate-100 hover:border-emerald-200',
              isMatchingLeftPaired(left.id) ? 'border-emerald-300 bg-emerald-50/60 shadow-sm' : ''
            ]"
          >
            <div class="flex items-start gap-3">
              <span
                class="mt-0.5 flex h-6 w-6 shrink-0 items-center justify-center rounded-full text-[11px] font-black"
                :class="isMatchingLeftPaired(left.id) || isMatchingLeftActive(left.id) ? 'bg-emerald-500 text-white' : 'bg-slate-100 text-slate-500'"
              >
                {{ left.orderNo }}
              </span>
              <div class="min-w-0 flex-1">
                <div class="font-bold text-slate-900 leading-snug" v-html="renderHtml(left.content)"></div>
                <div v-if="matchingPairsMap[left.id]" class="mt-2 flex items-center gap-2 text-[11px] font-bold text-emerald-700">
                  <span>{{ matchingRightItems.find((item) => item.id === matchingPairsMap[left.id])?.text || 'Terpasang' }}</span>
                  <span
                    role="button"
                    tabindex="0"
                    class="text-rose-500 hover:text-rose-600"
                    @click.stop="clearMatchingForLeft(left.id)"
                    @keydown.enter.prevent="clearMatchingForLeft(left.id)"
                    @keydown.space.prevent="clearMatchingForLeft(left.id)"
                  >
                    x
                  </span>
                </div>
              </div>
            </div>
          </div>
        </div>

        <div class="space-y-4">
          <div class="text-[10px] font-black uppercase tracking-widest text-slate-400">Pilihan Jawaban</div>
          <div
            v-for="opt in matchingRightItems"
            :key="opt.id"
            role="button"
            tabindex="0"
            :data-right-id="opt.id"
            class="w-full min-h-[58px] text-left px-5 py-4 rounded-xl border-2 transition-all bg-white cursor-pointer"
            :class="isMatchingRightUsed(opt.id) ? 'border-emerald-200 bg-emerald-50/60 text-emerald-800' : 'border-slate-100 hover:border-emerald-200'"
            @click="onMatchingRightClick(opt.id)"
            @keydown.enter.prevent="onMatchingRightClick(opt.id)"
            @keydown.space.prevent="onMatchingRightClick(opt.id)"
          >
            <div class="flex items-center justify-between gap-3">
              <span class="font-bold leading-snug">{{ opt.text || '(kosong)' }}</span>
              <span v-if="isMatchingRightUsed(opt.id)" class="text-emerald-500 font-black">✓</span>
              <span v-else class="text-slate-200 text-2xl leading-none">›</span>
            </div>
          </div>
          <div class="text-[10px] text-slate-400 font-semibold italic text-center">Klik pernyataan dulu</div>
        </div>
      </div>

      <div class="space-y-4 md:hidden">
        <div
          v-for="pair in normalizedLeftItems"
          :key="pair.id"
          role="button"
          tabindex="0"
          :data-left-id="pair.id"
          @click="setActiveMatchingLeft(pair.id)"
          @keydown.enter.prevent="setActiveMatchingLeft(pair.id)"
          @keydown.space.prevent="setActiveMatchingLeft(pair.id)"
          class="px-5 py-4 rounded-2xl border-2 transition-all bg-white"
          :class="[
            isMatchingLeftActive(pair.id) ? 'border-emerald-400 bg-emerald-50/70 ring-2 ring-emerald-100' : 'border-slate-100',
            isMatchingLeftPaired(pair.id) ? 'border-emerald-300 bg-emerald-50/60' : ''
          ]"
        >
          <div class="flex items-start gap-3">
            <span
              class="mt-0.5 flex h-7 w-7 shrink-0 items-center justify-center rounded-full text-[12px] font-black"
              :class="isMatchingLeftPaired(pair.id) || isMatchingLeftActive(pair.id) ? 'bg-emerald-500 text-white' : 'bg-slate-100 text-slate-500'"
            >
              {{ pair.orderNo }}
            </span>
            <div class="min-w-0 flex-1">
              <div class="font-bold text-slate-900 leading-snug" v-html="renderHtml(pair.content)"></div>
              <div
                v-if="matchingPairsMap[pair.id]"
                class="mt-2 inline-flex items-center gap-2 rounded-full bg-emerald-100 px-3 py-1 text-[12px] font-bold text-emerald-700"
              >
                <span>{{ matchingRightItems.find((item) => item.id === matchingPairsMap[pair.id])?.text || 'Terpasang' }}</span>
                <span
                  role="button"
                  tabindex="0"
                  class="text-rose-500"
                  @click.stop="clearMatchingForLeft(pair.id)"
                  @keydown.enter.prevent="clearMatchingForLeft(pair.id)"
                  @keydown.space.prevent="clearMatchingForLeft(pair.id)"
                >
                  x
                </span>
              </div>
            </div>
          </div>
        </div>

        <div class="rounded-2xl border-2 border-slate-100 bg-slate-50/70 p-4">
          <div class="mb-3 text-[10px] font-black uppercase tracking-widest text-slate-400">Pilihan Jawaban</div>
          <div class="flex flex-wrap gap-2">
            <button
              v-for="opt in matchingRightItems"
              :key="opt.id"
              type="button"
              :data-right-id="opt.id"
              class="inline-flex min-h-[40px] items-center gap-2 rounded-full border px-4 py-2 text-left text-[13px] font-bold transition-all"
              :class="[
                isMatchingRightUsed(opt.id) ? 'border-emerald-200 bg-emerald-50 text-emerald-700' : 'border-slate-200 bg-white text-slate-700',
                activeMatchingLeftId ? 'hover:border-emerald-300 active:scale-95' : 'opacity-80'
              ]"
              @click="onMatchingRightClick(opt.id)"
            >
              <span>{{ opt.text || '(kosong)' }}</span>
              <span v-if="isMatchingRightUsed(opt.id)" class="text-emerald-500">✓</span>
            </button>
          </div>
          <div class="mt-4 text-center text-[11px] font-medium italic text-slate-400">
            {{ activeLeftLabel ? `Pasangkan untuk: ${activeLeftLabel}` : 'Pilih pernyataan di atas terlebih dahulu' }}
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
