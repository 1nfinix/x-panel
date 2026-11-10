<script setup lang="ts">
import { computed } from 'vue'
import { NProgress } from 'naive-ui'
import { PanelPanelConfigStyleEnum } from '@/enums'
import { resolveNaiveColor, useThemeColors } from '@/hooks/useTheme'

interface Prop {
  textColor: string
  progressColor: string
  progressRailColor: string
  progressHeight?: number
  percentage: number
  cardTypeStyle: PanelPanelConfigStyleEnum
  infoCardLeftText?: string
  infoCardRightText?: string
  compact?: boolean
}

const props = defineProps<Prop>()
const { colors } = useThemeColors()
const safeProgressColor = computed(() => resolveNaiveColor(props.progressColor, colors.value, colors.value.primary))
const safeRailColor = computed(() => resolveNaiveColor(props.progressRailColor, colors.value, colors.value.sep))
</script>

<template>
  <div v-if="compact" class="compact-progress">
    <div class="compact-line">
      <span class="compact-label">{{ infoCardLeftText }}</span>
      <span class="compact-value">{{ Math.round(percentage) }}%</span>
    </div>
    <div class="compact-track">
      <span :style="{ width: `${Math.min(100, Math.max(0, percentage))}%`, background: percentage > 85 ? 'var(--x-red-fg)' : 'var(--x-primary)' }" />
    </div>
  </div>
  <template v-else-if="cardTypeStyle === PanelPanelConfigStyleEnum.info">
    <div class="w-full">
      <div class="mb-1 text-[13px]" :style="{ color: textColor }">
        <span>
          {{ infoCardLeftText }}
        </span>
        <span class="float-right">
          {{ infoCardRightText }}
        </span>
      </div>
      <NProgress
        type="line"
        :color="safeProgressColor"
        :rail-color="safeRailColor"
        :height="progressHeight"
        :percentage="percentage"
        :show-indicator="false"
        :stroke-width="15"
        style="max-width: 135px;"
      />
    </div>
  </template>
  <template v-else>
    <div class="w-full flex justify-center h-full w-full mt-3">
      <NProgress
        :color="safeProgressColor"
        :rail-color="safeRailColor"
        type="dashboard"
        :percentage="percentage" :stroke-width="15"
        style="width: 50px;"
      >
        <div class="text-[13px]" :style="{ color: textColor }">
          {{ percentage }}%
        </div>
      </NProgress>
    </div>
  </template>
</template>

<style scoped>
.compact-progress { min-width: 88px; width: 100%; }
.compact-line { display: flex; align-items: baseline; justify-content: space-between; gap: 12px; margin-bottom: 8px; }
.compact-label { overflow: hidden; color: var(--x-text2); font-size: 13px; text-overflow: ellipsis; white-space: nowrap; }
.compact-value { color: var(--x-text1); font-size: 17px; font-weight: 500; font-variant-numeric: tabular-nums; white-space: nowrap; }
.compact-track { height: 3px; overflow: hidden; border-radius: 2px; background: var(--x-sep); }
.compact-track span { display: block; height: 100%; border-radius: inherit; }
</style>
