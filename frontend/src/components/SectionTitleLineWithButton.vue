<script setup>
import { mdiCog } from '@mdi/js'
import { useSlots, computed } from 'vue'
import BaseIcon from '@/components/BaseIcon.vue'
import BaseButton from '@/components/BaseButton.vue'
import IconRounded from '@/components/IconRounded.vue'

defineProps({
  icon: String,
  title: {
    type: String,
    required: true,
  },
  main: Boolean,
})

const hasSlot = computed(() => useSlots().default)
</script>

<template>
  <section :class="{ 'pt-6': !main }" class="mb-6 flex items-center justify-between">
    <div class="flex items-center justify-start">
      <IconRounded v-if="icon && main" :icon="icon" color="light" class="mr-3" bg />
      <BaseIcon v-else-if="icon" :path="icon" class="mr-2" size="20" />
      <h1 :class="[main ? 'text-2xl sm:text-3xl font-extrabold tracking-tight' : 'text-xl sm:text-2xl font-bold']" class="leading-tight text-slate-900 dark:text-white">
        {{ title }}
      </h1>
    </div>
    <slot v-if="hasSlot" />
    <BaseButton v-else :icon="mdiCog" color="whiteDark" />
  </section>
</template>
