<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { NLayout, NLayoutContent, NLayoutSider } from 'naive-ui'
import { useAuthStore } from '@/store'
import { AppLoader, RoundCardModal, SvgIcon } from '@/components/common'
import { t } from '@/locales'

interface App {
  name: string
  componentName: string
  icon: string
  auth?: number
}
const props = defineProps<{
  visible: boolean
  initialPanel?: string
}>()

const emit = defineEmits<{
  (e: 'update:visible', visible: boolean): void
}>()

const componentName = ref(props.visible ? props.initialPanel ?? 'UserInfo' : 'UserInfo')
watch(() => props.visible, (visible) => {
  if (visible)
    componentName.value = props.initialPanel ?? 'UserInfo'
})
const collapsed = ref(false)
const screenWidth = ref(0)
const isSmallScreen = ref(false)
const defaultTitle = t('appLauncher.title')
const title = ref('')
const height = computed(() => isSmallScreen.value ? 'calc(100dvh - 112px)' : '500px')

const apps = ref<App[]>([
  {
    name: t('apps.userInfo.appName'),
    componentName: 'UserInfo',
    icon: 'lucide-user-round-pen',
  },
  {
    name: t('apps.baseSettings.appName'),
    componentName: 'Style',
    icon: 'lucide-palette',
  },
  {
    name: t('apps.itemGroupManage.appName'),
    componentName: 'ItemGroupManage',
    icon: 'lucide-folder-closed',
  },
  {
    name: t('apps.uploadsFileManager.appName'),
    componentName: 'UploadFileManager',
    icon: 'lucide-upload',
  },
  {
    name: t('apps.exportImport.appName'),
    componentName: 'ImportExport',
    icon: 'lucide-arrow-down-up',
  },
  {
    name: t('apps.about.appName'),
    componentName: 'About',
    icon: 'lucide-info',
  },
])

const authStore = useAuthStore()

const show = computed({
  get: () => props.visible,
  set: (visible: boolean) => {
    emit('update:visible', visible)
  },
})

function handleClickApp(item: App) {
  componentName.value = item.componentName
  if (isSmallScreen.value)
    collapsed.value = true
}

function getScreenWidth() {
  return window.innerWidth
}

function handleResize() {
  screenWidth.value = getScreenWidth()
  if (screenWidth.value < 768) {
    collapsed.value = true
    isSmallScreen.value = true
  }
  else {
    collapsed.value = false
    isSmallScreen.value = false
  }
}

onMounted(() => {
  const adminApp: App = {
    name: t('adminSettingUsers.appName'),
    componentName: 'Users',
    icon: 'lucide-users',
    auth: 1,
  }
  // 初始化
  if (authStore.userInfo?.role === 1)
    apps.value.push(adminApp)

  window.addEventListener('resize', handleResize)
  handleResize()
})

onUnmounted(() => {
  window.removeEventListener('resize', handleResize)
})
</script>

<template>
  <div>
    <RoundCardModal
      v-model:show="show"
      style="width: calc(100vw - 32px); max-width: 900px;"
      size="small"
    >
      <template #header>
        <div class="flex items-center gap-2 select-none">
          <button type="button" class="sidebar-toggle" :aria-label="t('appLauncher.toggleSidebar')" :aria-expanded="!collapsed" @click="collapsed = !collapsed">
            <SvgIcon class=" transition-all duration-500" :icon="collapsed ? 'lucide-panel-left-open' : 'lucide-panel-left-close'" />
          </button>
          <div>
            {{ title === '' ? defaultTitle : title }}
          </div>
        </div>
      </template>
      <div class="w-full app-starter-modal-content" :style="{ height }">
        <NLayout has-sider class="h-full rounded-card bg-bg0">
          <NLayoutSider
            v-model:collapsed="collapsed"
            collapse-mode="width"
            :collapsed-width="0"
            :width="isSmallScreen ? '100%' : 240"
            style="height: 100%;"
            content-style="overflow: hidden"
          >
            <div class="w-full h-full bg-bg0">
              <div
                class="p-2 overflow-auto"
                :style="{
                  width: isSmallScreen ? '100%' : '220px',
                  minWidth: '200px',
                  height: '100%',
                }"
              >
                <button
                  v-for="item in apps"
                  :key="item.componentName"
                  class="category-tab mb-1"
                  :class="{ active: componentName === item.componentName }"
                  type="button"
                  :aria-current="componentName === item.componentName ? 'page' : undefined"
                  @click="handleClickApp(item)"
                >
                  <SvgIcon class="text-[18px]" :icon="item.icon" />
                  <span>{{ item.name }}</span>
                </button>
              </div>
            </div>
          </NLayoutSider>
          <NLayoutContent class="bg-bg0" :content-style="{ height: '100%' }">
            <div class="h-full overflow-auto min-w-[300px]" :class="(isSmallScreen && !collapsed) ? 'opacity-0' : 'opacity-100'">
              <AppLoader :component-name="componentName" class="h-full" />
            </div>
          </NLayoutContent>
        </NLayout>
      </div>
    </RoundCardModal>
  </div>
</template>

<style scoped>
.sidebar-toggle { display: grid; place-items: center; flex: none; width: 32px; height: 32px; padding: 0; border: 0; border-radius: var(--x-r-glyph); background: transparent; color: var(--x-text1); font-size: 18px; cursor: pointer; }
.sidebar-toggle:hover { background: var(--x-hover); }
.sidebar-toggle:focus-visible { outline: 2px solid var(--x-primary); outline-offset: 2px; }
.category-tab { display: flex; align-items: center; gap: 8px; width: 100%; position: relative; height: 40px; padding: 0 16px; border: 0; border-radius: var(--x-r-glyph); background: transparent; color: var(--x-text2); font-size: 15px; font-weight: 500; text-align: left; cursor: pointer; }
.category-tab::before { position: absolute; inset: -2px 0; content: ""; }
.category-tab.active { background: var(--x-bg1); color: var(--x-text1); }
.category-tab:focus-visible { outline: 2px solid var(--x-primary); outline-offset: 2px; }
</style>
