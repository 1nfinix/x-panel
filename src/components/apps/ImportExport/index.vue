<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import type { UploadFileInfo } from 'naive-ui'
import { NAlert, NButton, NCard, NCheckbox, NCheckboxGroup, NDivider, NInput, NProgress, NSpace, NUpload, useMessage } from 'naive-ui'
import { RoundCardModal, SvgIcon } from '@/components/common'
import type { IconGroup, ImportJsonResult } from '@/utils/jsonImportExport'
import { ConfigVersionLowError, FormatError, exportJson, importJsonString } from '@/utils/jsonImportExport'
import type { BookmarkGroup } from '@/utils/bookmarkImport'
import { normalizeUrl, parseBookmarksHtml } from '@/utils/bookmarkImport'
import { get as getAbout } from '@/api/system/about'
import { edit as addGroup, getList as getGroupList } from '@/api/panel/itemIconGroup'
import { addMultiple as addMultipleIcons, getListByGroupId } from '@/api/panel/itemIcon'

import { t } from '@/locales'
import { useFaviconFill } from '@/hooks/useFaviconFill'
import { groupTints } from '@/utils/groupColors'

interface ItemGroup extends Panel.ItemIconGroup {
  items?: Panel.ItemInfo[]
}

interface BookmarkPreviewGroup extends BookmarkGroup {
  selected: boolean
  existingGroupId?: number
  skippedDuplicate: number
}

const ms = useMessage()
const bookmarkPreview = ref<BookmarkPreviewGroup[]>([])
const groupTintIndexById = new Map<number, number>()
const bookmarkSkippedInvalid = ref(0)
const bookmarkLoading = ref(false)
const bookmarkImporting = ref(false)
const bookmarkError = ref('')
const pendingFaviconIds = ref<number[]>([])
const bookmarkFill = useFaviconFill()
const bookmarkSelected = computed(() => bookmarkPreview.value.filter(group => group.selected && group.items.length))
const bookmarkSelectedCount = computed(() => bookmarkSelected.value.reduce((count, group) => count + group.items.length, 0))
const bookmarkAllSelected = computed(() => bookmarkPreview.value.length > 0 && bookmarkPreview.value.every(group => group.selected))
const bookmarkFillPercentage = computed(() => bookmarkFill.total.value ? Math.round(bookmarkFill.done.value / bookmarkFill.total.value * 100) : 0)

function toggleAllBookmarks(checked: boolean) {
  for (const group of bookmarkPreview.value)
    group.selected = checked
}

async function handleBookmarkFileChange(options: { file: UploadFileInfo }) {
  if (!options.file.file || bookmarkLoading.value || bookmarkImporting.value)
    return

  bookmarkLoading.value = true
  bookmarkError.value = ''
  bookmarkPreview.value = []
  pendingFaviconIds.value = []
  bookmarkFill.reset()
  let step = t('apps.exportImport.bookmarkReadFailed')
  try {
    const parsed = parseBookmarksHtml(await options.file.file.text())
    bookmarkSkippedInvalid.value = parsed.skippedInvalid
    step = t('apps.exportImport.bookmarkLoadGroupsFailed')
    const groupResponse = await getGroupList<Common.ListResponse<ItemGroup[]>>()
    if (groupResponse.code !== 0)
      throw new Error(groupResponse.msg)

    const existingGroups = new Map<string, number>()
    groupTintIndexById.clear()
    const existingUrls = new Set<string>()
    step = t('apps.exportImport.bookmarkLoadItemsFailed')
    await Promise.all(groupResponse.data.list.map(async (group) => {
      if (group.id)
        groupTintIndexById.set(group.id, groupTintIndexById.size)
      if (group.id && !existingGroups.has(group.title || ''))
        existingGroups.set(group.title || '', group.id)
      const response = await getListByGroupId<Common.ListResponse<Panel.ItemInfo[]>>(group.id)
      if (response.code !== 0)
        throw new Error(`${group.title}: ${response.msg}`)
      for (const item of response.data.list)
        existingUrls.add(normalizeUrl(item.url))
    }))

    const seenUrls = new Set(existingUrls)
    bookmarkPreview.value = parsed.groups.map((group) => {
      const items: BookmarkGroup['items'] = []
      let skippedDuplicate = 0
      for (const item of group.items) {
        const key = normalizeUrl(item.url)
        if (seenUrls.has(key)) {
          skippedDuplicate++
          continue
        }
        seenUrls.add(key)
        items.push(item)
      }
      return { title: group.title, items, selected: true, existingGroupId: existingGroups.get(group.title), skippedDuplicate }
    })
  }
  catch (error) {
    bookmarkError.value = `${step}: ${error instanceof Error ? error.message : t('common.unknownError')}`
  }
  finally {
    bookmarkLoading.value = false
  }
}

async function handleBookmarkImport() {
  if (bookmarkImporting.value || !bookmarkSelectedCount.value)
    return

  bookmarkImporting.value = true
  bookmarkError.value = ''
  let importedCount = 0
  let step = ''
  try {
    for (const group of bookmarkSelected.value) {
      if (!group.existingGroupId) {
        step = `${t('apps.exportImport.bookmarkCreateGroupFailed')} ${group.title}`
        const response = await addGroup<Panel.ItemIconGroup>({ title: group.title })
        if (response.code !== 0 || !response.data?.id)
          throw new Error(response.msg)
        group.existingGroupId = response.data.id
        groupTintIndexById.set(response.data.id, groupTintIndexById.size)
      }

      while (group.items.length) {
        step = `${t('apps.exportImport.bookmarkAddFailed')} ${group.title}`
        const batch = group.items.slice(0, 50)
        const payload: Panel.ItemInfo[] = batch.map(item => ({
          title: item.title,
          url: item.url,
          itemIconGroupId: group.existingGroupId,
          openMethod: 2,
          icon: { itemType: 1, text: Array.from(item.title)[0], backgroundColor: groupTints[(groupTintIndexById.get(group.existingGroupId as number) ?? 0) % groupTints.length] },
        }))
        const response = await addMultipleIcons<Panel.ItemInfo[]>(payload)
        if (response.code !== 0 || response.data?.length !== batch.length)
          throw new Error(response.msg)
        pendingFaviconIds.value.push(...response.data.map(item => item.id as number))
        importedCount += batch.length
        group.items.splice(0, batch.length)
      }
      group.selected = false
    }

    ms.success(t('apps.exportImport.bookmarkImportSuccess', { count: importedCount }))
    await bookmarkFill.run(pendingFaviconIds.value)
    pendingFaviconIds.value = []
    if (bookmarkFill.failures.value.length)
      ms.error(t('apps.exportImport.bookmarkFillFailed', { count: bookmarkFill.failures.value.length }))
    ms.info(t('apps.exportImport.bookmarkRefreshHint'))
  }
  catch (error) {
    bookmarkError.value = t('apps.exportImport.bookmarkImportFailed', {
      step: `${step}: ${error instanceof Error ? error.message : t('common.unknownError')}`,
      count: importedCount,
    })
  }
  finally {
    bookmarkImporting.value = false
  }
}

const jsonData = ref<string | null>(null)
const importWarning = ref<string[]>([])
const importRoundModalShow = ref(false)
const exportRoundModalShow = ref(false)
const loading = ref(false)
const uploadLoading = ref(false)
const ioError = ref('')
const version = ref('') // 当前软件版本
const versionLoading = ref(false)
const versionError = ref('')
const debug = ref(false)

const importObj = ref<ImportJsonResult | null> (null)

const importItems = ref<string[]>(['icons']) // 当前软件版本支持导入导出的项目
const checkedItems = ref<string[]>(['icons']) // 当前准备导入的项目

// 导入图标
async function importIcons(): Promise<string | null> {
  const groups = importObj.value?.geticons()
  const batchSize = 50

  if (!groups)
    return null

  try {
    for (let i = 0; i < groups.length; i++) {
      const element = groups[i]

      // 创建组得到组id
      const createGroupResponse = await addGroup<Panel.ItemIconGroup>({
        title: element.title,
        sort: element.sort,
      })

      if (createGroupResponse.code === 0) {
        const groupId = createGroupResponse.data?.id

        if (groupId) {
          let addIcons: Panel.ItemInfo[] = []

          // 批量添加子项
          for (let iconI = 0; iconI < element.children.length; iconI++) {
            const iconElement = element.children[iconI]

            addIcons.push({
              title: iconElement.title,
              sort: iconElement.sort,
              icon: iconElement.icon,
              url: iconElement.url,
              lanUrl: iconElement.lanUrl,
              description: iconElement.description,
              openMethod: iconElement.openMethod,
              itemIconGroupId: groupId,
            })

            // 每 batchSize 个添加一次
            if (addIcons.length === batchSize || iconI === element.children.length - 1) {
              const response = await addMultipleIcons(addIcons)

              if (response.code !== 0)
                return response.msg

              addIcons = []
            }
          }
        }
        else {
          return t('common.failed')
        }
      }
      else {
        return createGroupResponse.msg
      }
    }

    return null
  }
  catch (error) {
    if (error instanceof Error)
      return `${t('common.failed')}: ${error.message}`
    else
      return t('common.unknownError')
  }
}

// 导出图标
async function exportIcons(): Promise<IconGroup[]> {
  const iconGroups: IconGroup[] = []

  // 获取组数据
  const { code, data, msg } = await getGroupList<Common.ListResponse<ItemGroup[]>>()

  if (code === 0) {
    // 使用 Promise.all 等待所有异步操作完成
    await Promise.all(data.list.map(async (element) => {
      const group: IconGroup = {
        title: element.title as string,
        sort: element.sort as 0,
        children: [],
      }

      const res = await getListByGroupId<Common.ListResponse<Panel.ItemInfo[]>>(element.id)

      if (res.code !== 0)
        throw new Error(res.msg || t('common.failed'))
      for (const iconElement of res.data.list) {
        group.children.push({
          icon: iconElement.icon,
          sort: iconElement.sort || 99999,
          title: iconElement.title,
          url: iconElement.url,
          lanUrl: iconElement.lanUrl || '',
          description: iconElement.description || '',
          openMethod: iconElement.openMethod || 1,
        })
      }

      iconGroups.push(group)
    }))

    return iconGroups
  }
  throw new Error(msg || t('common.failed'))
}

async function loadVersion() {
  versionLoading.value = true
  versionError.value = ''
  interface Version {
    versionName: string
    versionCode: number
  }

  try {
    const res = await getAbout<Version>()
    if (res.code === 0)
      version.value = res.data.versionName
    else
      throw new Error(res.msg || t('common.failed'))
  }
  catch (error) {
    versionError.value = error instanceof Error ? error.message : t('common.failed')
  }
  finally {
    versionLoading.value = false
  }
}

onMounted(() => {
  loadVersion()
})

function handleFileChange(options: { file: UploadFileInfo; fileList: Array<UploadFileInfo> }) {
  if (!options.file.file)
    return
  uploadLoading.value = true
  ioError.value = ''
  const reader = new FileReader()
  reader.onload = () => {
    if (reader.result) {
      jsonData.value = reader.result as string
      importCheck()
    }
    else {
      ioError.value = `${t('common.failed')}: ${t('common.repeatLater')}`
    }
    uploadLoading.value = false
  }
  reader.onerror = () => {
    ioError.value = `${t('common.failed')}: ${t('common.repeatLater')}`
    uploadLoading.value = false
  }
  reader.readAsText(options.file.file)
}

// 验证导入文件
function importCheck() {
  importWarning.value = []
  ioError.value = ''
  if (jsonData.value) {
    try {
      importObj.value = importJsonString(jsonData.value)
      if (importObj.value) {
        if (!importObj.value.isPassCheckMd5())
          importWarning.value.push(t('apps.exportImport.fileModified'))

        if (!importObj.value.isPassCheckConfigVersionOld())
          importWarning.value.push(t('apps.exportImport.warnConfigFileLow'))

        if (!importObj.value.isPassCheckConfigVersionNew())
          importWarning.value.push(t('apps.exportImport.softwareVersionLow'))

        // （暂时不做）此处可以判断，当前的配置文件是否存在的导入项目（不存在隐藏importItems里面的值）操作变量：importItems

        // 通过了验证,打开弹窗
        importRoundModalShow.value = true

        // console.log(importObj.value.geticons())
      }
    }
    catch (error) {
      if (error instanceof ConfigVersionLowError)
        ioError.value = t('apps.exportImport.errorConfigFileLow')

      else if (error instanceof FormatError)
        ioError.value = t('apps.exportImport.errorConfigFileFormat')

      else
        ioError.value = t('apps.exportImport.errorConfigFileFormat')
    }
  }
  else {
    ioError.value = t('apps.exportImport.errorConfigFileFormat')
  }
}

// 开始导出
async function handleStartExport() {
  if (loading.value)
    return
  loading.value = true
  ioError.value = ''
  try {
    if (!version.value)
      throw new Error(versionError.value || t('common.failed'))
    const exportResult = exportJson(version.value)
    if (checkedItems.value.includes('icons'))
      exportResult.addIconsData(await exportIcons())
    jsonData.value = exportResult.string()
    exportResult.exportFile()
    exportRoundModalShow.value = false
    ms.success(t('common.success'))
  }
  catch (error) {
    ioError.value = error instanceof Error ? error.message : t('common.failed')
  }
  finally {
    loading.value = false
  }
}

// 开始导入
async function handleStartImport() {
  if (loading.value)
    return
  loading.value = true
  ioError.value = ''
  try {
    if (checkedItems.value.includes('icons')) {
      const errMsg = await importIcons()
      if (errMsg !== null) {
        ioError.value = `${t('common.failed')}: ${errMsg}`
        return
      }
    }
    importRoundModalShow.value = false
    ms.success(`${t('common.success')}, ${t('common.refreshPage')}`)
  }
  catch (error) {
    ioError.value = error instanceof Error ? error.message : t('common.failed')
  }
  finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="import-export-panel">
    <NCard size="small" class="import-export-card">
      <h2>{{ $t('apps.exportImport.bookmarkTitle') }}</h2>
      <p class="card-description">
        {{ $t('apps.exportImport.bookmarkDescription') }}
      </p>
      <NUpload accept=".html,.htm" :default-upload="false" :show-file-list="false" :disabled="bookmarkImporting" @change="handleBookmarkFileChange">
        <NButton secondary :loading="bookmarkLoading" :disabled="bookmarkImporting">
          <template #icon>
            <SvgIcon icon="lucide-upload" />
          </template>
          {{ $t('apps.exportImport.bookmarkChooseFile') }}
        </NButton>
      </NUpload>
      <p v-if="bookmarkError" class="async-error" role="alert">
        {{ bookmarkError }}
      </p>

      <div v-if="bookmarkPreview.length" class="mt-4">
        <p class="mb-2">
          {{ $t('apps.exportImport.bookmarkTotal', { groups: bookmarkPreview.length, count: bookmarkSelectedCount }) }}
          · {{ $t('apps.exportImport.bookmarkInvalid', { count: bookmarkSkippedInvalid }) }}
        </p>
        <NCheckbox :checked="bookmarkAllSelected" :indeterminate="!bookmarkAllSelected && bookmarkPreview.some(group => group.selected)" :disabled="bookmarkImporting" @update:checked="toggleAllBookmarks">
          {{ $t('apps.exportImport.bookmarkSelectAll') }}
        </NCheckbox>
        <div v-for="group in bookmarkPreview" :key="group.title" class="flex items-start gap-3 py-2 border-b border-sep">
          <NCheckbox v-model:checked="group.selected" :aria-label="group.title" :disabled="bookmarkImporting || !group.items.length" />
          <div class="min-w-0">
            <div class="font-medium">
              {{ group.title }}
              <span class="ml-2 text-sm text-text2">{{ group.existingGroupId ? $t('apps.exportImport.bookmarkMergeGroup') : $t('apps.exportImport.bookmarkNewGroup') }}</span>
            </div>
            <div class="text-sm text-text2">
              {{ $t('apps.exportImport.bookmarkGroupCount', { count: group.items.length, skipped: group.skippedDuplicate }) }}
            </div>
          </div>
        </div>
        <NButton class="mt-4" type="primary" :loading="bookmarkImporting" :disabled="!bookmarkSelectedCount" @click="handleBookmarkImport">
          {{ $t('apps.exportImport.bookmarkImport') }}
        </NButton>
      </div>
      <div v-if="bookmarkFill.total.value" class="mt-4">
        <p class="mb-2">
          {{ $t('apps.exportImport.bookmarkFilling') }} · {{ $t('apps.exportImport.bookmarkFillProgress', { success: bookmarkFill.success.value, failed: bookmarkFill.failures.value.length }) }}
        </p>
        <NProgress type="line" :percentage="bookmarkFillPercentage" />
        <div v-for="failure in bookmarkFill.failures.value" :key="failure.id" class="mt-1 text-sm">
          {{ failure.id }}: {{ failure.reason }}
        </div>
      </div>
    </NCard>
    <NCard size="small" class="import-export-card">
      <h2>{{ $t('apps.exportImport.configBackup') }}</h2>
      <p class="card-description">
        {{ $t('apps.exportImport.tip') }}
      </p>
      <p v-if="versionError" class="async-error" role="alert">
        {{ versionError }} <NButton size="small" :loading="versionLoading" @click="loadVersion">
          {{ t('panelHome.retry') }}
        </NButton>
      </p>
      <div class="card-actions">
        <NUpload
          accept=".sun-panel.json,.sunpanel.json"
          directory-dnd
          :default-upload="false"
          :show-file-list="false"
          @change="handleFileChange"
        >
          <NButton type="primary" :loading="uploadLoading">
            <template #icon>
              <SvgIcon icon="lucide-file-input" />
            </template>
            {{ $t('apps.exportImport.import') }}
          </NButton>
        </NUpload>
        <NButton secondary @click="exportRoundModalShow = !exportRoundModalShow">
          <template #icon>
            <SvgIcon icon="lucide-file-output" />
          </template>
          {{ $t('apps.exportImport.export') }}
        </NButton>
      </div>
      <p v-if="ioError && !importRoundModalShow && !exportRoundModalShow" class="async-error" role="alert">
        {{ ioError }}
      </p>
      <a class="converter-link" href="https://hslr-s.github.io/sun-panel-tool-page/#/" target="_blank" rel="noopener noreferrer">{{ $t('apps.exportImport.transmuteStandard') }}</a>
    </NCard>

    <!-- 调试模式 -->
    <div v-if="debug">
      <NButton @click="importCheck">
        检查导入
      </NButton>

      <!-- <NButton @click="exportJsonS">
      导出JSON
    </NButton> -->

      <NButton @click="jsonData = ''">
        清空导入数据
      </NButton>

      <NInput
        v-model:value="jsonData" :input-props="{ 'aria-label': $t('apps.exportImport.import') }"
        type="textarea"
        placeholder="基本的 Textarea"
      />

      <div v-if="jsonData">
        <h2>JSON 数据</h2>
        <pre>{{ jsonData }}</pre>
      </div>
    </div>

    <RoundCardModal v-model:show="importRoundModalShow" style="max-width: 400px;" :title=" $t('apps.exportImport.import')">
      <p v-if="ioError" class="async-error" role="alert">
        {{ ioError }}
      </p>
      <div v-if="importWarning.length > 0">
        <NAlert :title="$t('common.warning')" type="warning">
          <template #icon>
            <SvgIcon icon="lucide-triangle-alert" />
          </template>
          <div v-for="(text, index) in importWarning " :key="index">
            {{ text }}
          </div>
        </NAlert>
      </div>
      <NDivider title-placement="left">
        {{ $t('apps.exportImport.selectImportData') }}
      </NDivider>

      <NSpace justify="center" style="margin-top: 20px;">
        <NCheckboxGroup v-model:value="checkedItems">
          <NCheckbox v-if="importItems.includes('icons')" value="icons" :label="$t('apps.exportImport.moduleIcon')" />
          <NCheckbox v-if="importItems.includes('style')" value="style" :label="$t('apps.exportImport.moduleStyle')" />
        </NCheckboxGroup>
      </NSpace>
      <NSpace justify="center">
        <div class="mt-12">
          <NButton type="primary" :disabled="checkedItems.length === 0" :loading="loading" @click="handleStartImport">
            {{ $t('common.continue') }}
          </NButton>
        </div>
      </NSpace>
    </RoundCardModal>

    <RoundCardModal v-model:show="exportRoundModalShow" style="max-width: 400px;" :title=" $t('apps.exportImport.export')">
      <p v-if="ioError" class="async-error" role="alert">
        {{ ioError }}
      </p>
      <NDivider title-placement="left">
        {{ $t('apps.exportImport.selectExportData') }}
      </NDivider>

      <NSpace justify="center" style="margin-top: 20px;">
        <NCheckboxGroup v-model:value="checkedItems">
          <NCheckbox v-if="importItems.includes('icons')" value="icons" :label="$t('apps.exportImport.moduleIcon')" />
          <NCheckbox v-if="importItems.includes('style')" value="style" :label="$t('apps.exportImport.moduleStyle')" />
        </NCheckboxGroup>
      </NSpace>
      <NSpace justify="center">
        <div class="mt-12">
          <NButton type="primary" :disabled="checkedItems.length === 0" :loading="loading" @click="handleStartExport">
            {{ $t('common.continue') }}
          </NButton>
        </div>
      </NSpace>
    </RoundCardModal>
  </div>
</template>

<style scoped>
.async-error { margin-top: 12px; color: var(--x-red-fg); font-size: 13px; }
.import-export-panel { display: grid; align-content: start; gap: 16px; padding: 8px; }
.import-export-card :deep(.n-card__content) { padding: 20px; }
.import-export-card h2 { margin: 0 0 8px; color: var(--x-text1); font-size: 17px; font-weight: 600; }
.card-description { margin: 0 0 16px; color: var(--x-text2); font-size: 13px; }
.card-actions { display: flex; align-items: center; flex-wrap: wrap; gap: 8px; }
.card-actions :deep(.n-upload) { display: inline-flex; width: auto; }
.converter-link { display: inline-block; margin-top: 16px; color: var(--x-link); font-size: 13px; }
</style>
