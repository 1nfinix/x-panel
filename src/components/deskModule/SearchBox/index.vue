<script setup lang="ts">
import { defineEmits, onMounted, ref } from 'vue'
import { NCheckbox } from 'naive-ui'
import { SvgIcon } from '@/components/common'
import { useModuleConfig } from '@/store/modules'
import { useAuthStore } from '@/store'
import { VisitMode } from '@/enums/auth'

import SvgSrcBaidu from '@/assets/search_engine_svg/baidu.svg'
import SvgSrcBing from '@/assets/search_engine_svg/bing.svg'
import SvgSrcGoogle from '@/assets/search_engine_svg/google.svg'

withDefaults(defineProps<{
  background?: string
  textColor?: string
}>(), {
  background: 'var(--x-bg2)',
  textColor: 'var(--x-text1)',
})

const emits = defineEmits(['itemSearch'])

interface State {
  currentSearchEngine: DeskModule.SearchBox.SearchEngine
  searchEngineList: DeskModule.SearchBox.SearchEngine[]
  newWindowOpen: boolean
}

const moduleConfigName = 'deskModuleSearchBox'
const moduleConfig = useModuleConfig()
const authStore = useAuthStore()
const searchTerm = ref('')
const isFocused = ref(false)
const searchSelectListShow = ref(false)
const defaultSearchEngineList = ref<DeskModule.SearchBox.SearchEngine[]>([
  {
    iconSrc: SvgSrcGoogle,
    title: 'Google',
    url: 'https://www.google.com/search?q=%s',
  },
  {
    iconSrc: SvgSrcBaidu,
    title: 'Baidu',
    url: 'https://www.baidu.com/s?wd=%s',
  },
  {
    iconSrc: SvgSrcBing,
    title: 'Bing',
    url: 'https://www.bing.com/search?q=%s',
  },
])

const defaultState: State = {
  currentSearchEngine: defaultSearchEngineList.value[0],
  searchEngineList: [] || defaultSearchEngineList,
  newWindowOpen: false,
}

const state = ref<State>({ ...defaultState })

const onFocus = (): void => {
  isFocused.value = true
}

const onBlur = (): void => {
  isFocused.value = false
}

function handleEngineClick() {
  // 访客模式不允许修改
  if (authStore.visitMode === VisitMode.VISIT_MODE_PUBLIC)
    return
  searchSelectListShow.value = !searchSelectListShow.value
}

function handleEngineUpdate(engine: DeskModule.SearchBox.SearchEngine) {
  state.value.currentSearchEngine = engine
  if (authStore.visitMode !== VisitMode.VISIT_MODE_PUBLIC)
    moduleConfig.saveToCloud(moduleConfigName, state.value)
  searchSelectListShow.value = false
}

function handleSearchClick() {
  const url = state.value.currentSearchEngine.url
  const keyword = searchTerm
  // 如果网址中存在 %s，则直接替换为关键字
  const fullUrl = replaceOrAppendKeywordToUrl(url, keyword.value)
  handleClearSearchTerm()
  if (state.value.newWindowOpen)
    window.open(fullUrl)
  else
    window.location.href = fullUrl
}

function replaceOrAppendKeywordToUrl(url: string, keyword: string) {
  // 如果网址中存在 %s，则直接替换为关键字
  if (url.includes('%s'))
    return url.replace('%s', encodeURIComponent(keyword))

  // 如果网址中不存在 %s，则将关键字追加到末尾
  return url + (keyword ? `${encodeURIComponent(keyword)}` : '')
}

const handleItemSearch = () => {
  emits('itemSearch', searchTerm.value)
}

function handleClearSearchTerm() {
  searchTerm.value = ''
  emits('itemSearch', searchTerm.value)
}

onMounted(() => {
  moduleConfig.getValueByNameFromCloud<State>('deskModuleSearchBox').then(({ code, data }) => {
    if (code === 0)
      state.value = data || defaultState
    else
      state.value = defaultState
  })
})
</script>

<template>
  <div class="search-box" @keydown.enter="handleSearchClick" @keydown.esc="handleClearSearchTerm">
    <div class="search-controls">
      <button class="engine-button" type="button" :aria-label="state.currentSearchEngine.title" @click="handleEngineClick">
        <img :src="state.currentSearchEngine.iconSrc" alt=""><span>{{ state.currentSearchEngine.title }}</span><span class="chevron" aria-hidden="true" />
      </button>
      <label class="search-field" :class="{ focused: isFocused }"><input v-model="searchTerm" :placeholder="$t('deskModule.searchBox.inputPlaceholder')" @focus="onFocus" @blur="onBlur" @input="handleItemSearch"><button v-if="searchTerm" class="clear-button" type="button" :aria-label="$t('common.cancel')" @click="handleClearSearchTerm"><SvgIcon icon="line-md:close-small" /></button></label>
      <button class="submit-button" type="button" :aria-label="$t('deskModule.searchBox.inputPlaceholder')" @click="handleSearchClick">
        <SvgIcon icon="iconamoon:search-fill" />
      </button>
    </div>
    <div class="search-engines">
      <button v-for="item in defaultSearchEngineList" :key="item.title" class="quick-engine" :aria-pressed="state.currentSearchEngine.title === item.title" type="button" @click="handleEngineUpdate(item)">
        <img :src="item.iconSrc" alt="">{{ item.title }}
      </button>
    </div>
    <div v-if="searchSelectListShow" class="search-options">
      <NCheckbox v-model:checked="state.newWindowOpen" @update-checked="moduleConfig.saveToCloud(moduleConfigName, state)">
        {{ $t('deskModule.searchBox.openWithNewOpen') }}
      </NCheckbox>
    </div>
  </div>
</template>

<style scoped>
.search-box { width: 100%; }.search-controls { display: flex; align-items: center; gap: 8px; }.engine-button, .quick-engine { display: flex; align-items: center; justify-content: center; gap: 8px; height: 44px; padding: 0 14px; border: 0; border-radius: 999px; background: var(--x-bg2); color: var(--x-text1); cursor: pointer; white-space: nowrap; }.engine-button img, .quick-engine img { width: 18px; height: 18px; }.chevron { width: 7px; height: 7px; border-right: 1.5px solid currentColor; border-bottom: 1.5px solid currentColor; transform: translateY(-2px) rotate(45deg); }.search-field { display: flex; align-items: center; flex: 1; min-width: 0; height: 44px; padding: 0 16px; border-radius: 999px; background: var(--x-bg2); }.search-field input { flex: 1; min-width: 0; height: 100%; border: 0; outline: 0; background: transparent; color: var(--x-text1); font-size: 15px; user-select: text; }.search-field:focus-within { box-shadow: 0 0 0 2px var(--x-primary), 0 0 0 5px var(--x-focus); }.clear-button { border: 0; background: transparent; color: var(--x-text2); cursor: pointer; }.submit-button { display: grid; place-items: center; flex: none; width: 44px; height: 44px; border: 0; border-radius: 50%; background: var(--x-primary); color: var(--x-on-primary); cursor: pointer; }.search-engines { display: flex; flex-wrap: wrap; gap: 8px; margin-top: 12px; }.quick-engine { font-size: 13px; }.quick-engine[aria-pressed="true"] { background: var(--x-blue-tint); color: var(--x-blue-fg); }.search-options { margin-top: 12px; }.engine-button:focus-visible, .quick-engine:focus-visible, .submit-button:focus-visible, .clear-button:focus-visible { outline: 3px solid var(--x-focus); outline-offset: 2px; }
@media (max-width: 767px) { .engine-button span { display: none; }.quick-engine { padding: 0 10px; } }
</style>
