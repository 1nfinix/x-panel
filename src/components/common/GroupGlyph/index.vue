<script setup lang="ts">
import { computed } from 'vue'
import SvgIcon from '../SvgIcon/index.vue'
import { resolveGroupIcon } from './icons'

const props = withDefaults(defineProps<{
  icon?: string
  title?: string
  size?: number
}>(), { size: 20 })

const glyph = computed(() => resolveGroupIcon(props.icon, props.title))
</script>

<template>
  <SvgIcon :icon="glyph" :style="{ fontSize: `${size}px` }" />
</template>
