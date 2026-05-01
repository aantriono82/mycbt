<script setup>
import { ref } from 'vue'
import AsideMenuItem from '@/components/AsideMenuItem.vue'

defineProps({
  isDropdownList: Boolean,
  menu: {
    type: Array,
    required: true,
  },
})

const emit = defineEmits(['menu-click'])
const expandedKey = ref('')

const menuClick = (event, item) => {
  emit('menu-click', event, item)
}

const toggleItem = (itemKey) => {
  expandedKey.value = expandedKey.value === itemKey ? '' : itemKey
}
</script>

<template>
  <ul>
    <AsideMenuItem
      v-for="(item, index) in menu"
      :key="index"
      :item="item"
      :item-key="`${index}-${item.to || item.label || 'item'}`"
      :expanded-key="expandedKey"
      :is-dropdown-list="isDropdownList"
      @menu-click="menuClick"
      @toggle-dropdown="toggleItem"
    />
  </ul>
</template>
