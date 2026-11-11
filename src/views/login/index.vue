<script setup lang="ts">
import { NButton, NCard, NForm, NFormItem, NInput, NSelect, useMessage } from 'naive-ui'
import type { InputInst } from 'naive-ui'
import { nextTick, onMounted, ref } from 'vue'
import { login } from '@/api'
import { useAppStore, useAuthStore, usePanelState } from '@/store'
import { SvgIcon, XBrand } from '@/components/common'
import { router } from '@/router'
import { t } from '@/locales'
import { languageOptions } from '@/utils/defaultData'
import type { Language } from '@/store/modules/app/helper'
import Captcha from '@/components/common/Captcha/index.vue'
import { getLoginConfig } from '@/api/openness'
import { useWallpaper } from '@/hooks/useWallpaper'

// const userStore = useUserStore()
const authStore = useAuthStore()
const appStore = useAppStore()
const ms = useMessage()
const loading = ref(false)
const fieldError = ref<{ username?: string; password?: string; vcode?: string }>({})
const usernameInput = ref<InputInst | null>(null)
const passwordInput = ref<InputInst | null>(null)
const captchaInput = ref<InputInst | null>(null)
const languageValue = ref<Language>(appStore.language)

const isShowCaptcha = ref(true)
const captchaRef = ref<InstanceType<typeof Captcha> | null>(null)
const panelState = usePanelState()
const { source: wallpaperSource, loaded: wallpaperLoaded, show: showWallpaper } = useWallpaper()
// const isShowRegister = ref<boolean>(false)

const form = ref<Login.LoginReqest>({
  username: '',
  password: '',
  vcode: '',
})

async function focusFirstError() {
  await nextTick()
  if (fieldError.value.username)
    usernameInput.value?.focus()
  else if (fieldError.value.password)
    passwordInput.value?.focus()
  else if (fieldError.value.vcode)
    captchaInput.value?.focus()
}

onMounted(async () => {
  if (panelState.panelConfig.showWallpaper)
    showWallpaper(panelState.panelConfig.backgroundImageSrc, panelState.panelConfig.autoWallpaper !== false)
  try {
    const response = await getLoginConfig<{ loginCaptcha: boolean }>()
    if (response.code === 0)
      isShowCaptcha.value = response.data.loginCaptcha
  }
  catch {
    isShowCaptcha.value = true
  }
})

const loginPost = async () => {
  if (loading.value)
    return
  fieldError.value = {}
  if (!form.value.username) {
    fieldError.value.username = t('form.required')
    await focusFirstError()
    return
  }
  if (!form.value.password) {
    fieldError.value.password = t('form.required')
    await focusFirstError()
    return
  }
  loading.value = true
  try {
    const res = await login<Login.LoginResponse>(form.value)
    if (res.code === 0) {
      authStore.setToken(res.data.token)
      authStore.setUserInfo(res.data)

      setTimeout(() => {
        ms.success(`Hi ${res.data.name},${t('login.welcomeMessage')}`)
        loading.value = false
        router.push({ path: '/' })
      }, 500)
    }
    else {
      loading.value = false
      const message = res.code === 1101 ? t('login.captchaRequired') : res.code === 1102 ? t('login.captchaInvalid') : res.code === 1003 ? t('login.credentialsInvalid') : res.msg
      if (res.code === 1101 || res.code === 1102)
        fieldError.value.vcode = message
      else
        fieldError.value.password = message
      form.value.vcode = ''
      captchaRef.value?.refresh()
      await focusFirstError()
    }
  }
  catch (error) {
    loading.value = false
    fieldError.value.password = t('login.networkError')
    form.value.vcode = ''
    captchaRef.value?.refresh()
    await focusFirstError()
  }
}

function handleSubmit() {
  // 点击登录按钮触发
  loginPost()
}

function handleChangeLanuage(value: Language) {
  languageValue.value = value
  appStore.setLanguage(value)
}
</script>

<template>
  <main class="login-container">
    <div v-if="panelState.panelConfig.showWallpaper" class="login-wallpaper" :class="{ loaded: wallpaperLoaded }" :style="{ backgroundImage: wallpaperLoaded ? `url(${wallpaperSource})` : 'none' }" /><div v-if="panelState.panelConfig.showWallpaper" class="wallpaper-mask" />
    <NCard class="login-card">
      <div class="mb-5 flex items-center justify-end">
        <div class="mr-2">
          <SvgIcon icon="ion-language" style="width: 20px;height: 20px;" />
        </div>
        <div class="min-w-[100px]">
          <NSelect v-model:value="languageValue" size="small" :options="languageOptions" :aria-label="t('common.language')" @update-value="handleChangeLanuage" />
        </div>
      </div>

      <div class="login-title">
        <span class="login-brand"><XBrand /></span>
        <h1 class="login-heading">
          {{ panelState.panelConfig.logoText || $t('common.appName') }}
        </h1>
      </div>
      <NForm class="login-form" :model="form" :show-label="false" @keydown.enter="handleSubmit">
        <div class="login-fields">
          <div class="login-field">
            <label class="login-label" for="login-username">{{ $t('login.usernamePlaceholder') }}</label>
            <NFormItem :show-feedback="!!fieldError.username" :validation-status="fieldError.username ? 'error' : undefined">
              <NInput ref="usernameInput" v-model:value="form.username" :placeholder="$t('login.usernamePlaceholder')" :input-props="{ 'id': 'login-username', 'name': 'username', 'autocomplete': 'username', 'aria-invalid': fieldError.username ? 'true' : undefined, 'aria-describedby': fieldError.username ? 'login-username-error' : undefined }">
                <template #prefix>
                  <SvgIcon icon="ph:user-bold" />
                </template>
              </NInput>
              <template #feedback>
                <span v-if="fieldError.username" id="login-username-error" role="alert">{{ fieldError.username }}</span>
              </template>
            </NFormItem>
          </div>

          <div class="login-field">
            <label class="login-label" for="login-password">{{ $t('login.passwordPlaceholder') }}</label>
            <NFormItem :show-feedback="!!fieldError.password" :validation-status="fieldError.password ? 'error' : undefined">
              <NInput ref="passwordInput" v-model:value="form.password" type="password" :placeholder="$t('login.passwordPlaceholder')" :input-props="{ 'id': 'login-password', 'name': 'password', 'autocomplete': 'current-password', 'aria-invalid': fieldError.password ? 'true' : undefined, 'aria-describedby': fieldError.password ? 'login-password-error' : undefined }">
                <template #prefix>
                  <SvgIcon icon="mdi:password-outline" />
                </template>
              </NInput>
              <template #feedback>
                <span v-if="fieldError.password" id="login-password-error" role="alert">{{ fieldError.password }}</span>
              </template>
            </NFormItem>
          </div>

          <div v-if="isShowCaptcha" class="login-field">
            <label class="login-label" for="login-captcha">{{ $t('login.captchaLabel') }}</label>
            <NFormItem :show-feedback="!!fieldError.vcode" :validation-status="fieldError.vcode ? 'error' : undefined">
              <div class="captcha-row">
                <div class="captcha-image" :title="t('login.refreshCaptcha')">
                  <Captcha ref="captchaRef" src="/api/captcha/getImage" />
                </div>
                <NInput ref="captchaInput" v-model:value="form.vcode" type="text" maxlength="4" :placeholder="t('login.captchaShortPlaceholder')" :input-props="{ 'id': 'login-captcha', 'name': 'captcha', 'autocomplete': 'off', 'spellcheck': false, 'autocapitalize': 'off', 'autocorrect': 'off', 'aria-invalid': fieldError.vcode ? 'true' : undefined, 'aria-describedby': fieldError.vcode ? 'login-captcha-error' : undefined }" />
              </div>
              <template #feedback>
                <span v-if="fieldError.vcode" id="login-captcha-error" role="alert">{{ fieldError.vcode }}</span>
              </template>
            </NFormItem>
          </div>
        </div>
        <div class="login-submit">
          <NButton type="primary" block :loading="loading" @click="handleSubmit">
            {{ $t('login.loginButton') }}
          </NButton>
        </div>

        <!-- <div class="flex justify-end">
          <NButton v-if="isShowRegister" text type="primary" class="flex" @click="$router.push({ path: '/register' })">
            注册
          </NButton>
          <NButton text type="primary" class="flex" @click="$router.push({ path: '/resetPassword' })">
            忘记密码?
          </NButton>
        </div> -->

        <div class="login-footer flex justify-center text-text2 text-[13px]">
          Powered By <a href="https://github.com/1nfinix/x-panel" target="_blank" class="ml-1 text-text2">X-Panel</a>
        </div>
      </NForm>
    </NCard>
  </main>
</template>

<style scoped>
.login-container { position: relative; min-height: 100vh; display: flex; align-items: center; justify-content: center; padding: 24px; background-color: var(--x-bg0); }
.login-wallpaper { position: absolute; inset: 0; background-position: center; background-size: cover; opacity: 0; transition: opacity 600ms ease; }
.login-wallpaper.loaded { opacity: 1; }
.wallpaper-mask { position: absolute; inset: 0; background: rgba(18,18,20,.55); }
.login-card { position: relative; width: min(100%, 430px); border: 0; border-radius: var(--x-r-card); background: var(--x-bg1); color: var(--x-text1); box-shadow: none; }
.login-card :deep(.n-card__content) { padding: 40px; }
.login-card :deep(.n-input) { min-height: 44px; border-radius: var(--x-r-small); background: var(--x-bg2); }
.login-card :deep(.n-button--primary-type) { height: 44px; border-radius: var(--x-r-pill); }
.login-title { display: flex; align-items: center; justify-content: center; gap: 10px; margin: 16px 0 32px; }
.login-brand { display: grid; place-items: center; width: 44px; height: 44px; border-radius: var(--x-r-icon); background: var(--x-bg2); }.login-brand :deep(svg) { width: 28px; height: 28px; }
.login-fields { display: grid; gap: 16px; }
.login-field { display: grid; gap: 8px; min-width: 0; }
.login-label { color: var(--x-text1); font-size: 13px; font-weight: 600; }
.login-fields :deep(.n-form-item) { margin: 0; }
.login-fields :deep(.n-form-item-blank) { min-width: 0; }
.login-submit { margin-top: 24px; }
.login-submit .n-button { width: 100%; }
.login-footer { margin-top: 16px; }
.login-heading { margin: 0; font-family: var(--x-font); font-size: 22px; color: var(--x-text1); font-weight: 600; letter-spacing: -.02em; }
.captcha-row { display: flex; width: 100%; gap: 12px; min-width: 0; }
.captcha-row :deep(.n-input) { flex: 1; min-width: 0; }
.captcha-image { flex: none; width: 160px; height: 44px; cursor: pointer; border-radius: var(--x-r-icon); overflow: hidden; }
.captcha-image :deep(img) { display: block; width: 100%; height: 100%; object-fit: fill; }
@media (max-width: 480px) {
  .login-card :deep(.n-card__content) { padding: 24px; }
  .captcha-image { width: clamp(100px, 34vw, 138px); }
}
@media (prefers-reduced-motion: reduce) { .login-wallpaper { transition: none; } }
</style>
