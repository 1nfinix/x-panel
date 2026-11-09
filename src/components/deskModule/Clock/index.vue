<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from 'vue'
import { t } from '@/locales'

const props = defineProps<{
  hideSecond?: boolean
}>()

interface CurrentDate {
  time: string
  date: string
  week: string
}

const currentDate = ref<CurrentDate>({
  time: '--:--',
  date: '------',
  week: '--',
})

function updateCurrentDate() {
  const now = new Date()
  const hours = String(now.getHours()).padStart(2, '0')
  const minutes = String(now.getMinutes()).padStart(2, '0')

  if (!props.hideSecond) {
    const seconds = String(now.getSeconds()).padStart(2, '0')
    currentDate.value.time = `${hours}:${minutes}:${seconds}`
  }
  else {
    currentDate.value.time = `${hours}:${minutes}`
  }

  // 获取当前的日期
  const day = now.getDate()
  const month = now.getMonth() + 1 // 月份从0开始，所以要加1
  // const year = now.getFullYear()

  const daysOfWeek = [
    t('deskModule.clock.sun'),
    t('deskModule.clock.mon'),
    t('deskModule.clock.tue'),
    t('deskModule.clock.wed'),
    t('deskModule.clock.thu'),
    t('deskModule.clock.fri'),
    t('deskModule.clock.sat'),
  ]
  // const daysOfWeek = ['Sunday', 'Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday']
  currentDate.value.week = daysOfWeek[now.getDay()]
  currentDate.value.date = `${month}月${day}日`
}

const updateClock = () => {
  updateCurrentDate()
}

let intervalId: ReturnType<typeof setInterval>

onMounted(() => {
  updateClock()
  intervalId = setInterval(updateClock, 1000)
})

onBeforeUnmount(() => {
  clearInterval(intervalId)
})
</script>

<template>
  <div class="clock">
    <span class="clock-time">
      {{ currentDate.time }}
    </span>
    <div class="clock-calendar">
      <span class="clock-date">
        {{ currentDate.date }}
      </span>
      <span class="clock-week">
        {{ currentDate.week }}
      </span>
    </div>
  </div>
</template>

<style scoped>
.clock { display: flex; flex-direction: column; align-items: flex-start; gap: 4px; white-space: nowrap; }
.clock-time { color: var(--x-text1); font: 600 64px/1.1 var(--x-font); font-variant-numeric: tabular-nums; }
.clock-calendar { display: flex; gap: 6px; color: var(--x-text2); font-size: 14px; }
@media (max-width: 767px) {
  .clock { flex-direction: row; align-items: baseline; gap: 16px; }
  .clock-time { font-size: 48px; }
  .clock-calendar { font-size: 13px; }
}
</style>
