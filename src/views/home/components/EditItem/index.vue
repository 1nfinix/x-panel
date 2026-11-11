<script setup lang="ts">
import { computed, defineEmits, defineProps, ref, watch } from 'vue'
import type { FormInst, FormRules } from 'naive-ui'
import { NButton, NForm, NFormItem, NGrid, NGridItem, NInput, NModal, NSelect, useMessage } from 'naive-ui'
import IconEditor from './IconEditor.vue'
import { edit, getSiteFavicon } from '@/api/panel/itemIcon'
import { getList as getGroupList } from '@/api/panel/itemIconGroup'
import { t } from '@/locales'
import { groupTints } from '@/utils/groupColors'
import { SvgIcon } from '@/components/common'

interface Props {
  visible: boolean
  itemInfo: Panel.Info | null
  itemGroupId?: number
}

const props = defineProps<Props>()
const emit = defineEmits<Emit>()
const ms = useMessage()
const submitLoading = ref(false)
const saveError = ref('')
const getIconLoading = ref([false, false])
const itemIconGroupOptions = ref<{
  label: string
  value: number
}[]>([])

const restoreDefault: Panel.Info = {
  icon: null,
  title: '',
  url: '',
  lanUrl: '',
  description: '',
  openMethod: 2,
}

interface Emit {
  (e: 'update:visible', visible: boolean): void
  (e: 'done', item: Panel.Info): void// 创建完成
}

const model = ref<Panel.Info>(props.itemInfo ? { ...props.itemInfo } : { ...restoreDefault })
const formRef = ref<FormInst | null>(null)

const rules: FormRules = {
  title: {
    required: true,
    trigger: 'blur',
    message: t('form.required'),
  },
  url: {
    required: true,
    trigger: 'blur',
    type: 'string',
    message: t('form.required'),
  },
  // itemIconGroupId: {
  //   required: true,
  //   trigger: ['blur', 'change'],
  //   message: t('form.required'),
  // },
}

const defaultIconBackground = computed(() => groupTints[Math.max(0, itemIconGroupOptions.value.findIndex(group => group.value === model.value.itemIconGroupId)) % groupTints.length])

const options = [
  {
    default: true,
    label: t('iconItem.currentPageOpen'),
    value: 1,
  },
  {
    label: t('iconItem.newWindowOpen'),
    value: 2,
  },
  {
    label: t('iconItem.currentPageLayerOpen'),
    value: 3,
  },
]

// 更新值父组件传来的值
const show = computed({
  get: () => props.visible,
  set: (visible: boolean) => {
    emit('update:visible', visible)
  },
})

async function editApi() {
  if ((!model.value.icon || model.value.icon.itemType === 2) && !model.value.icon?.src?.trim()) {
    model.value.icon = {
      itemType: 1,
      text: model.value.title.trim().charAt(0).toUpperCase() || '?',
      backgroundColor: model.value.icon?.backgroundColor || defaultIconBackground.value,
    }
  }
  submitLoading.value = true
  saveError.value = ''
  try {
    const { code, data, msg } = await edit<Panel.ItemInfo>(model.value)
    if (code === 0) {
      show.value = false
      model.value = { ...restoreDefault }

      emit('done', data)
    }
    else {
      saveError.value = msg || t('common.saveFail')
    }
  }
  catch (error) {
    saveError.value = t('common.saveFail')
  }
  submitLoading.value = false
}

async function autoFetchIcon() {
  if ((model.value.icon && model.value.icon.itemType !== 2) || model.value.icon?.src?.trim() || !model.value.url?.trim())
    return
  const url = model.value.url.trim()
  if (!model.value.icon)
    model.value.icon = { itemType: 2, src: '' }
  try {
    const { code, data } = await getSiteFavicon<{ iconUrl: string }>(url)
    if (code === 0 && model.value.url.trim() === url && model.value.icon?.itemType === 2 && !model.value.icon.src?.trim())
      model.value.icon = { ...model.value.icon, src: data.iconUrl }
  }
  catch {
    // Automatic fetching never prevents saving.
  }
}

const handleValidateButtonClick = (e: MouseEvent) => {
  e.preventDefault()
  formRef.value?.validate((errors) => {
    if (!errors)
      editApi()
  })
}

async function getIconByUrl(url: string, loadingIndex: number) {
  getIconLoading.value[loadingIndex] = true
  try {
    const { code, data } = await getSiteFavicon<{ iconUrl: string }>(url)
    if (code === 0) {
      model.value.icon = {
        ...model.value.icon,
        itemType: 2,
        src: data.iconUrl,
      }
    }
    else {
      ms.error(t('iconItem.geticonFail'))
    }
  }
  catch (error) {
    ms.error(t('iconItem.geticonFail'))
  }
  getIconLoading.value[loadingIndex] = false
}

watch(() => props.visible, (newValue) => {
  if (newValue === true) {
    saveError.value = ''
    model.value = props.itemInfo ? { ...props.itemInfo } : { ...restoreDefault }
    if (props.itemGroupId)
      model.value.itemIconGroupId = props.itemGroupId
  }

  getGroupListOptions()
})

function getGroupListOptions() {
  getGroupList<Common.ListResponse<Panel.ItemIconGroup[]>>().then(({ data, code, msg }) => {
    if (code === 0) {
      itemIconGroupOptions.value = []

      for (let i = 0; i < data.list.length; i++) {
        const element = data.list[i]
        if (i === 0 && !model.value.itemIconGroupId) {
          model.value.itemIconGroupId = element.id
          restoreDefault.itemIconGroupId = element.id
        }

        itemIconGroupOptions.value.push({
          value: element.id as number,
          label: element.title as string,
        })
      }
    }
    else {
      ms.error(`${t('iconItem.getGroupFail')}:${msg}`)
    }
  })
}
</script>

<template>
  <NModal v-model:show="show" class="edit-link-modal" preset="card" size="small" style="width: min(600px, calc(100vw - 32px));border-radius: var(--x-r-card);" :title="itemInfo ? t('iconItem.edit') : t('iconItem.add')">
    <div class="edit-link-body">
      <NForm ref="formRef" :model="model" :rules="rules" class="edit-link-form">
        <div class="edit-link-section">
          <h3 class="edit-link-section-title">
            {{ $t('iconItem.basicInfo') }}
          </h3>
          <NGrid cols="2" :x-gap="16" :y-gap="16" item-responsive>
            <NGridItem span="2 500:1">
              <NFormItem path="itemIconGroupId" :label="t('iconItem.iconGroup')">
                <NSelect v-model:value="model.itemIconGroupId" :options="itemIconGroupOptions" :aria-label="t('iconItem.iconGroup')" />
              </NFormItem>
            </NGridItem>
            <NGridItem span="2 500:1">
              <NFormItem path="title" :label="$t('common.title')">
                <NInput v-model:value="model.title" type="text" show-count :maxlength="20" :input-props="{ 'aria-label': $t('common.title') }" />
              </NFormItem>
            </NGridItem>
          </NGrid>
        </div>

        <div class="edit-link-section">
          <h3 class="edit-link-section-title">
            {{ $t('common.icon') }}
          </h3>
          <NFormItem path="icon">
            <IconEditor v-model:item-icon="model.icon" :default-background="defaultIconBackground" />
          </NFormItem>
        </div>
        <div class="edit-link-section">
          <h3 class="edit-link-section-title">
            {{ $t('iconItem.addresses') }}
          </h3>
          <NFormItem path="url" :label="$t('iconItem.url')" :feedback="saveError" :validation-status="saveError ? 'error' : undefined">
            <div class="address-row">
              <NInput v-model:value="model.url" type="text" :maxlength="1000" placeholder="http(s)://" :input-props="{ 'aria-label': $t('iconItem.url') }" @blur="autoFetchIcon" />
              <NButton class="fetch-icon-button" secondary :disabled="!model.url" :loading="getIconLoading[0]" @click="getIconByUrl(model.url, 0)">
                <template #icon>
                  <SvgIcon icon="lucide-image" />
                </template>
                {{ $t('iconItem.getIcon') }}
              </NButton>
            </div>
          </NFormItem>
          <NFormItem path="lanUrl" :label="$t('iconItem.lanUrl')">
            <div class="address-row">
              <NInput v-model:value="model.lanUrl" type="text" :maxlength="1000" :placeholder="$t('iconItem.lanUrlInputPlaceholder')" :input-props="{ 'aria-label': $t('iconItem.lanUrl') }" />
              <NButton class="fetch-icon-button" secondary :disabled="!model.lanUrl" :loading="getIconLoading[1]" @click="getIconByUrl(model.lanUrl || '', 1)">
                <template #icon>
                  <SvgIcon icon="lucide-image" />
                </template>
                {{ $t('iconItem.getIcon') }}
              </NButton>
            </div>
          </NFormItem>
        </div>
        <div class="edit-link-section">
          <h3 class="edit-link-section-title">
            {{ $t('iconItem.otherInfo') }}
          </h3>
          <NFormItem path="description" :label="$t('common.description')">
            <NInput v-model:value="model.description" type="text" show-count :maxlength="100" :input-props="{ 'aria-label': $t('common.description') }" />
          </NFormItem>
          <NFormItem path="openMethod" :label="$t('iconItem.openMethod')">
            <NSelect v-model:value="model.openMethod" :options="options" :aria-label="$t('iconItem.openMethod')" />
          </NFormItem>
        </div>
      </NForm>
    </div>

    <template #footer>
      <div class="edit-link-actions">
        <NButton secondary :disabled="submitLoading" @click="show = false">
          {{ $t('common.cancel') }}
        </NButton><NButton type="primary" :loading="submitLoading" @click="handleValidateButtonClick">
          {{ $t('common.save') }}
        </NButton>
      </div>
    </template>
  </NModal>
</template>

<style>
.edit-link-body { max-height: min(600px, 65vh); overflow: auto; padding: 4px; }
.edit-link-section + .edit-link-section { margin-top: 24px; }
.edit-link-section-title { margin-bottom: 16px; color: var(--x-text1); font-size: 13px; font-weight: 600; }
.edit-link-section .n-form-item + .n-form-item { margin-top: 16px; }
.edit-link-form .n-form-item-label { color: var(--x-text1); font-size: 13px; font-weight: 600; }
.edit-link-form .n-form-item-label__asterisk { color: var(--x-red-fg) !important; }
.edit-link-form .n-form-item-feedback-wrapper, .edit-link-form .n-input__count { color: var(--x-text2); font-size: 12px; }
.address-row { display: flex; align-items: flex-start; gap: 8px; width: 100%; }
.address-row .n-input { flex: 1; min-width: 0; }
.fetch-icon-button.n-button, .edit-link-actions .n-button { height: 44px; border-radius: var(--x-r-pill); }
.fetch-icon-button.n-button { flex: none; }
.fetch-icon-button .svg-icon { width: 18px; height: 18px; }
.edit-link-actions { display: flex; justify-content: flex-end; gap: 8px; }
@media (max-width: 480px) { .address-row { flex-wrap: wrap; }.address-row .n-input { flex-basis: 100%; } }
</style>
