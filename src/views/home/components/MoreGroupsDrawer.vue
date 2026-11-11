<script setup lang="ts">
import { NDrawer, NDrawerContent } from 'naive-ui'
import { GroupGlyph } from '@/components/common'
import { t } from '@/locales'

interface Group extends Panel.ItemIconGroup {
  hoverStatus: boolean
  items?: Panel.ItemInfo[]
}

defineProps<{
  show: boolean
  groups: Group[]
  selectedGroupId: number | 'all' | null
  groupColorStyle: (group: Group) => Record<string, string>
}>()

const emit = defineEmits<{
  (e: 'update:show', show: boolean): void
  (e: 'select', id: number | 'all'): void
}>()
</script>

<template>
  <NDrawer :show="show" placement="bottom" height="min(70vh, 620px)" class="groups-drawer" @update:show="emit('update:show', $event)">
    <NDrawerContent :title="t('panelHome.groups')" closable>
      <button class="drawer-group all-group" :class="{ active: selectedGroupId === 'all' }" :aria-current="selectedGroupId === 'all' ? 'true' : undefined" type="button" @click="emit('select', 'all')">
        <span class="heading-glyph"><GroupGlyph icon="lucide-layout-grid" :title="t('panelHome.all')" :size="22" /></span><span>{{ t('panelHome.all') }}</span>
      </button>
      <button v-for="group in groups" :key="group.id" class="drawer-group" :class="{ active: selectedGroupId === group.id }" :style="groupColorStyle(group)" :aria-current="selectedGroupId === group.id ? 'true' : undefined" type="button" @click="emit('select', group.id as number)">
        <span class="heading-glyph"><GroupGlyph :icon="group.icon" :title="group.title" :size="22" /></span><span>{{ group.title }}</span><span class="drawer-count">{{ group.items?.length ?? 0 }}</span>
      </button>
    </NDrawerContent>
  </NDrawer>
</template>

<style scoped>
.drawer-group { display: flex; align-items: center; gap: 12px; width: 100%; min-height: 64px; padding: 8px 16px; border: 0; border-radius: var(--x-r-small); background: transparent; color: var(--x-text1); text-align: left; }
.drawer-group.active { background: var(--group-tint); color: var(--group-fg); }
.all-group { --group-tint: var(--x-blue-tint); --group-fg: var(--x-blue-fg); }
.drawer-group + .drawer-group { margin-top: 8px; }
.drawer-group:focus-visible { outline: 3px solid var(--x-focus); outline-offset: 2px; }
.heading-glyph { display: grid; flex: none; place-items: center; width: 32px; height: 32px; border-radius: var(--x-r-glyph); background: var(--group-tint); color: var(--group-fg); }
.drawer-count { margin-left: auto; color: var(--x-text2); }
</style>
