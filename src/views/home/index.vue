<script setup lang="ts">
import { VueDraggable } from 'vue-draggable-plus'
import { NBackTop, NButton, NDropdown, NModal, NSkeleton, NSpin, useDialog, useMessage } from 'naive-ui'
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { AppStarter, EditItem } from './components'
import { Clock, SearchBox, SystemMonitor } from '@/components/deskModule'
import { ItemIcon, SvgIcon } from '@/components/common'
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
const failedIcons = ref(new Set<number>())

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
const navRef = ref<HTMLElement | null>(null)
const indicatorStyle = ref({ left: '0px', width: '0px' })
const { source: wallpaperSource, loaded: wallpaperLoaded, show: showWallpaper } = useWallpaper()

function updateIndicator() {
  const selected = navRef.value?.querySelector<HTMLElement>('.nav-tab.active')
  if (selected)
    indicatorStyle.value = { left: `${selected.offsetLeft}px`, width: `${selected.offsetWidth}px` }
}

function selectGroup(id: number | 'all') {
  selectedGroupId.value = id
  listFilter.value = ''
  filterItems.value = items.value
  nextTick(updateIndicator)
}

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
    nextTick(updateIndicator)
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
  window.addEventListener('resize', updateIndicator)
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
  window.removeEventListener('resize', updateIndicator)
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
    nextTick(updateIndicator)
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
    <div ref="scrollContainerRef" class="absolute w-full h-full overflow-auto">
      <div
        class="px-6 max-sm:px-4 mx-auto"
        :style="{
          marginTop: `${panelState.panelConfig.marginTop}%`,
          marginBottom: `${panelState.panelConfig.marginBottom}%`,
          maxWidth: (panelState.panelConfig.maxWidth ?? '1200') + panelState.panelConfig.maxWidthUnit,
        }"
      >
        <header class="navigation-shell">
          <div class="navigation-row">
            <strong class="brand-name" :class="{ 'on-wallpaper': panelState.panelConfig.showWallpaper }">{{ panelState.panelConfig.logoText }}</strong>
            <label class="list-filter">
              <SvgIcon icon="iconamoon:search-fill" aria-hidden="true" />
              <input ref="listFilterRef" v-model="listFilter" type="search" :placeholder="t('panelHome.filterPlaceholder')" :aria-label="t('panelHome.filterPlaceholder')">
              <kbd>/</kbd>
            </label>
            <NButton v-if="authStore.visitMode === VisitMode.VISIT_MODE_LOGIN" :type="arrangeMode ? 'primary' : 'default'" :secondary="!arrangeMode" class="arrange-button" @click="toggleArrangeMode">
              {{ arrangeMode ? t('panelHome.arrangeDone') : t('panelHome.arrange') }}
            </NButton>
          </div>
          <nav ref="navRef" class="group-navigation" :aria-label="t('panelHome.groups')">
            <span class="nav-indicator" :style="indicatorStyle" aria-hidden="true" />
            <button class="nav-tab" :class="{ active: selectedGroupId === 'all' }" type="button" @click="selectGroup('all')">
              {{ t('panelHome.all') }} <b>{{ totalCount }}</b>
            </button>
            <button v-for="group in items" :key="group.id" class="nav-tab" :class="{ active: selectedGroupId === group.id }" type="button" @click="selectGroup(group.id as number)">
              {{ group.title }} <b>{{ group.items?.length ?? 0 }}</b>
            </button>
          </nav>
          <Clock v-if="panelState.panelConfig.clockShow" :hide-second="!panelState.panelConfig.clockShowSecond" />
          <SearchBox v-if="panelState.panelConfig.searchBoxShow" class="search-row" @item-search="itemFrontEndSearch" />
        </header>

        <!-- 应用盒子 -->
        <div :style="{ marginLeft: `${panelState.panelConfig.marginX}px`, marginRight: `${panelState.panelConfig.marginX}px` }">
          <!-- 系统监控状态 -->
          <div
            v-if="panelState.panelConfig.systemMonitorShow
              && ((panelState.panelConfig.systemMonitorPublicVisitModeShow && authStore.visitMode === VisitMode.VISIT_MODE_PUBLIC)
                || authStore.visitMode === VisitMode.VISIT_MODE_LOGIN)"
            class="flex mx-auto"
          >
            <SystemMonitor
              :allow-edit="authStore.visitMode === VisitMode.VISIT_MODE_LOGIN"
              :show-title="panelState.panelConfig.systemMonitorShowTitle"
            />
          </div>

          <div v-if="visibleCount || (arrangeMode && visibleGroups.length) || (!isGroupedView && visibleGroups.length)" :class="isGroupedView ? 'group-columns' : ''">
            <section v-for="group in visibleGroups" :key="group.id" class="group-section" :class="{ 'single-section': !isGroupedView, 'on-wallpaper': panelState.panelConfig.showWallpaper }">
              <div v-if="isGroupedView" class="group-heading">
                <h2>
                  {{ group.title }} <span>{{ group.items?.length ?? 0 }}</span>
                </h2>
              </div>
              <VueDraggable v-if="group.items" v-model="group.items" item-key="id" :animation="220" class="url-list" :class="{ 'arrange-drop-zone': arrangeMode }" group="items" draggable=".url-row:not(.add-row)" :data-group-id="group.id" :disabled="!arrangeMode || !!listFilter.trim()" :delay="200" :delay-on-touch-only="true" ghost-class="sortable-ghost" @end="handleEndDrag">
                <button v-for="item in group.items" :key="item.id" class="url-row" :class="{ sorting: arrangeMode }" type="button" :title="item.title" @click="handleItemClick(item)" @contextmenu="(e) => handleContextMenu(e, item)">
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
        <div class="mt-5 footer" v-html="panelState.panelConfig.footerHtml" />
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
@media (prefers-reduced-motion: reduce) { .n-modal { animation: none; } }
</style>

<style scoped>
.sun-main { position: relative; background: var(--xp-canvas); user-select: none; font-family: var(--xp-font-body); }
.cover, .mask { position: absolute; inset: 0; }
.cover { background-position: center; background-size: cover; transform: scale(1.03); transition: opacity 600ms ease; }
.navigation-shell { margin: 28px 0 24px; }
.navigation-row { display: flex; align-items: center; gap: 16px; flex-wrap: wrap; }
.brand-name { flex: none; font: 400 28px/1.2 var(--xp-font-display); letter-spacing: -.3px; color: var(--xp-ink); white-space: nowrap; }
.brand-name.on-wallpaper { color: var(--xp-wallpaper-ink); }
.list-filter { display: flex; align-items: center; gap: 8px; flex: 1; min-width: 220px; max-width: 420px; height: 40px; margin-left: auto; padding: 0 14px; border: 1px solid var(--xp-hairline); border-radius: var(--xp-radius-md); background: var(--xp-canvas); color: var(--xp-muted); }
.list-filter input { min-width: 0; flex: 1; border: 0; outline: 0; background: none; color: var(--xp-ink); font: inherit; user-select: text; }
.list-filter input::placeholder { color: var(--xp-muted); }
.list-filter kbd { padding: 0 5px; border: 1px solid var(--xp-hairline); border-radius: var(--xp-radius-sm); font-size: 13px; }
.list-filter:focus-within { border-color: var(--xp-primary); box-shadow: 0 0 0 3px var(--xp-focus-ring); }
.arrange-button { height: 40px; border-radius: var(--xp-radius-md); }
.group-navigation { position: relative; display: flex; gap: 6px; width: max-content; max-width: 100%; margin-top: 16px; overflow-x: auto; scrollbar-width: none; }
.group-navigation::-webkit-scrollbar { display: none; }
.nav-tab { position: relative; z-index: 1; flex: none; padding: 8px 14px; border: 0; border-radius: var(--xp-radius-md); color: var(--xp-muted); font-size: 14px; font-weight: 500; white-space: nowrap; cursor: pointer; background: transparent; }
.nav-tab b { margin-left: 4px; font-size: 13px; font-weight: 500; color: var(--xp-muted-soft); }
.nav-tab.active { color: var(--xp-ink); }
.nav-indicator { position: absolute; top: 0; bottom: 0; border-radius: var(--xp-radius-md); background: var(--xp-surface-card); transition: left 220ms ease, width 220ms ease; }
.nav-tab:focus-visible, .url-row:focus-visible { outline: 2px solid var(--xp-primary); outline-offset: 2px; }
.search-row { margin-top: 16px; max-width: 720px; }
.group-columns { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 16px; align-items: start; }
.group-section { min-width: 0; padding: 16px; border: 1px solid var(--xp-hairline); border-radius: var(--xp-radius-lg); background: var(--xp-canvas); }
.group-section.on-wallpaper { background: color-mix(in srgb, var(--xp-canvas) 92%, transparent); }
.group-heading { display: flex; align-items: center; justify-content: space-between; margin-bottom: 8px; }
.group-heading h2 { display: flex; justify-content: space-between; width: 100%; margin: 0; color: var(--xp-ink); font-size: 16px; font-weight: 500; letter-spacing: 0; }
.group-heading span { color: var(--xp-muted-soft); font-size: 13px; }
.url-list { min-width: 0; }
.url-list.arrange-drop-zone { min-height: 40px; border: 1px dashed var(--xp-hairline); border-radius: var(--xp-radius-md); }
.single-section .url-list { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 4px 12px; }
.url-row { display: flex; align-items: center; gap: 10px; width: 100%; min-width: 0; min-height: 40px; padding: 8px 10px; border: 1px solid transparent; border-radius: var(--xp-radius-md); background: transparent; color: var(--xp-ink); font: inherit; font-size: 14px; text-align: left; cursor: pointer; }
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
.row-arrow { flex: none; color: var(--xp-muted-soft); font-size: 13px; }
.add-row { color: var(--xp-muted); }
.add-symbol { font-size: 18px; }
.empty-list { padding: 32px; color: var(--xp-muted); text-align: center; }
.footer { color: var(--xp-muted-soft); font-size: 13px; }
.fixed-element { position: fixed; right: 16px; bottom: 40px; }
.floating-actions { display: flex; flex-direction: column; gap: 8px; }
.floating-button { width: 36px; height: 36px; border: 1px solid var(--xp-hairline); border-radius: var(--xp-radius-pill); background: var(--xp-canvas); color: var(--xp-ink); }
@media (max-width: 1023px) { .group-columns, .single-section .url-list { grid-template-columns: repeat(2, minmax(0, 1fr)); } }
@media (max-width: 639px) { .navigation-shell { margin-top: 16px; } .list-filter { order: 3; width: 100%; max-width: none; margin-left: 0; } .group-columns, .single-section .url-list { grid-template-columns: 1fr; } }
@media (prefers-reduced-motion: reduce) { .cover, .nav-indicator { transition: none !important; } }
</style>
