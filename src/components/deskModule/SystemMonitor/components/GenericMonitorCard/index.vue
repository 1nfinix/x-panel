<script setup lang="ts">
// -------------------
// -------------------

import { ref } from 'vue'
import { ItemCard, SvgIcon } from '@/components/common'
import type { PanelPanelConfigStyleEnum } from '@/enums'

interface Prop {
//   size?: number // 默认70
  extendParam?: any
  iconTextColor?: string
  iconTextIconHideTitle?: boolean
  iconText?: string
  textColor?: string
  cardTypeStyle: PanelPanelConfigStyleEnum
  // monitorType: string
  icon?: string
  class?: string
  backgroundColor?: string
}

const props = withDefaults(defineProps<Prop>(), {})
const propClass = ref(props.class)
</script>

<template>
  <div class="generic-monitor-card w-full">
    <ItemCard
      :card-type-style="cardTypeStyle"
      :icon-text="iconText"
      :icon-text-color="iconTextColor"
      :class="propClass"
      :background-color="backgroundColor"
      :icon-text-icon-hide-title="iconTextIconHideTitle"
    >
      <template #info>
        <!-- 图标 -->
        <div class="w-[60px] h-[70px]">
          <div class="w-[60px] h-full flex items-center justify-center text-text1">
            <slot name="icon">
              <SvgIcon :icon="icon ?? ''" style="width: 35px;height: 35px;" :style="{ color: textColor }" />
            </slot>
          </div>
        </div>

        <div
          class=" w-full text-text1 flex items-center"
          :style=" { maxWidth: 'calc(100% - 80px)' }"
        >
          <slot name="info" />
        </div>
      </template>
      <template #small>
        <slot name="small" />
      </template>
    </ItemCard>
  </div>
</template>
