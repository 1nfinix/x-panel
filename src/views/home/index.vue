<script setup lang="ts">
import { VueDraggable } from 'vue-draggable-plus'
import { NBackTop, NButton, NButtonGroup, NDropdown, NModal, NSkeleton, NSpin, useDialog, useMessage } from 'naive-ui'
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
  sortStatus?: boolean
  hoverStatus: boolean
  items?: Panel.ItemInfo[]
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
const listFilterRef = ref<HTMLInputElement | null>(null)
const failedIcons = ref(new Set<number>())

// Icon tile colors honor the item's own backgroundColor; unset means white.
function iconBackground(item: Panel.ItemInfo) {
  return item.icon?.backgroundColor || '#ffffff'
}

function iconTextColor(item: Panel.ItemInfo) {
  const color = iconBackground(item).trim()
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
    return '#f2f7f4'
  return (0.299 * r + 0.587 * g + 0.114 * b) / 255 > 0.6 ? '#111827' : '#f2f7f4'
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

function groupIndex(group: ItemGroup) {
  return items.value.findIndex(item => item.id === group.id)
}

watch(() => [panelState.panelConfig.autoWallpaper, panelState.panelConfig.backgroundImageSrc], () => {
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

function handleItemClick(itemGroupIndex: number, item: Panel.ItemInfo) {
  if (items.value[itemGroupIndex] && items.value[itemGroupIndex].sortStatus) {
    handleEditItem(item)
    return
  }

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

function handleContextMenu(e: MouseEvent, itemGroupIndex: number, item: Panel.ItemInfo) {
  if (items.value[itemGroupIndex] && items.value[itemGroupIndex].sortStatus)
    return

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

// 结束拖拽
// function handleEndDrag(event: any, itemIconGroup: Panel.ItemIconGroup) {
//   // console.log(event)
//   // console.log(items.value)
// }

function handleSaveSort(itemGroup: ItemGroup) {
  const saveItems: Common.SortItemRequest[] = []
  if (itemGroup.items) {
    for (let i = 0; i < itemGroup.items.length; i++) {
      const element = itemGroup.items[i]
      saveItems.push({
        id: element.id as number,
        sort: i + 1,
      })
    }

    saveSort({ itemIconGroupId: itemGroup.id as number, sortItems: saveItems }).then(({ code, msg }) => {
      if (code === 0) {
        ms.success(t('common.saveSuccess'))
        itemGroup.sortStatus = false
      }
      else {
        ms.error(`${t('common.saveFail')}:${msg}`)
      }
    })
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

function handleSetSortStatus(groupIndex: number, sortStatus: boolean) {
  if (items.value[groupIndex])
    items.value[groupIndex].sortStatus = sortStatus

  // 并未保存排序重新更新数据
  if (!sortStatus) {
    // 单独更新组
    if (items.value[groupIndex] && items.value[groupIndex].id)
      updateItemIconGroupByNet(groupIndex, items.value[groupIndex].id as number)
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
    <svg width="0" height="0" aria-hidden="true" style="position:absolute"><defs><filter id="goop-filter"><feGaussianBlur in="SourceGraphic" stdDeviation="2" result="blur" /><feColorMatrix in="blur" mode="matrix" values="1 0 0 0 0  0 1 0 0 0  0 0 1 0 0  0 0 0 18 -7" /></filter></defs></svg>
    <div
      class="cover wallpaper" :style="{
        filter: `blur(${panelState.panelConfig.backgroundBlur}px)`,
        backgroundImage: wallpaperLoaded ? `url(${wallpaperSource})` : 'none',
        opacity: wallpaperLoaded ? 1 : 0,
        backgroundSize: 'cover',
        backgroundPosition: 'center',
      }"
    />
    <div class="mask" :style="{ backgroundColor: `rgba(6,12,15,${Math.max(0.58, panelState.panelConfig.backgroundMaskNumber ?? 0.58)})` }" />
    <div ref="scrollContainerRef" class="absolute w-full h-full overflow-auto">
      <div
        class="p-2.5 mx-auto"
        :style="{
          marginTop: `${panelState.panelConfig.marginTop}%`,
          marginBottom: `${panelState.panelConfig.marginBottom}%`,
          maxWidth: (panelState.panelConfig.maxWidth ?? '1200') + panelState.panelConfig.maxWidthUnit,
        }"
      >
        <header class="navigation-shell">
          <div class="navigation-row">
            <strong class="brand-name">{{ panelState.panelConfig.logoText }}</strong>
            <label class="list-filter">
              <SvgIcon icon="material-symbols:search-rounded" aria-hidden="true" />
              <input ref="listFilterRef" v-model="listFilter" type="search" :placeholder="t('panelHome.filterPlaceholder')" :aria-label="t('panelHome.filterPlaceholder')">
              <kbd>/</kbd>
            </label>
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

          <div v-if="visibleCount || (!isGroupedView && visibleGroups.length)" :class="isGroupedView ? 'group-columns' : ''">
            <section v-for="group in visibleGroups" :key="group.id" class="group-section" :class="{ 'single-section': !isGroupedView }">
              <div v-if="isGroupedView || authStore.visitMode === VisitMode.VISIT_MODE_LOGIN" class="group-heading">
                <h2 v-if="isGroupedView">
                  {{ group.title }} <span>{{ group.items?.length ?? 0 }}</span>
                </h2>
                <div v-if="!isGroupedView && authStore.visitMode === VisitMode.VISIT_MODE_LOGIN" class="group-actions">
                  <NButton size="small" secondary @click="handleSetSortStatus(groupIndex(group), !group.sortStatus)">
                    {{ group.sortStatus ? t('common.cancel') : t('common.sort') }}
                  </NButton>
                  <NButton v-if="group.sortStatus" size="small" type="primary" @click="handleSaveSort(group)">
                    {{ t('common.saveSort') }}
                  </NButton>
                </div>
              </div>
              <VueDraggable v-if="group.items" v-model="group.items" item-key="id" :animation="220" class="url-list" filter=".add-row" :disabled="isGroupedView || !group.sortStatus">
                <button v-for="item in group.items" :key="item.id" class="url-row" :class="{ sorting: group.sortStatus }" type="button" :title="item.title" @click="handleItemClick(groupIndex(group), item)" @contextmenu="(e) => handleContextMenu(e, groupIndex(group), item)">
                  <img v-if="item.icon?.itemType === 2 && item.icon.src && !failedIcons.has(item.id as number)" class="row-icon" :style="{ backgroundColor: iconBackground(item) }" :src="item.icon.src" alt="" loading="lazy" @error="failedIcons.add(item.id as number)">
                  <ItemIcon v-else-if="item.icon?.itemType === 3" class="row-icon" :item-icon="item.icon" :size="20" :force-background="iconBackground(item)" />
                  <span v-else class="row-placeholder" :style="{ backgroundColor: iconBackground(item), color: iconTextColor(item) }">{{ (item.icon?.itemType === 1 && item.icon.text ? item.icon.text : item.title).charAt(0).toUpperCase() }}</span>
                  <span class="row-name" :title="item.title">{{ item.title }}</span>
                  <span class="row-domain" :title="item.url">{{ displayDomain(item.url) }}</span>
                  <span class="row-arrow" aria-hidden="true">↗</span>
                </button>
                <button v-if="!isGroupedView && authStore.visitMode === VisitMode.VISIT_MODE_LOGIN" class="url-row add-row" type="button" @click="handleAddItem(group.id)">
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
    <div class="fixed-element shadow-[0_0_10px_2px_rgba(0,0,0,0.2)]">
      <NButtonGroup vertical>
        <!-- 网络模式切换按钮组 -->
        <NButton
          v-if="panelState.networkMode === PanelStateNetworkModeEnum.lan && panelState.panelConfig.netModeChangeButtonShow" color="#2a2a2a6b"
          :title="t('panelHome.changeToWanModel')" @click="handleChangeNetwork(PanelStateNetworkModeEnum.wan)"
        >
          <template #icon>
            <SvgIcon class="text-white font-xl" icon="material-symbols:lan-outline-rounded" />
          </template>
        </NButton>

        <NButton
          v-if="panelState.networkMode === PanelStateNetworkModeEnum.wan && panelState.panelConfig.netModeChangeButtonShow" color="#2a2a2a6b"
          :title="t('panelHome.changeToLanModel')" @click="handleChangeNetwork(PanelStateNetworkModeEnum.lan)"
        >
          <template #icon>
            <SvgIcon class="text-white font-xl" icon="mdi:wan" />
          </template>
        </NButton>

        <NButton v-if="authStore.visitMode === VisitMode.VISIT_MODE_LOGIN" color="#2a2a2a6b" @click="settingModalShow = !settingModalShow">
          <template #icon>
            <SvgIcon class="text-white font-xl" icon="majesticons-applications" />
          </template>
        </NButton>

        <NButton v-if="authStore.visitMode === VisitMode.VISIT_MODE_PUBLIC" color="#2a2a2a6b" :title="$t('panelHome.goToLogin')" @click="router.push('/login')">
          <template #icon>
            <SvgIcon class="text-white font-xl" icon="material-symbols:account-circle" />
          </template>
        </NButton>
      </NButtonGroup>

      <AppStarter v-model:visible="settingModalShow" />
      <!-- <Setting v-model:visible="settingModalShow" /> -->
    </div>

    <NBackTop
      :listen-to="() => scrollContainerRef"
      :right="10"
      :bottom="10"
      style="background-color:transparent;border: none;box-shadow: none;"
    >
      <div class="shadow-[0_0_10px_2px_rgba(0,0,0,0.2)]">
        <NButton color="#2a2a2a6b">
          <template #icon>
            <SvgIcon class="text-white font-xl" icon="icon-park-outline:to-top" />
          </template>
        </NButton>
      </div>
    </NBackTop>

    <EditItem v-model:visible="editItemInfoShow" :item-info="editItemInfoData" :item-group-id="currentAddItenIconGroupId" @done="handleEditSuccess" />

    <!-- 弹窗 -->
    <NModal
      v-model:show="windowShow" :mask-closable="false" preset="card"
      style="max-width: 1000px;height: 600px;border-radius: 1rem;" :bordered="true" size="small" role="dialog"
      aria-modal="true"
    >
      <template #header>
        <div class="flex items-center">
          <span class="mr-[20px]">
            {{ windowTitle }}
          </span>

          <NSpin v-if="windowIframeIsLoad" size="small" />
        </div>
      </template>
      <div class="w-full h-full rounded-2xl overflow-hidden border dark:border-zinc-700">
        <div v-if="windowIframeIsLoad" class="flex flex-col p-5">
          <NSkeleton height="50px" width="100%" class="rounded-lg" />
          <NSkeleton height="180px" width="100%" class="mt-[20px] rounded-lg" />
          <NSkeleton height="180px" width="100%" class="mt-[20px] rounded-lg" />
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
:root { --goop-bg: #F7F5F0; --goop-panel: #FFFFFF; --goop-green: #237a49; --goop-text: #111827; --goop-border: #D9DEE7; --goop-panel-overlay: rgba(255,255,255,.8); --goop-muted: #334155; --goop-row-hover: rgba(47,158,95,.1); --goop-selection-text: #FFFFFF; --goop-active-text: #FFFFFF; --goop-placeholder-text: #FFFFFF; }
html.dark { --goop-bg: #111827; --goop-panel: #1b2734; --goop-green: #76d99b; --goop-text: #f2f7f4; --goop-border: rgba(220,255,231,.1); --goop-panel-overlay: rgba(17,24,39,.72); --goop-muted: #9fb1a8; --goop-row-hover: rgba(118,217,155,.12); --goop-selection-text: #111827; --goop-active-text: #111827; --goop-placeholder-text: #0d1a12; }
body, html { overflow: hidden; background: var(--goop-bg); color: var(--goop-text); }
::selection { background: var(--goop-green); color: var(--goop-selection-text); }
.n-modal, .n-card { border-radius: 8px !important; }
.n-modal { animation: modal-open 260ms cubic-bezier(.16,1,.3,1); }
@keyframes modal-open { from { transform: scale(.96); opacity: .7; } to { transform: scale(1); opacity: 1; } }
@media (prefers-reduced-motion: reduce) { .n-modal { animation: none; } }
</style>

<style scoped>
.sun-main { position: relative; background: var(--goop-bg); user-select: none; font-family: 'Arial Rounded MT Bold', 'PingFang SC', system-ui, sans-serif; }
.cover, .mask { position: absolute; inset: 0; }
.cover { background-position: center; background-size: cover; transform: scale(1.03); transition: opacity 600ms ease; }
.navigation-shell { margin: 28px 0 22px; }
.navigation-row { display: flex; align-items: center; gap: 16px; flex-wrap: wrap; }
.brand-name { flex: none; font-size: 26px; font-weight: 850; color: #f2f7f4; white-space: nowrap; }
.list-filter { display: flex; align-items: center; gap: 8px; flex: 1; min-width: 220px; max-width: 420px; margin-left: auto; padding: 8px 14px; border: 1px solid var(--goop-border); border-radius: 999px; background: var(--goop-panel-overlay); color: var(--goop-muted); }
.list-filter input { min-width: 0; flex: 1; border: 0; outline: 0; background: none; color: var(--goop-text); font: inherit; user-select: text; }
.list-filter input::placeholder { color: var(--goop-muted); }
.list-filter kbd { padding: 0 5px; border: 1px solid var(--goop-border); border-radius: 4px; font-size: 11px; }
.list-filter:focus-within { outline: 2px solid var(--goop-green); outline-offset: 2px; }
.group-navigation { position: relative; display: flex; gap: 6px; width: max-content; max-width: 100%; margin-top: 18px; overflow-x: auto; scrollbar-width: none; padding: 5px; border: 1px solid var(--goop-border); border-radius: 999px; background: var(--goop-panel-overlay); }
.group-navigation::-webkit-scrollbar { display: none; }
.nav-tab { position: relative; z-index: 1; flex: none; padding: 6px 13px; border: 0; border-radius: 999px; color: var(--goop-muted); font-weight: 500; white-space: nowrap; cursor: pointer; background: transparent; }
.nav-tab b { margin-left: 4px; font-size: 12px; font-weight: 500; opacity: .7; }
.nav-tab.active { color: var(--goop-active-text); }
.nav-indicator { position: absolute; top: 5px; bottom: 5px; border-radius: 999px; background: var(--goop-green); transition: left 430ms cubic-bezier(.16,1,.3,1), width 430ms cubic-bezier(.16,1,.3,1); filter: url(#goop-filter); }
.nav-tab:focus-visible, .url-row:focus-visible { outline: 2px solid var(--goop-green); outline-offset: 2px; }
.search-row { margin-top: 16px; max-width: 720px; }
.group-columns { columns: 3 300px; column-gap: 18px; }
.group-section { break-inside: avoid; margin-bottom: 18px; padding: 10px 6px 6px; border: 1px solid var(--goop-border); border-radius: 8px; background: var(--goop-panel-overlay); }
.single-section { padding: 8px; }
.group-heading { display: flex; align-items: center; justify-content: space-between; margin: 0 10px 4px; }
.group-heading h2 { display: flex; justify-content: space-between; width: 100%; margin: 0; color: var(--goop-muted); font-size: 12px; font-weight: 600; letter-spacing: .02em; }
.group-actions { display: flex; gap: 8px; margin-left: auto; }
.url-list { min-width: 0; }
.single-section .url-list { columns: 3 300px; column-gap: 18px; }
.url-row { display: flex; align-items: center; gap: 10px; width: 100%; min-width: 0; padding: 6px 10px; break-inside: avoid; border: 0; border-radius: 6px; background: transparent; color: var(--goop-text); font: inherit; text-align: left; cursor: pointer; }
.url-row:hover { background: var(--goop-row-hover); }
.url-row.sorting { cursor: grab; }
.row-icon, .row-placeholder { flex: none; width: 20px; height: 20px; border-radius: 5px; }
.row-icon { object-fit: contain; padding: 1px; }
.row-placeholder { display: grid; place-items: center; font-size: 11px; font-weight: 800; }
.add-row .row-placeholder { background: var(--goop-green); color: var(--goop-placeholder-text); }
.row-name, .row-domain { min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.row-name { flex: 1; }
.row-domain { flex: 0 1 auto; max-width: 46%; color: var(--goop-muted); font-size: 12px; }
.row-arrow { flex: none; color: var(--goop-muted); font-size: 12px; }
.add-row { color: var(--goop-muted); }
.add-symbol { font-size: 17px; }
.empty-list { padding: 30px; color: var(--goop-muted); text-align: center; }
.fixed-element { position: fixed; right: 16px; bottom: 40px; }
@media (max-width: 640px) { .navigation-shell { margin-top: 18px; } .list-filter { order: 3; width: 100%; max-width: none; margin-left: 0; } .group-columns, .single-section .url-list { columns: 1; } }
@media (prefers-reduced-motion: reduce) { .cover, .nav-indicator { transition: none !important; filter: none !important; } }
</style>
