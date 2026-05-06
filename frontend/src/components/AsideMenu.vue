<script setup>
import { computed } from 'vue'
import AsideMenuLayer from '@/components/AsideMenuLayer.vue'
import OverlayLayer from '@/components/OverlayLayer.vue'

const props = defineProps({
  menu: {
    type: Array,
    required: true,
  },
  menuBottom: Array,
  isAsideLgActive: Boolean,
  isAsideDesktopHidden: Boolean,
})

const emit = defineEmits(['menu-click', 'aside-lg-close-click'])

const menuClick = (event, item) => {
  emit('menu-click', event, item)
}

const asideLgCloseClick = (event) => {
  emit('aside-lg-close-click', event)
}

// Sidebar desktop aktif mulai md; mobile tetap pakai BottomNavigation.
const asidePositionClass = computed(() => {
  if (props.isAsideDesktopHidden) {
    return '-left-60 md:-left-60'
  }
  return '-left-60 md:left-0'
})

const asideVisibilityClass = computed(() => 'hidden md:flex')
</script>

<template>
  <AsideMenuLayer
    :menu="menu"
    :menu-bottom="menuBottom"
    :class="[asidePositionClass, asideVisibilityClass]"
    @menu-click="menuClick"
    @aside-lg-close-click="asideLgCloseClick"
  />
</template>
