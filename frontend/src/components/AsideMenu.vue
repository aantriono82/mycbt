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
  isAsideMobileExpanded: Boolean,
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

const asidePositionClass = computed(() => {
  if (props.isAsideMobileExpanded || props.isAsideLgActive) {
    return 'left-0'
  }

  if (props.isAsideDesktopHidden) {
    return '-left-60 lg:-left-60 xl:-left-60'
  }

  return '-left-60 lg:-left-60 xl:left-0'
})

const asideVisibilityClass = computed(() => {
  if (props.isAsideLgActive) {
    return 'lg:flex xl:flex'
  }

  return 'lg:hidden xl:flex'
})
</script>

<template>
  <AsideMenuLayer
    :menu="menu"
    :menu-bottom="menuBottom"
    :class="[asidePositionClass, asideVisibilityClass]"
    @menu-click="menuClick"
    @aside-lg-close-click="asideLgCloseClick"
  />
  <OverlayLayer v-if="isAsideLgActive" z-index="z-30" @overlay-click="asideLgCloseClick" />
</template>
