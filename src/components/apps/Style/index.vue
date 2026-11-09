<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import type { UploadFileInfo } from 'naive-ui'
import { NButton, NCard, NColorPicker, NGrid, NGridItem, NInput, NInputGroup, NPopconfirm, NSelect, NSlider, NSwitch, NUpload, NUploadDragger, useMessage } from 'naive-ui'
import { useAuthStore, usePanelState } from '@/store'
import { set as setUserConfig } from '@/api/panel/userConfig'
import { faviconCandidates } from '@/api/panel/itemIcon'
import { PanelPanelConfigStyleEnum } from '@/enums/panel'
import { t } from '@/locales'
import { ItemIcon } from '@/components/common'
import { useFaviconFill } from '@/hooks/useFaviconFill'

const authStore = useAuthStore()
const panelState = usePanelState()
const ms = useMessage()
const showWallpaperInput = ref(false)
interface FaviconCandidate { id: number; title: string; url: string; icon: Panel.ItemIcon }
const candidates = ref<FaviconCandidate[]>([])
const selectedIds = ref<number[]>([])
const { running: fillRunning, done: fillDone, success: fillSuccess, failures, reset: resetFill, run: runFaviconFill } = useFaviconFill()
const fillFailures = computed(() => failures.value.map(failure => `${candidates.value.find(item => item.id === failure.id)?.title || failure.id}: ${failure.reason}`))

async function loadCandidates() {
  const response = await faviconCandidates<Common.ListResponse<FaviconCandidate[]>>()
  if (response.code !== 0) {
    ms.error(response.msg)
    return
  }
  candidates.value = response.data.list
  selectedIds.value = candidates.value.map(item => item.id)
  resetFill()
}

async function runFill() {
  await runFaviconFill(selectedIds.value, (result) => {
    const item = candidates.value.find(candidate => candidate.id === result.id)
    if (item && result.iconUrl)
      item.icon = { ...item.icon, itemType: 2, src: result.iconUrl }
  })
}

const isSaveing = ref(false)

const iconTypeOptions = [
  {
    label: t('apps.baseSettings.detailIcon'),
    value: PanelPanelConfigStyleEnum.info,
  },
  {
    label: t('apps.baseSettings.smallIcon'),
    value: PanelPanelConfigStyleEnum.icon,
  },
]

const maxWidthUnitOption = [
  {
    label: 'px',
    value: 'px',
  },
  {
    label: '%',
    value: '%',
  },
]

watch(panelState.panelConfig, () => {
  if (!isSaveing.value) {
    isSaveing.value = true

    setTimeout(() => {
      panelState.recordState()// 本地记录
      isSaveing.value = false
      uploadCloud()
    }, 1000)
  }
})

function handleUploadBackgroundFinish({
  file,
  event,
}: {
  file: UploadFileInfo
  event?: ProgressEvent
}) {
  const res = JSON.parse((event?.target as XMLHttpRequest).response)
  panelState.panelConfig.backgroundImageSrc = res.data.imageUrl
  return file
}

function uploadCloud() {
  setUserConfig({ panel: panelState.panelConfig }).then((res) => {
    if (res.code === 0)
      ms.success(t('apps.baseSettings.configSaved'))
    else
      ms.error(t('apps.baseSettings.configFailed', { message: res.msg }))
  })
}

function resetPanelConfig() {
  panelState.resetPanelConfig()
  uploadCloud()
}
</script>

<template>
  <div class="bg-bg1 rounded-card p-2 overflow-auto">
    <NCard size="small">
      <div class="text-text1 text-[16px] font-medium mb-2">
        LOGO
      </div>

      <div>
        <div>
          {{ $t('apps.baseSettings.textContent') }}
        </div>
        <div class="flex items-center mt-1">
          <NInput v-model:value="panelState.panelConfig.logoText" type="text" show-count :maxlength="20" placeholder="请输入文字" />
        </div>
      </div>
    </NCard>

    <NCard class="mt-3" size="small">
      <div class="text-text1 text-[16px] font-medium mb-2">
        {{ $t('apps.baseSettings.clock') }}
      </div>
      <div class="flex items-center mt-1">
        <span class="mr-3">{{ $t('common.show') }}</span>
        <NSwitch v-model:value="panelState.panelConfig.clockShow" />
      </div>
      <div v-if="panelState.panelConfig.clockShow" class="flex items-center mt-1">
        <span class="mr-3">{{ $t('apps.baseSettings.clockSecondShow') }}</span>
        <NSwitch v-model:value="panelState.panelConfig.clockShowSecond" />
      </div>
    </NCard>

    <NCard class="mt-3" size="small">
      <div class="text-text1 text-[16px] font-medium mb-2">
        {{ $t('apps.baseSettings.searchBar') }}
      </div>
      <div class="flex items-center mt-1">
        <span class="mr-3">{{ $t('common.show') }}</span>
        <NSwitch v-model:value="panelState.panelConfig.searchBoxShow" />
      </div>
      <div v-if="panelState.panelConfig.searchBoxShow" class="flex items-center mt-1">
        <span class="mr-3">{{ $t('apps.baseSettings.searchBarSearchItem') }}</span>
        <NSwitch v-model:value="panelState.panelConfig.searchBoxSearchIcon" />
      </div>
    </NCard>

    <NCard class="mt-3" size="small">
      <div class="text-text1 text-[16px] font-medium mb-2">
        {{ $t('apps.baseSettings.systemMonitorStatus') }}
      </div>
      <div class="flex items-center mt-1">
        <span class="mr-3">{{ $t('common.show') }}</span>
        <NSwitch v-model:value="panelState.panelConfig.systemMonitorShow" />
      </div>
      <div v-if="panelState.panelConfig.systemMonitorShow" class="flex items-center mt-1">
        <span class="mr-3">{{ $t('apps.baseSettings.showTitle') }}</span>
        <NSwitch v-model:value="panelState.panelConfig.systemMonitorShowTitle" />
      </div>
      <div v-if="panelState.panelConfig.systemMonitorShow" class="flex items-center mt-1">
        <span class="mr-3">{{ $t('apps.baseSettings.publicVisitModeShow') }}</span>
        <NSwitch v-model:value="panelState.panelConfig.systemMonitorPublicVisitModeShow" />
      </div>
    </NCard>

    <NCard class="mt-3" size="small">
      <div class="text-text1 text-[16px] font-medium mb-2">
        {{ $t('common.icon') }}
      </div>
      <div class="mt-1">
        <div>
          {{ $t('common.style') }}
        </div>
        <div class="flex items-center mt-1">
          <NSelect v-model:value="panelState.panelConfig.iconStyle" :options="iconTypeOptions" />
        </div>
      </div>

      <div v-if="panelState.panelConfig.iconStyle === PanelPanelConfigStyleEnum.info" class="mt-1">
        <div>
          {{ $t('apps.baseSettings.hideDescription') }}
        </div>
        <div class="flex items-center mt-1">
          <NSwitch v-model:value="panelState.panelConfig.iconTextInfoHideDescription" />
        </div>
      </div>

      <div v-if="panelState.panelConfig.iconStyle === PanelPanelConfigStyleEnum.icon" class="mt-1">
        <div>
          {{ $t('apps.baseSettings.hideTitle') }}
        </div>
        <div class="flex items-center mt-1">
          <NSwitch v-model:value="panelState.panelConfig.iconTextIconHideTitle" />
        </div>
      </div>

      <div class="mt-1">
        <div>
          {{ $t('common.textColor') }}
        </div>
        <div class="flex items-center mt-1">
          <NColorPicker
            v-model:value="panelState.panelConfig.iconTextColor"
            :show-alpha="false"
            size="small"
            :modes="['hex']"
            :swatches="[
              '#cc785c',
              '#5db8a6',
              '#e8a55a',
              '#141413',
              '#6c6a64',
            ]"
          />
        </div>
      </div>
    </NCard>
    <NCard class="mt-3" size="small">
      <div class="text-text1 text-[16px] font-medium mb-2">
        {{ $t('apps.baseSettings.wallpaper') }}
      </div>
      <div class="flex items-center gap-3 mb-4">
        <span>{{ $t('apps.baseSettings.showWallpaper') }}</span>
        <NSwitch v-model:value="panelState.panelConfig.showWallpaper" />
      </div>
      <div v-if="panelState.panelConfig.showWallpaper">
        <NUpload
          action="/api/file/uploadImg"
          :show-file-list="false"
          name="imgfile"
          :headers="{
            token: authStore.token as string,
          }"
          :directory-dnd="true"
          @finish="handleUploadBackgroundFinish"
        >
          <NUploadDragger style="width: 100%;">
            <div
              class="h-[200px] w-full bg-bg2 flex justify-center items-center cursor-pointer rounded-small"
              :style="{ background: `url(${panelState.panelConfig.backgroundImageSrc}) no-repeat`, backgroundSize: 'cover' }"
            >
              <div class="text-[color:var(--x-on-primary)]">
                {{ $t('apps.baseSettings.uploadOrDragText') }}
              </div>
            </div>
          </NUploadDragger>
        </NUpload>

        <div class="flex items-center mt-1">
          <span class="mr-3">{{ $t('apps.baseSettings.customImageAddress') }}</span>
          <NSwitch v-model:value="showWallpaperInput" />
        </div>

        <div class="flex items-center mt-3 gap-3">
          <span>{{ $t('apps.baseSettings.autoWallpaper') }}</span>
          <NSwitch v-model:value="panelState.panelConfig.autoWallpaper" />
        </div>
        <div v-if="showWallpaperInput" class="mt-1">
          <NInput v-model:value="panelState.panelConfig.backgroundImageSrc" type="text" size="small" clearable />
        </div>

        <div class="flex items-center mt-3">
          <span class="mr-3">{{ $t('apps.baseSettings.vague') }}</span>
          <NSlider v-model:value="panelState.panelConfig.backgroundBlur" class="max-w-[200px]" :step="2" :max="20" />
        </div>

        <div class="flex items-center mt-3">
          <span class="mr-3">{{ $t('apps.baseSettings.mask') }}</span>
          <NSlider v-model:value="panelState.panelConfig.backgroundMaskNumber" class="max-w-[200px]" :step="0.1" :max="1" />
        </div>
      </div>
    </NCard>

    <NCard size="small" class="mt-3">
      <div class="text-text1 text-[16px] font-medium mb-3">
        {{ $t('apps.baseSettings.fillFavicons') }}
      </div>
      <NButton :disabled="fillRunning" @click="loadCandidates">
        {{ $t('apps.baseSettings.findCandidates') }}
      </NButton>
      <div v-for="item in candidates" :key="item.id" class="flex gap-3 items-center mt-2">
        <input v-model="selectedIds" type="checkbox" :value="item.id" :disabled="fillRunning">
        <ItemIcon v-if="item.icon.itemType === 1 || item.icon.src" :item-icon="item.icon" :size="28" />
        <span v-else class="w-[28px] text-center" aria-hidden="true">—</span>
        <span>{{ item.title }}</span><span class="truncate text-text2">{{ item.url }}</span>
      </div>
      <NPopconfirm v-if="candidates.length" @positive-click="runFill">
        <template #trigger>
          <NButton class="mt-3" type="primary" :disabled="!selectedIds.length || fillRunning">
            {{ $t('common.confirm') }}
          </NButton>
        </template>
        {{ $t('apps.baseSettings.fillConfirm') }}
      </NPopconfirm>
      <div v-if="fillDone" class="mt-3">
        {{ fillDone }}/{{ selectedIds.length }} · {{ $t('apps.baseSettings.fillSuccess') }} {{ fillSuccess }} · {{ $t('apps.baseSettings.fillFailed') }} {{ fillFailures.length }}
      </div>
      <div v-for="failure in fillFailures" :key="failure" class="text-[color:var(--x-red-fg)]">
        {{ failure }}
      </div>
    </NCard>

    <NCard class="mt-3" size="small">
      <div class="text-text1 text-[16px] font-medium mb-2">
        {{ $t('apps.baseSettings.contentArea') }}
      </div>

      <NGrid cols="2">
        <NGridItem span="12 400:12">
          <div class="flex items-center mt-1">
            <span class="mr-3">{{ $t('apps.baseSettings.netModeChangeButtonShow') }}</span>
            <NSwitch v-model:value="panelState.panelConfig.netModeChangeButtonShow" />
          </div>
        </NGridItem>

        <NGridItem span="12 400:12">
          <div class="flex items-center mt-3">
            <span class="mr-3">{{ $t('apps.baseSettings.maxWidth') }}</span>
            <div class="flex">
              <NInputGroup>
                <NInput v-model:value="panelState.panelConfig.maxWidth" size="small" type="number" :maxlength="10" :style="{ width: '100px' }" placeholder="1200" />
                <NSelect v-model:value="panelState.panelConfig.maxWidthUnit" :style="{ width: '80px' }" :options="maxWidthUnitOption" size="small" />
              </NInputGroup>
            </div>
          </div>
        </NGridItem>
        <NGridItem span="12 400:12">
          <div class="flex items-center mt-3">
            <span class="mr-3">{{ $t('apps.baseSettings.leftRightMargin') }}</span>
            <NSlider v-model:value="panelState.panelConfig.marginX" class="max-w-[200px]" :step="1" :max="100" />
          </div>
        </NGridItem>
        <NGridItem span="12 400:12">
          <div class="flex items-center mt-3">
            <span class="mr-3">{{ $t('apps.baseSettings.topMargin') }} (%)</span>
            <NSlider v-model:value="panelState.panelConfig.marginTop" class="max-w-[200px]" :step="1" :max="50" />
          </div>
        </NGridItem>
        <NGridItem span="12 400:6">
          <div class="flex items-center mt-3">
            <span class="mr-3">{{ $t('apps.baseSettings.bottomMargin') }} (%)</span>
            <NSlider v-model:value="panelState.panelConfig.marginBottom" class="max-w-[200px]" :step="1" :max="50" />
          </div>
        </NGridItem>
      </NGrid>
    </NCard>

    <NCard class="mt-3" size="small">
      <div class="text-text1 text-[16px] font-medium mb-2">
        {{ $t('apps.baseSettings.customFooter') }}
      </div>

      <NInput
        v-model:value="panelState.panelConfig.footerHtml"
        type="textarea"
        clearable
      />
    </NCard>

    <NCard class="mt-3" size="small">
      <NPopconfirm
        @positive-click="resetPanelConfig"
      >
        <template #trigger>
          <NButton size="small" quaternary type="error">
            {{ $t('common.reset') }}
          </NButton>
        </template>
        {{ $t('apps.baseSettings.resetWarnText') }}
      </NPopconfirm>

      <NButton size="small" secondary class="ml-3" @click="uploadCloud">
        {{ $t('common.save') }}
      </NButton>
    </NCard>
  </div>
</template>
