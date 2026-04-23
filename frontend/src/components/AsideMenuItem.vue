<script setup>
import { ref, computed } from 'vue'
import { RouterLink } from 'vue-router'
import { mdiMinus, mdiPlus } from '@mdi/js'
import { getButtonColor } from '@/colors.js'
import BaseIcon from '@/components/BaseIcon.vue'
import AsideMenuList from '@/components/AsideMenuList.vue'

const props = defineProps({
  item: {
    type: Object,
    required: true,
  },
  isDropdownList: Boolean,
})

const emit = defineEmits(['menu-click'])

const hasColor = computed(() => props.item && props.item.color)

const asideMenuItemActiveStyle = computed(() =>
  hasColor.value ? '' : 'aside-menu-item-active font-extrabold shadow-lg shadow-blue-500/20',
)

const isDropdownActive = ref(false)

const componentClass = computed(() => [
  props.isDropdownList ? 'py-2.5 px-6 text-[16px]' : 'py-3 text-[17px]',
  'rounded-2xl transition-all duration-300 mb-1 px-2',
  hasColor.value
    ? getButtonColor(props.item.color, false, true)
    : `text-slate-400 hover:text-white hover:bg-white/5 hover:translate-x-1`,
])

const activeClass = (isExactActive) => isExactActive ? asideMenuItemActiveStyle.value : ''

const hasDropdown = computed(() => !!props.item.menu)

const menuClick = (event) => {
  emit('menu-click', event, props.item)

  if (hasDropdown.value) {
    isDropdownActive.value = !isDropdownActive.value
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

    <a
      v-else
      :href="item.href ?? null"
      :target="item.target ?? null"
      class="flex cursor-pointer transition-colors"
      :class="componentClass"
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
    </a>
    <AsideMenuList
      v-if="hasDropdown"
      :menu="item.menu"
      :class="['aside-menu-dropdown', isDropdownActive ? 'block bg-white/5 rounded-2xl mb-2' : 'hidden']"
      is-dropdown-list
    />
  </li>
</template>
