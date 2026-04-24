<script setup>
import { computed } from 'vue'

const props = defineProps({
  questions: { type: Array, default: () => [] },
  answers: { type: Object, default: () => ({}) },
  flagged: { type: Object, default: () => ({}) },
  currentIndex: { type: Number, default: 0 },
})

const emit = defineEmits(['jump-to'])

const statusOf = (question, idx) => {
  const qid = String(question?.id ?? '')
  if (props.flagged?.[qid]) return 'flagged'
  if (props.answers?.[qid] !== undefined && props.answers?.[qid] !== null && props.answers?.[qid] !== '') return 'answered'
  if (idx === props.currentIndex) return 'current'
  return 'unanswered'
}

const items = computed(() =>
  (props.questions || []).map((q, idx) => ({
    idx,
    id: String(q?.id ?? idx),
    label: idx + 1,
    status: statusOf(q, idx),
  })),
)

const jumpTo = (idx) => emit('jump-to', idx)
</script>

<template>
  <div class="grid grid-cols-5 gap-2" data-testid="question-navigator">
    <button
      v-for="item in items"
      :key="item.id"
      type="button"
      :data-testid="`qnav-${item.idx}`"
      :data-status="item.status"
      class="relative h-10 w-10 rounded border text-sm font-semibold"
      @click="jumpTo(item.idx)"
    >
      {{ item.label }}
      <span
        v-if="item.status === 'flagged'"
        data-testid="flag-icon"
        class="absolute -right-1 -top-1 h-3 w-3 rounded-full bg-amber-400"
      />
    </button>
  </div>
</template>

