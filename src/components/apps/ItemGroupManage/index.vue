<script setup lang="ts">
import { onMounted, ref } from 'vue'
import type { FormInst, FormRules } from 'naive-ui'
import { NButton, NCard, NForm, NFormItem, NInput, useDialog, useMessage } from 'naive-ui'
import { VueDraggable } from 'vue-draggable-plus'
import { deletes, edit, getList, saveSort } from '@/api/panel/itemIconGroup'
import { getListByGroupId } from '@/api/panel/itemIcon'
import { GroupGlyph, RoundCardModal, SvgIcon } from '@/components/common'
import { groupIcons, isValidGroupIcon } from '@/components/common/GroupGlyph/icons'
import { t } from '@/locales'
import { confirmDelete } from '@/utils/confirmDelete'

interface EditModalArg {
  show: boolean
  editStatus: number // 1.添加 2.编辑
  model: Panel.ItemIconGroup
  rules: FormRules
}

const formRef = ref<FormInst | null>(null)
const ms = useMessage()
const dialog = useDialog()
const sortStatus = ref(false)
const sortSaving = ref(false)
const saving = ref(false)
const saveError = ref('')
const deletingId = ref<number | null>(null)
const listLoading = ref(false)
const listError = ref('')

const defaultMNodal = {
  title: '',
  icon: '',
  sort: 9999,
}

const editModalArg = ref<EditModalArg>({
  show: false,
  editStatus: 1,
  model: defaultMNodal,
  rules: {
    title: [
      {
        required: true,
        trigger: 'blur',
        message: t('form.required'),
      },
    ],
  },
})

const groups = ref<Panel.ItemIconGroup[]>([])

function handleAddGroup() {
  saveError.value = ''
  editModalArg.value.model = { ...defaultMNodal }
  editModalArg.value.editStatus = 1
  editModalArg.value.show = true
}

function handleEditGroup(groupInfo: Panel.ItemIconGroup) {
  saveError.value = ''
  editModalArg.value.show = true
  editModalArg.value.model = { ...groupInfo }
  editModalArg.value.editStatus = 2
}

function handleDragSort() {
  sortStatus.value = true
}

async function handleSaveSort() {
  const saveItems: Common.SortItemRequest[] = []
  for (let i = 0; i < groups.value.length; i++) {
    const element = groups.value[i]
    saveItems.push({
      id: element.id as number,
      sort: i + 1,
    })
  }
  sortSaving.value = true
  try {
    const { code, msg } = await saveSort(saveItems)
    if (code === 0) {
      ms.success(t('common.saveSuccess'))
      sortStatus.value = false
    }
    else {
      saveError.value = `${t('common.saveFail')}: ${msg}`
    }
  }
  catch {
    saveError.value = t('common.saveFail')
  }
  finally {
    sortSaving.value = false
  }
}

async function handleDelete(groupInfo: Panel.ItemIconGroup) {
  let detail = t('common.groupDeleteDetailUnknown')
  if (groupInfo.id) {
    try {
      const { code, data } = await getListByGroupId<Common.ListResponse<Panel.ItemInfo[]>>(groupInfo.id)
      if (code === 0)
        detail = t('common.groupDeleteDetail', { count: data.list.length })
    }
    catch {}
  }
  confirmDelete(dialog, {
    object: t('common.groupObject'),
    name: groupInfo.title || '',
    detail,
    onDelete: async () => {
      if (!groupInfo.id)
        return t('common.deleteFail')
      deletingId.value = groupInfo.id
      try {
        const { code, msg } = await deletes([groupInfo.id])
        if (code !== 0)
          return msg || t('common.deleteFail')
        await refreshList()
      }
      finally {
        deletingId.value = null
      }
    },
  })
}

function handleSaveGroup() {
  formRef.value?.validate(async (errors) => {
    if (!errors) {
      saving.value = true
      saveError.value = ''
      try {
        const { code, msg } = await edit(editModalArg.value.model)
        if (code !== 0) {
          saveError.value = msg || t('common.saveFail')
          return
        }
        refreshList()
        editModalArg.value.show = false
        editModalArg.value.model = { ...defaultMNodal }
      }
      catch {
        saveError.value = t('common.saveFail')
      }
      finally {
        saving.value = false
      }
    }
  })
}

async function refreshList() {
  listLoading.value = true
  listError.value = ''
  try {
    const { code, data, msg } = await getList<Common.ListResponse<Panel.ItemIconGroup[]>>()
    if (code !== 0)
      throw new Error(msg || t('common.failed'))
    groups.value = data.list
  }
  catch (error) {
    listError.value = error instanceof Error ? error.message : t('common.failed')
  }
  finally {
    listLoading.value = false
  }
}

onMounted(() => {
  refreshList()
})
</script>

<template>
  <div class="h-full p-2">
    <div class="panel-actions">
      <NButton type="primary" @click="handleAddGroup">
        <template #icon>
          <SvgIcon icon="lucide-plus" />
        </template>{{ $t('common.add') }}
      </NButton>
      <NButton v-if="!sortStatus" secondary @click="handleDragSort">
        <template #icon>
          <SvgIcon icon="lucide-arrow-down-up" />
        </template>{{ $t('common.sort') }}
      </NButton>
      <NButton v-else secondary :loading="sortSaving" @click="handleSaveSort">
        <template #icon>
          <SvgIcon icon="lucide-arrow-down-up" />
        </template>{{ $t('common.saveSort') }}
      </NButton>
    </div>

    <p v-if="saveError && !editModalArg.show" class="form-error">
      {{ saveError }}
    </p>
    <p v-if="listError" class="form-error" role="alert">
      {{ listError }} <NButton size="small" :loading="listLoading" @click="refreshList">
        {{ t('panelHome.retry') }}
      </NButton>
    </p>
    <div class="overflow-auto w-full rounded-small">
      <VueDraggable
        v-model="groups"
        item-key="sort" :animation="300"
        :style="{ padding: sortStatus ? '20px' : '10px' }"
        :disabled="!sortStatus"
      >
        <div v-for="(item, index) in groups" :key="index" class="w-full">
          <NCard size="small" class="mb-3 rounded-small">
            <div class="flex" :class="sortStatus ? 'cursor-move' : ''">
              <div class="flex items-center">
                <span class="group-list-glyph mr-3" :style="{ '--group-tint': `var(--x-${['blue', 'indigo', 'purple', 'pink', 'orange', 'mint', 'brown', 'teal', 'green', 'red'][index % 10]}-tint)`, '--group-fg': `var(--x-${['blue', 'indigo', 'purple', 'pink', 'orange', 'mint', 'brown', 'teal', 'green', 'red'][index % 10]}-fg)` }">
                  <GroupGlyph :icon="item.icon" :title="item.title" :size="18" />
                </span>
                <span>
                  {{ item.title }}
                </span>
              </div>
              <div class="group-row-actions ml-auto">
                <NButton quaternary class="group-icon-action" :aria-label="t('apps.itemGroupManage.editGroup')" @click="handleEditGroup(item)">
                  <template #icon>
                    <SvgIcon icon="lucide-pencil" />
                  </template>
                </NButton>
                <NButton class="group-icon-action danger-button" :loading="deletingId === item.id" :aria-label="t('apps.itemGroupManage.deleteGroup')" @click="handleDelete(item)">
                  <template #icon>
                    <SvgIcon icon="lucide-trash-2" />
                  </template>
                </NButton>
              </div>
            </div>
          </NCard>
        </div>
      </VueDraggable>
    </div>

    <RoundCardModal v-model:show="editModalArg.show" size="small" type="small" :title="editModalArg.editStatus === 1 ? '添加' : '编辑'" style="width: 400px;">
      <NForm ref="formRef" :model="editModalArg.model" :rules="editModalArg.rules">
        <NFormItem path="title" :label="$t('apps.itemGroupManage.groupName')" :feedback="saveError" :validation-status="saveError ? 'error' : undefined">
          <NInput v-model:value="editModalArg.model.title" type="text" :maxlength="20" show-count :input-props="{ 'aria-label': $t('apps.itemGroupManage.groupName') }" />
        </NFormItem>

        <NFormItem :label="$t('apps.itemGroupManage.icon')">
          <div class="group-icon-grid" role="group" :aria-label="$t('apps.itemGroupManage.icon')">
            <button class="group-icon-option" :class="{ active: !isValidGroupIcon(editModalArg.model.icon) }" type="button" :title="t('apps.itemGroupManage.auto')" :aria-label="t('apps.itemGroupManage.auto')" :aria-pressed="!isValidGroupIcon(editModalArg.model.icon)" @click="editModalArg.model.icon = ''">
              <GroupGlyph :title="editModalArg.model.title" :size="16" /><span class="auto-label">{{ t('apps.itemGroupManage.auto') }}</span>
            </button>
            <button v-for="icon in groupIcons" :key="icon" class="group-icon-option" :class="{ active: editModalArg.model.icon === icon }" type="button" :title="icon.replace('lucide-', '')" :aria-label="icon.replace('lucide-', '')" :aria-pressed="editModalArg.model.icon === icon" @click="editModalArg.model.icon = icon">
              <GroupGlyph :icon="icon" :title="editModalArg.model.title" />
            </button>
          </div>
        </NFormItem>
      </NForm>
      <p v-if="saveError" class="form-error" role="alert">
        {{ saveError }}
      </p>
      <template #footer>
        <NButton type="primary" size="small" class="float-right" :loading="saving" @click="handleSaveGroup">
          {{ $t('common.confirm') }}
        </NButton>
      </template>
    </RoundCardModal>
  </div>
</template>

<style scoped>
.group-list-glyph { display: grid; place-items: center; width: 44px; height: 44px; border-radius: var(--x-r-icon); background: var(--group-tint); color: var(--group-fg); }
.auto-label { font-size: 12px; line-height: 1; }
.group-icon-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(44px, 1fr)); gap: 6px; width: 100%; }
.group-icon-option { display: grid; place-items: center; width: 44px; height: 44px; border: 0; border-radius: var(--x-r-icon); background: transparent; color: var(--x-text2); cursor: pointer; }
.group-icon-option:hover { background: var(--x-bg1); color: var(--x-text1); }
.group-icon-option:first-child { align-content: center; gap: 1px; }
.group-icon-option.active { background: var(--x-blue-tint); color: var(--x-blue-fg); }
.form-error { padding: 8px 12px; color: var(--x-red-fg); font-size: 13px; }
.group-row-actions { display: flex; align-items: center; gap: 8px; }
.group-icon-action { width: 36px; height: 36px; min-width: 36px; min-height: 36px; border-radius: var(--x-r-small); font-size: 18px; }
.group-icon-action::before { position: absolute; inset: -4px; content: ""; }
.danger-button { background: var(--x-red-tint); color: var(--x-red-fg); }
.group-icon-option:focus-visible { outline: 2px solid var(--x-primary); outline-offset: 2px; }
</style>
