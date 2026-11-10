<script setup lang="ts">
import { h, nextTick, onBeforeUnmount, onMounted, ref } from 'vue'
import { VueDraggable } from 'vue-draggable-plus'
import { NButton, NDropdown, useDialog, useMessage } from 'naive-ui'
import { type MonitorData, MonitorType } from './typings'
import Edit from './Edit/index.vue'
import { deleteByIndex, getAll, saveAll } from './common'
import { usePanelState } from '@/store'
import { getCpuState, getDiskStateByPath, getMemonyState } from '@/api/system/systemMonitor'
import { SvgIcon } from '@/components/common'
import { t } from '@/locales'
import { useThemeColors } from '@/hooks/useTheme'

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
const { colors } = useThemeColors()

const dialog = useDialog()
const ms = useMessage()

const dropdownMenuX = ref(0)
const dropdownMenuY = ref(0)
const dropdownShow = ref(false)
const currentRightSelectIndex = ref<number | null>(null)
const sortSaving = ref(false)
const sortError = ref('')
const dataError = ref('')
const dataLoading = ref(false)

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
  dataLoading.value = true
  dataError.value = ''
  try {
    monitorDatas.value = await getAll()

    if (monitorDatas.value.length === 0) {
    // 防止空 - 默认数据
      monitorDatas.value.push(
        {
          extendParam: {
            backgroundColor: colors.value.bg2,
            color: colors.value.text1,
            progressColor: colors.value.primary,
            progressRailColor: colors.value.sep,
          },
          monitorType: MonitorType.cpu,
        },
      )

      // 生成并保存
      if (props.allowEdit) {
        const response = await saveAll(monitorDatas.value)
        if (response.code !== 0)
          throw new Error(response.msg || t('common.saveFail'))
      }
    }
    metricStates.value = {}
    await refreshMetrics()
  }
  catch (error) {
    dataError.value = error instanceof Error ? error.message : t('common.failed')
  }
  finally {
    dataLoading.value = false
  }
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
  if (sortSaving.value)
    return
  sortSaving.value = true
  sortError.value = ''
  try {
    const { code, msg } = await saveAll(monitorDatas.value)
    if (code !== 0) {
      sortError.value = msg || t('common.saveFail')
      return
    }
    monitorGroup.value.sortStatus = false
    ms.success(t('common.saveSuccess'))
  }
  catch {
    sortError.value = t('common.saveFail')
  }
  finally {
    sortSaving.value = false
  }
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

function handleRightMenuSelect(key: string | number) {
  dropdownShow.value = false

  switch (key) {
    case 'delete':
    {
      const deleteError = ref('')
      const confirmDialog = dialog.warning({
        title: t('common.warning'),
        content: () => h('div', [h('p', t('common.deleteConfirm')), deleteError.value ? h('p', { style: 'color:var(--x-red-fg);margin-top:12px' }, deleteError.value) : null]),
        positiveText: t('common.confirm'),
        negativeText: t('common.cancel'),
        onPositiveClick: async () => {
          if (monitorDatas.value.length <= 1) {
            deleteError.value = t('common.leastOne')
            return false
          }
          if (currentRightSelectIndex.value === null)
            return false
          confirmDialog.loading = true
          deleteError.value = ''
          try {
            const deleted = await deleteByIndex(currentRightSelectIndex.value)
            if (!deleted) {
              deleteError.value = t('common.deleteFail')
              return false
            }
            await getData()
            ms.success(t('common.deleteSuccess'))
          }
          catch {
            deleteError.value = t('common.deleteFail')
            return false
          }
          finally {
            confirmDialog.loading = false
          }
        },
      })

      break
    }
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
    <p v-if="dataLoading" class="sort-error" aria-live="polite">
      {{ t('panelHome.loading') }}
    </p>
    <p v-if="dataError" class="sort-error" role="alert">
      {{ dataError }} <NButton size="small" @click="getData">
        {{ t('panelHome.retry') }}
      </NButton>
    </p>
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
          <span>{{ metricLabel(item) }}</span><span v-if="metricStates[index]?.state === 'error' || (metricStates[index]?.state === 'ok' && metricStates[index]?.value > 85)" class="metric-chip" :class="metricStates[index]?.state === 'error' ? 'error' : 'high'">{{ metricStates[index]?.state === 'error' ? t('deskModule.systemMonitor.unavailable') : t('deskModule.systemMonitor.high') }}</span>
        </div>
        <div class="metric-value">
          {{ metricStates[index]?.state === 'ok' ? `${Math.round(metricStates[index].value)}%` : '—' }}
        </div>
        <div class="metric-track">
          <span v-if="metricStates[index]?.state === 'ok'" :class="{ high: metricStates[index].value > 85 }" :style="{ width: `${metricStates[index].value}%` }" />
        </div>
      </div>
    </VueDraggable>
    <p v-if="sortError" class="sort-error" role="alert">
      {{ sortError }}
    </p>
    <NButton v-if="monitorGroup.sortStatus && allowEdit" secondary size="small" class="save-sort" :loading="sortSaving" @click="handleSaveSort">
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
.system-monitor.on-wallpaper .monitor-title { color: var(--x-on-primary); }
.monitor-actions { display: flex; gap: 8px; margin-left: auto; opacity: 0; transition: opacity 120ms; }
.monitor-actions.visible, .monitor-actions:focus-within { opacity: 1; }
.monitor-actions button { display: grid; width: 24px; height: 24px; place-items: center; color: var(--x-text2); }
.monitor-actions button:focus-visible { outline: 2px solid var(--x-primary); outline-offset: 2px; }
.monitor-items { display: grid; grid-template-columns: repeat(4, minmax(0, 1fr)); min-height: 116px; overflow: hidden; border-radius: var(--x-r-small); background: var(--x-bg1); }
.monitor-item { position: relative; min-width: 0; padding: 20px; background: transparent; cursor: pointer; }
.monitor-item + .monitor-item::before { position: absolute; top: 20px; bottom: 20px; left: 0; width: 1px; background: var(--x-sep); content: ''; }
.monitor-item:nth-child(4n + 1)::before { display: none; }
.metric-head { display: flex; align-items: center; justify-content: space-between; gap: 4px; color: var(--x-text2); font-size: 13px; }
.metric-chip { color: var(--x-orange-fg); white-space: nowrap; }
.metric-chip.error { color: var(--x-red-fg); }
.metric-value { margin: 14px 0 12px; font-size: 22px; font-weight: 600; font-variant-numeric: tabular-nums; }
.metric-track { width: 100%; height: 3px; overflow: hidden; border-radius: var(--x-r-pill); background: var(--x-sep); }
.metric-track span { display: block; height: 100%; border-radius: inherit; background: var(--x-primary); transition: width 150ms; }
.metric-track span.high { background: var(--x-orange-sys); }
.monitor-item.sorting { cursor: grab; }
.save-sort { margin-top: 8px; }
.sort-error { margin-top: 12px; color: var(--x-red-fg); font-size: 13px; }
@media (max-width: 1179px) { .monitor-items { grid-template-columns: repeat(2, minmax(0, 1fr)); }.monitor-item:nth-child(4n + 1)::before { display: block; }.monitor-item:nth-child(2n + 1)::before { display: none; } }
@media (max-width: 767px) { .monitor-item + .monitor-item::before { display: none; } }
@media (prefers-reduced-motion: reduce) { .monitor-actions { transition: none; } }
</style>
