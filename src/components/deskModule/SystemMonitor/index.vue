<script setup lang="ts">
import { nextTick, onMounted, ref } from 'vue'
import { VueDraggable } from 'vue-draggable-plus'
import { NButton, NDropdown, useDialog, useMessage } from 'naive-ui'
import CardCPU from './AppIconSystemMonitor/CPU.vue'
import Memory from './AppIconSystemMonitor/Memory.vue'
import Disk from './AppIconSystemMonitor/Disk.vue'
import { type MonitorData, MonitorType } from './typings'
import Edit from './Edit/index.vue'
import { deleteByIndex, getAll, saveAll } from './common'
import { usePanelState } from '@/store'
import { PanelPanelConfigStyleEnum } from '@/enums'
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
          backgroundColor: 'var(--xp-surface-card)',
          color: 'var(--xp-ink)',
          progressColor: 'var(--xp-primary)',
          progressRailColor: 'var(--xp-hairline)',
        },
        monitorType: MonitorType.cpu,
      },
    )

    // 生成并保存
    saveAll(monitorDatas.value)
  }
}

onMounted(() => {
  getData()
})

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
        <CardCPU
          v-if="item.monitorType === MonitorType.cpu" compact
          :card-type-style="PanelPanelConfigStyleEnum.info" :refresh-interval="5000"
          text-color="var(--xp-ink)" progress-color="var(--xp-primary)" progress-rail-color="var(--xp-hairline)"
        />
        <Memory
          v-else-if="item.monitorType === MonitorType.memory" compact
          :card-type-style="PanelPanelConfigStyleEnum.info" :refresh-interval="5000"
          text-color="var(--xp-ink)" progress-color="var(--xp-primary)" progress-rail-color="var(--xp-hairline)"
        />
        <Disk
          v-else-if="item.monitorType === MonitorType.disk" compact :path="item.extendParam?.path"
          :card-type-style="PanelPanelConfigStyleEnum.info" :refresh-interval="5000"
          text-color="var(--xp-ink)" progress-color="var(--xp-primary)" progress-rail-color="var(--xp-hairline)"
        />
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
.system-monitor { min-width: 0; width: 100%; padding: 12px 16px; border: 1px solid var(--xp-hairline); border-radius: var(--xp-radius-md); background: var(--xp-surface); box-shadow: var(--xp-card-shadow); }
.monitor-toolbar { display: flex; align-items: center; justify-content: space-between; min-height: 16px; }
.monitor-toolbar:has(.monitor-title) { margin-bottom: 8px; }
.monitor-title { color: var(--xp-muted); font-size: 13px; }
.monitor-actions { display: flex; gap: 8px; margin-left: auto; opacity: 0; transition: opacity 120ms; }
.monitor-actions.visible, .monitor-actions:focus-within { opacity: 1; }
.monitor-actions button { display: grid; width: 24px; height: 24px; place-items: center; color: var(--xp-muted); }
.monitor-actions button:focus-visible { outline: 2px solid var(--xp-primary); outline-offset: 2px; }
.monitor-items { display: flex; flex-wrap: wrap; gap: 16px; }
.monitor-item { flex: 1 1 88px; min-width: 88px; cursor: pointer; }
.monitor-item.sorting { cursor: grab; }
.save-sort { margin-top: 8px; }
.on-wallpaper { background: var(--xp-surface); }
.on-wallpaper :deep(.compact-label), .on-wallpaper .monitor-title { color: var(--xp-muted); }
@media (prefers-reduced-motion: reduce) { .monitor-actions { transition: none; } }
</style>
