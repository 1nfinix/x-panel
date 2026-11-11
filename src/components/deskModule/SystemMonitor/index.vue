<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref } from 'vue'
import { VueDraggable } from 'vue-draggable-plus'
import { NDropdown, useDialog, useMessage } from 'naive-ui'
import { type MonitorData, MonitorType } from './typings'
import Edit from './Edit/index.vue'
import { deleteByIndex, getAll, saveAll } from './common'
import { getCpuState, getDiskStateByPath, getMemonyState } from '@/api/system/systemMonitor'
import { SvgIcon } from '@/components/common'
import { t } from '@/locales'
import { confirmDelete } from '@/utils/confirmDelete'
import { useThemeColors } from '@/hooks/useTheme'

interface MonitorGroup extends Panel.ItemIconGroup {
  sortStatus?: boolean
  items?: Panel.ItemInfo[]
}

const props = defineProps<{
  allowEdit?: boolean
  showTitle?: boolean
}>()
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

const monitorDatas = ref<MonitorData[]>([])
const actionOptions = computed(() => [
  { label: t('common.add'), key: 'add' },
  ...monitorDatas.value.map((item, index) => ({ label: `${t('common.edit')} ${metricLabel(item)}`, key: `edit-${index}` })),
  { label: t(monitorGroup.value.sortStatus ? 'common.saveSort' : 'common.sort'), key: monitorGroup.value.sortStatus ? 'save-sort' : 'sort' },
  ...(monitorGroup.value.sortStatus ? [{ label: t('common.cancel'), key: 'cancel-sort' }] : []),
  ...monitorDatas.value.map((item, index) => ({ label: `${t('common.delete')} ${metricLabel(item)}`, key: `delete-${index}` })),
  ...(dataError.value ? [{ label: t('panelHome.retry'), key: 'retry' }] : []),
])
function handleAction(key: string | number) {
  const action = String(key)
  if (action === 'add') {
    handleAddItem()
  }
  else if (action === 'sort') {
    handleSetSortStatus(true)
  }
  else if (action === 'save-sort') {
    handleSaveSort()
  }
  else if (action === 'cancel-sort') {
    handleSetSortStatus(false)
  }
  else if (action === 'retry') {
    getData()
  }
  else if (action.startsWith('edit-')) {
    const index = Number(action.slice(5))
    handleClick(index, monitorDatas.value[index])
  }
  else if (action.startsWith('delete-')) {
    currentRightSelectIndex.value = Number(action.slice(7))
    handleRightMenuSelect('delete')
  }
}
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
      confirmDelete(dialog, {
        object: t('common.metricObject'),
        name: currentRightSelectIndex.value === null ? '' : metricLabel(monitorDatas.value[currentRightSelectIndex.value]),
        onDelete: async () => {
          if (monitorDatas.value.length <= 1)
            return t('common.leastOne')
          if (currentRightSelectIndex.value === null)
            return t('common.deleteFail')
          try {
            const deleted = await deleteByIndex(currentRightSelectIndex.value)
            if (!deleted)
              return t('common.deleteFail')
            await getData()
            ms.success(t('common.deleteSuccess'))
          }
          catch { return t('common.deleteFail') }
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
  <div class="system-monitor" :aria-label="showTitle ? t('deskModule.systemMonitor.systemState') : undefined" :title="showTitle ? t('deskModule.systemMonitor.systemState') : undefined">
    <VueDraggable v-model="monitorDatas" item-key="sort" :animation="220" class="monitor-items" :disabled="!monitorGroup.sortStatus">
      <div v-for="item, index in monitorDatas" :key="index" class="monitor-item" :class="{ sorting: monitorGroup.sortStatus }" :title="item.description" @contextmenu="(e) => handleContextMenu(e, index)">
        <div class="metric-head">
          <span>{{ metricLabel(item) }}</span><span v-if="metricStates[index]?.state === 'error' || dataError" class="metric-error-dot" :title="t('deskModule.systemMonitor.unavailable')" :aria-label="t('deskModule.systemMonitor.unavailable')" />
        </div>
        <div class="metric-value">
          {{ !dataError && metricStates[index]?.state === 'ok' ? `${Math.round(metricStates[index].value)}%` : '—' }}
        </div>
        <div class="metric-track">
          <span v-if="metricStates[index]?.state === 'ok'" :class="{ high: metricStates[index].value > 85 }" :style="{ width: `${metricStates[index].value}%` }" />
        </div>
      </div>
      <div v-if="dataError && !monitorDatas.length" class="monitor-item" :title="dataError">
        <div class="metric-head">
          <span>{{ t('deskModule.systemMonitor.systemState') }}</span><span class="metric-error-dot" :title="t('deskModule.systemMonitor.unavailable')" />
        </div>
        <div class="metric-value">
          —
        </div>
        <div class="metric-track" />
      </div>
    </VueDraggable>
    <NDropdown v-if="allowEdit" trigger="click" placement="bottom-end" :options="actionOptions" @select="handleAction">
      <button class="monitor-more" type="button" :aria-label="t('deskModule.systemMonitor.systemStateActions')" :disabled="sortSaving" :title="sortError || dataError || t('deskModule.systemMonitor.systemStateActions')">
        <SvgIcon icon="lucide-ellipsis" />
      </button>
    </NDropdown>
    <button v-else-if="dataError" class="monitor-retry" type="button" :title="dataError" :aria-label="t('panelHome.retry')" @click="getData">
      <SvgIcon icon="lucide-refresh-cw" />
    </button>
    <span v-if="dataLoading" class="sr-only" aria-live="polite">{{ t('panelHome.loading') }}</span>
    <span v-if="sortError" class="sr-only" role="alert">{{ sortError }}</span>
    <Edit v-model:visible="editShowStatus" :monitor-data="editData" :index="editIndex" @done="handleSaveDone" />
    <NDropdown placement="bottom-start" trigger="manual" :x="dropdownMenuX" :y="dropdownMenuY" :options="getDropdownMenuOptions()" :show="dropdownShow" :on-clickoutside="onClickoutside" @select="handleRightMenuSelect" />
  </div>
</template>

<style scoped>
.system-monitor { display: flex; align-items: center; gap: 12px; min-width: 0; }
.monitor-items { display: flex; align-items: center; gap: 16px; min-width: 0; overflow-x: auto; }
.monitor-item { width: 72px; min-width: 72px; }
.monitor-item.sorting { cursor: grab; }
.monitor-more:focus-visible, .monitor-retry:focus-visible { outline: 3px solid var(--x-focus); outline-offset: 2px; }
.metric-head { display: flex; align-items: center; gap: 4px; overflow: hidden; color: var(--x-text2); font-size: 12px; line-height: 16px; white-space: nowrap; text-overflow: ellipsis; }
.metric-head span:first-child { overflow: hidden; text-overflow: ellipsis; }
.metric-error-dot { flex: none; width: 5px; height: 5px; border-radius: 50%; background: var(--x-red-sys); }
.metric-value { color: var(--x-text1); font-size: 15px; font-weight: 600; line-height: 18px; font-variant-numeric: tabular-nums; }
.metric-track { width: 72px; height: 3px; margin-top: 3px; overflow: hidden; border-radius: var(--x-r-pill); background: var(--x-sep); }
.metric-track span { display: block; height: 100%; border-radius: inherit; background: var(--x-primary); transition: width 150ms; }
.metric-track span.high { background: var(--x-orange-sys); }
.monitor-more, .monitor-retry { display: grid; flex: none; place-items: center; width: 32px; height: 32px; padding: 0; border: 0; border-radius: var(--x-r-glyph); background: transparent; color: var(--x-text2); cursor: pointer; }
.monitor-more:hover, .monitor-retry:hover { background: var(--x-hover); color: var(--x-text1); }
.sr-only { position: absolute; width: 1px; height: 1px; overflow: hidden; clip: rect(0, 0, 0, 0); white-space: nowrap; }
@media (max-width: 767px) { .system-monitor { flex: 1; } .monitor-items { display: grid; flex: 1; grid-template-columns: repeat(2, minmax(0, 1fr)); } .monitor-item { width: 100%; min-width: 0; } }
@media (prefers-reduced-motion: reduce) { .metric-track span { transition: none; } }
</style>
