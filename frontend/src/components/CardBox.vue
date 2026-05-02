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
  hasComponentLayout: Boolean,
  hasTable: Boolean,
  isForm: Boolean,
  isHoverable: Boolean,
  isModal: Boolean,
})

const emit = defineEmits(['submit'])

const slots = useSlots()

const hasFooterSlot = computed(() => slots.footer && !!slots.footer())

const componentClass = computed(() => {
  const base = [
    props.rounded || 'rounded-[2rem]',
    props.flex,
    props.isModal ? 'dark:bg-slate-900' : 'dark:bg-slate-900/80 backdrop-blur-md',
    'border border-slate-100/50 dark:border-indigo-500/10 shadow-sm animate-fade-in-up'
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
