<script setup lang="ts">
import { nextTick, onBeforeUnmount, onMounted, ref } from 'vue'
import { VueDraggable } from 'vue-draggable-plus'
import { NButton, NDropdown, useDialog, useMessage } from 'naive-ui'
import { type MonitorData, MonitorType } from './typings'
import Edit from './Edit/index.vue'
import { deleteByIndex, getAll, saveAll } from './common'
import { usePanelState } from '@/store'
import { getCpuState, getDiskStateByPath, getMemonyState } from '@/api/system/systemMonitor'
import { SvgIcon } from '@/components/common'
import { t } from '@/locales'

interface MonitorGroup extends Panel.ItemIconGroup {
  sortStatus?: boolean
  hoverStatus: boolean
  items?: Panel.ItemInfo[]
}

const props = defineProps<{
  allowEdit?: boolean
  showTitle?: boolean
}>()
const panelState = usePanelState()

const dialog = useDialog()
const ms = useMessage()

const dropdownMenuX = ref(0)
const dropdownMenuY = ref(0)
const dropdownShow = ref(false)
const currentRightSelectIndex = ref<number | null>(null)

const monitorGroup = ref<MonitorGroup>({
  hoverStatus: false,
  sortStatus: false,
})

const editShowStatus = ref<boolean>(false)
const editData = ref<MonitorData | null>(null)
const editIndex = ref<number | null>(null)

function handleAddItem() {
  editShowStatus.value = true
  editData.value = null
  editIndex.value = null
}

function handleSetSortStatus(sortStatus: boolean) {
  monitorGroup.value.sortStatus = sortStatus

  // 并未保存排序重新更新数据
  if (!sortStatus)
    getData()
}

function handleSetHoverStatus(hoverStatus: boolean) {
  monitorGroup.value.hoverStatus = hoverStatus
}

const monitorDatas = ref<MonitorData[]>([])
const metricStates = ref<Record<number, { state: 'loading' | 'ok' | 'error'; value: number; retryAt: number }>>({})
let metricsTimer: ReturnType<typeof setInterval>
async function refreshMetrics() {
  await Promise.all(monitorDatas.value.map(async (item, index) => {
    const previous = metricStates.value[index]
    if (previous?.state === 'error' && Date.now() < previous.retryAt)
      return
    if (!previous)
      metricStates.value[index] = { state: 'loading', value: 0, retryAt: 0 }
    try {
      const response = item.monitorType === MonitorType.cpu
        ? await getCpuState<SystemMonitor.CPUInfo>()
        : item.monitorType === MonitorType.memory
          ? await getMemonyState<SystemMonitor.MemoryInfo>()
          : await getDiskStateByPath<SystemMonitor.DiskInfo>(item.extendParam?.path ?? '/')
      if (response.code !== 0)
        throw new Error('monitor unavailable')
      const data = response.data
      const value = item.monitorType === MonitorType.cpu
        ? (data as SystemMonitor.CPUInfo).usages[0]
        : (data as SystemMonitor.MemoryInfo).usedPercent
      metricStates.value[index] = { state: 'ok', value: Math.max(0, Math.min(100, Number(value) || 0)), retryAt: 0 }
    }
    catch {
      metricStates.value[index] = { state: 'error', value: 0, retryAt: Date.now() + 30000 }
    }
  }))
}
function metricLabel(item: MonitorData) {
  if (item.monitorType === MonitorType.cpu)
    return 'CPU'
  if (item.monitorType === MonitorType.memory)
    return 'RAM'
  return item.extendParam?.path || 'Disk'
}

function handleClick(index: number, item: MonitorData) {
  if (!props.allowEdit)
    return
  editShowStatus.value = true
  editData.value = item
  editIndex.value = index
}

async function getData() {
  monitorDatas.value = await getAll()

  if (monitorDatas.value.length === 0) {
    // 防止空 - 默认数据
    monitorDatas.value.push(
      {
        extendParam: {
          backgroundColor: 'var(--x-bg2)',
          color: 'var(--x-text1)',
          progressColor: 'var(--x-primary)',
          progressRailColor: 'var(--x-sep)',
        },
        monitorType: MonitorType.cpu,
      },
    )

    // 生成并保存
    if (props.allowEdit)
      saveAll(monitorDatas.value)
  }
  metricStates.value = {}
  refreshMetrics()
}

onMounted(() => {
  getData()
  metricsTimer = setInterval(refreshMetrics, 5000)
})
onBeforeUnmount(() => clearInterval(metricsTimer))

function handleSaveDone() {
  getData()
}

async function handleSaveSort() {
  const { code } = await saveAll(monitorDatas.value)
  if (code === 0)
    monitorGroup.value.sortStatus = false
}

function handleContextMenu(e: MouseEvent, index: number | null) {
  if (index !== null) {
    e.preventDefault()
    currentRightSelectIndex.value = index
  }

  nextTick().then(() => {
    dropdownShow.value = true
    dropdownMenuX.value = e.clientX
    dropdownMenuY.value = e.clientY
  })
}

function getDropdownMenuOptions() {
  const dropdownMenuOptions = [
    {
      label: t('common.delete'),
      key: 'delete',
    },
  ]

  return dropdownMenuOptions
}

function onClickoutside() {
  // message.info('clickoutside')
  dropdownShow.value = false
}

async function deleteOneByIndex(index: number) {
  const res = await deleteByIndex(index)
  if (res)
    getData()
}

function handleRightMenuSelect(key: string | number) {
  dropdownShow.value = false

  switch (key) {
    case 'delete':
      dialog.warning({
        title: t('common.warning'),
        content: t('common.deleteConfirm'),
        positiveText: t('common.confirm'),
        negativeText: t('common.cancel'),
        onPositiveClick: () => {
          if (monitorDatas.value.length <= 1) {
            ms.warning(t('common.leastOne'))
            return
          }
          if (currentRightSelectIndex.value !== null)
            deleteOneByIndex(currentRightSelectIndex.value)
        },
      })

      break
    default:
      break
  }
}
</script>

<template>
  <div class="system-monitor" :class="{ 'on-wallpaper': panelState.panelConfig.showWallpaper }" @mouseenter="handleSetHoverStatus(true)" @mouseleave="handleSetHoverStatus(false)">
    <div class="monitor-toolbar">
      <span v-if="showTitle" class="monitor-title">{{ $t('deskModule.systemMonitor.systemState') }}</span>
      <div v-if="allowEdit" class="monitor-actions" :class="{ visible: monitorGroup.hoverStatus || monitorGroup.sortStatus }">
        <button type="button" :aria-label="t('common.add')" @click="handleAddItem">
          <SvgIcon icon="typcn:plus" />
        </button>
        <button type="button" :aria-label="t('common.saveSort')" @click="handleSetSortStatus(!monitorGroup.sortStatus)">
          <SvgIcon icon="ri:drag-drop-line" />
        </button>
      </div>
    </div>
    <VueDraggable
      v-model="monitorDatas" item-key="sort" :animation="220" class="monitor-items"
      :disabled="!monitorGroup.sortStatus"
    >
      <div
        v-for="item, index in monitorDatas" :key="index" class="monitor-item"
        :class="{ sorting: monitorGroup.sortStatus }" :title="item.description"
        @click="handleClick(index, item)" @contextmenu="(e) => handleContextMenu(e, index)"
      >
        <div class="metric-head">
          <span>{{ metricLabel(item) }}</span><span class="metric-chip" :class="metricStates[index]?.state === 'error' ? 'error' : metricStates[index]?.value > 85 ? 'high' : 'ok'">{{ metricStates[index]?.state === 'error' ? t('deskModule.systemMonitor.unavailable') : metricStates[index]?.state === 'loading' || !metricStates[index] ? '…' : metricStates[index]?.value > 85 ? t('deskModule.systemMonitor.high') : t('deskModule.systemMonitor.normal') }}</span>
        </div>
        <div class="metric-value">
          {{ metricStates[index]?.state === 'ok' ? `${Math.round(metricStates[index].value)}%` : '—' }}
        </div>
        <div class="metric-track">
          <span :class="{ high: metricStates[index]?.value > 85 }" :style="{ width: metricStates[index]?.state === 'ok' ? `${metricStates[index].value}%` : '0%' }" />
        </div>
      </div>
    </VueDraggable>
    <NButton v-if="monitorGroup.sortStatus && allowEdit" secondary size="small" class="save-sort" @click="handleSaveSort">
      {{ $t('common.saveSort') }}
    </NButton>
    <Edit v-model:visible="editShowStatus" :monitor-data="editData" :index="editIndex" @done="handleSaveDone" />
    <NDropdown
      placement="bottom-start" trigger="manual" :x="dropdownMenuX" :y="dropdownMenuY"
      :options="getDropdownMenuOptions()" :show="dropdownShow" :on-clickoutside="onClickoutside" @select="handleRightMenuSelect"
    />
  </div>
</template>

<style scoped>
.system-monitor { min-width: 0; width: 100%; }
.monitor-toolbar { display: flex; align-items: center; justify-content: space-between; min-height: 16px; }
.monitor-toolbar:has(.monitor-title) { margin-bottom: 8px; }
.monitor-title { color: var(--x-text2); font-size: 13px; }
.monitor-actions { display: flex; gap: 8px; margin-left: auto; opacity: 0; transition: opacity 120ms; }
.monitor-actions.visible, .monitor-actions:focus-within { opacity: 1; }
.monitor-actions button { display: grid; width: 24px; height: 24px; place-items: center; color: var(--x-text2); }
.monitor-actions button:focus-visible { outline: 2px solid var(--x-primary); outline-offset: 2px; }
.monitor-items { display: grid; grid-template-columns: repeat(auto-fit, minmax(145px, 1fr)); gap: 12px; }
.monitor-item { min-width: 0; padding: 20px; border-radius: var(--x-r-small); background: var(--x-bg2); cursor: pointer; }
.metric-head { display: flex; align-items: center; justify-content: space-between; gap: 4px; color: var(--x-text2); font-size: 13px; }.metric-chip { padding: 4px 7px; border-radius: 999px; background: var(--x-green-tint); color: var(--x-green-fg); white-space: nowrap; }.metric-chip.high { background: var(--x-orange-tint); color: var(--x-orange-fg); }.metric-chip.error { background: var(--x-red-tint); color: var(--x-red-fg); }.metric-value { margin: 20px 0 16px; font-size: 28px; font-weight: 600; font-variant-numeric: tabular-nums; }.metric-track { height: 8px; overflow: hidden; border-radius: 999px; background: var(--x-bg1); }.metric-track span { display: block; height: 100%; border-radius: inherit; background: var(--x-primary); transition: width 150ms; }.metric-track span.high { background: var(--x-orange-sys); }
.monitor-item.sorting { cursor: grab; }
.save-sort { margin-top: 8px; }
@media (prefers-reduced-motion: reduce) { .monitor-actions { transition: none; } }
</style>
