<script setup lang="ts">
import { VueDraggable } from 'vue-draggable-plus'
import { NBackTop, NButton, NDropdown, NModal, NSkeleton, NSpin, useDialog, useMessage } from 'naive-ui'
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { AppStarter, EditItem } from './components'
import { Clock, SearchBox, SystemMonitor } from '@/components/deskModule'
import { GroupGlyph, ItemIcon, SvgIcon } from '@/components/common'
import logo from '@/assets/logo.svg'
import { deletes, getListByGroupId, saveSort } from '@/api/panel/itemIcon'
import { getList as getGroupList } from '@/api/panel/itemIconGroup'

import { setTitle, updateLocalUserInfo } from '@/utils/cmn'
import { useAuthStore, usePanelState } from '@/store'
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
const panelState = usePanelState()
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
const filterItems = ref<ItemGroup[]>([])
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
  return item.icon?.backgroundColor || 'var(--xp-surface-card)'
}

function iconTextColor(item: Panel.ItemInfo) {
  const color = iconBackground(item).trim()
  if (color === 'var(--xp-surface-card)')
    return 'var(--xp-ink)'
  let r = 255
  let g = 255
  let b = 255
  let a = 1
  const hex = color.match(/^#([0-9a-f]{3,8})$/i)?.[1]
  const rgba = color.match(/^rgba?\(\s*([\d.]+)\s*,\s*([\d.]+)\s*,\s*([\d.]+)\s*(?:,\s*([\d.]+)\s*)?\)$/i)
  if (hex && (hex.length === 3 || hex.length === 6 || hex.length === 8)) {
    const full = hex.length === 3 ? hex.split('').map(c => c + c).join('') : hex
    r = parseInt(full.slice(0, 2), 16)
    g = parseInt(full.slice(2, 4), 16)
    b = parseInt(full.slice(4, 6), 16)
    if (full.length === 8)
      a = parseInt(full.slice(6, 8), 16) / 255
  }
  else if (rgba) {
    r = Number(rgba[1])
    g = Number(rgba[2])
    b = Number(rgba[3])
    a = rgba[4] === undefined ? 1 : Number(rgba[4])
  }
  // A mostly transparent tile sits on the dark panel, so treat it as dark.
  if (a < 0.5)
    return '#faf9f5'
  return (0.299 * r + 0.587 * g + 0.114 * b) / 255 > 0.6 ? '#141413' : '#faf9f5'
}
const selectedGroupId = ref<number | 'all' | null>('all')
const isGroupedView = computed(() => selectedGroupId.value === 'all' || !!listFilter.value.trim())
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
const railScrollRef = ref<HTMLElement | null>(null)
const railFocusKey = ref<string>('all')
const railIndicator = ref({ transform: 'translateY(0px)', opacity: 0 })
const tooltipKey = ref<string | null>(null)
const tooltipPosition = ref({ top: '0px', left: '0px' })
const { source: wallpaperSource, loaded: wallpaperLoaded, show: showWallpaper } = useWallpaper()
let tooltipTarget: HTMLElement | null = null
let tooltipOpenTimer: ReturnType<typeof setTimeout>
let tooltipCloseTimer: ReturnType<typeof setTimeout>
let railHot = false

const selectedGroup = computed(() => items.value.find(group => group.id === selectedGroupId.value))
const tooltipContent = computed(() => {
  if (tooltipKey.value === 'brand')
    return { title: t('panelHome.backToTop'), count: null }
  if (tooltipKey.value === 'all')
    return { title: t('panelHome.all'), count: totalCount.value }
  const group = items.value.find(item => String(item.id) === tooltipKey.value)
  return { title: group?.title ?? '', count: group?.items?.length ?? 0 }
})

function updateRailIndicator() {
  const rail = railRef.value
  const selected = Array.from(rail?.querySelectorAll<HTMLElement>('[data-rail-key]') ?? [])
    .find(button => button.dataset.railKey === String(selectedGroupId.value))
  if (!rail || !selected)
    return
  const railRect = rail.getBoundingClientRect()
  const rect = selected.getBoundingClientRect()
  const scrollRect = railScrollRef.value?.getBoundingClientRect()
  const mobile = window.innerWidth < 640
  let visible = false
  if (scrollRect)
    visible = mobile ? (rect.right > scrollRect.left && rect.left < scrollRect.right) : (rect.bottom > scrollRect.top && rect.top < scrollRect.bottom)
  railIndicator.value = {
    transform: mobile ? `translateX(${rect.left - railRect.left + 10}px)` : `translateY(${rect.top - railRect.top + 10}px)`,
    opacity: visible ? 1 : 0,
  }
}

function positionTooltip() {
  if (!tooltipTarget)
    return
  const rect = tooltipTarget.getBoundingClientRect()
  tooltipPosition.value = { top: `${rect.top + rect.height / 2}px`, left: `${rect.right + 12}px` }
}

function showRailTooltip(key: string, event: FocusEvent | MouseEvent) {
  if (window.innerWidth < 640 || (event instanceof MouseEvent && !window.matchMedia('(hover: hover)').matches))
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

watch([selectedGroupId, items], () => nextTick(() => {
  const selected = Array.from(railRef.value?.querySelectorAll<HTMLElement>('[data-rail-key]') ?? [])
    .find(button => button.dataset.railKey === String(selectedGroupId.value))
  if (window.innerWidth < 640)
    selected?.scrollIntoView({ block: 'nearest', inline: 'nearest' })
  updateRailIndicator()
}), { deep: true })

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
  filterItems.value = items.value
  nextTick(updateRailIndicator)
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

function getList() {
  // 获取组数据
  getGroupList<Common.ListResponse<ItemGroup[]>>().then(({ code, data }) => {
    if (code !== 0)
      return
    items.value = data.list.sort((a, b) => (a.sort ?? 0) - (b.sort ?? 0))
    failedIcons.value = new Set()
    if (selectedGroupId.value === null || (selectedGroupId.value !== 'all' && !items.value.some(group => group.id === selectedGroupId.value)))
      selectedGroupId.value = 'all'
    for (let i = 0; i < data.list.length; i++) {
      const element = data.list[i]
      if (element.id)
        updateItemIconGroupByNet(i, element.id)
    }
    filterItems.value = items.value
    nextTick(updateRailIndicator)
    // console.log(items)
  })
}

// 从后端获取组下面的图标
function updateItemIconGroupByNet(itemIconGroupIndex: number, itemIconGroupId: number) {
  getListByGroupId<Common.ListResponse<Panel.ItemInfo[]>>(itemIconGroupId).then((res) => {
    if (res.code === 0)
      items.value[itemIconGroupIndex].items = res.data.list
  })
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
      dialog.warning({
        title: t('common.warning'),
        content: t('common.deleteConfirmByName', { name: currentRightSelectItem.value?.title }),
        positiveText: t('common.confirm'),
        negativeText: t('common.cancel'),
        onPositiveClick: () => {
          deletes([currentRightSelectItem.value?.id as number]).then(({ code, msg }) => {
            if (code === 0) {
              ms.success(t('common.deleteSuccess'))
              getList()
            }
            else {
              ms.error(`${t('common.deleteFail')}:${msg}`)
            }
          })
        },
      })

      break
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

  try {
    const { code } = await saveSort({
      itemIconGroupId: targetId,
      sortItems: target.items.map((item, index) => ({ id: item.id as number, sort: index + 1 })),
    })
    if (code === 0)
      return
  }
  catch {}

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
  window.addEventListener('resize', updateRailIndicator)
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
  window.removeEventListener('resize', updateRailIndicator)
  window.removeEventListener('keydown', onFilterShortcut)
})

// 前端搜索过滤
function itemFrontEndSearch(keyword?: string) {
  keyword = keyword?.trim()
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
    selectedGroupId.value = 'all'
    nextTick(updateRailIndicator)
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
  <div class="w-full h-full sun-main">
    <div
      v-if="panelState.panelConfig.showWallpaper" class="cover wallpaper" :style="{
        filter: `blur(${panelState.panelConfig.backgroundBlur}px)`,
        backgroundImage: wallpaperLoaded ? `url(${wallpaperSource})` : 'none',
        opacity: wallpaperLoaded ? 1 : 0,
        backgroundSize: 'cover',
        backgroundPosition: 'center',
      }"
    />
    <div v-if="panelState.panelConfig.showWallpaper" class="mask" :style="{ backgroundColor: `rgba(24,23,21,${Math.max(0.58, panelState.panelConfig.backgroundMaskNumber ?? 0.58)})` }" />
    <nav ref="railRef" class="group-rail" :class="{ 'on-wallpaper': panelState.panelConfig.showWallpaper }" :aria-label="t('panelHome.groups')" @mouseleave="leaveRail" @focusout="onRailFocusOut" @keydown="onRailKeydown">
      <button class="rail-brand" type="button" :aria-label="t('panelHome.backToTop')" @click="scrollToTop" @mouseenter="showRailTooltip('brand', $event)" @focus="showRailTooltip('brand', $event)">
        <img :src="logo" alt="" width="28" height="28">
      </button>
      <div ref="railScrollRef" class="rail-scroll" @scroll="updateRailIndicator(); positionTooltip()">
        <button class="rail-button" :class="{ active: selectedGroupId === 'all' }" type="button" data-rail-key="all" :tabindex="railFocusKey === 'all' ? 0 : -1" :aria-label="t('panelHome.groupAria', { name: t('panelHome.all'), count: totalCount })" :aria-current="selectedGroupId === 'all' ? 'page' : undefined" @click="selectGroup('all')" @focus="railFocusKey = 'all'; showRailTooltip('all', $event)" @mouseenter="showRailTooltip('all', $event)">
          <GroupGlyph icon="lucide-layout-grid" :title="t('panelHome.all')" />
        </button>
        <span class="rail-divider" aria-hidden="true" />
        <button v-for="group in items" :key="group.id" class="rail-button" :class="{ active: selectedGroupId === group.id }" type="button" :data-rail-key="String(group.id)" :tabindex="railFocusKey === String(group.id) ? 0 : -1" :aria-label="t('panelHome.groupAria', { name: group.title, count: group.items?.length ?? 0 })" :aria-current="selectedGroupId === group.id ? 'page' : undefined" @click="selectGroup(group.id as number)" @focus="railFocusKey = String(group.id); showRailTooltip(String(group.id), $event)" @mouseenter="showRailTooltip(String(group.id), $event)">
          <GroupGlyph :icon="group.icon" :title="group.title" />
        </button>
      </div>
      <span class="rail-indicator" :style="railIndicator" aria-hidden="true" />
    </nav>
    <Teleport to="body">
      <Transition name="rail-tooltip">
        <div v-if="tooltipKey" class="rail-tooltip" :style="tooltipPosition" role="tooltip">
          <span>{{ tooltipContent.title }}</span><span v-if="tooltipContent.count !== null" class="rail-tooltip-count"> · {{ tooltipContent.count }}</span>
        </div>
      </Transition>
    </Teleport>
    <div ref="scrollContainerRef" class="main-scroll absolute w-full h-full overflow-auto">
      <div
        class="px-6 max-sm:px-4 mx-auto"
        :style="{
          marginTop: `${panelState.panelConfig.marginTop}%`,
          marginBottom: `${panelState.panelConfig.marginBottom}%`,
          maxWidth: (panelState.panelConfig.maxWidth ?? '1200') + panelState.panelConfig.maxWidthUnit,
        }"
      >
        <header class="navigation-shell" :style="{ marginLeft: `${panelState.panelConfig.marginX}px`, marginRight: `${panelState.panelConfig.marginX}px` }">
          <div class="navigation-row">
            <strong class="brand-name" :class="{ 'on-wallpaper': panelState.panelConfig.showWallpaper }">{{ panelState.panelConfig.logoText }}</strong>
            <label class="list-filter">
              <SvgIcon icon="iconamoon:search-fill" aria-hidden="true" />
              <input ref="listFilterRef" v-model="listFilter" type="search" :placeholder="t('panelHome.filterPlaceholder')" :aria-label="t('panelHome.filterPlaceholder')" @keydown="onFilterKeydown">
              <span v-if="listFilter.trim()" class="filter-count">{{ $t('panelHome.resultsCount', { count: visibleCount }) }}</span>
              <kbd v-else>/</kbd>
            </label>
            <NButton v-if="authStore.visitMode === VisitMode.VISIT_MODE_LOGIN" :type="arrangeMode ? 'primary' : 'default'" :secondary="!arrangeMode" class="arrange-button" @click="toggleArrangeMode">
              {{ arrangeMode ? t('panelHome.arrangeDone') : t('panelHome.arrange') }}
            </NButton>
          </div>
        </header>
        <div :style="{ marginLeft: `${panelState.panelConfig.marginX}px`, marginRight: `${panelState.panelConfig.marginX}px` }">
          <div v-if="overviewVisible" class="overview-band" :class="{ 'has-clock': panelState.panelConfig.clockShow, 'has-search': panelState.panelConfig.searchBoxShow, 'has-monitor': systemMonitorVisible }">
            <Clock v-if="panelState.panelConfig.clockShow" class="overview-clock" :class="{ 'on-wallpaper': panelState.panelConfig.showWallpaper }" :hide-second="!panelState.panelConfig.clockShowSecond" />
            <SearchBox v-if="panelState.panelConfig.searchBoxShow" class="overview-search" @item-search="itemFrontEndSearch" />
            <SystemMonitor v-if="systemMonitorVisible" class="overview-monitor" :allow-edit="authStore.visitMode === VisitMode.VISIT_MODE_LOGIN" :show-title="panelState.panelConfig.systemMonitorShowTitle" />
          </div>
          <h1 v-if="selectedGroupId !== 'all' && selectedGroup" class="selected-group-title">
            {{ selectedGroup.title }} <span>{{ selectedGroup.items?.length ?? 0 }}</span>
          </h1>

          <div v-if="visibleCount || (arrangeMode && visibleGroups.length) || (!isGroupedView && visibleGroups.length)" class="content-groups" :class="[isGroupedView ? 'group-columns' : '', { 'group-fade': groupFade, 'first-load-motion': firstLoadMotion }]">
            <section v-for="group in visibleGroups" :key="group.id" class="group-section" :style="firstLoadMotion ? { animationDelay: `${Math.min(9, visibleGroups.indexOf(group)) * 30}ms` } : undefined" :class="{ 'single-section': !isGroupedView, 'on-wallpaper': panelState.panelConfig.showWallpaper }">
              <div v-if="isGroupedView" class="group-heading">
                <h2>
                  {{ group.title }} <span>{{ group.items?.length ?? 0 }}</span>
                </h2>
              </div>
              <VueDraggable v-if="group.items" v-model="group.items" item-key="id" :animation="220" class="url-list" :class="{ 'arrange-drop-zone': arrangeMode }" group="items" draggable=".url-row:not(.add-row)" :data-group-id="group.id" :disabled="!arrangeMode || !!listFilter.trim()" :delay="200" :delay-on-touch-only="true" ghost-class="sortable-ghost" @end="handleEndDrag">
                <button v-for="item in group.items" :key="item.id" class="url-row" :class="{ sorting: arrangeMode }" :data-item-id="item.id" :tabindex="(rovingItemId ?? firstVisibleItemId) === item.id ? 0 : -1" type="button" :title="item.title" @focus="rovingItemId = item.id as number" @click="handleItemClick(item)" @keydown.down="moveRowFocus($event, 1)" @keydown.up="moveRowFocus($event, -1)" @contextmenu="(e) => handleContextMenu(e, item)">
                  <img v-if="item.icon?.itemType === 2 && item.icon.src && !failedIcons.has(item.id as number)" class="row-icon" :style="{ backgroundColor: iconBackground(item) }" :src="item.icon.src" alt="" loading="lazy" @error="failedIcons.add(item.id as number)">
                  <ItemIcon v-else-if="item.icon?.itemType === 3" class="row-icon" :item-icon="item.icon" :size="20" :force-background="iconBackground(item)" />
                  <span v-else class="row-placeholder" :style="{ backgroundColor: iconBackground(item), color: iconTextColor(item) }">{{ (item.icon?.itemType === 1 && item.icon.text ? item.icon.text : item.title).charAt(0).toUpperCase() }}</span>
                  <span class="row-name" :title="item.title">{{ item.title }}</span>
                  <span class="row-domain" :title="item.url">{{ displayDomain(item.url) }}</span>
                  <span class="row-arrow" aria-hidden="true">↗</span>
                </button>
                <button v-if="!arrangeMode && !isGroupedView && authStore.visitMode === VisitMode.VISIT_MODE_LOGIN" class="url-row add-row" type="button" @click="handleAddItem(group.id)">
                  <span class="row-placeholder add-symbol">+</span><span>{{ t('common.add') }}</span>
                </button>
              </VueDraggable>
            </section>
          </div>
          <div v-else class="empty-list">
            {{ t('panelHome.noMatches') }}
          </div>
        </div>
        <div class="mt-5 footer" :class="{ 'on-wallpaper': panelState.panelConfig.showWallpaper }" v-html="panelState.panelConfig.footerHtml" />
      </div>
    </div>

    <!-- 右键菜单 -->
    <NDropdown
      placement="bottom-start" trigger="manual" :x="dropdownMenuX" :y="dropdownMenuY"
      :options="getDropdownMenuOptions()" :show="dropdownShow" :on-clickoutside="onClickoutside" @select="handleRightMenuSelect"
    />

    <!-- 悬浮按钮 -->
    <div class="fixed-element">
      <div class="floating-actions">
        <!-- 网络模式切换按钮组 -->
        <NButton
          v-if="panelState.networkMode === PanelStateNetworkModeEnum.lan && panelState.panelConfig.netModeChangeButtonShow" circle secondary class="floating-button"
          :title="t('panelHome.changeToWanModel')" @click="handleChangeNetwork(PanelStateNetworkModeEnum.wan)"
        >
          <template #icon>
            <SvgIcon icon="material-symbols:lan-outline-rounded" />
          </template>
        </NButton>

        <NButton
          v-if="panelState.networkMode === PanelStateNetworkModeEnum.wan && panelState.panelConfig.netModeChangeButtonShow" circle secondary class="floating-button"
          :title="t('panelHome.changeToLanModel')" @click="handleChangeNetwork(PanelStateNetworkModeEnum.lan)"
        >
          <template #icon>
            <SvgIcon icon="mdi:wan" />
          </template>
        </NButton>

        <NButton v-if="authStore.visitMode === VisitMode.VISIT_MODE_LOGIN" circle secondary class="floating-button" @click="settingModalShow = !settingModalShow">
          <template #icon>
            <SvgIcon icon="majesticons-applications" />
          </template>
        </NButton>

        <NButton v-if="authStore.visitMode === VisitMode.VISIT_MODE_PUBLIC" circle secondary class="floating-button" :title="$t('panelHome.goToLogin')" @click="router.push('/login')">
          <template #icon>
            <SvgIcon icon="material-symbols:account-circle" />
          </template>
        </NButton>
      </div>

      <AppStarter v-model:visible="settingModalShow" />
      <!-- <Setting v-model:visible="settingModalShow" /> -->
    </div>

    <NBackTop
      class="rail-back-top"
      :listen-to="() => scrollContainerRef"
      :right="10"
      :bottom="10"
      style="background-color:transparent;border: none;box-shadow: none;"
    >
      <div>
        <NButton circle secondary class="floating-button">
          <template #icon>
            <SvgIcon icon="icon-park-outline:to-top" />
          </template>
        </NButton>
      </div>
    </NBackTop>

    <EditItem v-model:visible="editItemInfoShow" :item-info="editItemInfoData" :item-group-id="currentAddItenIconGroupId" @done="handleEditSuccess" />

    <!-- 弹窗 -->
    <NModal
      v-model:show="windowShow" :mask-closable="false" preset="card"
      style="max-width: 1000px;height: 600px;" :bordered="true" size="small" role="dialog"
      aria-modal="true"
    >
      <template #header>
        <div class="flex items-center">
          <span class="mr-5">
            {{ windowTitle }}
          </span>

          <NSpin v-if="windowIframeIsLoad" size="small" />
        </div>
      </template>
      <div class="w-full h-full rounded-xp-lg overflow-hidden border border-hairline">
        <div v-if="windowIframeIsLoad" class="flex flex-col p-5">
          <NSkeleton height="50px" width="100%" class="rounded-xp-md" />
          <NSkeleton height="180px" width="100%" class="mt-5 rounded-xp-md" />
          <NSkeleton height="180px" width="100%" class="mt-5 rounded-xp-md" />
        </div>
        <iframe
          v-show="!windowIframeIsLoad" id="windowIframeId" ref="windowIframeRef" :src="windowSrc"
          class="w-full h-full" frameborder="0" @load="handWindowIframeIdLoad"
        />
      </div>
    </NModal>
  </div>
</template>

<style>
body, html { overflow: hidden; background: var(--xp-canvas); color: var(--xp-ink); }
::selection { background: var(--xp-primary); color: var(--xp-on-primary); }
.n-modal { animation: modal-open 260ms cubic-bezier(.16,1,.3,1); }
@keyframes modal-open { from { transform: scale(.96); opacity: .7; } to { transform: scale(1); opacity: 1; } }
.rail-tooltip { position: fixed; z-index: 3000; transform: translateY(-50%); padding: 6px 10px; border-radius: 6px; background: var(--xp-ink); color: var(--xp-canvas); font: 13px/1.4 var(--xp-font-body); white-space: nowrap; pointer-events: none; box-shadow: none; }
.rail-tooltip-count { color: color-mix(in srgb, var(--xp-canvas) 65%, var(--xp-ink)); font-variant-numeric: tabular-nums; }
.rail-tooltip-enter-active, .rail-tooltip-leave-active { transition: opacity 120ms ease, translate 120ms ease; }
.rail-tooltip-enter-from, .rail-tooltip-leave-to { opacity: 0; translate: -4px 0; }
@media (max-width: 639px) { .rail-tooltip { display: none; } .rail-back-top { bottom: calc(66px + env(safe-area-inset-bottom)) !important; } }
@media (prefers-reduced-motion: reduce) { .n-modal { animation: none; } .rail-tooltip-enter-active, .rail-tooltip-leave-active { transition: none; } }
</style>

<style scoped>
.sun-main { position: relative; background: var(--xp-canvas); user-select: none; font-family: var(--xp-font-body); }
.cover, .mask { position: absolute; inset: 0; }
.cover { background-position: center; background-size: cover; transform: scale(1.03); transition: opacity 600ms ease; }
.navigation-shell { margin: 28px 0 0; }
.navigation-row { display: flex; align-items: center; gap: 16px; flex-wrap: wrap; }
.brand-name { flex: none; font: 600 28px/1.2 var(--xp-font-display); letter-spacing: 0; color: var(--xp-ink); white-space: nowrap; }
.brand-name.on-wallpaper { color: var(--xp-wallpaper-ink); }
.list-filter { display: flex; align-items: center; gap: 8px; flex: 1; min-width: 220px; max-width: 420px; height: 40px; margin-left: auto; padding: 0 14px; border: 1px solid var(--xp-hairline); border-radius: var(--xp-radius-md); background: var(--xp-canvas); color: var(--xp-muted); }
.list-filter input { min-width: 0; flex: 1; border: 0; outline: 0; background: none; color: var(--xp-ink); font: inherit; user-select: text; }
.list-filter input::placeholder { color: var(--xp-muted); }
.filter-count { flex: none; color: var(--xp-muted); font-size: 13px; white-space: nowrap; }
.list-filter kbd { padding: 0 5px; border: 1px solid var(--xp-hairline); border-radius: var(--xp-radius-sm); font-size: 13px; }
.list-filter:focus-within { border-color: var(--xp-primary); box-shadow: 0 0 0 3px var(--xp-focus-ring); }
.arrange-button { height: 40px; border-radius: var(--xp-radius-md); }
.overview-band { display: grid; grid-template-columns: auto minmax(220px, 1fr) minmax(180px, .75fr); align-items: center; gap: 24px; margin-top: 24px; }
.overview-band:not(.has-clock) { grid-template-columns: minmax(220px, 1fr) minmax(180px, .75fr); }
.overview-band:not(.has-search) { grid-template-columns: auto minmax(180px, .75fr); }
.overview-band:not(.has-monitor) { grid-template-columns: auto minmax(220px, 1fr); }
.overview-band.has-clock:not(.has-search):not(.has-monitor), .overview-band.has-search:not(.has-clock):not(.has-monitor), .overview-band.has-monitor:not(.has-clock):not(.has-search) { grid-template-columns: minmax(0, 1fr); }
.overview-clock.on-wallpaper { padding: 8px 12px; border-radius: var(--xp-radius-md); background: color-mix(in srgb, var(--xp-canvas) 92%, transparent); }
.overview-search { min-width: 0; }
.main-scroll { box-sizing: border-box; padding-left: 64px; }
.group-rail { position: fixed; z-index: 10; inset: 0 auto 0 0; display: flex; flex-direction: column; align-items: center; width: 64px; padding: 16px 11px; border-right: 1px solid var(--xp-hairline); background: var(--xp-canvas); }
.group-rail.on-wallpaper { background: color-mix(in srgb, var(--xp-canvas) 88%, transparent); }
.rail-brand, .rail-button { display: grid; flex: none; place-items: center; width: 40px; height: 40px; padding: 0; border: 0; border-radius: 8px; background: transparent; color: var(--xp-muted); cursor: pointer; transition: background-color 120ms ease, color 120ms ease; }
.rail-brand { margin-bottom: 16px; }
.rail-brand img { width: 28px; height: 28px; border-radius: 8px; }
.rail-scroll { display: flex; flex: 1; flex-direction: column; align-items: center; gap: 4px; width: 100%; min-height: 0; overflow-y: auto; scrollbar-width: none; }
.rail-scroll::-webkit-scrollbar { display: none; }
.rail-divider { display: block; flex: none; width: 24px; height: 1px; margin: 6px 0; background: var(--xp-hairline); }
.rail-button:hover, .rail-brand:hover { background: var(--xp-surface-soft); color: var(--xp-ink); }
.rail-button.active { background: var(--xp-surface-card); color: var(--xp-ink); }
.rail-indicator { position: absolute; top: 0; left: 0; width: 3px; height: 20px; border-radius: 3px; background: var(--xp-primary); pointer-events: none; transition: transform 180ms ease, opacity 120ms ease; }
.rail-brand:focus-visible, .rail-button:focus-visible, .url-row:focus-visible { outline: 2px solid var(--xp-primary); outline-offset: 2px; }
.selected-group-title { display: flex; align-items: baseline; gap: 8px; margin: 24px 0 16px; color: var(--xp-ink); font: 600 18px/1.4 var(--xp-font-display); }
.selected-group-title span { color: var(--xp-muted); font: 13px/1 var(--xp-font-body); font-variant-numeric: tabular-nums; }
.content-groups { margin-top: 24px; }
.selected-group-title + .content-groups { margin-top: 0; }
.group-columns { columns: 320px; column-gap: 16px; }
.group-columns .group-section { display: inline-block; width: 100%; margin-bottom: 16px; break-inside: avoid; }
.group-section { min-width: 0; padding: 16px; border: 1px solid var(--xp-hairline); border-radius: var(--xp-radius-lg); background: var(--xp-canvas); }
.group-section.on-wallpaper { background: color-mix(in srgb, var(--xp-canvas) 92%, transparent); }
.group-heading { display: flex; align-items: center; justify-content: space-between; margin-bottom: 8px; }
.group-heading h2 { display: flex; justify-content: space-between; width: 100%; margin: 0; color: var(--xp-ink); font-size: 16px; font-weight: 500; letter-spacing: 0; }
.group-heading span { color: var(--xp-muted-soft); font-size: 13px; }
.url-list { min-width: 0; }
.url-list.arrange-drop-zone { min-height: 40px; border: 1px dashed var(--xp-hairline); border-radius: var(--xp-radius-md); }
.single-section .url-list { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 4px 12px; }
.url-row { display: flex; align-items: center; gap: 10px; width: 100%; min-width: 0; min-height: 40px; padding: 8px 10px; border: 1px solid transparent; border-radius: var(--xp-radius-md); background: transparent; color: var(--xp-ink); font: inherit; font-size: 14px; text-align: left; cursor: pointer; }
.url-row { transition: background-color 120ms ease; }
.url-row:hover { background: var(--xp-surface-soft); }
.url-row.sorting { border: 1px dashed var(--xp-hairline); cursor: grab; }
.url-row.sorting:active { cursor: grabbing; }
.url-row.sortable-ghost { background: var(--xp-surface-strong); opacity: .6; }
.row-icon, .row-placeholder { flex: none; width: 20px; height: 20px; border-radius: var(--xp-radius-sm); }
.row-icon { object-fit: contain; padding: 1px; }
.row-placeholder { display: grid; place-items: center; font-size: 11px; font-weight: 800; }
.add-row .row-placeholder { background: var(--xp-surface-card); color: var(--xp-ink); }
.row-name, .row-domain { min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.row-name { flex: 1; }
.row-domain { flex: 0 1 auto; max-width: 46%; color: var(--xp-muted-soft); font-size: 13px; }
.row-arrow { flex: none; color: var(--xp-muted-soft); font-size: 13px; transition: color 120ms ease, transform 120ms ease; }
.url-row:hover .row-arrow { color: var(--xp-ink); transform: translate(2px, -2px); }
.add-row { color: var(--xp-muted); }
.add-symbol { font-size: 18px; }
.empty-list { padding: 32px; color: var(--xp-muted); text-align: center; }
.footer { color: var(--xp-muted-soft); font-size: 13px; }
.footer.on-wallpaper, .footer.on-wallpaper :deep(*) { color: var(--xp-wallpaper-ink); }
.fixed-element { position: fixed; right: 16px; bottom: 40px; }
.floating-actions { display: flex; flex-direction: column; gap: 8px; }
.floating-button { width: 36px; height: 36px; border: 1px solid var(--xp-hairline); border-radius: var(--xp-radius-pill); background: var(--xp-canvas); color: var(--xp-ink); }
@keyframes first-card { from { opacity: .2; transform: translateY(6px); } to { opacity: 1; transform: translateY(0); } }
@keyframes group-appear { from { opacity: .4; } to { opacity: 1; } }
.first-load-motion .group-section { animation: first-card 130ms ease-out both; }
.group-fade { animation: group-appear 150ms ease-out; }
@media (min-width: 640px) and (max-width: 1023px) {
  .overview-band { grid-template-columns: minmax(0, 1fr) minmax(0, 1fr); }
  .overview-band.has-search .overview-search { grid-column: 1 / -1; grid-row: 2; }
  .overview-band:not(.has-clock):not(.has-monitor) { grid-template-columns: minmax(0, 1fr); }
  .single-section .url-list { grid-template-columns: repeat(2, minmax(0, 1fr)); }
}
@media (max-width: 639px) {
  .main-scroll { padding-left: 0; padding-bottom: calc(56px + env(safe-area-inset-bottom)); }
  .group-rail { inset: auto 0 0; flex-direction: row; align-items: stretch; width: 100%; height: calc(56px + env(safe-area-inset-bottom)); padding: 0; border-top: 1px solid var(--xp-hairline); border-right: 0; }
  .rail-brand { display: none; }
  .rail-scroll { flex-direction: row; gap: 4px; overflow-x: auto; overflow-y: hidden; padding: 7px 12px calc(8px + env(safe-area-inset-bottom)); }
  .rail-divider { width: 1px; height: 24px; margin: 0 6px; }
  .rail-indicator { top: 0; left: 0; width: 20px; height: 3px; }
  .fixed-element { bottom: calc(96px + env(safe-area-inset-bottom)); }
  .navigation-shell { margin-top: 16px; }
  .arrange-button { margin-left: auto; }
  .brand-name { font-size: 24px; }
  .list-filter { order: 3; width: 100%; max-width: none; margin-left: 0; }
  .overview-band { grid-template-columns: minmax(0, 1fr) !important; gap: 16px; }
  .group-columns { columns: 1; }
  .single-section .url-list { grid-template-columns: 1fr; }
}
@media (prefers-reduced-motion: reduce) { .cover, .rail-indicator, .rail-button, .rail-brand, .url-row, .row-arrow { transition: none !important; } .first-load-motion .group-section, .group-fade { animation: none !important; } }
</style>
