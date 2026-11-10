<script setup lang="ts">
import { VueDraggable } from 'vue-draggable-plus'
import { NBackTop, NButton, NDrawer, NDrawerContent, NDropdown, NModal, NSkeleton, NSpin, useDialog, useMessage, useNotification } from 'naive-ui'
import { computed, h, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { AppStarter, EditItem } from './components'
import { Clock, SearchBox, SystemMonitor } from '@/components/deskModule'
import { GroupGlyph, ItemIcon, SvgIcon, XBrand } from '@/components/common'
import { deletes, edit as editItem, getListByGroupId, saveSort } from '@/api/panel/itemIcon'
import { getList as getGroupList } from '@/api/panel/itemIconGroup'

import { setTitle, updateLocalUserInfo } from '@/utils/cmn'
import { useAppStore, useAuthStore, usePanelState } from '@/store'
import { PanelStateNetworkModeEnum } from '@/enums'
import { VisitMode } from '@/enums/auth'
import { router } from '@/router'
import { t } from '@/locales'
import { useWallpaper } from '@/hooks/useWallpaper'

interface ItemGroup extends Panel.ItemIconGroup {
  hoverStatus: boolean
  items?: Panel.ItemInfo[]
}

interface DragEndEvent {
  from: HTMLElement
  to: HTMLElement
  oldDraggableIndex?: number
  newDraggableIndex?: number
}

const ms = useMessage()
const dialog = useDialog()
const notification = useNotification()
const panelState = usePanelState()
const appStore = useAppStore()
const authStore = useAuthStore()

const scrollContainerRef = ref<HTMLElement | undefined>(undefined)

const editItemInfoShow = ref<boolean>(false)
const editItemInfoData = ref<Panel.ItemInfo | null>(null)
const windowShow = ref<boolean>(false)
const windowSrc = ref<string>('')
const windowTitle = ref<string>('')

const windowIframeRef = ref(null)
const windowIframeIsLoad = ref<boolean>(false)

const dropdownMenuX = ref(0)
const dropdownMenuY = ref(0)
const dropdownShow = ref(false)
const currentRightSelectItem = ref<Panel.ItemInfo | null>(null)
const currentAddItenIconGroupId = ref<number | undefined>()

const settingModalShow = ref(false)

const items = ref<ItemGroup[]>([])
const groupsState = ref<'loading' | 'ready' | 'error'>('loading')
const filterItems = ref<ItemGroup[]>([])
const searchBoxKeyword = ref('')
const listFilter = ref('')
const arrangeMode = ref(false)
const listFilterRef = ref<HTMLInputElement | null>(null)
const rovingItemId = ref<number | null>(null)
const failedIcons = ref(new Set<number>())
const firstLoadMotion = ref(true)
const groupFade = ref(false)
let groupFadeTimer: ReturnType<typeof setTimeout>

// Icon tile colors honor the item's own backgroundColor; unset means white.
function iconBackground(item: Panel.ItemInfo) {
  return item.icon?.backgroundColor || 'var(--x-bg2)'
}

const selectedGroupId = ref<number | 'all' | null>('all')
const isGroupedView = computed(() => selectedGroupId.value === 'all' || !!listFilter.value.trim() || !!searchBoxKeyword.value)
const visibleGroups = computed(() => {
  const keyword = listFilter.value.trim().toLowerCase()
  const source = keyword ? items.value : filterItems.value
  const groups = isGroupedView.value ? source : source.filter(group => group.id === selectedGroupId.value)
  if (!keyword)
    return groups
  return groups.map(group => ({ ...group, items: group.items?.filter(item => `${item.title} ${item.url} ${item.lanUrl ?? ''}`.toLowerCase().includes(keyword)) })).filter(group => group.items?.length)
})
const totalCount = computed(() => items.value.reduce((total, group) => total + (group.items?.length ?? 0), 0))
const visibleCount = computed(() => visibleGroups.value.reduce((total, group) => total + (group.items?.length ?? 0), 0))
const firstVisibleItemId = computed(() => visibleGroups.value.find(group => group.items?.length)?.items?.[0]?.id ?? null)
const systemMonitorVisible = computed(() => panelState.panelConfig.systemMonitorShow
  && ((panelState.panelConfig.systemMonitorPublicVisitModeShow && authStore.visitMode === VisitMode.VISIT_MODE_PUBLIC)
    || authStore.visitMode === VisitMode.VISIT_MODE_LOGIN))
const overviewVisible = computed(() => panelState.panelConfig.clockShow || panelState.panelConfig.searchBoxShow || systemMonitorVisible.value)
const railRef = ref<HTMLElement | null>(null)
const railFocusKey = ref<string>('all')
const moreDrawerShow = ref(false)
const initialSettingsPanel = ref('UserInfo')
const now = ref(new Date())
const saveStatuses = ref<Record<number, 'saving' | 'saved' | 'error'>>({})
const saveStatusTimers = new Map<number, ReturnType<typeof setTimeout>>()
let clockTimer: ReturnType<typeof setInterval>
const tooltipKey = ref<string | null>(null)
const tooltipPosition = ref({ top: '0px', left: '0px' })
const { source: wallpaperSource, loaded: wallpaperLoaded, show: showWallpaper } = useWallpaper()
let tooltipTarget: HTMLElement | null = null
let tooltipOpenTimer: ReturnType<typeof setTimeout>
let tooltipCloseTimer: ReturnType<typeof setTimeout>
let railHot = false

const selectedGroup = computed(() => items.value.find(group => group.id === selectedGroupId.value))
const groupColors = ['blue', 'indigo', 'purple', 'pink', 'orange', 'mint', 'brown', 'teal', 'green', 'red'] as const
function groupColor(group: ItemGroup) {
  const index = items.value.findIndex(item => item.id === group.id)
  return groupColors[Math.max(0, index) % groupColors.length]
}
function groupColorStyle(group: ItemGroup) {
  const color = groupColor(group)
  return {
    '--group-fg': `var(--x-${color}-fg)`,
    '--group-tint': `var(--x-${color}-tint)`,
    '--group-sys': `var(--x-${color}-sys)`,
  }
}
function groupColorForId(id: number | 'all' | null) {
  const group = items.value.find(item => item.id === id)
  return group ? `var(--x-${groupColor(group)}-sys)` : 'var(--x-primary)'
}
const tooltipDotColor = computed(() => groupColorForId(Number(tooltipKey.value)))
const selectedGroupColor = computed(() => selectedGroup.value ? groupColorStyle(selectedGroup.value) : {})
const currentDate = computed(() => new Intl.DateTimeFormat(appStore.language === 'en-US' ? 'en-US' : 'zh-CN', { month: 'long', day: 'numeric', weekday: 'long' }).format(now.value))
const dockGroups = computed(() => {
  const selected = items.value.findIndex(group => group.id === selectedGroupId.value)
  if (selected < 0)
    return items.value.slice(0, 3)
  const nearby = items.value.map((group, index) => ({ group, index }))
    .filter(entry => entry.index !== selected)
    .sort((a, b) => Math.abs(a.index - selected) - Math.abs(b.index - selected))
    .slice(0, 2)
  return [...nearby.map(entry => entry.group), items.value[selected]]
    .sort((a, b) => items.value.indexOf(a) - items.value.indexOf(b))
})
function displayedItems(group: ItemGroup) {
  if (selectedGroupId.value === 'all' && !listFilter.value.trim() && !searchBoxKeyword.value && !arrangeMode.value)
    return group.items?.slice(0, 8) ?? []
  return group.items ?? []
}
function openImportSettings() {
  initialSettingsPanel.value = 'ImportExport'
  settingModalShow.value = true
}
function selectDockGroup(id: number | 'all') {
  selectGroup(id)
  moreDrawerShow.value = false
}

const tooltipContent = computed(() => {
  if (tooltipKey.value === 'brand')
    return { title: t('panelHome.backToTop'), count: null }
  if (tooltipKey.value === 'all')
    return { title: t('panelHome.all'), count: totalCount.value }
  const group = items.value.find(item => String(item.id) === tooltipKey.value)
  return { title: group?.title ?? '', count: group?.items?.length ?? 0 }
})

function positionTooltip() {
  if (!tooltipTarget)
    return
  const rect = tooltipTarget.getBoundingClientRect()
  tooltipPosition.value = { top: `${rect.top + rect.height / 2}px`, left: `${rect.right + 16}px` }
}

function showRailTooltip(key: string, event: FocusEvent | MouseEvent) {
  if (window.innerWidth < 768 || (event instanceof MouseEvent && !window.matchMedia('(hover: hover)').matches))
    return
  clearTimeout(tooltipOpenTimer)
  clearTimeout(tooltipCloseTimer)
  tooltipTarget = event.currentTarget as HTMLElement
  positionTooltip()
  if (event instanceof FocusEvent || railHot || tooltipKey.value)
    tooltipKey.value = key
  else
    tooltipOpenTimer = setTimeout(() => tooltipKey.value = key, 300)
  railHot = true
}

function leaveRail() {
  clearTimeout(tooltipOpenTimer)
  clearTimeout(tooltipCloseTimer)
  tooltipCloseTimer = setTimeout(() => {
    tooltipKey.value = null
    tooltipTarget = null
    railHot = false
  }, 100)
}

function onRailFocusOut(event: FocusEvent) {
  if (!railRef.value?.contains(event.relatedTarget as Node))
    leaveRail()
}

function onRailKeydown(event: KeyboardEvent) {
  const buttons = Array.from(railRef.value?.querySelectorAll<HTMLButtonElement>('[data-rail-key]') ?? [])
  const current = buttons.indexOf(document.activeElement as HTMLButtonElement)
  let next = current
  if (event.key === 'ArrowDown' || event.key === 'ArrowRight')
    next = Math.min(current + 1, buttons.length - 1)
  else if (event.key === 'ArrowUp' || event.key === 'ArrowLeft')
    next = Math.max(current - 1, 0)
  else if (event.key === 'Home')
    next = 0
  else if (event.key === 'End')
    next = buttons.length - 1
  else
    return
  event.preventDefault()
  if (next >= 0 && buttons[next]) {
    railFocusKey.value = buttons[next].dataset.railKey || 'all'
    buttons[next].focus()
    buttons[next].scrollIntoView({ block: 'nearest', inline: 'nearest' })
  }
}

function scrollToTop() {
  selectGroup('all')
  scrollContainerRef.value?.scrollTo({ top: 0, behavior: 'smooth' })
}

function selectGroup(id: number | 'all') {
  if (selectedGroupId.value !== id) {
    groupFade.value = false
    clearTimeout(groupFadeTimer)
    nextTick(() => {
      groupFade.value = true
      groupFadeTimer = setTimeout(() => groupFade.value = false, 150)
    })
  }
  selectedGroupId.value = id
  rovingItemId.value = null
  listFilter.value = ''
  if (!searchBoxKeyword.value)
    filterItems.value = items.value
}

function moveRowFocus(event: KeyboardEvent, direction: number) {
  if (arrangeMode.value)
    return
  const rows = Array.from(document.querySelectorAll<HTMLButtonElement>('.url-row:not(.add-row)'))
  if (!rows.length)
    return
  const current = rows.indexOf(document.activeElement as HTMLButtonElement)
  const next = current < 0 ? (direction > 0 ? 0 : rows.length - 1) : Math.max(0, Math.min(rows.length - 1, current + direction))
  event.preventDefault()
  rovingItemId.value = Number(rows[next].dataset.itemId)
  rows[next].focus()
}

watch(listFilter, () => rovingItemId.value = null)

function onFilterKeydown(event: KeyboardEvent) {
  if (arrangeMode.value)
    return
  if (event.key === 'ArrowDown' || event.key === 'ArrowUp') {
    moveRowFocus(event, event.key === 'ArrowDown' ? 1 : -1)
  }
  else if (event.key === 'Escape') {
    listFilter.value = ''
    listFilterRef.value?.blur()
  }
}

const stopFirstLoadWatch = watch(visibleCount, (count) => {
  if (count && firstLoadMotion.value) {
    setTimeout(() => firstLoadMotion.value = false, 400)
    stopFirstLoadWatch()
  }
})

function onFilterShortcut(event: KeyboardEvent) {
  if (event.key === '/' && !event.ctrlKey && !event.metaKey && !event.altKey && !(event.target instanceof HTMLInputElement) && !(event.target instanceof HTMLTextAreaElement) && !(event.target instanceof HTMLElement && event.target.isContentEditable)) {
    event.preventDefault()
    listFilterRef.value?.focus()
  }
}

function displayDomain(url: string) {
  try {
    return new URL(url).hostname.replace(/^www\./, '') || url
  }
  catch {
    return url
  }
}

watch(() => [panelState.panelConfig.showWallpaper, panelState.panelConfig.autoWallpaper, panelState.panelConfig.backgroundImageSrc], () => {
  if (panelState.panelConfig.showWallpaper)
    showWallpaper(panelState.panelConfig.backgroundImageSrc, panelState.panelConfig.autoWallpaper !== false)
}, { immediate: true })

watch(settingModalShow, (visible, previous) => {
  if (previous && !visible)
    getList()
})

function openPage(openMethod: number, url: string, title?: string) {
  switch (openMethod) {
    case 1:
      window.location.href = url
      break
    case 2:
      window.open(url)
      break
    case 3:
      windowShow.value = true
      windowSrc.value = url
      windowTitle.value = title || url
      windowIframeIsLoad.value = true
      break

    default:
      break
  }
}

function handleItemClick(item: Panel.ItemInfo) {
  if (arrangeMode.value)
    return

  let jumpUrl = ''

  if (item)
    jumpUrl = (panelState.networkMode === PanelStateNetworkModeEnum.lan ? item.lanUrl : item.url) as string
  if (item.lanUrl === '')
    jumpUrl = item.url

  openPage(item.openMethod, jumpUrl, item.title)
}

function handWindowIframeIdLoad(_payload: Event) {
  windowIframeIsLoad.value = false
}

async function getList() {
  groupsState.value = 'loading'
  try {
    const { code, data } = await getGroupList<Common.ListResponse<ItemGroup[]>>()
    if (code !== 0)
      throw new Error('group load failed')
    items.value = data.list.sort((a, b) => (a.sort ?? 0) - (b.sort ?? 0))
    failedIcons.value = new Set()
    if (selectedGroupId.value === null || (selectedGroupId.value !== 'all' && !items.value.some(group => group.id === selectedGroupId.value)))
      selectedGroupId.value = 'all'
    const results = await Promise.all(data.list.map((element, index) => element.id ? updateItemIconGroupByNet(index, element.id) : Promise.resolve(true)))
    if (results.some(result => !result))
      throw new Error('item load failed')
    filterItems.value = items.value
    groupsState.value = 'ready'
  }
  catch {
    groupsState.value = 'error'
  }
}

// 从后端获取组下面的图标
async function updateItemIconGroupByNet(itemIconGroupIndex: number, itemIconGroupId: number) {
  try {
    const res = await getListByGroupId<Common.ListResponse<Panel.ItemInfo[]>>(itemIconGroupId)
    if (res.code !== 0)
      return false
    items.value[itemIconGroupIndex].items = res.data.list
    return true
  }
  catch {
    return false
  }
}

function handleRightMenuSelect(key: string | number) {
  dropdownShow.value = false
  // console.log(currentRightSelectItem, key)
  let jumpUrl = panelState.networkMode === PanelStateNetworkModeEnum.lan ? currentRightSelectItem.value?.lanUrl : currentRightSelectItem.value?.url
  if (currentRightSelectItem.value?.lanUrl === '')
    jumpUrl = currentRightSelectItem.value.url
  switch (key) {
    case 'newWindows':
      window.open(jumpUrl)
      break
    case 'openWanUrl':
      if (currentRightSelectItem.value)
        openPage(currentRightSelectItem.value?.openMethod, currentRightSelectItem.value?.url, currentRightSelectItem.value?.title)
      break
    case 'openLanUrl':
      if (currentRightSelectItem.value && currentRightSelectItem.value.lanUrl)
        openPage(currentRightSelectItem.value?.openMethod, currentRightSelectItem.value.lanUrl, currentRightSelectItem.value?.title)
      break
    case 'edit':
      // 这里有个奇怪的问题，如果不使用{...}的方式 父组件的值会同步修改 标记一下
      handleEditItem({ ...currentRightSelectItem.value } as Panel.ItemInfo)
      break
    case 'delete':
    {
      const deletedItem = { ...currentRightSelectItem.value } as Panel.ItemInfo
      const deleteError = ref('')
      const deleting = ref(false)
      const confirmDialog = dialog.warning({
        title: t('common.warning'),
        content: () => h('div', [h('p', t('common.deleteConfirmByName', { name: deletedItem.title })), deleteError.value ? h('p', { style: 'color:var(--x-red-fg);margin-top:12px' }, deleteError.value) : null]),
        positiveText: t('common.confirm'),
        negativeText: t('common.cancel'),
        onPositiveClick: async () => {
          if (deleting.value)
            return false
          deleting.value = true
          confirmDialog.loading = true
          try {
            const { code, msg } = await deletes([deletedItem.id as number])
            if (code === 0) {
              getList()
              const toast = notification.success({
                title: t('common.deleteSuccess'),
                duration: 5000,
                action: () => h(NButton, {
                  onClick: async () => {
                    try {
                      const { code } = await editItem({ ...deletedItem, id: undefined })
                      if (code !== 0)
                        throw new Error(t('common.saveFail'))
                      toast.destroy()
                      getList()
                    }
                    catch {
                      ms.error(t('common.saveFail'))
                    }
                  },
                }, () => t('panelHome.undo')),
              })
            }
            else {
              deleteError.value = `${t('common.deleteFail')}: ${msg}`
              return false
            }
          }
          catch {
            deleteError.value = t('common.deleteFail')
            return false
          }
          finally {
            deleting.value = false
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

function handleContextMenu(e: MouseEvent, item: Panel.ItemInfo) {
  if (arrangeMode.value) {
    e.preventDefault()
    return
  }

  e.preventDefault()
  currentRightSelectItem.value = item
  dropdownShow.value = false
  nextTick().then(() => {
    dropdownShow.value = true
    dropdownMenuX.value = e.clientX
    dropdownMenuY.value = e.clientY
  })
}

function onClickoutside() {
  // message.info('clickoutside')
  dropdownShow.value = false
}

function handleEditSuccess(_item: Panel.ItemInfo) {
  getList()
}

function handleChangeNetwork(mode: PanelStateNetworkModeEnum) {
  panelState.setNetworkMode(mode)
  if (mode === PanelStateNetworkModeEnum.lan)
    ms.success(t('panelHome.changeToLanModelSuccess'))

  else
    ms.success(t('panelHome.changeToWanModelSuccess'))
}

function toggleArrangeMode() {
  arrangeMode.value = !arrangeMode.value
  if (arrangeMode.value) {
    listFilter.value = ''
    filterItems.value = items.value
    dropdownShow.value = false
  }
}

async function handleEndDrag(event: DragEndEvent) {
  if (event.from === event.to && event.oldDraggableIndex === event.newDraggableIndex)
    return

  const targetId = Number(event.to.dataset.groupId)
  const target = items.value.find(group => group.id === targetId)
  if (!target?.items)
    return

  saveStatuses.value[targetId] = 'saving'
  try {
    const { code } = await saveSort({
      itemIconGroupId: targetId,
      sortItems: target.items.map((item, index) => ({ id: item.id as number, sort: index + 1 })),
    })
    if (code === 0) {
      saveStatuses.value[targetId] = 'saved'
      saveStatusTimers.set(targetId, setTimeout(() => delete saveStatuses.value[targetId], 1500))
      return
    }
  }
  catch {}

  saveStatuses.value[targetId] = 'error'
  ms.error(t('panelHome.arrangeSaveFailed'))
  const sourceId = Number(event.from.dataset.groupId)
  for (const id of new Set([sourceId, targetId])) {
    const index = items.value.findIndex(group => group.id === id)
    if (index !== -1)
      updateItemIconGroupByNet(index, id)
  }
}

function getDropdownMenuOptions() {
  const dropdownMenuOptions = [
    {
      label: t('iconItem.newWindowOpen'),
      key: 'newWindows',
    },

  ]

  if (currentRightSelectItem.value?.lanUrl && panelState.networkMode === PanelStateNetworkModeEnum.wan) {
    dropdownMenuOptions.push({
      label: t('panelHome.openLanUrl'),
      key: 'openLanUrl',
    })
  }

  if (currentRightSelectItem.value?.lanUrl && panelState.networkMode === PanelStateNetworkModeEnum.lan) {
    dropdownMenuOptions.push({
      label: t('panelHome.openWanUrl'),
      key: 'openWanUrl',
    })
  }

  if (authStore.visitMode === VisitMode.VISIT_MODE_LOGIN) {
    dropdownMenuOptions.push({
      label: t('common.edit'),
      key: 'edit',
    }, {
      label: t('common.delete'),
      key: 'delete',
    })
  }

  return dropdownMenuOptions
}

onMounted(() => {
  clockTimer = setInterval(() => now.value = new Date(), 60000)
  window.addEventListener('keydown', onFilterShortcut)
  // 更新用户信息
  updateLocalUserInfo()
  getList()

  // 更新同步云端配置
  panelState.updatePanelConfigByCloud()

  // 设置标题
  if (panelState.panelConfig.logoText)
    setTitle(panelState.panelConfig.logoText)
})
onBeforeUnmount(() => {
  clearTimeout(groupFadeTimer)
  clearTimeout(tooltipOpenTimer)
  clearTimeout(tooltipCloseTimer)
  clearInterval(clockTimer)
  for (const timer of saveStatusTimers.values())
    clearTimeout(timer)
  window.removeEventListener('keydown', onFilterShortcut)
})

// 前端搜索过滤
function itemFrontEndSearch(keyword?: string) {
  keyword = keyword?.trim()
  searchBoxKeyword.value = keyword ?? ''
  if (keyword !== '' && panelState.panelConfig.searchBoxSearchIcon) {
    const filteredData = ref<ItemGroup[]>([])
    for (let i = 0; i < items.value.length; i++) {
      const element = items.value[i].items?.filter((item: Panel.ItemInfo) => {
        return (
          item.title.toLowerCase().includes(keyword?.toLowerCase() ?? '')
          || item.url.toLowerCase().includes(keyword?.toLowerCase() ?? '')
          || item.description?.toLowerCase().includes(keyword?.toLowerCase() ?? '')
        )
      })
      if (element && element.length > 0)
        filteredData.value.push({ ...items.value[i], items: element })
    }
    filterItems.value = filteredData.value
  }
  else {
    filterItems.value = items.value
  }
}

function handleEditItem(item: Panel.ItemInfo) {
  editItemInfoData.value = item
  editItemInfoShow.value = true
  currentAddItenIconGroupId.value = undefined
}

function handleAddItem(itemIconGroupId?: number) {
  editItemInfoData.value = null
  editItemInfoShow.value = true
  if (itemIconGroupId)
    currentAddItenIconGroupId.value = itemIconGroupId
}
</script>

<template>
  <div class="sun-main">
    <div v-if="panelState.panelConfig.showWallpaper" class="wallpaper" :style="{ filter: `blur(${panelState.panelConfig.backgroundBlur}px)`, backgroundImage: wallpaperLoaded ? `url(${wallpaperSource})` : 'none' }" />
    <div v-if="panelState.panelConfig.showWallpaper" class="wallpaper-mask" :style="{ backgroundColor: `rgba(18,18,20,${Math.max(0.35, panelState.panelConfig.backgroundMaskNumber ?? 0.35)})` }" />
    <div ref="scrollContainerRef" class="main-scroll">
      <div class="page-layout">
        <nav ref="railRef" class="group-rail" :aria-label="t('panelHome.groups')" @mouseleave="leaveRail" @focusout="onRailFocusOut" @keydown="onRailKeydown">
          <button class="rail-brand" type="button" :aria-label="t('panelHome.backToTop')" @click="scrollToTop" @mouseenter="showRailTooltip('brand', $event)" @focus="showRailTooltip('brand', $event)">
            <XBrand />
          </button>
          <button class="rail-button is-all" :class="{ active: selectedGroupId === 'all' }" type="button" data-rail-key="all" :tabindex="railFocusKey === 'all' ? 0 : -1" :aria-label="t('panelHome.groupAria', { name: t('panelHome.all'), count: totalCount })" :aria-current="selectedGroupId === 'all' ? 'page' : undefined" @click="selectGroup('all')" @focus="railFocusKey = 'all'; showRailTooltip('all', $event)" @mouseenter="showRailTooltip('all', $event)">
            <GroupGlyph icon="lucide-layout-grid" :title="t('panelHome.all')" :size="22" />
          </button>
          <span class="rail-divider" aria-hidden="true" />
          <div class="rail-scroll" @scroll="positionTooltip">
            <button v-for="group in items" :key="group.id" class="rail-button" :style="groupColorStyle(group)" :class="{ active: selectedGroupId === group.id }" type="button" :data-rail-key="String(group.id)" :tabindex="railFocusKey === String(group.id) ? 0 : -1" :aria-label="t('panelHome.groupAria', { name: group.title, count: group.items?.length ?? 0 })" :aria-current="selectedGroupId === group.id ? 'page' : undefined" @click="selectGroup(group.id as number)" @focus="railFocusKey = String(group.id); showRailTooltip(String(group.id), $event)" @mouseenter="showRailTooltip(String(group.id), $event)">
              <GroupGlyph :icon="group.icon" :title="group.title" :size="22" />
            </button>
          </div>
        </nav>
        <main class="page-content" :style="{ maxWidth: (panelState.panelConfig.maxWidth ?? '1200') + panelState.panelConfig.maxWidthUnit, paddingTop: `calc(24px + ${panelState.panelConfig.marginTop}%)`, paddingBottom: `calc(40px + ${panelState.panelConfig.marginBottom}%)` }">
          <header class="navigation-shell" :style="{ marginLeft: `${panelState.panelConfig.marginX}px`, marginRight: `${panelState.panelConfig.marginX}px` }">
            <div class="brand-block" :class="{ 'on-wallpaper': panelState.panelConfig.showWallpaper }">
              <h1>{{ panelState.panelConfig.logoText }}</h1><p>{{ currentDate }} · {{ t('panelHome.groupCount', { count: items.length }) }} · {{ t('panelHome.bookmarkCount', { count: totalCount }) }}</p>
            </div>
            <div class="header-actions">
              <label class="list-filter"><SvgIcon icon="iconamoon:search-fill" aria-hidden="true" /><input ref="listFilterRef" v-model="listFilter" type="search" :placeholder="t('panelHome.filterPlaceholder')" :aria-label="t('panelHome.filterPlaceholder')" @keydown="onFilterKeydown"><span v-if="listFilter.trim()" class="filter-count">{{ t('panelHome.resultsCount', { count: visibleCount }) }}</span><kbd v-else>/</kbd></label>
              <NButton v-if="authStore.visitMode === VisitMode.VISIT_MODE_LOGIN" class="quiet-button" :type="arrangeMode ? 'primary' : 'default'" @click="toggleArrangeMode">
                {{ arrangeMode ? t('panelHome.arrangeDone') : t('panelHome.arrange') }}
              </NButton>
              <NButton v-if="authStore.visitMode === VisitMode.VISIT_MODE_LOGIN" class="import-button" type="primary" @click="openImportSettings">
                {{ t('panelHome.importBookmarks') }}
              </NButton>
            </div>
          </header>
          <div class="content-shell" :style="{ marginLeft: `${panelState.panelConfig.marginX}px`, marginRight: `${panelState.panelConfig.marginX}px` }">
            <section v-if="overviewVisible" class="overview-band" :class="{ 'has-clock': panelState.panelConfig.clockShow, 'has-search': panelState.panelConfig.searchBoxShow, 'has-monitor': systemMonitorVisible }">
              <Clock v-if="panelState.panelConfig.clockShow" class="overview-clock" :hide-second="!panelState.panelConfig.clockShowSecond" />
              <SearchBox v-if="panelState.panelConfig.searchBoxShow" class="overview-search" @item-search="itemFrontEndSearch" />
              <SystemMonitor v-if="systemMonitorVisible" class="overview-monitor" :allow-edit="authStore.visitMode === VisitMode.VISIT_MODE_LOGIN" :show-title="panelState.panelConfig.systemMonitorShowTitle" />
              <div v-if="panelState.panelConfig.showWallpaper" class="wallpaper-card">
                <div class="wallpaper-preview" :style="wallpaperLoaded ? { backgroundImage: `url(${wallpaperSource})` } : {}">
                  <span v-if="!wallpaperLoaded">{{ t('panelHome.todayWallpaper') }}<small>320 × 180</small></span>
                </div><span>{{ t('panelHome.todayWallpaper') }}</span>
              </div>
            </section>
            <h2 v-if="selectedGroupId !== 'all' && selectedGroup" class="selected-group-title" :style="selectedGroupColor">
              <span class="title-glyph"><GroupGlyph :icon="selectedGroup.icon" :title="selectedGroup.title" :size="20" /></span>{{ selectedGroup.title }} <span class="title-count">{{ selectedGroup.items?.length ?? 0 }}</span>
            </h2>
            <div v-if="groupsState === 'loading'" class="content-groups" aria-busy="true">
              <NSkeleton v-for="index in 2" :key="index" height="220px" class="rounded-card" />
            </div>
            <div v-else-if="groupsState === 'error'" class="empty-list" role="alert">
              {{ t('panelHome.loadFailed') }} <button class="raised-button" type="button" @click="getList">
                {{ t('panelHome.retry') }}
              </button>
            </div>
            <div v-else-if="visibleCount || (arrangeMode && visibleGroups.length) || (!isGroupedView && visibleGroups.length)" class="content-groups" :class="{ 'group-fade': groupFade, 'first-load-motion': firstLoadMotion }">
              <section v-for="group in visibleGroups" :key="group.id" class="group-section" :style="{ ...groupColorStyle(group), ...(firstLoadMotion ? { animationDelay: `${Math.min(9, visibleGroups.indexOf(group)) * 30}ms` } : {}) }">
                <div class="group-heading">
                  <span class="heading-glyph"><GroupGlyph :icon="group.icon" :title="group.title" :size="22" /></span><h2>{{ group.title }}</h2><span class="heading-count">{{ group.items?.length ?? 0 }}</span><span v-if="saveStatuses[group.id as number]" class="save-status" :class="saveStatuses[group.id as number]">{{ t(`panelHome.${saveStatuses[group.id as number]}`) }}</span><button v-if="authStore.visitMode === VisitMode.VISIT_MODE_LOGIN" class="group-add raised-button" type="button" :aria-label="t('common.add')" @click="handleAddItem(group.id)">
                    <SvgIcon icon="typcn:plus" />
                  </button>
                </div>
                <VueDraggable v-if="group.items" v-model="group.items" item-key="id" :animation="150" class="url-list" :class="{ 'arrange-drop-zone': arrangeMode }" group="items" draggable=".url-row:not(.add-row)" :data-group-id="group.id" :disabled="!arrangeMode || !!listFilter.trim()" :delay="200" :delay-on-touch-only="true" ghost-class="sortable-ghost" @end="handleEndDrag">
                  <button v-for="item in displayedItems(group)" :key="item.id" class="url-row" :class="{ sorting: arrangeMode }" :data-item-id="item.id" :tabindex="(rovingItemId ?? firstVisibleItemId) === item.id ? 0 : -1" type="button" :title="item.title" @focus="rovingItemId = item.id as number" @click="handleItemClick(item)" @keydown.down="moveRowFocus($event, 1)" @keydown.up="moveRowFocus($event, -1)" @contextmenu="(e) => handleContextMenu(e, item)">
                    <img v-if="item.icon?.itemType === 2 && item.icon.src && !failedIcons.has(item.id as number)" class="row-icon" :style="{ backgroundColor: iconBackground(item) }" :src="item.icon.src" alt="" loading="lazy" @error="failedIcons.add(item.id as number)"><ItemIcon v-else-if="item.icon?.itemType === 3" class="row-icon" :item-icon="item.icon" :size="20" :force-background="iconBackground(item)" /><span v-else class="row-placeholder"><GroupGlyph icon="lucide-globe" :title="item.title" :size="18" /></span>
                    <span class="row-copy"><span class="row-name">{{ item.title }}</span><span class="row-domain">{{ displayDomain(item.url) }}</span></span><span class="row-arrow" aria-hidden="true">↗</span>
                  </button>
                  <div v-if="arrangeMode && !group.items.length" class="drop-target">
                    {{ t('panelHome.dropHere') }}
                  </div>
                </VueDraggable>
                <button v-if="selectedGroupId === 'all' && !listFilter.trim() && !searchBoxKeyword && !arrangeMode && (group.items?.length ?? 0) > 8" class="view-all raised-button" type="button" @click="selectGroup(group.id as number)">
                  {{ t('panelHome.viewAll', { count: group.items?.length ?? 0 }) }}
                </button>
              </section>
            </div>
            <div v-else class="empty-list">
              {{ t('panelHome.noMatches') }}
            </div>
          </div>
          <div class="footer" :class="{ 'on-wallpaper': panelState.panelConfig.showWallpaper }" v-html="panelState.panelConfig.footerHtml" />
        </main>
      </div>
    </div>
    <Teleport to="body">
      <Transition name="rail-tooltip">
        <div v-if="tooltipKey" class="rail-tooltip" :style="tooltipPosition" role="tooltip">
          <span v-if="tooltipKey !== 'brand'" class="rail-tooltip-dot" :style="{ backgroundColor: tooltipDotColor }" aria-hidden="true" />{{ tooltipContent.title }}<span v-if="tooltipContent.count !== null" class="rail-tooltip-count"> · {{ tooltipContent.count }}</span>
        </div>
      </Transition>
    </Teleport>
    <div class="mobile-dock-band" /><nav class="mobile-dock" :aria-label="t('panelHome.groups')">
      <button class="dock-button is-all" :class="{ active: selectedGroupId === 'all' }" type="button" :aria-label="t('panelHome.all')" @click="selectDockGroup('all')">
        <GroupGlyph icon="lucide-layout-grid" :title="t('panelHome.all')" :size="22" /><span v-if="selectedGroupId === 'all'">{{ t('panelHome.all') }}</span>
      </button><button v-for="group in dockGroups" :key="group.id" class="dock-button" :class="{ active: selectedGroupId === group.id }" :style="groupColorStyle(group)" type="button" :aria-label="group.title" @click="selectDockGroup(group.id as number)">
        <GroupGlyph :icon="group.icon" :title="group.title" :size="22" /><span v-if="selectedGroupId === group.id">{{ group.title }}</span>
      </button><button class="dock-button more-button" type="button" :aria-label="t('panelHome.moreGroups')" @click="moreDrawerShow = true">
        <SvgIcon icon="mingcute:more-1-fill" />
      </button>
    </nav>
    <NDrawer v-model:show="moreDrawerShow" placement="bottom" height="min(70vh, 620px)" class="groups-drawer">
      <NDrawerContent :title="t('panelHome.groups')" closable>
        <button v-for="group in items" :key="group.id" class="drawer-group" :style="groupColorStyle(group)" type="button" @click="selectDockGroup(group.id as number)">
          <span class="heading-glyph"><GroupGlyph :icon="group.icon" :title="group.title" :size="22" /></span><span>{{ group.title }}</span><span class="drawer-count">{{ group.items?.length ?? 0 }}</span>
        </button>
      </NDrawerContent>
    </NDrawer>
    <NDropdown placement="bottom-start" trigger="manual" :x="dropdownMenuX" :y="dropdownMenuY" :options="getDropdownMenuOptions()" :show="dropdownShow" :on-clickoutside="onClickoutside" @select="handleRightMenuSelect" />
    <div class="fixed-element">
      <div class="floating-actions">
        <NButton v-if="panelState.networkMode === PanelStateNetworkModeEnum.lan && panelState.panelConfig.netModeChangeButtonShow" class="floating-button" :title="t('panelHome.changeToWanModel')" @click="handleChangeNetwork(PanelStateNetworkModeEnum.wan)">
          <template #icon>
            <SvgIcon icon="material-symbols:lan-outline-rounded" />
          </template>
        </NButton><NButton v-if="panelState.networkMode === PanelStateNetworkModeEnum.wan && panelState.panelConfig.netModeChangeButtonShow" class="floating-button" :title="t('panelHome.changeToLanModel')" @click="handleChangeNetwork(PanelStateNetworkModeEnum.lan)">
          <template #icon>
            <SvgIcon icon="mdi:wan" />
          </template>
        </NButton><NButton v-if="authStore.visitMode === VisitMode.VISIT_MODE_LOGIN" class="floating-button" :aria-label="t('appLauncher.title')" @click="initialSettingsPanel = 'UserInfo'; settingModalShow = true">
          <template #icon>
            <SvgIcon icon="majesticons-applications" />
          </template>
        </NButton><NButton v-if="authStore.visitMode === VisitMode.VISIT_MODE_PUBLIC" class="floating-button" :title="t('panelHome.goToLogin')" @click="router.push('/login')">
          <template #icon>
            <SvgIcon icon="material-symbols:account-circle" />
          </template>
        </NButton>
      </div>
    </div>
    <AppStarter v-model:visible="settingModalShow" :initial-panel="initialSettingsPanel" />
    <NBackTop class="rail-back-top" :listen-to="() => scrollContainerRef" :right="16" :bottom="16">
      <div class="back-top-button">
        <SvgIcon icon="icon-park-outline:to-top" />
      </div>
    </NBackTop>
    <EditItem v-model:visible="editItemInfoShow" :item-info="editItemInfoData" :item-group-id="currentAddItenIconGroupId" @done="handleEditSuccess" />
    <NModal v-model:show="windowShow" :mask-closable="false" preset="card" style="max-width: 1000px;height: 600px" :bordered="false" size="small" role="dialog" aria-modal="true">
      <template #header>
        <div class="flex items-center">
          <span class="mr-5">{{ windowTitle }}</span><NSpin v-if="windowIframeIsLoad" size="small" />
        </div>
      </template><div class="w-full h-full rounded-small overflow-hidden bg-bg1">
        <div v-if="windowIframeIsLoad" class="flex flex-col p-5">
          <NSkeleton height="50px" width="100%" class="rounded-icon" /><NSkeleton height="180px" width="100%" class="mt-5 rounded-icon" />
        </div><iframe v-show="!windowIframeIsLoad" id="windowIframeId" ref="windowIframeRef" :src="windowSrc" class="w-full h-full" frameborder="0" @load="handWindowIframeIdLoad" />
      </div>
    </NModal>
  </div>
</template>

<style>
body, html { overflow: hidden; background: var(--x-bg0); color: var(--x-text1); }
::selection { background: var(--x-primary); color: var(--x-on-primary); }
.rail-tooltip { position: fixed; z-index: 3000; transform: translateY(-50%); padding: 8px 14px; border-radius: 999px; background: var(--x-inverse-bg); color: var(--x-inverse-fg); font: 13px/1.4 var(--x-font); white-space: nowrap; pointer-events: none; }
.rail-tooltip-dot { display: inline-block; width: 6px; height: 6px; margin-right: 7px; border-radius: 50%; vertical-align: 2px; }
.rail-tooltip-count { opacity: .65; font-variant-numeric: tabular-nums; }
.rail-tooltip-enter-active, .rail-tooltip-leave-active { transition: opacity 120ms ease, translate 120ms ease; }
.rail-tooltip-enter-from, .rail-tooltip-leave-to { opacity: 0; translate: -4px 0; }
.groups-drawer .n-drawer-content { border-radius: 32px 32px 0 0; background: var(--x-bg0); }
@media (max-width: 767px) { .rail-tooltip { display: none; } .rail-back-top { bottom: calc(136px + env(safe-area-inset-bottom)) !important; } }
@media (prefers-reduced-motion: reduce) { .rail-tooltip-enter-active, .rail-tooltip-leave-active { transition: none; } }
</style>

<style scoped>
.sun-main { position: relative; width: 100%; height: 100%; background: var(--x-bg0); color: var(--x-text1); font-family: var(--x-font); user-select: none; }
.wallpaper, .wallpaper-mask { position: absolute; inset: 0; pointer-events: none; }
.wallpaper { background-position: center; background-size: cover; }
.main-scroll { position: relative; width: 100%; height: 100%; overflow: auto; }
.page-layout { display: flex; align-items: flex-start; gap: 16px; min-height: 100%; padding: 16px; }
.group-rail { position: sticky; top: 16px; z-index: 10; display: flex; flex: none; flex-direction: column; align-items: center; box-sizing: border-box; width: 88px; height: calc(100vh - 32px); padding: 16px 0; border-radius: var(--x-r-card); background: var(--x-bg1); }
.rail-brand, .rail-button { display: grid; flex: none; place-items: center; width: 44px; height: 44px; padding: 0; border: 0; border-radius: var(--x-r-icon); cursor: pointer; transition: background-color 150ms, color 150ms; }
.rail-brand { margin-bottom: 16px; background: var(--x-bg2); color: var(--x-text1); }
.rail-brand :deep(svg) { width: 27px; height: 27px; }
.rail-scroll { display: flex; flex: 1 1 auto; flex-direction: column; align-items: center; gap: 8px; box-sizing: border-box; width: 100%; min-height: 0; padding: 8px 0 16px; overflow-x: hidden; overflow-y: auto; overscroll-behavior: contain; scrollbar-width: none; }
.rail-scroll::-webkit-scrollbar { display: none; }
.rail-divider { flex: none; width: 24px; height: 2px; margin: 12px 0 4px; background: var(--x-sep); }
.rail-button { background: var(--group-tint, var(--x-blue-tint)); color: var(--group-fg, var(--x-primary)); }
.rail-button.is-all { background: var(--x-blue-tint); color: var(--x-blue-fg); }
.rail-button:hover { background: var(--x-bg2); }
.rail-button.active { background: var(--group-fg, var(--x-primary)); color: var(--x-bg0); }
.rail-button.is-all.active { background: var(--x-primary); color: var(--x-on-primary); }
.page-content { flex: 1; min-width: 0; width: 100%; margin: 0 auto; padding-right: 56px; padding-left: 56px; }
.navigation-shell { display: flex; justify-content: space-between; align-items: center; gap: 24px; min-height: 80px; }
.brand-block h1 { margin: 0; font-size: 34px; font-weight: 700; line-height: 1.15; letter-spacing: -.02em; }
.brand-block p { margin: 8px 0 0; color: var(--x-text2); font-size: 13px; font-variant-numeric: tabular-nums; }
.brand-block.on-wallpaper h1, .brand-block.on-wallpaper p { color: #fff; }
.header-actions { display: flex; align-items: center; gap: 12px; }
.list-filter { display: flex; align-items: center; gap: 12px; min-width: 280px; width: 320px; height: 44px; padding: 0 16px; border: 0; border-radius: var(--x-r-pill); background: var(--x-bg1); color: var(--x-text2); }
.list-filter input { flex: 1; min-width: 0; border: 0; outline: 0; background: transparent; color: var(--x-text1); user-select: text; }
.list-filter input::placeholder { color: var(--x-text2); }
.list-filter kbd { padding: 2px 8px; border-radius: 7px; background: var(--x-bg2); color: var(--x-text2); font-size: 13px; }
.list-filter:focus-within { outline: 2px solid var(--x-primary); box-shadow: 0 0 0 5px var(--x-focus); }
.filter-count { color: var(--x-text2); font-size: 13px; white-space: nowrap; }
.quiet-button, .import-button { height: 44px; white-space: nowrap; }
.quiet-button { background: var(--x-bg1); }
.content-shell { padding-top: 48px; }
.overview-band { display: grid; grid-template-columns: repeat(12, minmax(0, 1fr)); align-items: center; gap: 24px; padding: 32px; border-radius: var(--x-r-card); background: var(--x-bg1); }
.overview-clock { grid-column: span 4; }
.overview-search { grid-column: span 8; min-width: 0; }
.overview-monitor { grid-column: span 12; }
.overview-band:has(.wallpaper-card) .overview-monitor { grid-column: span 9; }
.overview-band.has-clock:not(.has-search) .overview-clock, .overview-band.has-search:not(.has-clock) .overview-search { grid-column: span 12; }
.wallpaper-card { grid-column: span 3; min-width: 0; padding: 16px; border-radius: var(--x-r-small); background: var(--x-bg2); font-size: 13px; }
.wallpaper-preview { display: grid; place-items: center; width: 100%; aspect-ratio: 16 / 9; margin-bottom: 10px; border-radius: 14px; background: var(--x-ph); background-position: center; background-size: cover; color: var(--x-text2); text-align: center; }
.wallpaper-preview small { display: block; margin-top: 4px; }
.selected-group-title { display: flex; align-items: center; gap: 12px; margin: 40px 0 0; font-size: 18px; font-weight: 600; }
.title-glyph, .heading-glyph { display: grid; flex: none; place-items: center; background: var(--group-tint); color: var(--group-fg); }
.title-glyph { width: 28px; height: 28px; border-radius: 10px; }
.title-count { color: var(--x-text2); font-size: 13px; font-variant-numeric: tabular-nums; }
.content-groups { display: grid; grid-template-columns: repeat(auto-fill, minmax(min(100%, 460px), 1fr)); gap: 40px; margin-top: 48px; }
.selected-group-title + .content-groups { margin-top: 24px; }
.group-section { min-width: 0; padding: 24px; border-radius: var(--x-r-card); background: var(--x-bg1); }
.group-heading { display: flex; align-items: center; gap: 12px; margin-bottom: 20px; }
.heading-glyph { width: 44px; height: 44px; border-radius: var(--x-r-icon); }
.group-heading h2 { flex: 1; min-width: 0; margin: 0; overflow: hidden; color: var(--x-text1); font-size: 17px; font-weight: 600; text-overflow: ellipsis; white-space: nowrap; }
.heading-count, .save-status { padding: 5px 9px; border-radius: 999px; background: var(--x-bg2); color: var(--x-text2); font-size: 13px; font-variant-numeric: tabular-nums; }
.save-status.saved { color: var(--x-green-fg); }.save-status.error { color: var(--x-red-fg); }
.raised-button { display: inline-flex; justify-content: center; align-items: center; min-height: 44px; padding: 0 16px; border: 0; border-radius: 999px; background: var(--x-bg2); color: var(--x-text1); cursor: pointer; }
.raised-button:hover { background: var(--x-hover); }
.group-add { flex: none; width: 44px; padding: 0; font-size: 20px; }
.url-list { display: grid; grid-template-columns: repeat(auto-fill, minmax(min(100%, 208px), 1fr)); gap: 8px; min-height: 64px; }
.url-row { display: flex; align-items: center; gap: 12px; min-width: 0; height: 64px; padding: 0 16px; border: 0; border-radius: var(--x-r-small); background: var(--x-bg2); color: var(--x-text1); text-align: left; cursor: pointer; transition: background-color 150ms; }
.url-row:hover { background: var(--x-hover); }
.url-row.sorting { cursor: grab; }.url-row.sorting:active { cursor: grabbing; }.url-row.sortable-ghost { opacity: .5; }
.row-icon, .row-placeholder { flex: none; width: 32px; height: 32px; border-radius: 10px; }
.row-icon { object-fit: contain; padding: 3px; }
.row-placeholder { display: grid; place-items: center; background: var(--group-tint); color: var(--group-fg); }
.row-copy { display: flex; flex: 1; flex-direction: column; gap: 2px; min-width: 0; }
.row-name, .row-domain { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.row-name { font-size: 15px; font-weight: 500; }.row-domain { color: var(--x-text2); font-size: 13px; }
.row-arrow { flex: none; color: var(--x-icon3); transition: color 150ms, transform 150ms; }.url-row:hover .row-arrow { color: var(--x-text1); transform: translate(2px, -2px); }
.drop-target { display: grid; place-items: center; height: 64px; border-radius: var(--x-r-small); background: var(--x-bg2); color: var(--x-text2); }
.view-all { margin-top: 20px; }.empty-list { padding: 48px; color: var(--x-text2); text-align: center; }
.footer { margin: 40px 0 0; color: var(--x-text2); font-size: 13px; }.footer.on-wallpaper, .footer.on-wallpaper :deep(*) { color: #fff; }
.fixed-element { position: fixed; z-index: 20; right: 16px; bottom: 72px; }.floating-actions { display: flex; flex-direction: column; gap: 8px; }.floating-button, .back-top-button { display: grid; place-items: center; width: 44px; height: 44px; border: 0; border-radius: var(--x-r-icon); background: var(--x-bg1); color: var(--x-text1); cursor: pointer; }
.rail-back-top { border-radius: var(--x-r-icon); background: transparent; }.back-top-button { font-size: 20px; }
.mobile-dock-band, .mobile-dock { display: none; }.drawer-group { display: flex; align-items: center; gap: 12px; width: 100%; min-height: 64px; padding: 8px 16px; border: 0; border-radius: var(--x-r-small); background: var(--x-bg1); color: var(--x-text1); text-align: left; }.drawer-group + .drawer-group { margin-top: 8px; }.drawer-count { margin-left: auto; color: var(--x-text2); }
.rail-brand:focus-visible, .rail-button:focus-visible, .dock-button:focus-visible, .drawer-group:focus-visible, .url-row:focus-visible, .raised-button:focus-visible, .floating-button:focus-visible { outline: 3px solid var(--x-focus); outline-offset: 2px; }
@keyframes first-card { from { opacity: .2; transform: translateY(6px); } to { opacity: 1; transform: translateY(0); } }
.first-load-motion .group-section { animation: first-card 150ms ease-out both; }.group-fade { animation: first-card 150ms ease-out; }
@media (max-width: 1179px) { .overview-clock, .overview-search, .overview-monitor, .overview-band:has(.wallpaper-card) .overview-monitor { grid-column: span 12; }.wallpaper-card { grid-column: span 6; } }
@media (max-width: 767px) { .page-content { padding-right: 0; padding-left: 0; }.page-layout { padding: 16px 16px calc(136px + env(safe-area-inset-bottom)); }.group-rail { display: none; }.navigation-shell { display: block; }.header-actions { flex-wrap: wrap; margin-top: 24px; }.list-filter { order: -1; width: 100%; }.content-shell { padding-top: 32px; }.overview-band { display: flex; flex-direction: column; align-items: stretch; padding: 24px; }.content-groups { grid-template-columns: 1fr; gap: 24px; margin-top: 40px; }.url-list { grid-template-columns: 1fr; }.mobile-dock-band { position: fixed; z-index: 19; right: 0; bottom: 0; left: 0; display: block; height: calc(112px + env(safe-area-inset-bottom)); background: var(--x-bg0); }.mobile-dock { position: fixed; z-index: 21; right: 16px; bottom: calc(24px + env(safe-area-inset-bottom)); left: 16px; display: flex; align-items: center; justify-content: space-around; gap: 4px; height: 64px; padding: 8px; border-radius: 24px; background: var(--x-bg1); }.dock-button { display: flex; align-items: center; justify-content: center; gap: 8px; min-width: 44px; max-width: 46%; height: 44px; padding: 0 10px; border: 0; border-radius: var(--x-r-icon); background: var(--group-tint, var(--x-blue-tint)); color: var(--group-fg, var(--x-primary)); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }.dock-button.active { flex: 1; background: var(--x-primary); color: var(--x-on-primary); }.dock-button.active span { overflow: hidden; text-overflow: ellipsis; }.more-button { background: var(--x-bg2); color: var(--x-text1); }.fixed-element { bottom: calc(136px + env(safe-area-inset-bottom)); }.brand-block h1 { font-size: 30px; } }
@media (prefers-reduced-motion: reduce) { .rail-brand, .rail-button, .url-row, .row-arrow { transition: none; }.first-load-motion .group-section, .group-fade { animation: none; }.url-row:hover .row-arrow { transform: none; } }
</style>
