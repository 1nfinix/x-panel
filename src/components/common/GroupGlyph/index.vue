<script setup lang="ts">
import { computed } from 'vue'
import SvgIcon from '../SvgIcon/index.vue'

const props = withDefaults(defineProps<{
  icon?: string
  title?: string
  size?: number
}>(), { size: 20 })

const spriteNames = new Set(Object.keys(import.meta.glob('../../../assets/svg-icons/*.svg'))
  .map(path => path.split('/').pop()?.replace(/\.svg$/, '')))
const legacyIcons = new Set([
  'material-symbols:ad-group-outline-rounded',
  'material-symbols:ad-group-outline',
])
const validIcon = computed(() => {
  const icon = props.icon?.replace(/:/g, '-')
  if (props.icon && icon && !legacyIcons.has(props.icon) && spriteNames.has(icon))
    return props.icon
  return ''
})
const initial = computed(() => Array.from(props.title?.trim() || '?')[0].toLocaleUpperCase())
</script>

<template>
  <SvgIcon v-if="validIcon" :icon="validIcon" :style="{ fontSize: `${size}px` }" />
  <span v-else class="group-initial" :style="{ fontSize: `${size * 0.75}px` }" aria-hidden="true">{{ initial }}</span>
</template>

<style scoped>
.group-initial { display: inline-grid; place-items: center; width: 1.333em; height: 1.333em; font-family: var(--xp-font-display); font-weight: 600; line-height: 1; }
</style>
