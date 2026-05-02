<script setup>
import { computed, ref } from 'vue'
import { mdiEye, mdiEyeOff } from '@mdi/js'
import FormControl from '@/components/FormControl.vue'
import BaseIcon from '@/components/BaseIcon.vue'

const props = defineProps({
  modelValue: {
    type: String,
    default: '',
  },
  name: String,
  placeholder: String,
  autocomplete: String,
  icon: String,
  disabled: Boolean,
})

const emit = defineEmits(['update:modelValue'])

const revealed = ref(false)

const computedValue = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value),
})
</script>

<template>
  <div class="relative">
    <FormControl
      v-model="computedValue"
      :name="name"
      :placeholder="placeholder"
      :autocomplete="autocomplete"
      :icon="icon"
      :disabled="disabled"
      :type="revealed ? 'text' : 'password'"
    />
    <button
      type="button"
      class="absolute inset-y-0 right-0 z-10 flex items-center pr-3 text-slate-400 hover:text-slate-600"
      :aria-label="revealed ? 'Sembunyikan password' : 'Tampilkan password'"
      @click="revealed = !revealed"
    >
      <BaseIcon :path="revealed ? mdiEyeOff : mdiEye" size="18" />
    </button>
  </div>
</template>
