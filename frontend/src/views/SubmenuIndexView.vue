<script setup>
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { mdiViewListOutline, mdiChevronRight } from '@mdi/js'
import LayoutAuthenticated from '@/layouts/LayoutAuthenticated.vue'
import SectionMain from '@/components/SectionMain.vue'
import SectionTitleLineWithButton from '@/components/SectionTitleLineWithButton.vue'
import CardBox from '@/components/CardBox.vue'
import BaseIcon from '@/components/BaseIcon.vue'
import { getMenuAsideMain } from '@/menuAside.js'
import { useAuthStore } from '@/stores/auth.js'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()

const menuAsideMain = computed(() => getMenuAsideMain(authStore.role))
const currentGroup = computed(() =>
  menuAsideMain.value.find((item) => item.to === route.path && Array.isArray(item.menu)),
)
const title = computed(() => currentGroup.value?.label || 'Submenu')
const items = computed(() => currentGroup.value?.menu || [])
</script>

<template>
  <LayoutAuthenticated>
    <SectionMain>
      <SectionTitleLineWithButton :icon="mdiViewListOutline" :title="title" main />
      <CardBox>
        <div class="space-y-2">
          <button
            v-for="item in items"
            :key="item.to"
            type="button"
            class="flex w-full items-center justify-between rounded-xl border border-slate-200 bg-white px-4 py-3 text-left transition hover:bg-slate-50 dark:border-slate-700 dark:bg-slate-900 dark:hover:bg-slate-800"
            @click="router.push(item.to)"
          >
            <span class="flex items-center gap-3">
              <BaseIcon :path="item.icon" size="18" class="text-blue-600 dark:text-blue-300" />
              <span class="font-semibold text-slate-800 dark:text-slate-100">{{ item.label }}</span>
            </span>
            <BaseIcon :path="mdiChevronRight" size="18" class="text-slate-400" />
          </button>
        </div>
      </CardBox>
    </SectionMain>
  </LayoutAuthenticated>
</template>
