<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from 'vue'
import type { UploadFileInfo } from 'naive-ui'
import { NButton, NCard, NGrid, NGridItem, NInput, NInputGroup, NPopconfirm, NSelect, NSlider, NSwitch, NUpload, NUploadDragger } from 'naive-ui'
import { useAuthStore, usePanelState } from '@/store'
import { set as setUserConfig } from '@/api/panel/userConfig'
import { faviconCandidates } from '@/api/panel/itemIcon'
import { t } from '@/locales'
import { ItemIcon } from '@/components/common'
import { useFaviconFill } from '@/hooks/useFaviconFill'

const authStore = useAuthStore()
const panelState = usePanelState()
const showWallpaperInput = ref(false)
interface FaviconCandidate { id: number; title: string; url: string; icon: Panel.ItemIcon }
const candidates = ref<FaviconCandidate[]>([])
const selectedIds = ref<number[]>([])
const candidateLoading = ref(false)
const candidateError = ref('')
const saveState = ref<'idle' | 'saving' | 'saved' | 'error'>('idle')
const saveError = ref('')
let savedTimer: ReturnType<typeof setTimeout>
const { running: fillRunning, done: fillDone, success: fillSuccess, failures, reset: resetFill, run: runFaviconFill } = useFaviconFill()
const fillFailures = computed(() => failures.value.map(failure => `${candidates.value.find(item => item.id === failure.id)?.title || failure.id}: ${failure.reason}`))

async function loadCandidates() {
  candidateLoading.value = true
  candidateError.value = ''
  try {
    const response = await faviconCandidates<Common.ListResponse<FaviconCandidate[]>>()
    if (response.code !== 0)
      throw new Error(response.msg || t('common.failed'))
    candidates.value = response.data.list
    selectedIds.value = candidates.value.map(item => item.id)
    resetFill()
  }
  catch (error) {
    candidateError.value = error instanceof Error ? error.message : t('common.failed')
  }
  finally {
    candidateLoading.value = false
  }
}

async function runFill() {
  await runFaviconFill(selectedIds.value, (result) => {
    const item = candidates.value.find(candidate => candidate.id === result.id)
    if (item && result.iconUrl)
      item.icon = { ...item.icon, itemType: 2, src: result.iconUrl }
  })
}

const isSaveing = ref(false)

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

async function uploadCloud() {
  clearTimeout(savedTimer)
  saveState.value = 'saving'
  saveError.value = ''
  try {
    const res = await setUserConfig({ panel: panelState.panelConfig })
    if (res.code !== 0)
      throw new Error(t('apps.baseSettings.configFailed', { message: res.msg }))
    saveState.value = 'saved'
    savedTimer = setTimeout(() => saveState.value = 'idle', 1500)
  }
  catch (error) {
    saveError.value = error instanceof Error ? error.message : t('common.saveFail')
    saveState.value = 'error'
  }
}

function resetPanelConfig() {
  panelState.resetPanelConfig()
}
onBeforeUnmount(() => clearTimeout(savedTimer))
</script>

<template>
  <div class="p-2 overflow-auto">
    <div class="panel-actions">
      <NButton type="primary" :loading="saveState === 'saving'" @click="uploadCloud">
        {{ $t('common.save') }}
      </NButton>
      <NPopconfirm @positive-click="resetPanelConfig">
        <template #trigger>
          <NButton secondary>
            {{ $t('common.reset') }}
          </NButton>
        </template>
        {{ $t('apps.baseSettings.resetWarnText') }}
      </NPopconfirm>
      <span v-if="saveState === 'saved'" class="setting-status" role="status">{{ t('panelHome.saved') }}</span>
    </div>
    <p v-if="saveState === 'error'" class="setting-error" role="alert">
      {{ saveError }} <NButton size="small" @click="uploadCloud">
        {{ t('panelHome.retry') }}
      </NButton>
    </p>
    <NCard size="small">
      <div class="text-text1 text-[17px] font-medium mb-2">
        LOGO
      </div>

      <div>
        <div>
          {{ $t('apps.baseSettings.textContent') }}
        </div>
        <div class="flex items-center mt-1">
          <NInput v-model:value="panelState.panelConfig.logoText" type="text" show-count :maxlength="20" placeholder="请输入文字" :input-props="{ 'aria-label': $t('apps.baseSettings.textContent') }" />
        </div>
      </div>
    </NCard>

    <NCard class="mt-3" size="small">
      <div class="text-text1 text-[17px] font-medium mb-2">
        {{ $t('apps.baseSettings.clock') }}
      </div>
      <div class="flex items-center mt-1">
        <span class="mr-3">{{ $t('common.show') }}</span>
        <NSwitch v-model:value="panelState.panelConfig.clockShow" :aria-label="$t('apps.baseSettings.clock')" />
      </div>
      <div v-if="panelState.panelConfig.clockShow" class="flex items-center mt-1">
        <span class="mr-3">{{ $t('apps.baseSettings.clockSecondShow') }}</span>
        <NSwitch v-model:value="panelState.panelConfig.clockShowSecond" :aria-label="$t('apps.baseSettings.clockSecondShow')" />
      </div>
    </NCard>

    <NCard class="mt-3" size="small">
      <div class="text-text1 text-[17px] font-medium mb-2">
        {{ $t('apps.baseSettings.searchBar') }}
      </div>
      <div class="flex items-center mt-1">
        <span class="mr-3">{{ $t('common.show') }}</span>
        <NSwitch v-model:value="panelState.panelConfig.searchBoxShow" :aria-label="$t('apps.baseSettings.searchBar')" />
      </div>
      <div v-if="panelState.panelConfig.searchBoxShow" class="flex items-center mt-1">
        <span class="mr-3">{{ $t('apps.baseSettings.searchBarSearchItem') }}</span>
        <NSwitch v-model:value="panelState.panelConfig.searchBoxSearchIcon" :aria-label="$t('apps.baseSettings.searchBarSearchItem')" />
      </div>
    </NCard>

    <NCard class="mt-3" size="small">
      <div class="text-text1 text-[17px] font-medium mb-2">
        {{ $t('apps.baseSettings.systemMonitorStatus') }}
      </div>
      <div class="flex items-center mt-1">
        <span class="mr-3">{{ $t('common.show') }}</span>
        <NSwitch v-model:value="panelState.panelConfig.systemMonitorShow" :aria-label="$t('apps.baseSettings.systemMonitorStatus')" />
      </div>
      <div v-if="panelState.panelConfig.systemMonitorShow" class="flex items-center mt-1">
        <span class="mr-3">{{ $t('apps.baseSettings.showTitle') }}</span>
        <NSwitch v-model:value="panelState.panelConfig.systemMonitorShowTitle" :aria-label="$t('apps.baseSettings.showTitle')" />
      </div>
      <div v-if="panelState.panelConfig.systemMonitorShow" class="flex items-center mt-1">
        <span class="mr-3">{{ $t('apps.baseSettings.publicVisitModeShow') }}</span>
        <NSwitch v-model:value="panelState.panelConfig.systemMonitorPublicVisitModeShow" :aria-label="$t('apps.baseSettings.publicVisitModeShow')" />
      </div>
    </NCard>

    <NCard class="mt-3" size="small">
      <div class="text-text1 text-[17px] font-medium mb-2">
        {{ $t('apps.baseSettings.wallpaper') }}
      </div>
      <div class="flex items-center gap-3 mb-4">
        <span>{{ $t('apps.baseSettings.showWallpaper') }}</span>
        <NSwitch v-model:value="panelState.panelConfig.showWallpaper" :aria-label="$t('apps.baseSettings.showWallpaper')" />
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
          <NSwitch v-model:value="showWallpaperInput" :aria-label="$t('apps.baseSettings.customImageAddress')" />
        </div>

        <div class="flex items-center mt-3 gap-3">
          <span>{{ $t('apps.baseSettings.autoWallpaper') }}</span>
          <NSwitch v-model:value="panelState.panelConfig.autoWallpaper" :aria-label="$t('apps.baseSettings.autoWallpaper')" />
        </div>
        <div v-if="showWallpaperInput" class="mt-1">
          <NInput v-model:value="panelState.panelConfig.backgroundImageSrc" :input-props="{ 'aria-label': $t('apps.baseSettings.customImageAddress') }" type="text" size="small" clearable />
        </div>

        <div class="flex items-center mt-3">
          <span id="style-background-blur-label" class="mr-3">{{ $t('apps.baseSettings.vague') }}</span>
          <div role="group" aria-labelledby="style-background-blur-label" class="w-full max-w-[200px]">
            <NSlider v-model:value="panelState.panelConfig.backgroundBlur" :step="2" :max="20" />
          </div>
        </div>

        <div class="flex items-center mt-3">
          <span id="style-background-mask-label" class="mr-3">{{ $t('apps.baseSettings.mask') }}</span>
          <div role="group" aria-labelledby="style-background-mask-label" class="w-full max-w-[200px]">
            <NSlider v-model:value="panelState.panelConfig.backgroundMaskNumber" :step="0.1" :max="1" />
          </div>
        </div>
      </div>
    </NCard>

    <NCard size="small" class="mt-3">
      <div class="text-text1 text-[17px] font-medium mb-3">
        {{ $t('apps.baseSettings.fillFavicons') }}
      </div>
      <NButton :disabled="fillRunning" :loading="candidateLoading" @click="loadCandidates">
        {{ $t('apps.baseSettings.findCandidates') }}
      </NButton>
      <p v-if="candidateError" class="setting-error" role="alert">
        {{ candidateError }} <NButton size="small" @click="loadCandidates">
          {{ t('panelHome.retry') }}
        </NButton>
      </p>
      <div v-for="item in candidates" :key="item.id" class="flex gap-3 items-center mt-2">
        <input v-model="selectedIds" type="checkbox" :aria-label="item.title" :value="item.id" :disabled="fillRunning">
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
      <div class="text-text1 text-[17px] font-medium mb-2">
        {{ $t('apps.baseSettings.contentArea') }}
      </div>

      <NGrid cols="2">
        <NGridItem span="12 400:12">
          <div class="flex items-center mt-1">
            <span class="mr-3">{{ $t('apps.baseSettings.netModeChangeButtonShow') }}</span>
            <NSwitch v-model:value="panelState.panelConfig.netModeChangeButtonShow" :aria-label="$t('apps.baseSettings.netModeChangeButtonShow')" />
          </div>
        </NGridItem>

        <NGridItem span="12 400:12">
          <div class="flex items-center mt-3">
            <span class="mr-3">{{ $t('apps.baseSettings.maxWidth') }}</span>
            <div class="flex">
              <NInputGroup>
                <NInput v-model:value="panelState.panelConfig.maxWidth" size="small" type="number" :maxlength="10" :style="{ width: '100px' }" placeholder="1200" :input-props="{ 'aria-label': $t('apps.baseSettings.maxWidth') }" />
                <NSelect v-model:value="panelState.panelConfig.maxWidthUnit" :style="{ width: '80px' }" :options="maxWidthUnitOption" size="small" :aria-label="$t('apps.baseSettings.maxWidthUnit')" />
              </NInputGroup>
            </div>
          </div>
        </NGridItem>
        <NGridItem span="12 400:12">
          <div class="flex items-center mt-3">
            <span id="style-horizontal-margin-label" class="mr-3">{{ $t('apps.baseSettings.leftRightMargin') }}</span>
            <div role="group" aria-labelledby="style-horizontal-margin-label" class="w-full max-w-[200px]">
              <NSlider v-model:value="panelState.panelConfig.marginX" :step="1" :max="100" />
            </div>
          </div>
        </NGridItem>
        <NGridItem span="12 400:12">
          <div class="flex items-center mt-3">
            <span id="style-top-margin-label" class="mr-3">{{ $t('apps.baseSettings.topMargin') }} (%)</span>
            <div role="group" aria-labelledby="style-top-margin-label" class="w-full max-w-[200px]">
              <NSlider v-model:value="panelState.panelConfig.marginTop" :step="1" :max="50" />
            </div>
          </div>
        </NGridItem>
        <NGridItem span="12 400:6">
          <div class="flex items-center mt-3">
            <span id="style-bottom-margin-label" class="mr-3">{{ $t('apps.baseSettings.bottomMargin') }} (%)</span>
            <div role="group" aria-labelledby="style-bottom-margin-label" class="w-full max-w-[200px]">
              <NSlider v-model:value="panelState.panelConfig.marginBottom" :step="1" :max="50" />
            </div>
          </div>
        </NGridItem>
      </NGrid>
    </NCard>

    <NCard class="mt-3" size="small">
      <div class="text-text1 text-[17px] font-medium mb-2">
        {{ $t('apps.baseSettings.customFooter') }}
      </div>

      <NInput
        v-model:value="panelState.panelConfig.footerHtml"
        :input-props="{ 'aria-label': $t('apps.baseSettings.customFooter') }"
        type="textarea"
        clearable
      />
    </NCard>
  </div>
</template>

<style scoped>
.setting-status { margin-left: 12px; color: var(--x-green-fg); font-size: 13px; }
.setting-error { margin-top: 12px; color: var(--x-red-fg); font-size: 13px; }
</style>
