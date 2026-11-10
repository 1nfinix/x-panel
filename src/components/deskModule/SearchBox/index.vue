<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { NCheckbox, NPopover } from 'naive-ui'
import { SvgIcon } from '@/components/common'
import { useModuleConfig } from '@/store/modules'
import { useAuthStore } from '@/store'
import { VisitMode } from '@/enums/auth'
import { t } from '@/locales'

import SvgSrcBaidu from '@/assets/search_engine_svg/baidu.svg'
import SvgSrcBing from '@/assets/search_engine_svg/bing.svg'
import SvgSrcGoogle from '@/assets/search_engine_svg/google.svg'

const emit = defineEmits<{
  (e: 'itemSearch', keyword: string): void
}>()

interface State {
  currentSearchEngine: DeskModule.SearchBox.SearchEngine
  searchEngineList: DeskModule.SearchBox.SearchEngine[]
  newWindowOpen: boolean
}

const moduleConfigName = 'deskModuleSearchBox'
const moduleConfig = useModuleConfig()
const authStore = useAuthStore()
const searchTerm = ref('')
const searchSelectListShow = ref(false)
const engineSaving = ref(false)
const engineError = ref('')
const defaultSearchEngineList: DeskModule.SearchBox.SearchEngine[] = [
  { iconSrc: SvgSrcGoogle, title: 'Google', url: 'https://www.google.com/search?q=%s' },
  { iconSrc: SvgSrcBaidu, title: 'Baidu', url: 'https://www.baidu.com/s?wd=%s' },
  { iconSrc: SvgSrcBing, title: 'Bing', url: 'https://www.bing.com/search?q=%s' },
]
const defaultState: State = {
  currentSearchEngine: defaultSearchEngineList[0],
  searchEngineList: [],
  newWindowOpen: false,
}
const state = ref<State>({ ...defaultState })

async function saveEngineSettings() {
  if (authStore.visitMode === VisitMode.VISIT_MODE_PUBLIC)
    return true
  try {
    const { code, msg } = await moduleConfig.saveToCloud(moduleConfigName, state.value)
    if (code !== 0)
      throw new Error(msg || t('common.saveFail'))
    return true
  }
  catch (error) {
    engineError.value = error instanceof Error ? error.message : t('common.saveFail')
    return false
  }
}

async function handleEngineUpdate(engine: DeskModule.SearchBox.SearchEngine) {
  if (engineSaving.value)
    return
  const previous = state.value.currentSearchEngine
  state.value.currentSearchEngine = engine
  engineSaving.value = true
  engineError.value = ''
  const saved = await saveEngineSettings()
  engineSaving.value = false
  if (saved)
    searchSelectListShow.value = false
  else
    state.value.currentSearchEngine = previous
}

async function handleNewWindowUpdate(checked: boolean) {
  state.value.newWindowOpen = checked
  engineSaving.value = true
  engineError.value = ''
  const saved = await saveEngineSettings()
  engineSaving.value = false
  if (!saved)
    state.value.newWindowOpen = !checked
}

function handleSearchClick() {
  const url = state.value.currentSearchEngine.url
  const fullUrl = url.includes('%s')
    ? url.replace('%s', encodeURIComponent(searchTerm.value))
    : url + encodeURIComponent(searchTerm.value)
  handleClearSearchTerm()
  if (state.value.newWindowOpen)
    window.open(fullUrl)
  else
    window.location.href = fullUrl
}

function handleItemSearch() {
  emit('itemSearch', searchTerm.value)
}

function handleClearSearchTerm() {
  searchTerm.value = ''
  emit('itemSearch', '')
}

onMounted(async () => {
  try {
    const { code, data } = await moduleConfig.getValueByNameFromCloud<State>(moduleConfigName)
    state.value = (code === 0 && data) ? data : { ...defaultState }
  }
  catch {
    state.value = { ...defaultState }
  }
})
</script>

<template>
  <div class="search-box" @keydown.esc="handleClearSearchTerm">
    <div class="search-controls">
      <NPopover v-model:show="searchSelectListShow" trigger="click" placement="bottom-start" :show-arrow="false" class="engine-popover">
        <template #trigger>
          <button class="engine-button" type="button" :aria-label="state.currentSearchEngine.title" :aria-expanded="searchSelectListShow">
            <img :src="state.currentSearchEngine.iconSrc" alt=""><span>{{ state.currentSearchEngine.title }}</span><span class="chevron" aria-hidden="true" />
          </button>
        </template>
        <div class="engine-menu" role="group" :aria-label="t('deskModule.searchBox.inputPlaceholder')">
          <button v-for="engine in defaultSearchEngineList" :key="engine.title" class="engine-option" :class="{ active: state.currentSearchEngine.title === engine.title }" type="button" :disabled="engineSaving" :aria-pressed="state.currentSearchEngine.title === engine.title" @click="handleEngineUpdate(engine)">
            <img :src="engine.iconSrc" alt="">{{ engine.title }}
          </button>
          <NCheckbox :checked="state.newWindowOpen" :disabled="engineSaving" @update:checked="handleNewWindowUpdate">
            {{ t('deskModule.searchBox.openWithNewOpen') }}
          </NCheckbox>
          <p v-if="engineSaving" class="engine-status" aria-live="polite">
            {{ t('panelHome.saving') }}
          </p>
          <p v-if="engineError" class="engine-error" role="alert">
            {{ engineError }}
          </p>
        </div>
      </NPopover>
      <div class="search-field">
        <input v-model="searchTerm" :placeholder="t('deskModule.searchBox.inputPlaceholder')" :aria-label="t('deskModule.searchBox.inputPlaceholder')" @input="handleItemSearch" @keydown.enter="handleSearchClick">
        <button v-if="searchTerm" class="clear-button" type="button" :aria-label="t('common.cancel')" @click="handleClearSearchTerm">
          <SvgIcon icon="line-md:close-small" />
        </button>
      </div>
      <button class="submit-button" type="button" :aria-label="t('deskModule.searchBox.inputPlaceholder')" @click="handleSearchClick">
        <SvgIcon icon="iconamoon:search-fill" />
      </button>
    </div>
  </div>
</template>

<style scoped>
.search-box { width: 100%; }
.search-controls { display: flex; align-items: center; gap: 8px; }
.engine-button, .engine-option { display: flex; align-items: center; gap: 8px; min-height: 44px; padding: 0 14px; border: 0; border-radius: var(--x-r-pill); background: var(--x-bg2); color: var(--x-text1); cursor: pointer; white-space: nowrap; }
.engine-button { justify-content: center; }
.engine-button img, .engine-option img { width: 18px; height: 18px; }
.chevron { width: 7px; height: 7px; border-right: 1.5px solid currentColor; border-bottom: 1.5px solid currentColor; transform: translateY(-2px) rotate(45deg); }
.search-field { display: flex; align-items: center; flex: 1; min-width: 0; height: 44px; padding: 0 16px; border-radius: var(--x-r-pill); background: var(--x-bg2); }
.search-field input { flex: 1; min-width: 0; height: 100%; border: 0; outline: 0; background: transparent; color: var(--x-text1); font-size: 15px; user-select: text; }
.search-field:focus-within { box-shadow: 0 0 0 2px var(--x-primary), 0 0 0 5px var(--x-focus); }
.clear-button { border: 0; background: transparent; color: var(--x-text2); cursor: pointer; }
.submit-button { display: grid; place-items: center; flex: none; width: 44px; height: 44px; border: 0; border-radius: 50%; background: var(--x-primary); color: var(--x-on-primary); cursor: pointer; }
.engine-menu { display: flex; flex-direction: column; gap: 8px; min-width: 180px; }
.engine-option { width: 100%; border-radius: var(--x-r-icon); text-align: left; }
.engine-option.active { background: var(--x-blue-tint); color: var(--x-blue-fg); }
.engine-error { margin: 0; color: var(--x-red-fg); font-size: 13px; }
.engine-status { margin: 0; color: var(--x-text2); font-size: 13px; }
.engine-button:focus-visible, .engine-option:focus-visible, .submit-button:focus-visible, .clear-button:focus-visible { outline: 3px solid var(--x-focus); outline-offset: 2px; }
@media (max-width: 767px) { .engine-button > span:not(.chevron) { display: none; } }
</style>
