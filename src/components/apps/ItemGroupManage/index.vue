<script setup lang="ts">
import { onMounted, ref } from 'vue'
import type { FormInst, FormRules } from 'naive-ui'
import { NButton, NCard, NForm, NFormItem, NInput, useDialog, useMessage } from 'naive-ui'
import { VueDraggable } from 'vue-draggable-plus'
import { deletes, edit, getList, saveSort } from '@/api/panel/itemIconGroup'
import { GroupGlyph, RoundCardModal, SvgIcon } from '@/components/common'
import { groupIcons, isValidGroupIcon } from '@/components/common/GroupGlyph/icons'
import { t } from '@/locales'

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

function handleDelete(groupInfo: Panel.ItemIconGroup) {
  dialog.warning({
    title: t('common.warning'),
    content: t('apps.itemGroupManage.deleteWarnText', { name: groupInfo.title }),
    positiveText: t('common.confirm'),
    negativeText: t('common.cancel'),
    onPositiveClick: async () => {
      if (groupInfo.id) {
        deletingId.value = groupInfo.id
        try {
          const { code, msg } = await deletes([groupInfo.id])
          if (code !== 0)
            saveError.value = msg || t('common.deleteFail')
          else
            refreshList()
        }
        catch {
          saveError.value = t('common.deleteFail')
        }
        finally {
          deletingId.value = null
        }
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
    else { console.log(errors) }
  })
}

function refreshList() {
  getList<Common.ListResponse<Panel.ItemIconGroup[]>>().then(({ data }) => {
    groups.value = data.list
  })
}

onMounted(() => {
  refreshList()
})
</script>

<template>
  <div class="h-full">
    <div class="p-2">
      <NButton type="primary" size="small" class="mr-3" @click="handleAddGroup">
        {{ $t('common.add') }}
      </NButton>

      <NButton v-if="!sortStatus" size="small" @click="handleDragSort">
        {{ $t('common.sort') }}
      </NButton>

      <NButton v-else type="primary" size="small" :loading="sortSaving" @click="handleSaveSort">
        {{ $t('common.saveSort') }}
      </NButton>
    </div>

    <p v-if="saveError && !editModalArg.show" class="form-error">
      {{ saveError }}
    </p>
    <div class="overflow-auto w-full mt-5 bg-bg1 rounded-small" style="height:calc(100% - 65px)">
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
              <div class="ml-auto">
                <span>
                  <NButton secondary size="small" @click="handleEditGroup(item)">
                    <template #icon>
                      <SvgIcon icon="basil:edit-solid" />
                    </template>
                  </NButton>
                </span>
                <span class="ml-3">
                  <NButton type="error" size="small" class="ml-3 danger-button" :loading="deletingId === item.id" @click="handleDelete(item)">
                    <template #icon>
                      <SvgIcon icon="material-symbols:delete" />
                    </template>
                  </NButton>
                </span>
              </div>
            </div>
          </NCard>
        </div>
      </VueDraggable>
    </div>

    <RoundCardModal v-model:show="editModalArg.show" size="small" type="small" :title="editModalArg.editStatus === 1 ? '添加' : '编辑'" style="width: 400px;">
      <NForm ref="formRef" :model="editModalArg.model" :rules="editModalArg.rules">
        <NFormItem path="title" :label="$t('apps.itemGroupManage.groupName')" :feedback="saveError" :validation-status="saveError ? 'error' : undefined">
          <NInput v-model:value="editModalArg.model.title" type="text" :maxlength="20" show-count />
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
.auto-label { font-size: 9px; line-height: 1; }
.group-icon-grid { display: grid; grid-template-columns: repeat(8, 36px); gap: 6px; max-width: 100%; }
.group-icon-option { display: grid; place-items: center; width: 36px; height: 36px; border: 0; border-radius: 8px; background: transparent; color: var(--x-text2); cursor: pointer; }
.group-icon-option:hover { background: var(--x-bg1); color: var(--x-text1); }
.group-icon-option:first-child { align-content: center; gap: 1px; }
.group-icon-option.active { background: var(--x-blue-tint); color: var(--x-blue-fg); }
.form-error { padding: 8px 12px; color: var(--x-red-fg); font-size: 13px; }
.danger-button { background: var(--x-red-tint); color: var(--x-red-fg); }
.group-icon-option:focus-visible { outline: 2px solid var(--x-primary); outline-offset: 2px; }
</style>
