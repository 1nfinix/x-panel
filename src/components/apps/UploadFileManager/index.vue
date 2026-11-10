<script setup lang="ts">
import { NAlert, NButton, NButtonGroup, NCard, NEllipsis, NGrid, NGridItem, NImage, NImageGroup, NSpin, useDialog, useMessage } from 'naive-ui'
import { h, onMounted, ref } from 'vue'
import { deletes, getList } from '@/api/system/file'
import { set as savePanelConfig } from '@/api/panel/userConfig'
import { RoundCardModal, SvgIcon } from '@/components/common'
import { copyToClipboard, timeFormat } from '@/utils/cmn'
import { t } from '@/locales'
import { usePanelState } from '@/store'

interface InfoModalState {
  title: string
  show: boolean
  fileInfo: File.Info | null
}
const imageList = ref<File.Info[]>([])
const ms = useMessage()
const dialog = useDialog()
const panelStore = usePanelState()
const loading = ref(false)
const operationError = ref('')
const listError = ref('')
const deleteLoadingId = ref<number | null>(null)
const wallpaperLoadingId = ref<number | null>(null)
const infoModalState = ref<InfoModalState>({
  show: false,
  title: '',
  fileInfo: null,
})

async function getFileList() {
  loading.value = true
  try {
    const { code, data } = await getList<Common.ListResponse<File.Info[]>>()
    if (code !== 0)
      throw new Error(t('common.failed'))
    imageList.value = data.list
    listError.value = ''
  }
  catch {
    listError.value = t('common.failed')
  }
  finally {
    loading.value = false
  }
}

async function copyImageUrl(text: string) {
  const res = await copyToClipboard(text)
  if (res)
    ms.success(t('apps.uploadsFileManager.copySuccess'))

  else
    ms.error(t('apps.uploadsFileManager.copyFailed'))
}

function handleDelete(id: number) {
  const deleteError = ref('')
  const confirmDialog = dialog.warning({
    title: t('common.warning'),
    content: () => h('div', [h('p', t('apps.uploadsFileManager.deleteWarningText')), deleteError.value ? h('p', { style: 'color:var(--x-red-fg);margin-top:12px' }, deleteError.value) : null]),
    positiveText: t('common.confirm'),
    negativeText: t('common.cancel'),
    onPositiveClick: async () => {
      deleteLoadingId.value = id
      confirmDialog.loading = true
      try {
        const { code, msg } = await deletes([id])
        if (code !== 0) {
          deleteError.value = `${t('common.failed')}: ${msg}`
          return false
        }
        await getFileList()
        ms.success(t('common.success'))
      }
      catch {
        deleteError.value = t('common.failed')
        return false
      }
      finally {
        deleteLoadingId.value = null
        confirmDialog.loading = false
      }
    },
  })
}

function handleInfoClick(fileInfo: File.Info) {
  infoModalState.value.fileInfo = fileInfo
  infoModalState.value.show = true
}

async function handleSetWallpaper(imgSrc: string, id: number) {
  const previous = panelStore.panelConfig.backgroundImageSrc
  wallpaperLoadingId.value = id
  panelStore.panelConfig.backgroundImageSrc = imgSrc
  try {
    const { code, msg } = await savePanelConfig({ panel: panelStore.panelConfig })
    if (code !== 0)
      throw new Error(msg)
    operationError.value = ''
    ms.success(t('common.saveSuccess'))
  }
  catch (error) {
    panelStore.panelConfig.backgroundImageSrc = previous
    operationError.value = error instanceof Error ? error.message : t('common.saveFail')
  }
  finally {
    wallpaperLoadingId.value = null
  }
}

onMounted(() => {
  getFileList()
})
</script>

<template>
  <div class="p-2 h-full">
    <NSpin v-show="loading" size="small" />
    <p v-if="operationError" class="operation-error" role="alert">
      {{ operationError }}
    </p>
    <p v-if="listError" class="operation-error" role="alert">
      {{ listError }} <NButton size="small" @click="getFileList">
        {{ t('panelHome.retry') }}
      </NButton>
    </p>
    <NAlert type="info" :bordered="false">
      {{ $t('apps.uploadsFileManager.alertText') }}
    </NAlert>
    <div class="flex justify-center mt-2">
      <div v-if="imageList.length === 0 && !loading" class="flex">
        {{ $t('apps.uploadsFileManager.nothingText') }}
      </div>
      <NImageGroup v-else>
        <NGrid cols="2 300:2 600:4 900:6 1100:9" :x-gap="5" :y-gap="5">
          <NGridItem v-for=" item, index in imageList" :key="index">
            <NCard size="small" style="border-radius: var(--x-r-small);" :bordered="false">
              <template #cover>
                <div class="card transparent-grid">
                  <NImage :lazy="true" style="object-fit: contain;height: 100%;" :src="item.src" />
                </div>
              </template>
              <template #footer>
                <span class="text-[13px]">
                  <NEllipsis>
                    {{ item.fileName }}
                  </NEllipsis>
                </span>
                <div class="flex justify-center mt-3">
                  <NButtonGroup>
                    <NButton size="tiny" tertiary style="cursor: pointer;" :title="$t('apps.uploadsFileManager.copyLink')" :aria-label="$t('apps.uploadsFileManager.copyLink')" @click="copyImageUrl(item.src)">
                      <template #icon>
                        <SvgIcon icon="ion-copy" />
                      </template>
                    </NButton>
                    <NButton size="tiny" tertiary style="cursor: pointer;" :title="timeFormat(item.createTime)" :aria-label="$t('apps.uploadsFileManager.infoTitle')" @click="handleInfoClick(item)">
                      <template #icon>
                        <SvgIcon icon="mdi-information-box-outline" />
                      </template>
                    </NButton>
                    <NButton size="tiny" tertiary style="cursor: pointer;" :loading="wallpaperLoadingId === item.id" :title="$t('apps.uploadsFileManager.setWallpaper')" :aria-label="$t('apps.uploadsFileManager.setWallpaper')" @click="handleSetWallpaper(item.src, item.id as number)">
                      <template #icon>
                        <SvgIcon icon="lucide:wallpaper" />
                      </template>
                    </NButton>
                    <NButton size="tiny" type="error" style="cursor: pointer;" :loading="deleteLoadingId === item.id" :title="$t('common.delete')" :aria-label="$t('common.delete')" @click="handleDelete(item.id as number)">
                      <template #icon>
                        <SvgIcon icon="material-symbols-delete" />
                      </template>
                    </NButton>
                  </NButtonGroup>
                </div>
              </template>
            </NCard>
          </NGridItem>
        </NGrid>
      </NImageGroup>
    </div>

    <RoundCardModal v-model:show="infoModalState.show" style="max-width: 300px;" size="small" :title="$t('apps.uploadsFileManager.infoTitle')">
      <div>
        <div>
          <div class="mb-2">
            <span class="text-text2">
              {{ $t('apps.uploadsFileManager.fileName') }}
            </span>
            <div class="text-[13px]">
              {{ infoModalState.fileInfo?.fileName }}
            </div>
          </div>
          <div class="mb-2">
            <span class="text-text2">
              {{ $t('apps.uploadsFileManager.path') }}
            </span>
            <div class="text-[13px]">
              {{ infoModalState.fileInfo?.src }}
            </div>
          </div>
          <div class="mb-2">
            <span class="text-text2">
              {{ $t('apps.uploadsFileManager.uploadTime') }}
            </span>
            <div class="text-[13px]">
              {{ timeFormat(infoModalState.fileInfo?.createTime) }}
            </div>
          </div>
        </div>
      </div>
    </RoundCardModal>
  </div>
</template>

<style scoped>
.operation-error { margin: 8px 0; color: var(--x-red-fg); font-size: 13px; }
.card {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 80px;
}

.transparent-grid { background: var(--x-ph); }
</style>
