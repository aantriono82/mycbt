<script setup>
import { computed, useSlots } from 'vue'
import CardBoxComponentBody from '@/components/CardBoxComponentBody.vue'
import CardBoxComponentFooter from '@/components/CardBoxComponentFooter.vue'

const props = defineProps({
  rounded: {
    type: String,
    default: 'rounded-3xl',
  },
  flex: {
    type: String,
    default: 'flex-col',
  },
  color: {
    type: String,
    default: 'slate',
  },
  hasComponentLayout: Boolean,
  hasTable: Boolean,
  isForm: Boolean,
  isHoverable: Boolean,
  isModal: Boolean,
  hasShadow: {
    type: Boolean,
    default: true,
  },
})

const emit = defineEmits(['submit'])

const slots = useSlots()

const hasFooterSlot = computed(() => slots.footer && !!slots.footer())

const colorMap = {
  slate: 'border-slate-300 dark:border-slate-700',
  blue: 'border-blue-400/60 dark:border-blue-800/80',
  indigo: 'border-indigo-400/60 dark:border-indigo-800/80',
  emerald: 'border-emerald-400/60 dark:border-emerald-800/80',
  amber: 'border-amber-400/60 dark:border-amber-800/80',
  purple: 'border-purple-400/60 dark:border-purple-800/80',
  orange: 'border-orange-400/60 dark:border-orange-800/80',
  cyan: 'border-cyan-400/60 dark:border-cyan-800/80',
  rose: 'border-rose-400/60 dark:border-rose-800/80',
}

const componentClass = computed(() => {
  const base = [
    props.rounded || 'rounded-[2rem]',
    props.flex,
    props.isModal ? 'dark:bg-slate-900' : 'dark:bg-slate-900/80 backdrop-blur-md',
    'border animate-fade-in-up',
    props.hasShadow ? 'shadow-sm' : 'shadow-none',
    colorMap[props.color] || colorMap.slate
  ]

  if (props.isHoverable) {
    base.push('hover:shadow-lg hover:scale-[1.01] transition-all duration-300 cursor-pointer')
  }

  return base
})

const submit = (event) => {
  emit('submit', event)
}
</script>

<template>
  <component
    :is="isForm ? 'form' : 'div'"
    :class="componentClass"
    class="flex min-w-0 max-w-full bg-white"
    @submit="submit"
  >
    <slot v-if="hasComponentLayout" />
    <template v-else>
      <CardBoxComponentBody :no-padding="hasTable">
        <slot />
      </CardBoxComponentBody>
      <CardBoxComponentFooter v-if="hasFooterSlot">
        <slot name="footer" />
      </CardBoxComponentFooter>
    </template>
  </component>
</template>
