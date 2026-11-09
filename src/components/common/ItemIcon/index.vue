<script setup lang="ts">
import { NAvatar, NImage } from 'naive-ui'
import { computed, ref, withDefaults } from 'vue'
import { SvgIconOnline } from '@/components/common'

interface Prop {
  itemIcon?: Panel.ItemIcon | null
  size?: number // 默认70
  forceBackground?: string // 强制背景色
}

const props = withDefaults(defineProps<Prop>(), { size: 70 })
const defaultBackground = 'var(--xp-surface-card)'
const defaultStyle = ref({
  width: `${props.size}px`,
  height: `${props.size}px`,
})
const iconExt = computed(() => {
  return props.itemIcon?.src?.split('.').pop()
})
</script>

<template>
  <div class="item-icon" :style="defaultStyle">
    <slot>
      <template v-if="itemIcon">
        <template v-if="itemIcon?.itemType === 1">
          <NAvatar :size="props.size" :style="{ backgroundColor: (forceBackground ?? itemIcon?.backgroundColor) || defaultBackground }">
            {{ itemIcon.text }}
          </NAvatar>
        </template>

        <template v-else-if="itemIcon?.itemType === 2">
          <div v-if="iconExt === 'svg'" :style="{ backgroundColor: (forceBackground ?? itemIcon?.backgroundColor) || defaultBackground, ...defaultStyle }" class="flex justify-center items-center">
            <img :src="itemIcon?.src" class="w-[75%] h-[75%] object-contain">
          </div>
          <NImage v-else :style="{ backgroundColor: (forceBackground ?? itemIcon?.backgroundColor) || defaultBackground, ...defaultStyle }" :src="itemIcon?.src" preview-disabled />
        </template>

        <template v-else-if="itemIcon?.itemType === 3">
          <NAvatar :size="props.size" :style="{ backgroundColor: (forceBackground ?? itemIcon?.backgroundColor) || defaultBackground }">
            <SvgIconOnline :style="{ fontSize: `${Math.round(props.size * 0.68)}px` }" :icon="itemIcon.text" />
          </NAvatar>
        </template>
      </template>

      <template v-else>
        <NAvatar :size="props.size" />
      </template>
    </slot>
  </div>
</template>
