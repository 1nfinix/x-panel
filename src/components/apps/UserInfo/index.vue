<script setup lang="ts">
import type { FormInst, FormRules } from 'naive-ui'
import { NButton, NCard, NDivider, NForm, NFormItem, NInput, NSelect, useDialog, useMessage } from 'naive-ui'
import { h, ref, watch } from 'vue'
import { useAppStore, useAuthStore, usePanelState, useUserStore } from '@/store'
import { languageOptions } from '@/utils/defaultData'
import type { Language, Theme } from '@/store/modules/app/helper'
import { logout } from '@/api'
import { RoundCardModal, SvgIcon } from '@/components/common/'
import { updateInfo, updatePassword } from '@/api/system/user'
import { updateLocalUserInfo } from '@/utils/cmn'
import { t } from '@/locales'

const userStore = useUserStore()
const authStore = useAuthStore()
const appStore = useAppStore()
const panelState = usePanelState()
const ms = useMessage()
const dialog = useDialog()

const languageValue = ref(appStore.language)
const themeValue = ref(appStore.theme)
const nickName = ref(authStore.userInfo?.name || '')
const isEditNickNameStatus = ref(false)
const nickSaving = ref(false)
const nickError = ref('')
const passwordError = ref('')
const confirmPasswordError = ref('')
const formRef = ref<FormInst | null>(null)
const themeOptions: { label: string; key: string; value: Theme }[] = [
  { label: t('apps.userInfo.themeStyle.dark'), key: 'dark', value: 'dark' },
  { label: t('apps.userInfo.themeStyle.light'), key: 'light', value: 'light' },
  { label: t('apps.userInfo.themeStyle.auto'), key: 'Auto', value: 'auto' },
]
const updatePasswordModalState = ref({
  show: false,
  loading: false,
  form: {
    password: '',
    oldPassword: '',
    confirmPassword: '',
  },
})

const updatePasswordModalFormRules: FormRules = {
  oldPassword: {
    required: true,
    trigger: 'blur',
    min: 6,
    max: 20,
    message: t('adminSettingUsers.formRules.passwordLimit'),
  },
  password: {
    required: true,
    trigger: 'blur',
    min: 6,
    max: 20,
    message: t('adminSettingUsers.formRules.passwordLimit'),
  },
  confirmPassword: {
    required: true,
    trigger: 'blur',
    min: 6,
    max: 20,
    message: t('adminSettingUsers.formRules.passwordLimit'),
  },
}

async function logoutApi() {
  const { code, msg } = await logout()
  if (code !== 0)
    throw new Error(msg || t('common.failed'))
  userStore.resetUserInfo()
  authStore.removeToken()
  panelState.removeState()
  appStore.removeToken()
  ms.success(t('settingUserInfo.logoutSuccess'))
  // router.push({ path: '/login' })
  location.reload()// 强制刷新一下页面
}

async function handleSaveInfo() {
  if (nickSaving.value)
    return
  nickSaving.value = true
  nickError.value = ''
  try {
    const { code, msg } = await updateInfo(nickName.value)
    if (code === 0) {
      await updateLocalUserInfo()
      isEditNickNameStatus.value = false
      ms.success(t('common.saveSuccess'))
    }
    else { nickError.value = msg || t('common.editFail') }
  }
  catch {
    nickError.value = t('common.editFail')
  }
  finally {
    nickSaving.value = false
  }
}

async function handleUpdatePassword(e: MouseEvent) {
  e.preventDefault()
  try {
    await formRef.value?.validate()
  }
  catch {
    return
  }
  confirmPasswordError.value = ''
  passwordError.value = ''
  if (updatePasswordModalState.value.form.password !== updatePasswordModalState.value.form.confirmPassword) {
    confirmPasswordError.value = t('settingUserInfo.confirmPasswordInconsistentMsg')
    return
  }
  if (updatePasswordModalState.value.loading)
    return
  updatePasswordModalState.value.loading = true
  try {
    const { code, msg } = await updatePassword(updatePasswordModalState.value.form.oldPassword, updatePasswordModalState.value.form.password)
    if (code !== 0) {
      passwordError.value = msg || t('common.saveFail')
      return
    }
    updatePasswordModalState.value.show = false
    ms.success(t('common.success'))
  }
  catch {
    passwordError.value = t('common.serverError')
  }
  finally {
    updatePasswordModalState.value.loading = false
  }
}

watch(() => updatePasswordModalState.value.show, (show) => {
  if (show) {
    passwordError.value = ''
    confirmPasswordError.value = ''
  }
})

function handleLogout() {
  const logoutError = ref('')
  const confirmDialog = dialog.warning({
    title: t('common.warning'),
    content: () => h('div', [h('p', t('settingUserInfo.confirmLogoutText')), logoutError.value ? h('p', { style: 'color:var(--x-red-fg);margin-top:12px' }, logoutError.value) : null]),
    positiveText: t('common.confirm'),
    negativeText: t('common.cancel'),
    onPositiveClick: async () => {
      confirmDialog.loading = true
      logoutError.value = ''
      try {
        await logoutApi()
      }
      catch (error) {
        logoutError.value = error instanceof Error ? error.message : t('common.failed')
        return false
      }
      finally {
        confirmDialog.loading = false
      }
    },
  })
}

function handleChangeLanuage(value: Language) {
  languageValue.value = value
  appStore.setLanguage(value)
  location.reload()
}

function handleChangeTheme(value: Theme) {
  themeValue.value = value
  appStore.setTheme(value)
  // location.reload()
}
</script>

<template>
  <div class="bg-bg1 p-2 h-full">
    <NCard size="small">
      <div>
        <div class="text-text1 font-medium">
          {{ $t('common.username') }}
        </div>
        {{ authStore.userInfo?.username }}
      </div>

      <div class="mt-3">
        <div class="text-text1 font-medium">
          {{ $t('common.nikeName') }}
        </div>

        <div v-if="!isEditNickNameStatus">
          {{ authStore.userInfo?.name }}

          <NButton size="small" text type="primary" @click="isEditNickNameStatus = !isEditNickNameStatus">
            {{ $t('common.edit') }}
          </NButton>
        </div>

        <div v-else class="flex items-center">
          <div class="max-w-[150px]">
            <NInput v-model:value="nickName" type="text" :placeholder="$t('common.inputPlaceholder')" :input-props="{ 'aria-label': $t('common.nikeName') }" />
          </div>
          <NButton size="small" secondary :loading="nickSaving" @click="handleSaveInfo">
            {{ $t('common.save') }}
          </NButton>
        </div>
        <p v-if="nickError" class="field-error" role="alert">
          {{ nickError }}
        </p>
      </div>

      <div class="mt-3">
        <div class="text-text1 font-medium">
          {{ $t('common.language') }}
        </div>
        <div class="max-w-[200px]">
          <NSelect v-model:value="languageValue" :options="languageOptions" :aria-label="$t('common.language')" @update-value="handleChangeLanuage" />
        </div>
      </div>

      <div class="mt-3">
        <div class="text-text1 font-medium">
          {{ $t('apps.userInfo.theme') }}
        </div>
        <div class="max-w-[200px]">
          <NSelect v-model:value="themeValue" :options="themeOptions" :aria-label="$t('apps.userInfo.theme')" @update-value="handleChangeTheme" />
        </div>
      </div>

      <NDivider style="margin: 10px 0;" dashed />
      <div>
        <NButton size="small" text type="primary" @click="updatePasswordModalState.show = !updatePasswordModalState.show">
          {{ $t('settingUserInfo.updatePassword') }}
        </NButton>
      </div>
    </NCard>

    <NCard class="mt-3" size="small">
      <NButton type="error" @click="handleLogout">
        <template #icon>
          <SvgIcon icon="tabler:logout" />
        </template>
        {{ $t('settingUserInfo.logout') }}
      </NButton>
    </NCard>

    <RoundCardModal v-model:show="updatePasswordModalState.show" size="small" preset="card" style="width: 400px" :title="$t('settingUserInfo.updatePassword')">
      <NForm ref="formRef" :model="updatePasswordModalState.form" :rules="updatePasswordModalFormRules">
        <NFormItem path="oldPassword" :label="$t('settingUserInfo.oldPassword')" :feedback="passwordError" :validation-status="passwordError ? 'error' : undefined">
          <NInput v-model:value="updatePasswordModalState.form.oldPassword" :maxlength="20" type="password" :placeholder="$t('settingUserInfo.oldPassword')" :input-props="{ 'aria-label': $t('settingUserInfo.oldPassword') }" />
        </NFormItem>

        <NFormItem path="password" :label="$t('settingUserInfo.newPassword')">
          <NInput v-model:value="updatePasswordModalState.form.password" :maxlength="20" type="password" :placeholder="$t('settingUserInfo.newPassword')" :input-props="{ 'aria-label': $t('settingUserInfo.newPassword') }" />
        </NFormItem>

        <NFormItem path="confirmPassword" :label="$t('settingUserInfo.confirmPassword')" :feedback="confirmPasswordError" :validation-status="confirmPasswordError ? 'error' : undefined">
          <NInput v-model:value="updatePasswordModalState.form.confirmPassword" :maxlength="20" type="password" :placeholder="$t('settingUserInfo.confirmPassword')" :input-props="{ 'aria-label': $t('settingUserInfo.confirmPassword') }" />
        </NFormItem>
      </NForm>

      <template #footer>
        <div class="float-right">
          <NButton type="primary" size="small" :loading="updatePasswordModalState.loading" @click="handleUpdatePassword">
            {{ $t('common.save') }}
          </NButton>
        </div>
      </template>
    </RoundCardModal>
  </div>
</template>

<style scoped>
.field-error { margin-top: 8px; color: var(--x-red-fg); font-size: 13px; }
</style>
