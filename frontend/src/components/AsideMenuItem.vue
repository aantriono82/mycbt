<script setup>
import { computed } from 'vue'
import { RouterLink, useRoute } from 'vue-router'
import { mdiMinus, mdiPlus } from '@mdi/js'
import { getButtonColor } from '@/colors.js'
import BaseIcon from '@/components/BaseIcon.vue'
import AsideMenuList from '@/components/AsideMenuList.vue'

const props = defineProps({
  item: {
    type: Object,
    required: true,
  },
  itemKey: {
    type: String,
    default: '',
  },
  expandedKey: {
    type: String,
    default: '',
  },
  isDropdownList: Boolean,
})

const emit = defineEmits(['menu-click', 'toggle-dropdown'])
const route = useRoute()

const hasColor = computed(() => props.item && props.item.color)

const asideMenuItemActiveStyle = computed(() =>
  hasColor.value ? '' : 'aside-menu-item-active font-extrabold shadow-lg shadow-blue-400/20 dark:shadow-blue-500/20',
)

const componentClass = computed(() => [
  props.isDropdownList ? 'py-2.5 px-6 text-[16px]' : 'py-3 text-[17px]',
  'rounded-2xl transition-all duration-300 mb-1 px-2',
  hasColor.value
    ? getButtonColor(props.item.color, false, true)
    : `text-slate-700 dark:text-slate-400 hover:text-blue-600 dark:hover:text-white hover:bg-blue-50 dark:hover:bg-white/5 hover:translate-x-1`,
])

const activeClass = (isExactActive) => isExactActive ? asideMenuItemActiveStyle.value : ''

const hasDropdown = computed(() => !!props.item.menu)
const hasActiveChild = computed(() => {
  if (!hasDropdown.value || !Array.isArray(props.item.menu)) return false

  const currentPath = String(route.path || '')
  return props.item.menu.some((child) => {
    const target = String(child?.to || '')
    if (!target) return false
    return currentPath === target || currentPath.startsWith(`${target}/`)
  })
})
const isDropdownActive = computed(() => {
  const expandedMatch = props.expandedKey === props.itemKey
  const hasManualSelection = Boolean(props.expandedKey)
  if (hasManualSelection) {
    return hasDropdown.value && expandedMatch
  }
  return hasDropdown.value && hasActiveChild.value
})
const groupActiveClass = computed(() =>
  hasDropdown.value && (isDropdownActive.value || hasActiveChild.value)
    ? asideMenuItemActiveStyle.value
    : '',
)

const menuClick = (event) => {
  emit('menu-click', event, props.item)

  if (hasDropdown.value) {
    emit('toggle-dropdown', props.itemKey)
  }
}
</script>

<template>
  <li>
    <RouterLink v-if="item.to" :to="item.to" custom v-slot="{ href, navigate, isExactActive }">
      <a
        :href="href"
        class="flex w-full cursor-pointer transition-colors"
        :class="[componentClass, activeClass(isExactActive)]"
        @click="(e) => { navigate(e); menuClick(e) }"
      >
        <BaseIcon
          v-if="item.icon"
          :path="item.icon"
          class="flex-none"
          w="w-16"
          :size="18"
        />
        <span class="line-clamp-1 grow text-ellipsis" :class="[{ 'pr-6': !hasDropdown }]">
          {{ item.label }}
        </span>
        <BaseIcon
          v-if="hasDropdown"
          :path="isDropdownActive ? mdiMinus : mdiPlus"
          class="flex-none"
          w="w-12"
        />
      </a>
    </RouterLink>

    <button
      v-else
      type="button"
      class="flex w-full cursor-pointer text-left transition-colors"
      :class="[componentClass, groupActiveClass]"
      @click="menuClick"
    >
      <BaseIcon
        v-if="item.icon"
        :path="item.icon"
        class="flex-none"
        w="w-16"
        :size="18"
      />
      <span
        class="line-clamp-1 grow text-ellipsis"
        :class="[{ 'pr-6': !hasDropdown }]"
        >{{ item.label }}</span
      >
      <BaseIcon
        v-if="hasDropdown"
        :path="isDropdownActive ? mdiMinus : mdiPlus"
        class="flex-none"
        w="w-12"
      />
    </button>
    <AsideMenuList
      v-if="hasDropdown && isDropdownActive"
      :menu="item.menu"
      class="aside-menu-dropdown bg-blue-50 dark:bg-white/5 rounded-2xl mb-2"
      is-dropdown-list
    />
  </li>
</template>
