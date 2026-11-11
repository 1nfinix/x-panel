<script setup lang="ts">
import { NAlert, NButton, NCard, NGrid, NGridItem, NImage, NImageGroup, NSpin, useDialog, useMessage } from 'naive-ui'
import { onMounted, ref } from 'vue'
import { deletes, getList } from '@/api/system/file'
import { set as savePanelConfig } from '@/api/panel/userConfig'
import { RoundCardModal, SvgIcon } from '@/components/common'
import { copyToClipboard, timeFormat } from '@/utils/cmn'
import { t } from '@/locales'
import { confirmDelete } from '@/utils/confirmDelete'
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
  confirmDelete(dialog, {
    object: t('common.fileObject'),
    name: imageList.value.find(item => item.id === id)?.fileName || '',
    onDelete: async () => {
      deleteLoadingId.value = id
      try {
        const { code, msg } = await deletes([id])
        if (code !== 0)
          return `${t('common.deleteFail')}: ${msg}`
        await getFileList()
        ms.success(t('common.success'))
      }
      finally {
        deleteLoadingId.value = null
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
    <NAlert class="panel-alert" type="info" :bordered="false">
      <template #icon>
        <SvgIcon icon="lucide-info" />
      </template>
      {{ $t('apps.uploadsFileManager.alertText') }}
    </NAlert>
    <div class="upload-grid-wrap">
      <div v-if="imageList.length === 0 && !loading" class="flex">
        {{ $t('apps.uploadsFileManager.nothingText') }}
      </div>
      <NImageGroup v-else>
        <NGrid cols="1 420:2 600:3" :x-gap="12" :y-gap="12">
          <NGridItem v-for=" item, index in imageList" :key="index">
            <NCard size="small" style="border-radius: var(--x-r-small);" :bordered="false">
              <template #cover>
                <div class="file-thumbnail">
                  <NImage :alt="item.fileName ?? ''" :lazy="true" :src="item.src" />
                </div>
              </template>
              <template #footer>
                <div class="file-name" :title="item.fileName">
                  {{ item.fileName }}
                </div>
                <div class="file-actions">
                  <NButton quaternary class="file-action" :title="$t('apps.uploadsFileManager.copyLink')" :aria-label="$t('apps.uploadsFileManager.copyLink')" @click="copyImageUrl(item.src)">
                    <template #icon>
                      <SvgIcon icon="lucide-copy" />
                    </template>
                  </NButton>
                  <NButton quaternary class="file-action" :title="timeFormat(item.createTime)" :aria-label="$t('apps.uploadsFileManager.infoTitle')" @click="handleInfoClick(item)">
                    <template #icon>
                      <SvgIcon icon="lucide-info" />
                    </template>
                  </NButton>
                  <NButton quaternary class="file-action" :loading="wallpaperLoadingId === item.id" :title="$t('apps.uploadsFileManager.setWallpaper')" :aria-label="$t('apps.uploadsFileManager.setWallpaper')" @click="handleSetWallpaper(item.src, item.id as number)">
                    <template #icon>
                      <SvgIcon icon="lucide-image" />
                    </template>
                  </NButton>
                  <NButton class="file-action file-delete" :loading="deleteLoadingId === item.id" :title="$t('common.delete')" :aria-label="$t('common.delete')" @click="handleDelete(item.id as number)">
                    <template #icon>
                      <SvgIcon icon="lucide-trash-2" />
                    </template>
                  </NButton>
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
.upload-grid-wrap { margin-top: 16px; }
.file-thumbnail { display: block; height: 96px; overflow: hidden; border-radius: var(--x-r-small); background: var(--x-bg2); }
.file-thumbnail :deep(.n-image) { display: block; width: 100%; height: 100% !important; }
.file-thumbnail :deep(img) { display: block; width: 100% !important; height: 100% !important; object-fit: contain !important; }
.file-name { min-width: 0; overflow: hidden; color: var(--x-text1); font-size: 13px; text-overflow: ellipsis; white-space: nowrap; }
.file-actions { display: flex; justify-content: flex-start; gap: 4px; margin-top: 12px; }
.file-action { width: 32px; min-width: 32px; height: 32px; min-height: 32px; padding: 0; border-radius: var(--x-r-glyph); font-size: 18px; }
.file-action:hover { background: var(--x-hover); }
.file-delete { background: var(--x-red-tint); color: var(--x-red-fg); }
.file-delete:hover { background: var(--x-danger-hover); }
</style>
